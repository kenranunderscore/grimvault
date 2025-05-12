package rawreader

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

type T struct {
	data   []byte
	cursor uint32
}

func New(data []byte) *T {
	return &T{data, 0}
}

func FromFile(file string) (*T, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", file, err)
	}

	return New(data), nil
}

func (reader *T) Seek(cursor uint32) {
	reader.cursor = cursor
}

func (reader *T) Advance(offset uint32) {
	reader.cursor += offset
}

func (reader *T) BytesFrom(start uint32, count uint32) []byte {
	bytes := reader.data[start : start+count]
	reader.cursor += count
	return bytes
}

func (reader *T) Bytes(count uint32) []byte {
	return reader.BytesFrom(reader.cursor, count)
}

func (reader *T) Uint16() uint16 {
	bytes := reader.Bytes(2)
	return binary.LittleEndian.Uint16(bytes)
}

func (r *T) Uint32() uint32 {
	bytes := r.Bytes(4)
	return binary.LittleEndian.Uint32(bytes)
}

func (r *T) Uint64() uint64 {
	b := r.Bytes(8)
	return binary.LittleEndian.Uint64(b)
}

func (r *T) Float32() float32 {
	n := r.Uint32()
	return math.Float32frombits(n)
}

func (r *T) String() string {
	length := r.Uint32()
	remaining := uint32(len(r.data)) - r.cursor
	if length > remaining || length > 10*1024 {
		return ""
	}
	return string(r.Bytes(length))
}

// FIXME: which one of the string variants is really necessary?
func (r *T) CString() string {
	start := r.cursor
	for ; r.data[r.cursor] != 0; r.cursor++ {
	}
	// skip the 0
	r.cursor++
	return string(r.data[start : r.cursor-1])
}
