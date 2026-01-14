package indexing

import (
	"minibank/internal/storage"
	"sync"
)

type HashIndex struct {
	Items map[interface{}][]storage.RID
	mu    sync.RWMutex
}

func NewHashIndex() *HashIndex {
	return &HashIndex{
		Items: make(map[interface{}][]storage.RID),
	}
}

func (idx *HashIndex) Insert(key interface{}, rid storage.RID) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.Items[key] = append(idx.Items[key], rid)
}

func (idx *HashIndex) Get(key interface{}) []storage.RID {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.Items[key]
}

func (idx *HashIndex) Delete(key interface{}, rid storage.RID) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	rids := idx.Items[key]
	for i, r := range rids {
		if r == rid {
			idx.Items[key] = append(rids[:i], rids[i+1:]...)
			return
		}
	}
}
