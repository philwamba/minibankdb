package storage

import (
	"encoding/binary"
	"errors"
)

const (
	PageHeaderSize = 4
	SlotSize       = 4
)

var ErrPageFull = errors.New("page full")
var ErrSlotInvalid = errors.New("invalid slot id")

type SlottedPage struct {
	Header *PageHeader
	Body   []byte
}

type PageHeader struct {
	SlotCount        uint16
	FreeSpacePointer uint16
}

func CastPage(p *Page) *SlottedPage {
	sp := &SlottedPage{
		Body: p.Data[:],
	}
	sp.Header = &PageHeader{}
	sp.readHeader()

	if sp.Header.SlotCount == 0 && sp.Header.FreeSpacePointer == 0 {
		sp.Header.FreeSpacePointer = PageSize
		sp.writeHeader()
	}
	return sp
}

func (sp *SlottedPage) readHeader() {
	sp.Header.SlotCount = binary.BigEndian.Uint16(sp.Body[0:2])
	sp.Header.FreeSpacePointer = binary.BigEndian.Uint16(sp.Body[2:4])
}

func (sp *SlottedPage) writeHeader() {
	binary.BigEndian.PutUint16(sp.Body[0:2], sp.Header.SlotCount)
	binary.BigEndian.PutUint16(sp.Body[2:4], sp.Header.FreeSpacePointer)
}

func (sp *SlottedPage) FreeSpace() int {
	slotsEnd := PageHeaderSize + int(sp.Header.SlotCount)*SlotSize
	return int(sp.Header.FreeSpacePointer) - slotsEnd
}

func (sp *SlottedPage) InsertTuple(data []byte) (int, error) {
	required := len(data) + SlotSize
	if sp.FreeSpace() < required {
		return -1, ErrPageFull
	}

	slotID := int(sp.Header.SlotCount)
	sp.Header.SlotCount++

	newOffset := int(sp.Header.FreeSpacePointer) - len(data)
	sp.Header.FreeSpacePointer = uint16(newOffset)

	copy(sp.Body[newOffset:], data)

	sp.writeSlot(slotID, uint16(newOffset), uint16(len(data)))
	sp.writeHeader()

	return slotID, nil
}

func (sp *SlottedPage) writeSlot(id int, offset, length uint16) {
	pos := PageHeaderSize + id*SlotSize
	binary.BigEndian.PutUint16(sp.Body[pos:pos+2], offset)
	binary.BigEndian.PutUint16(sp.Body[pos+2:pos+4], length)
}

func (sp *SlottedPage) GetSlot(id int) (uint16, uint16) {
	pos := PageHeaderSize + id*SlotSize
	offset := binary.BigEndian.Uint16(sp.Body[pos : pos+2])
	length := binary.BigEndian.Uint16(sp.Body[pos+2 : pos+4])
	return offset, length
}

func (sp *SlottedPage) GetTuple(id int) []byte {
	if id >= int(sp.Header.SlotCount) {
		return nil
	}
	offset, length := sp.GetSlot(id)
	if length == 0 {
		return nil
	}
	return sp.Body[offset : offset+length]
}

func (sp *SlottedPage) DeleteTuple(id int) error {
	if id >= int(sp.Header.SlotCount) {
		return ErrSlotInvalid
	}
	offset, _ := sp.GetSlot(id)
	sp.writeSlot(id, offset, 0)
	sp.writeHeader()
	return nil
}
