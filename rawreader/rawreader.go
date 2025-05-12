package rawreader

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

// A read-only view on some binary data.
type T struct {
	data   []byte
	cursor uint32
}

// Create a new reader.
func New(data []byte) *T {
	return &T{data, 0}
}

// Create a new reader from a file by reading in all its data at once.
func FromFile(file string) (*T, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", file, err)
	}

	return New(data), nil
}

// Jump to the specified `cursor` position.
func (reader *T) Seek(cursor uint32) {
	reader.cursor = cursor
}

// Move `offset` bytes relative to the current cursor position.
func (reader *T) Advance(offset uint32) {
	reader.cursor += offset
}

// From the position `start`, get a slice of `count` bytes.
func (reader *T) BytesFrom(start uint32, count uint32) []byte {
	bytes := reader.data[start : start+count]
	reader.cursor += count
	return bytes
}

// Get a slice of `count` bytes from the current cursor position.
func (reader *T) Bytes(count uint32) []byte {
	return reader.BytesFrom(reader.cursor, count)
}

// Read a `uint16` at the current position.
func (reader *T) Uint16() uint16 {
	bytes := reader.Bytes(2)
	return binary.LittleEndian.Uint16(bytes)
}

// Read a `uint32` at the current position.
func (r *T) Uint32() uint32 {
	bytes := r.Bytes(4)
	return binary.LittleEndian.Uint32(bytes)
}

// Read a `uint64` at the current position.
func (r *T) Uint64() uint64 {
	b := r.Bytes(8)
	return binary.LittleEndian.Uint64(b)
}

// Read a `float32` at the current position.
func (r *T) Float32() float32 {
	n := r.Uint32()
	return math.Float32frombits(n)
}

// Read a `string` at the current position.
func (r *T) String() string {
	length := r.Uint32()
	remaining := uint32(len(r.data)) - r.cursor
	if length > remaining || length > 10*1024 {
		return ""
	}
	return string(r.Bytes(length))
}

// Read a C string at the current position; that is, read until encountering a
// terminal '\0'.
//
// FIXME: which one of the string variants is really necessary?
func (r *T) CString() string {
	start := r.cursor
	for ; r.data[r.cursor] != 0; r.cursor++ {
	}
	// skip the 0
	r.cursor++
	return string(r.data[start : r.cursor-1])
}
