package arc

import (
	"encoding/binary"
	"fmt"
	"os"
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

func (r *reader) readCString() string {
	start := r.cursor
	for ; r.data[r.cursor] != 0; r.cursor++ {
	}
	// skip the 0
	r.cursor++
	return string(r.data[start:r.cursor])
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

type filePart struct {
	offset           int32
	compressedSize   int32
	uncompressedSize int32
}

func (r *reader) readFileParts(header header) []filePart {
	parts := make([]filePart, 0, header.recordCount)
	fmt.Printf("  trying to read %d parts\n", header.recordCount)
	r.cursor = uint32(header.recordOffset)
	for range header.recordCount {
		p := filePart{r.readInt32(), r.readInt32(), r.readInt32()}
		parts = append(parts, p)
	}
	return parts
}

func (r *reader) readStrings(header header) []string {
	strings := make([]string, 0, header.stringCount)
	r.cursor = uint32(header.recordOffset + header.recordSize)
	// bytes := r.data[r.cursor : r.cursor+uint32(header.stringSize)]

	// r.cursor = 0
	for range header.stringCount {
		s := r.readCString()
		strings = append(strings, s)
	}
	return strings
}

func ReadFile(file string) (int, error) {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return 0, err
	}

	r := newReader(bytes)
	header := r.readHeader()
	fmt.Printf("got arc header: %+v\n", header)
	if header.version != 3 {
		return 0, fmt.Errorf("unknown arc header version: %d\n", header.version)
	}

	parts := r.readFileParts(header)
	fmt.Printf("found %d parts\n", len(parts))
	for _, p := range parts {
		fmt.Printf("  part: %v\n", p)
	}

	strings := r.readStrings(header)
	fmt.Printf("found %d strings\n", len(strings))
	for _, s := range strings {
		fmt.Printf("  string: %s\n", s)
	}

	return len(strings), nil
}
