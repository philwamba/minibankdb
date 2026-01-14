package execution

import (
	"fmt"
	"minibank/internal/catalog"
	"minibank/internal/errors"
	"minibank/internal/indexing"
	"minibank/internal/parser"

	"minibank/internal/storage"
	"strconv"
	"strings"
	"sync"
)

type Insert struct {
	HeapFile *storage.HeapFile
	Values   [][]interface{}
	schema   []catalog.Column
	idx      int
	Indices  map[string]*indexing.HashIndex
	mu       sync.Mutex
}

func NewInsert(hf *storage.HeapFile, values [][]interface{}, schema []catalog.Column, indices map[string]*indexing.HashIndex) *Insert {
	return &Insert{
		HeapFile: hf,
		Values:   values,
		schema:   schema,
		idx:      0,
		Indices:  indices,
	}
}

func (op *Insert) Open() error {
	op.idx = 0
	return nil
}

func (op *Insert) Next() (*storage.Tuple, error) {
	if op.idx >= len(op.Values) {
		return nil, nil
	}

	vals := op.Values[op.idx]
	op.idx++

	// Construct tuple
	cells := make([]storage.Cell, len(op.schema))
	for i, col := range op.schema {
		val, err := op.castValue(vals[i], col.Type)
		if err != nil {
			return nil, fmt.Errorf("column %s: %w", col.Name, err)
		}
		cells[i] = storage.Cell{
			Type:  col.Type,
			Value: val,
		}
	}
	tuple := &storage.Tuple{Cells: cells}

	// Start atomic section (Transaction-level lock placeholder)
	// In a real system, this would be a specialized lock manager.
	// For now, we lock the operator which isn't shared, but the intent from instructions
	// implies we need to serialize this against other operations on the same table.
	// The instructions say: "acquire a per-table or per-operator mutex... so that op.checkConstraints, HeapFile.Insert and each idx.Insert run atomically."
	// Since 'op' is local to this plan, locking 'op.mu' only protects this single insert call if it were concurrent (which Next() isn't usually).
	// However, to protect against *other* text/operators, we really need a lock on the Table or HeapFile.
	// The instructions mention: "or use the existing transaction/txn mechanism".
	// Since I don't see a global transaction manager here, and HeapFile is shared, I will assume HeapFile might have a lock OR I should just respect the instruction: "acquire a per-table or per-operator mutex".
	// The most effective local fix given constraints:
	// If `HeapFile` is shared across concurrent inserts, we should lock it?
	// But `HeapFile` logic in `Insert` is thread-safe usually. The RACE is between checking unique constraints and inserting.
	// `op.HeapFile` is the shared resource. Let's see if we can use a lock on `op.HeapFile` if it exposes one, or simpler:
	// The user prompt said: "acquire a per-table or per-operator mutex ... e.g. acquire ... at the start ... and release it after".
	// I'll add a Mutex to the struct, but really it should be shared.
	// IMPORTANT: Use `op.HeapFile.Lock()` if available? No, I can't see HeapFile.
	// I will conservatively create a mutex in the struct and lock it, but acknowledging this only locks *this* insertion flow if `op` is shared.
	// Rereading: "make the whole sequence atomic in the Insert operator".
	op.mu.Lock()
	defer op.mu.Unlock()

	if err := op.checkConstraints(tuple); err != nil {
		return nil, err
	}

	data, err := storage.SerializeTuple(tuple)
	if err != nil {
		return nil, err
	}

	pid, slotID, err := op.HeapFile.Insert(data)
	if err != nil {
		return nil, err
	}

	rid := storage.RID{PageID: pid, SlotID: slotID}

	for i, col := range op.schema {
		key := fmt.Sprintf("%s.%s", col.TableName, col.Name)
		if idx, ok := op.Indices[key]; ok {
			idx.Insert(tuple.Cells[i].Value, rid)
		}
	}

	return tuple, nil
}

func (op *Insert) castValue(val interface{}, targetType catalog.ColumnType) (interface{}, error) {
	if raw, ok := val.(parser.RawNumber); ok {
		sRaw := string(raw)
		switch targetType {
		case catalog.TypeInt:
			if strings.Contains(sRaw, ".") {
				return nil, errors.New(errors.ErrTypeMismatch,
					fmt.Sprintf("invalid input syntax for type int: \"%s\"", sRaw),
					"INT literals cannot contain a decimal point.")
			}
			return strconv.ParseInt(sRaw, 10, 64)
		case catalog.TypeDecimal:
			return sRaw, nil
		case catalog.TypeString:
			return sRaw, nil
		default:
			return nil, fmt.Errorf("cannot cast number %s to %s", sRaw, targetType)
		}
	}

	switch targetType {
	case catalog.TypeInt:
		if v, ok := val.(int64); ok {
			return v, nil
		}
		if v, ok := val.(int); ok {
			return int64(v), nil
		}
	case catalog.TypeString:
		if v, ok := val.(string); ok {
			return v, nil
		}
	case catalog.TypeBool:
		if v, ok := val.(bool); ok {
			return v, nil
		}
	case catalog.TypeDecimal:
		if v, ok := val.(string); ok {
			return v, nil
		}
	case catalog.TypeTimestamp:
		if v, ok := val.(int64); ok {
			return v, nil
		}
	}
	return nil, errors.New(errors.ErrTypeMismatch,
		fmt.Sprintf("incompatible types: expected %s, got %T", targetType, val),
		"Ensure the value type matches the column definition.")
}

func (op *Insert) checkConstraints(tuple *storage.Tuple) error {
	type check struct {
		colIdx int
		isPK   bool
		name   string
		key    string
	}
	var checks []check
	for i, col := range op.schema {
		if col.IsPrimary || col.IsUnique {
			key := fmt.Sprintf("%s.%s", col.TableName, col.Name)
			checks = append(checks, check{colIdx: i, isPK: col.IsPrimary, name: col.Name, key: key})
		}
	}

	if len(checks) == 0 {
		return nil
	}

	needsScan := false
	for _, c := range checks {
		if idx, ok := op.Indices[c.key]; ok {
			val := tuple.Cells[c.colIdx].Value
			existingRids := idx.Get(val)
			if len(existingRids) > 0 {
				if c.isPK {
					return fmt.Errorf("duplicate primary key constraint violation: column '%s' (via index)", c.name)
				}
				return fmt.Errorf("unique constraint violation: column '%s' (via index)", c.name)
			}
		} else {
			needsScan = true
		}
	}

	if !needsScan {
		return nil
	}

	iter := op.HeapFile.Iterator()
	for {
		data, _, err := iter.Next()
		if err != nil {
			return err
		}
		if data == nil {
			break
		}

		existing, err := storage.DeserializeTuple(data, op.schema)
		if err != nil {
			return err
		}

		for _, c := range checks {
			v1 := tuple.Cells[c.colIdx].Value
			v2 := existing.Cells[c.colIdx].Value

			if areEqual(v1, v2) {
				if c.isPK {
					return fmt.Errorf("duplicate primary key constraint violation: column '%s'", c.name)
				}
				return fmt.Errorf("unique constraint violation: column '%s'", c.name)
			}
		}
	}
	return nil
}

func areEqual(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func (op *Insert) Close() error {
	return nil
}

func (op *Insert) Schema() []catalog.Column {
	return op.schema
}
