package arc

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kenranunderscore/grimvault/rawreader"
	"github.com/pierrec/lz4"
)

type header struct {
	fileCount    uint32
	recordCount  uint32
	recordSize   uint32
	stringSize   uint32
	recordOffset uint32
}

func readHeader(r *rawreader.T) (header, error) {
	_ = r.Uint32()
	version := r.Uint32()
	if version != 3 {
		return header{}, fmt.Errorf("unknown header version: %d", version)
	}

	return header{
		r.Uint32(),
		r.Uint32(),
		r.Uint32(),
		r.Uint32(),
		r.Uint32(),
	}, nil
}

type part struct {
	offset           uint32
	compressedSize   uint32
	uncompressedSize uint32
}

func readFileParts(r *rawreader.T, header header) []part {
	parts := make([]part, 0, header.recordCount)
	r.Seek(header.recordOffset)
	for range header.recordCount {
		p := part{r.Uint32(), r.Uint32(), r.Uint32()}
		parts = append(parts, p)
	}
	return parts
}

func readFileNames(r *rawreader.T, header header) []string {
	files := make([]string, 0, header.fileCount)
	r.Seek(uint32(header.recordOffset + header.recordSize))
	for range header.fileCount {
		s := r.CString()
		files = append(files, s)
	}
	return files
}

type record struct {
	typ              uint32
	offset           uint32
	compressedSize   uint32
	uncompressedSize uint32
	unknown          uint32
	time             uint64
	partCount        uint32
	index            uint32
	stringSize       uint32
	stringOffset     uint32
	text             string
}

func readRecord(r *rawreader.T) record {
	typ := r.Uint32()
	offset := r.Uint32()
	compressedSize := r.Uint32()
	uncompressedSize := r.Uint32()
	unknown := r.Uint32()
	time := r.Uint64()
	partCount := r.Uint32()
	index := r.Uint32()
	stringSize := r.Uint32()
	stringOffset := r.Uint32()
	return record{
		typ,
		offset,
		compressedSize,
		uncompressedSize,
		unknown,
		time,
		partCount,
		index,
		stringSize,
		stringOffset,
		"",
	}
}

func readRecords(r *rawreader.T, header header) []record {
	r.Seek(uint32(header.recordOffset + header.recordSize + header.stringSize))
	records := make([]record, 0, header.recordCount)
	for range header.fileCount {
		rec := readRecord(r)
		// NOTE: Sometimes we hit "records" with uncompressed size 0. In those
		// cases the subsequent record has always had the same index as this one
		// so far, and there have always been `header.recordCount` many records.
		// We thus maintain the invariant `records[i].index == i`.
		if rec.uncompressedSize > 0 {
			records = append(records, rec)
		}
	}
	return records
}

type Tag struct {
	Tag  string
	Name string
}

func uncompress(r *rawreader.T, parts []part, record record) []byte {
	data := make([]byte, record.uncompressedSize)
	offset := 0
	for i := range int(record.partCount) {
		part := parts[int(record.index)+i]
		compressed := r.BytesFrom(part.offset, part.compressedSize)
		if part.compressedSize == part.uncompressedSize {
			for j := range int(part.uncompressedSize) {
				data[offset+j] = compressed[j]
			}
		} else {
			lz4.UncompressBlock(compressed, data)
		}
		offset += int(part.uncompressedSize)
	}
	return data
}

func readTags(r *rawreader.T, record *record) []Tag {
	var tags []Tag
	lines := strings.SplitSeq(record.text, "\n")
	for line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && strings.HasPrefix(strings.ToLower(trimmed), "tag") {
			kv := strings.SplitN(trimmed, "=", 2)
			if len(kv) != 2 {
				fmt.Println("    fail: key value could not be parsed")
				fmt.Printf("       kv == %v\n", kv)
			} else {
				tag := Tag{kv[0], kv[1]}
				tags = append(tags, tag)
			}
		}
	}

	return tags
}

func readAllTags(r *rawreader.T, files []string, records []record) []Tag {
	var tags []Tag
	for i, file := range files {
		ext := filepath.Ext(file)
		if ext != ".txt" {
			continue
		}

		ts := readTags(r, &records[i])
		tags = append(tags, ts...)
	}
	return tags
}

func ReadFile(file string) ([]Tag, error) {
	r, err := rawreader.FromFile(file)
	if err != nil {
		return nil, err
	}

	header, err := readHeader(r)
	if err != nil {
		return nil, fmt.Errorf("could not read header: %+v\n", err)
	}

	parts := readFileParts(r, header)
	files := readFileNames(r, header)
	records := readRecords(r, header)

	for i := range records {
		records[i].text = string(uncompress(r, parts, records[i]))
	}

	tags := readAllTags(r, files, records)
	return tags, nil
}
