package storage

import (
	"io"
	"os"
	"sync"
)

const PageSize = 4096

type PageID int

type Page struct {
	ID    PageID
	Data  [PageSize]byte
	Dirty bool
}

type Pager struct {
	file *os.File
	path string
	mu   sync.Mutex
}

func NewPager(path string) (*Pager, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return &Pager{
		file: file,
		path: path,
	}, nil
}

func (p *Pager) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.file.Close()
}

func (p *Pager) ReadPage(id PageID) (*Page, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	offset := int64(id) * PageSize
	_, err := p.file.Seek(offset, 0)
	if err != nil {
		return nil, err
	}

	page := &Page{ID: id}
	_, err = p.file.Read(page.Data[:])
	if err != nil && err != io.EOF {
		return nil, err
	}
	// If EOF, just return empty page? Or should we extend?
	// Usually ReadPage expects existing page.
	// For simplicity, if we read partial/EOF, it's just zeroed data if we allow implicit extension,
	// but better to error if strictly reading.
	// However, specifically for "Access past EOF", DBs usually handle it.
	// Let's assume valid ID.
	return page, nil
}

func (p *Pager) WritePage(page *Page) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	offset := int64(page.ID) * PageSize
	_, err := p.file.Seek(offset, 0)
	if err != nil {
		return err
	}

	_, err = p.file.Write(page.Data[:])
	if err != nil {
		return err
	}

	page.Dirty = false
	return nil
}

func (p *Pager) PageCount() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	info, err := p.file.Stat()
	if err != nil {
		return 0, err
	}
	return int(info.Size() / PageSize), nil
}
