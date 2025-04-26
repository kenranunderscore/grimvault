package database

import (
	"encoding/binary"
	"fmt"
	"os"
)

type reader struct {
	data   *[]byte
	cursor uint
}

func newReader(file string) (error, *reader) {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("could not read %s: %w", file, err), nil
	}

	return nil, &reader{&bytes, 0}
}

func (r *reader) getBytes(count uint) []byte {
	bytes := (*r.data)[r.cursor : r.cursor+count]
	r.cursor += count
	return bytes
}

func (r *reader) readUint16() uint16 {
	bytes := r.getBytes(2)
	return binary.LittleEndian.Uint16(bytes)
}

func (r *reader) readUint32() uint32 {
	bytes := r.getBytes(4)
	return binary.LittleEndian.Uint32(bytes)
}

func (r *reader) readString() string {
	length := uint(r.readUint32())
	remaining := uint(len(*r.data)) - r.cursor
	if length > remaining || length > 10*1024 {
		// fmt.Printf("  encountered strange length: %d, remaining: %d\n", length, remaining)
		return ""
	}
	return string(r.getBytes(length))
}

func (r *reader) getStringTable(start uint, byteCount uint) []string {
	end := start + byteCount
	var strings []string
	r.cursor = start
	for r.cursor < end {
		n := r.readUint32()
		for range n {
			strings = append(strings, r.readString())
		}
	}
	return strings
}

type itemDeclarationMeta struct {
	stringIndex      uint32
	name             string
	pos              uint32
	compressedSize   uint32
	uncompressedSize uint32
}

func (r *reader) readItemDeclarationMeta() itemDeclarationMeta {
	stringIndex := r.readUint32()
	name := r.readString()
	offset := r.readUint32()
	compressedSize := r.readUint32()
	uncompressedSize := r.readUint32()
	r.cursor += 8
	return itemDeclarationMeta{stringIndex, name, offset, compressedSize, uncompressedSize}
}

func (r *reader) readItemDeclarations(start uint, count uint) []itemDeclarationMeta {
	r.cursor = start
	fmt.Printf("trying to read %d meta records\n", count)
	var metas []itemDeclarationMeta
	for range count {
		metas = append(metas, r.readItemDeclarationMeta())
	}
	return metas
}

func GetItemDeclarations(file string) (error, []itemDeclarationMeta) {
	err, reader := newReader(file)
	if err != nil {
		return err, nil
	}

	tag := reader.readUint16()
	if tag != 2 {
		return fmt.Errorf("unexpected tag: %d", tag), nil
	}

	version := reader.readUint16()
	if version != 3 {
		return fmt.Errorf("unsupported version: %d", version), nil
	}

	recordStart := reader.readUint32()
	_ = reader.readUint32()
	recordCount := reader.readUint32()
	stringStart := reader.readUint32()
	stringByteCount := reader.readUint32()

	strings := reader.getStringTable(uint(stringStart), uint(stringByteCount))
	fmt.Printf("found %d strings in %s\n", len(strings), file)

	// FIXME: this is inefficient/naive as a first pass. use more efficient
	// arguments (iterator, pointer), but measure beforehand to learn things
	// about go!
	metas := reader.readItemDeclarations(uint(recordStart), uint(recordCount))

	return nil, metas
}
