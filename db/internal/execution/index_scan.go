package execution

import (
	"minibank/internal/catalog"
	"minibank/internal/indexing"
	"minibank/internal/storage"
)

type IndexScan struct {
	Index    *indexing.HashIndex
	HeapFile *storage.HeapFile
	Key      interface{}
	schema   []catalog.Column

	// Runtime
	rids []storage.RID
	curr int
}

func NewIndexScan(idx *indexing.HashIndex, hf *storage.HeapFile, key interface{}, schema []catalog.Column) *IndexScan {
	return &IndexScan{
		Index:    idx,
		HeapFile: hf,
		Key:      key,
		schema:   schema,
	}
}

func (scan *IndexScan) Open() error {
	scan.rids = scan.Index.Get(scan.Key)
	scan.curr = 0
	return nil
}

func (scan *IndexScan) Next() (*storage.Tuple, error) {
	if scan.curr >= len(scan.rids) {
		return nil, nil
	}

	rid := scan.rids[scan.curr]
	scan.curr++

	bytes, err := scan.HeapFile.ReadTuple(rid.PageID, rid.SlotID)
	if err != nil {
		return nil, err
	}
	if bytes == nil {
		return scan.Next()
	}

	tuple, err := storage.DeserializeTuple(bytes, scan.schema)
	if err != nil {
		return nil, err
	}
	tuple.RID = rid
	return tuple, nil
}

func (scan *IndexScan) Close() error {
	return nil
}

func (scan *IndexScan) Schema() []catalog.Column {
	return scan.schema
}
