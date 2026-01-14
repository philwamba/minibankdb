package storage

import (
	"fmt"
	"path/filepath"
	"sync"
)

type Engine struct {
	DataDir string
	pagers  map[string]*Pager
	heaps   map[string]*HeapFile
	mu      sync.Mutex
}

func NewEngine(dataDir string) *Engine {
	return &Engine{
		DataDir: dataDir,
		pagers:  make(map[string]*Pager),
		heaps:   make(map[string]*HeapFile),
	}
}

func (e *Engine) GetHeapFile(tableName string) (*HeapFile, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if hf, ok := e.heaps[tableName]; ok {
		return hf, nil
	}

	path := filepath.Join(e.DataDir, tableName+".data")
	pager, err := NewPager(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open table file %s: %w", path, err)
	}

	hf := NewHeapFile(pager)
	e.pagers[tableName] = pager
	e.heaps[tableName] = hf

	return hf, nil
}

func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, p := range e.pagers {
		p.Close()
	}
	return nil
}
