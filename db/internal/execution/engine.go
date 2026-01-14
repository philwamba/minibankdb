package execution

import (
	"minibank/internal/catalog"
	"minibank/internal/storage"
)

type Iterator interface {
	Open() error
	Next() (*storage.Tuple, error)
	Close() error
	Schema() []catalog.Column
}

type ExecutionContext struct {
	Catalog *catalog.Catalog
	Storage *storage.Engine
}

type Result struct {
	Columns []catalog.Column
	Rows    []*storage.Tuple
}
