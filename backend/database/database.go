package database

import (
	"fmt"
	"math"

	"github.com/kenranunderscore/grimvault/backend/rawreader"
	"github.com/pierrec/lz4"
)

type stringTable []string

func getStringTable(r *rawreader.T, start uint32) stringTable {
	var strings []string
	r.Seek(start)
	count := r.Uint32()
	fmt.Printf("    reading %d strings\n", count)
	for range count {
		strings = append(strings, r.String())
	}
	return strings
}

type record struct {
	stringIndex      uint32
	name             string
	offset           uint32
	compressedSize   uint32
	uncompressedSize uint32
	data             []byte
}

func readRecord(r *rawreader.T) record {
	stringIndex := r.Uint32()
	name := r.String()
	offset := r.Uint32()
	compressedSize := r.Uint32()
	uncompressedSize := r.Uint32()
	r.Advance(8)
	return record{stringIndex, name, offset, compressedSize, uncompressedSize, nil}
}

func readRecords(r *rawreader.T, start uint32, count uint32) []record {
	r.Seek(start)
	fmt.Printf("reading %d records\n", count)
	records := make([]record, 0, count)
	for range count {
		records = append(records, readRecord(r))
	}
	return records
}

func uncompress(r *rawreader.T, rec *record) error {
	r.Seek(rec.offset + 24)
	compressed := r.Bytes(rec.compressedSize)
	rec.data = make([]byte, rec.uncompressedSize)
	_, err := lz4.UncompressBlock(compressed, rec.data)
	if err != nil {
		return err
	}
	return nil
}

type Stat struct {
	Name string
	// FIXME: go has no sum types, so what's the idiom here?
	Value any
}

type Entry struct {
	Key   string
	Stats []Stat
}

func (rec *record) toEntry(strings stringTable) (Entry, error) {
	key := strings[rec.stringIndex]
	r := rawreader.New(rec.data)
	var i uint32
	var offset uint32
	var stats []Stat
	for int(i) < len(rec.data)/4 {
		r.Seek(offset)
		typeId := r.Uint16()
		entryCount := r.Uint16()
		stringIndex := r.Uint32()

		i += 2 + uint32(entryCount)
		name := strings[stringIndex]
		for n := uint32(0); n < uint32(entryCount); n++ {
			r.Seek(offset + 8 + 4*n)
			switch typeId {
			case 1:
				f := r.Float32()
				if math.Abs(float64(f)) > 0.01 {
					stats = append(stats, Stat{name, f})
				}
			case 2:
				index := r.Uint32()
				if int(index) < len(strings) {
					value := strings[int(index)]
					if value != "" {
						stats = append(stats, Stat{name, value})
					}
				}
			default:
				value := r.Uint32()
				if value > 0 {
					stats = append(stats, Stat{name, value})
				}
			}
		}
		offset += 8 + 4*uint32(entryCount)
	}
	return Entry{key, stats}, nil
}

func GetEntries(file string) ([]Entry, error) {
	r, err := rawreader.FromFile(file)
	if err != nil {
		return nil, err
	}

	tag := r.Uint16()
	if tag != 2 {
		return nil, fmt.Errorf("unexpected tag: %d", tag)
	}

	version := r.Uint16()
	if version != 3 {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}

	recordStart := r.Uint32()
	_ = r.Uint32()
	recordCount := r.Uint32()
	stringStart := r.Uint32()
	_ = r.Uint32()

	strings := getStringTable(r, stringStart)
	fmt.Printf("  found %d strings in %s\n", len(strings), file)

	records := readRecords(r, recordStart, recordCount)
	for i := range records {
		err := uncompress(r, &records[i])
		if err != nil {
			return nil, err
		}
	}

	var items []Entry
	for i := range records {
		it, err := records[i].toEntry(strings)
		if err != nil {
			return items, err
		}
		items = append(items, it)
	}

	return items, nil
}
