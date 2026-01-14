package storage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"minibank/internal/catalog"
)

type Cell struct {
	Type  catalog.ColumnType
	Value interface{}
}

type RID struct {
	PageID PageID
	SlotID int
}

type Tuple struct {
	RID   RID
	Cells []Cell
}

func SerializeTuple(t *Tuple) ([]byte, error) {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.BigEndian, uint16(len(t.Cells))); err != nil {
		return nil, err
	}

	for _, cell := range t.Cells {
		switch cell.Type {
		case catalog.TypeInt:
			val, ok := cell.Value.(int64)
			if !ok {
				if v, ok2 := cell.Value.(int); ok2 {
					val = int64(v)
				} else {
					return nil, fmt.Errorf("expected int64 for INT column")
				}
			}
			if err := binary.Write(&buf, binary.BigEndian, val); err != nil {
				return nil, err
			}
		case catalog.TypeString:
			val, ok := cell.Value.(string)
			if !ok {
				return nil, fmt.Errorf("expected string for STRING column")
			}
			length := uint16(len(val))
			if err := binary.Write(&buf, binary.BigEndian, length); err != nil {
				return nil, err
			}
			buf.WriteString(val)
		case catalog.TypeBool:
			val, ok := cell.Value.(bool)
			if !ok {
				return nil, fmt.Errorf("expected bool for BOOL column")
			}
			var b byte = 0
			if val {
				b = 1
			}
			buf.WriteByte(b)
		case catalog.TypeDecimal:
			val, ok := cell.Value.(string)
			if !ok {
				return nil, fmt.Errorf("expected string (decimal) for DECIMAL column")
			}
			length := uint16(len(val))
			if err := binary.Write(&buf, binary.BigEndian, length); err != nil {
				return nil, err
			}
			buf.WriteString(val)
		case catalog.TypeTimestamp:
			val, ok := cell.Value.(int64)
			if !ok {
				return nil, fmt.Errorf("expected int64 (unix) for TIMESTAMP column")
			}
			if err := binary.Write(&buf, binary.BigEndian, val); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported type for serialization: %s", cell.Type)
		}
	}
	return buf.Bytes(), nil
}

func DeserializeTuple(data []byte, columns []catalog.Column) (*Tuple, error) {
	buf := bytes.NewReader(data)
	var numCells uint16
	if err := binary.Read(buf, binary.BigEndian, &numCells); err != nil {
		return nil, err
	}

	if int(numCells) != len(columns) {
		return nil, fmt.Errorf("tuple cell count %d does not match schema column count %d", numCells, len(columns))
	}

	cells := make([]Cell, len(columns))
	for i, col := range columns {
		cells[i].Type = col.Type
		switch col.Type {
		case catalog.TypeInt:
			var val int64
			if err := binary.Read(buf, binary.BigEndian, &val); err != nil {
				return nil, err
			}
			cells[i].Value = val
		case catalog.TypeString:
			var length uint16
			if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
				return nil, err
			}
			strBytes := make([]byte, length)
			if _, err := buf.Read(strBytes); err != nil {
				return nil, err
			}
			cells[i].Value = string(strBytes)
		case catalog.TypeBool:
			b, err := buf.ReadByte()
			if err != nil {
				return nil, err
			}
			cells[i].Value = (b == 1)
		case catalog.TypeDecimal:
			var length uint16
			if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
				return nil, err
			}
			strBytes := make([]byte, length)
			if _, err := buf.Read(strBytes); err != nil {
				return nil, err
			}
			cells[i].Value = string(strBytes)
		case catalog.TypeTimestamp:
			var val int64
			if err := binary.Read(buf, binary.BigEndian, &val); err != nil {
				return nil, err
			}
			cells[i].Value = val
		}
	}
	return &Tuple{Cells: cells}, nil
}
