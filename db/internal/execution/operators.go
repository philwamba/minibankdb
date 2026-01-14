package execution

import (
	"fmt"
	"minibank/internal/catalog"
	"minibank/internal/parser"
	"minibank/internal/storage"
)

// SeqScan
type SeqScan struct {
	HeapFile *storage.HeapFile
	Iterator *storage.HeapIterator
	schema   []catalog.Column
}

func NewSeqScan(hf *storage.HeapFile, schema []catalog.Column) *SeqScan {
	return &SeqScan{HeapFile: hf, schema: schema}
}

func (s *SeqScan) Open() error {
	s.Iterator = s.HeapFile.Iterator()
	return nil
}

func (s *SeqScan) Next() (*storage.Tuple, error) {
	bytes, rid, err := s.Iterator.Next()
	if err != nil {
		return nil, err
	}
	if bytes == nil {
		return nil, nil
	}
	// Deserialize
	tuple, err := storage.DeserializeTuple(bytes, s.schema)
	if err != nil {
		return nil, err
	}
	tuple.RID = rid
	return tuple, nil
}

func (s *SeqScan) Close() error {
	return nil
}

func (s *SeqScan) Schema() []catalog.Column {
	return s.schema
}

// Filter
type Filter struct {
	Child Iterator
	Pred  parser.Expression
}

func NewFilter(child Iterator, pred parser.Expression) *Filter {
	return &Filter{Child: child, Pred: pred}
}

func (f *Filter) Open() error {
	return f.Child.Open()
}

func (f *Filter) Next() (*storage.Tuple, error) {
	for {
		tuple, err := f.Child.Next()
		if err != nil {
			return nil, err
		}
		if tuple == nil {
			return nil, nil
		}

		match, err := Evaluate(tuple, f.Pred, f.Child.Schema())
		if err != nil {
			return nil, err
		}
		if match {
			return tuple, nil
		}
	}
}

func (f *Filter) Close() error {
	return f.Child.Close()
}

func (f *Filter) Schema() []catalog.Column {
	return f.Child.Schema()
}

// Project
type Project struct {
	Child  Iterator
	Fields []string
	schema []catalog.Column
}

func NewProject(child Iterator, fields []string, schema []catalog.Column) *Project {
	return &Project{Child: child, Fields: fields, schema: schema}
}

func (p *Project) Open() error {
	return p.Child.Open()
}

func (p *Project) Next() (*storage.Tuple, error) {
	t, err := p.Child.Next()
	if err != nil || t == nil {
		return t, err
	}

	outCells := make([]storage.Cell, len(p.schema))
	inputSchema := p.Child.Schema()

	for i, col := range p.schema {
		found := false
		for j, inCol := range inputSchema {
			if inCol.Name == col.Name {
				outCells[i] = t.Cells[j]
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("column %s not found", col.Name)
		}
	}
	return &storage.Tuple{Cells: outCells}, nil
}

func (p *Project) Close() error {
	return p.Child.Close()
}

func (p *Project) Schema() []catalog.Column {
	return p.schema
}
