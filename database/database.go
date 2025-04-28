package database

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/pierrec/lz4"
)

type reader struct {
	data   []byte
	cursor uint32
}

func newReader(data []byte) *reader {
	return &reader{data, 0}
}

func readFile(file string) (*reader, error) {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", file, err)
	}

	return newReader(bytes), nil
}

func (r *reader) getBytes(count uint32) []byte {
	bytes := r.data[r.cursor : r.cursor+count]
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

func (r *reader) readFloat32() float32 {
	n := r.readUint32()
	return math.Float32frombits(n)
}

func (r *reader) readString() string {
	length := r.readUint32()
	remaining := uint32(len(r.data)) - r.cursor
	if length > remaining || length > 10*1024 {
		return ""
	}
	return string(r.getBytes(length))
}

type stringTable []string

func (r *reader) getStringTable(start uint32, byteCount uint32) stringTable {
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
	metas := make([]recordMeta, 0, count)
	for range count {
		metas = append(metas, r.readRecordMeta())
	}
	return metas
}

type uncompressedRecord struct {
	typ         string
	stringIndex uint32
	data        []byte
}

func (r *reader) readRecord(meta recordMeta) (uncompressedRecord, error) {
	r.cursor = meta.pos + 24
	compressed := r.getBytes(meta.compressedSize)
	uncompressed := make([]byte, meta.uncompressedSize)
	_, err := lz4.UncompressBlock(compressed, uncompressed)
	if err != nil {
		return uncompressedRecord{}, err
	}
	return uncompressedRecord{meta.name, meta.stringIndex, uncompressed}, nil
}

type stat struct {
	name string
	// FIXME: go has no sum types, so what's the idiom here?
	value any
}

type Entry struct {
	Key   string
	Stats []stat
}

func (rec *uncompressedRecord) toEntry(strings stringTable) (Entry, error) {
	key := strings[rec.stringIndex]
	r := newReader(rec.data)
	var i uint32
	var offset uint32
	var stats []stat
	for int(i) < len(rec.data)/4 {
		r.cursor = offset
		typeId := r.readUint16()
		entryCount := r.readUint16()
		stringIndex := r.readUint32()

		i += 2 + uint32(entryCount)
		name := strings[stringIndex]
		for n := uint32(0); n < uint32(entryCount); n++ {
			r.cursor = offset + 8 + 4*n
			switch typeId {
			case 1:
				f := r.readFloat32()
				if math.Abs(float64(f)) > 0.01 {
					stats = append(stats, stat{name, f})
				}
			case 2:
				index := r.readUint32()
				if int(index) < len(strings) {
					value := strings[int(index)]
					if value != "" {
						stats = append(stats, stat{name, value})
					}
				}
			default:
				value := r.readUint32()
				if value > 0 {
					stats = append(stats, stat{name, value})
				}
			}
		}
		offset += 8 + 4*uint32(entryCount)
	}
	return Entry{key, stats}, nil
}

func GetEntries(file string) ([]Entry, error) {
	reader, err := readFile(file)
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
	var decls []uncompressedRecord
	for _, meta := range metas {
		item, err := reader.readRecord(meta)
		if err != nil {
			return nil, err
		}
		decls = append(decls, item)
	}

	var items []Entry
	for _, decl := range decls {
		it, err := decl.toEntry(strings)
		if err != nil {
			return items, err
		}
		// fmt.Printf("    ---> got entry %s with %d stats\n", it.key, len(it.stats))
		items = append(items, it)
	}

	return items, nil
}
