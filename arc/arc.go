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

func (r *reader) readInt32() int32 {
	return int32(r.readUint32())
}

func (r *reader) readUint64() uint64 {
	b := r.data[r.cursor : r.cursor+8]
	r.cursor += 8
	return binary.LittleEndian.Uint64(b)
}

func (r *reader) readInt64() int64 {
	return int64(r.readUint64())
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
	unknown      int32
	version      int32
	stringCount  int32
	recordCount  int32
	recordSize   int32
	stringSize   int32
	recordOffset int32
}

// FIXME: should these really be ints? everything else uses uints almost
// exclusively
func (r *reader) readHeader() header {
	return header{
		r.readInt32(),
		r.readInt32(),
		r.readInt32(),
		r.readInt32(),
		r.readInt32(),
		r.readInt32(),
		r.readInt32(),
	}
}

type part struct {
	offset           int32
	compressedSize   int32
	uncompressedSize int32
}

func (r *reader) readFileParts(header header) []part {
	parts := make([]part, 0, header.recordCount)
	r.cursor = uint32(header.recordOffset)
	for range header.recordCount {
		p := part{r.readInt32(), r.readInt32(), r.readInt32()}
		parts = append(parts, p)
	}
	return parts
}

func (r *reader) readFileNames(header header) []string {
	files := make([]string, 0, header.stringCount)
	r.cursor = uint32(header.recordOffset + header.recordSize)
	for range header.stringCount {
		s := r.readCString()
		files = append(files, s)
	}
	return files
}

type record struct {
	typ              int32
	offset           int32
	compressedSize   int32
	uncompressedSize int32
	unknown          int32
	time             int64
	partCount        int32
	index            int32
	stringSize       int32
	stringOffset     int32
	data             []byte
	text             string
}

func (r *reader) readRecord() record {
	typ := r.readInt32()
	offset := r.readInt32()
	compressedSize := r.readInt32()
	uncompressedSize := r.readInt32()
	unknown := r.readInt32()
	time := r.readInt64()
	partCount := r.readInt32()
	index := r.readInt32()
	stringSize := r.readInt32()
	stringOffset := r.readInt32()
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
	for range header.stringCount {
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

func (r *reader) readTags(parts []part, files []string, records []record, file string) []Tag {
	blob := ""
	for i := range files {
		if files[i] != file {
			continue
		}

		var record = records[i]
		if record.text == "" {
			// FIXME: hasn't this been uncompressed already and resides in result.data?
			// --> add test and try changing
			data := r.uncompress(parts, record)
			size := len(data)
			var sb strings.Builder
			// sb.Grow(size)
			var lineb strings.Builder
			// lineb.Grow(size >> 3)

			for j := 0; j < size; {
				eof := j == size-1
				current := rune(data[j])
				var next rune
				if eof {
					next = 0
				} else {
					next = rune(data[j+1])
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
	}

	var tags []Tag
	// FIXME:
	lines := strings.SplitSeq(blob, "\n")
	for line := range lines {
		// TODO: trim before this
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && strings.HasPrefix(strings.ToLower(trimmed), "tag") {
			kv := strings.SplitN(trimmed, "=", 2)
			if len(kv) != 2 {
				fmt.Println("    fail: key value could not be parsed")
			} else {
				tag := Tag{kv[0], kv[1]}
				tags = append(tags, tag)
			}
		}
	}

	return tags
}

func (r *reader) readAllTags(parts []part, files []string, records []record) []Tag {
	var tags []Tag
	for _, file := range files {
		ext := filepath.Ext(file)
		if ext != ".txt" {
			continue
		}

		ts := r.readTags(parts, files, records, file)
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
	header := r.readHeader()
	if header.version != 3 {
		return nil, fmt.Errorf("unknown arc header version: %d\n", header.version)
	}

	parts := r.readFileParts(header)
	files := r.readFileNames(header)
	records := r.readRecords(header)

	for i := range records {
		records[i].data = r.uncompress(parts, records[i])
	}

	tags := r.readAllTags(parts, files, records)
	return tags, nil
}
