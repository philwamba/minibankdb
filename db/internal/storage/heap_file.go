package storage

import (
	"fmt"
	"sync"
)

type HeapFile struct {
	pager *Pager
	mu    sync.Mutex
}

func NewHeapFile(pager *Pager) *HeapFile {
	return &HeapFile{pager: pager}
}

func (hf *HeapFile) Insert(data []byte) (PageID, int, error) {
	hf.mu.Lock()
	defer hf.mu.Unlock()

	count, err := hf.pager.PageCount()
	if err != nil {
		return 0, 0, err
	}

	var page *Page
	var pid PageID

	if count > 0 {
		pid = PageID(count - 1)
		page, err = hf.pager.ReadPage(pid)
		if err != nil {
			return 0, 0, err
		}
	} else {
		pid = 0
		page = &Page{ID: 0, Dirty: true}
	}

	sp := CastPage(page)
	slotID, err := sp.InsertTuple(data)
	if err == ErrPageFull {
		pid = PageID(count)
		if pid == 0 && count == 0 {
			if len(data) > PageSize-PageHeaderSize-SlotSize {
				return 0, 0, fmt.Errorf("tuple too large for page")
			}
		}

		pid = PageID(count)
		page = &Page{ID: pid, Dirty: true}
		sp = CastPage(page)
		slotID, err = sp.InsertTuple(data)
		if err != nil {
			return 0, 0, err
		}
	}

	page.Dirty = true
	if err := hf.pager.WritePage(page); err != nil {
		return 0, 0, err
	}

	return pid, slotID, nil
}

func (hf *HeapFile) ReadTuple(pid PageID, slotID int) ([]byte, error) {
	hf.mu.Lock()
	defer hf.mu.Unlock()

	page, err := hf.pager.ReadPage(pid)
	if err != nil {
		return nil, err
	}
	sp := CastPage(page)
	return sp.GetTuple(slotID), nil
}

type HeapIterator struct {
	hf      *HeapFile
	curPage PageID
	curSlot int
}

func (hf *HeapFile) Iterator() *HeapIterator {
	return &HeapIterator{hf: hf, curPage: 0, curSlot: 0}
}

func (it *HeapIterator) Next() ([]byte, RID, error) {
	it.hf.mu.Lock()
	defer it.hf.mu.Unlock()

	pageCount, err := it.hf.pager.PageCount()
	if err != nil {
		return nil, RID{}, err
	}

	for int(it.curPage) < pageCount {
		page, err := it.hf.pager.ReadPage(it.curPage)
		if err != nil {
			return nil, RID{}, err
		}

		sp := CastPage(page)
		slots := int(sp.Header.SlotCount)

		for it.curSlot < slots {
			tupleBytes := sp.GetTuple(it.curSlot)
			slot := it.curSlot
			it.curSlot++
			if tupleBytes != nil {
				return tupleBytes, RID{PageID: it.curPage, SlotID: slot}, nil
			}
		}

		it.curPage++
		it.curSlot = 0
	}

	return nil, RID{}, nil 
}

func (hf *HeapFile) DeleteTuple(rid RID) error {
	hf.mu.Lock()
	defer hf.mu.Unlock()

	page, err := hf.pager.ReadPage(rid.PageID)
	if err != nil {
		return err
	}

	sp := CastPage(page)
	if err := sp.DeleteTuple(rid.SlotID); err != nil {
		return err
	}

	page.Dirty = true
	return hf.pager.WritePage(page)
}
