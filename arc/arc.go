package arc

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pierrec/lz4"
)

type reader struct {
	data   []byte
	cursor uint32
}

func newReader(data []byte) reader {
	return reader{data, 0}
}

func (r *reader) readUint32() uint32 {
	b := r.data[r.cursor : r.cursor+4]
	r.cursor += 4
	return binary.LittleEndian.Uint32(b)
}

func (r *reader) readUint64() uint64 {
	b := r.data[r.cursor : r.cursor+8]
	r.cursor += 8
	return binary.LittleEndian.Uint64(b)
}

func (r *reader) readCString() string {
	start := r.cursor
	for ; r.data[r.cursor] != 0; r.cursor++ {
	}
	// skip the 0
	r.cursor++
	return string(r.data[start : r.cursor-1])
}

type header struct {
	fileCount    uint32
	recordCount  uint32
	recordSize   uint32
	stringSize   uint32
	recordOffset uint32
}

func (r *reader) readHeader() (header, error) {
	_ = r.readUint32()
	version := r.readUint32()
	if version != 3 {
		return header{}, fmt.Errorf("unknown header version: %d", version)
	}

	return header{
		r.readUint32(),
		r.readUint32(),
		r.readUint32(),
		r.readUint32(),
		r.readUint32(),
	}, nil
}

type part struct {
	offset           uint32
	compressedSize   uint32
	uncompressedSize uint32
}

func (r *reader) readFileParts(header header) []part {
	parts := make([]part, 0, header.recordCount)
	r.cursor = uint32(header.recordOffset)
	for range header.recordCount {
		p := part{r.readUint32(), r.readUint32(), r.readUint32()}
		parts = append(parts, p)
	}
	return parts
}

func (r *reader) readFileNames(header header) []string {
	files := make([]string, 0, header.fileCount)
	r.cursor = uint32(header.recordOffset + header.recordSize)
	for range header.fileCount {
		s := r.readCString()
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
	data             []byte
	text             string
}

func (r *reader) readRecord() record {
	typ := r.readUint32()
	offset := r.readUint32()
	compressedSize := r.readUint32()
	uncompressedSize := r.readUint32()
	unknown := r.readUint32()
	time := r.readUint64()
	partCount := r.readUint32()
	index := r.readUint32()
	stringSize := r.readUint32()
	stringOffset := r.readUint32()
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
		nil,
		"",
	}
}

func (r *reader) readRecords(header header) []record {
	r.cursor = uint32(header.recordOffset + header.recordSize + header.stringSize)
	records := make([]record, 0, header.recordCount)
	for range header.fileCount {
		rec := r.readRecord()
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

func (r *reader) uncompress(parts []part, record record) []byte {
	data := make([]byte, record.uncompressedSize)
	offset := 0
	for i := range int(record.partCount) {
		part := parts[int(record.index)+i]
		compressed := r.data[part.offset : part.offset+part.compressedSize]
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

func (r *reader) readTags(record *record) []Tag {
	blob := ""
	if record.text == "" {
		size := len(record.data)
		var sb strings.Builder
		sb.Grow(size)
		var lineb strings.Builder
		lineb.Grow(size >> 3)

		for j := 0; j < size; {
			eof := j == size-1
			current := rune(record.data[j])
			var next rune
			if eof {
				next = 0
			} else {
				next = rune(record.data[j+1])
			}

			switch current {
			case '\r', '\n':
				if lineb.Len() > 0 {
					sb.WriteString(lineb.String())
					lineb.Reset()
				}
				lineb.WriteByte('\n')
				if current == '\r' && next == '\n' {
					j++
				}
			case '^':
				j++
			default:
				lineb.WriteRune(current)
			}

			j++
		}

		if lineb.Len() > 0 {
			sb.WriteString(lineb.String())
			lineb.Reset()
		}

		blob = sb.String()
		// FIXME: do I need the blob inside the record even?
		record.text = blob
	}

	var tags []Tag
	lines := strings.SplitSeq(blob, "\n")
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

func (r *reader) readAllTags(files []string, records []record) []Tag {
	var tags []Tag
	for i, file := range files {
		ext := filepath.Ext(file)
		if ext != ".txt" {
			continue
		}

		ts := r.readTags(&records[i])
		tags = append(tags, ts...)
	}
	return tags
}

func ReadFile(file string) ([]Tag, error) {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	r := newReader(bytes)
	header, err := r.readHeader()
	if err != nil {
		return nil, fmt.Errorf("could not read header: %+v\n", err)
	}

	parts := r.readFileParts(header)
	files := r.readFileNames(header)
	records := r.readRecords(header)

	for i := range records {
		records[i].data = r.uncompress(parts, records[i])
	}

	tags := r.readAllTags(files, records)
	return tags, nil
}
