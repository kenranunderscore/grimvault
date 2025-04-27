package database

import (
	"encoding/binary"
	"fmt"
	"github.com/pierrec/lz4"
	"os"
)

// FIXME: try bytes.NewReader: does that replace it?
type reader struct {
	data   *[]byte
	cursor uint32
}

func newReader(file string) (*reader, error) {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", file, err)
	}

	return &reader{&bytes, 0}, nil
}

func (r *reader) getBytes(count uint32) []byte {
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
	length := r.readUint32()
	remaining := uint32(len(*r.data)) - r.cursor
	if length > remaining || length > 10*1024 {
		// fmt.Printf("  encountered strange length: %d, remaining: %d\n", length, remaining)
		return ""
	}
	return string(r.getBytes(length))
}

func (r *reader) getStringTable(start uint32, byteCount uint32) []string {
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

type recordMeta struct {
	stringIndex      uint32
	name             string
	pos              uint32
	compressedSize   uint32
	uncompressedSize uint32
}

func (r *reader) readRecordMeta() recordMeta {
	stringIndex := r.readUint32()
	name := r.readString()
	offset := r.readUint32()
	compressedSize := r.readUint32()
	uncompressedSize := r.readUint32()
	r.cursor += 8
	return recordMeta{stringIndex, name, offset, compressedSize, uncompressedSize}
}

func (r *reader) readRecordMetas(start uint32, count uint32) []recordMeta {
	r.cursor = start
	fmt.Printf("trying to read %d meta records\n", count)
	var metas []recordMeta
	for range count {
		metas = append(metas, r.readRecordMeta())
	}
	return metas
}

type itemDeclaration struct {
	typ  string
	stringIndex uint32
	data []byte
}

func (r *reader) readItemDeclaration(meta recordMeta) (itemDeclaration, error) {
	r.cursor = meta.pos + 24
	compressed := r.getBytes(meta.compressedSize)
	uncompressed := make([]byte, meta.uncompressedSize)
	_, err := lz4.UncompressBlock(compressed, uncompressed)
	if err != nil {
		return itemDeclaration{}, err
	}
	return itemDeclaration{meta.name, meta.stringIndex, uncompressed}, nil
}

func GetItemDeclarations(file string) ([]itemDeclaration, error) {
	reader, err := newReader(file)
	if err != nil {
		return nil, err
	}

	tag := reader.readUint16()
	if tag != 2 {
		return nil, fmt.Errorf("unexpected tag: %d", tag)
	}

	version := reader.readUint16()
	if version != 3 {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}

	recordStart := reader.readUint32()
	_ = reader.readUint32()
	recordCount := reader.readUint32()
	stringStart := reader.readUint32()
	stringByteCount := reader.readUint32()

	strings := reader.getStringTable(stringStart, stringByteCount)
	fmt.Printf("found %d strings in %s\n", len(strings), file)

	// FIXME: this is inefficient/naive as a first pass. use more efficient
	// arguments (iterator, pointer), but measure beforehand to learn things
	// about go!
	metas := reader.readRecordMetas(recordStart, recordCount)
	var decls []itemDeclaration
	for _, meta := range metas {
		item, err := reader.readItemDeclaration(meta)
		if err != nil {
			return decls, err
		}
		decls = append(decls, item)
	}

	for _, decl := range decls {
		name := strings[decl.stringIndex]
		fmt.Printf("  got item '%s'\n", name)
	}

	return decls, nil
}
