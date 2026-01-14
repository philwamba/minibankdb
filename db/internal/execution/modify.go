package execution

import (
	"fmt"
	"minibank/internal/catalog"
	"minibank/internal/storage"
)

type Update struct {
	HeapFile *storage.HeapFile
	Child    Iterator
	SetPairs map[string]interface{}
}

func NewUpdate(hf *storage.HeapFile, child Iterator, setPairs map[string]interface{}) *Update {
	return &Update{
		HeapFile: hf,
		Child:    child,
		SetPairs: setPairs,
	}
}

func (op *Update) Open() error {
	return op.Child.Open()
}

func (op *Update) Next() (*storage.Tuple, error) {
	t, err := op.Child.Next()
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, nil
	}

	schema := op.Child.Schema()
	cells := make([]storage.Cell, len(t.Cells))
	copy(cells, t.Cells)

	for colName, val := range op.SetPairs {
		found := false
		for i, col := range schema {
			if col.Name == colName {
				cells[i].Value = val
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("column %s not found", colName)
		}
	}

	newTuple := &storage.Tuple{Cells: cells}

	data, err := storage.SerializeTuple(newTuple)
	if err != nil {
		return nil, err
	}
	_, _, err = op.HeapFile.Insert(data)
	if err != nil {
		return nil, err
	}

	if err := op.HeapFile.DeleteTuple(t.RID); err != nil {
		return nil, err
	}

	return newTuple, nil
}

func (op *Update) Close() error {
	return op.Child.Close()
}

func (op *Update) Schema() []catalog.Column {
	return op.Child.Schema()
}

type Delete struct {
	HeapFile *storage.HeapFile
	Child    Iterator
}

func NewDelete(hf *storage.HeapFile, child Iterator) *Delete {
	return &Delete{HeapFile: hf, Child: child}
}

func (op *Delete) Open() error {
	return op.Child.Open()
}

func (op *Delete) Next() (*storage.Tuple, error) {
	t, err := op.Child.Next()
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, nil
	}

	if err := op.HeapFile.DeleteTuple(t.RID); err != nil {
		return nil, err
	}

	return t, nil
}

func (op *Delete) Close() error {
	return op.Child.Close()
}

func (op *Delete) Schema() []catalog.Column {
	return op.Child.Schema()
}
