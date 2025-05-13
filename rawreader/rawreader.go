package rawreader

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

// A read-only view on some binary data.
type T struct {
	Data   []byte
	Cursor uint32
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
	reader.Cursor = cursor
}

// Move `offset` bytes relative to the current cursor position.
func (reader *T) Advance(offset uint32) {
	reader.Cursor += offset
}

// From the position `start`, get a slice of `count` bytes.
func (reader *T) BytesFrom(start uint32, count uint32) []byte {
	bytes := reader.Data[start : start+count]
	reader.Cursor += count
	return bytes
}

// Get a slice of `count` bytes from the current cursor position.
func (reader *T) Bytes(count uint32) []byte {
	return reader.BytesFrom(reader.Cursor, count)
}

// Read a `byte` at the current position.
func (reader *T) Byte() byte {
	b := reader.Data[reader.Cursor]
	reader.Advance(1)
	return b
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
//
// This expects the data at the current position to start with a `uint32`
// containing the length of the string to read.
func (r *T) String() string {
	length := r.Uint32()
	return string(r.Bytes(length))
}

// Read a C string at the current position; that is, read until encountering a
// terminal '\0'.
func (r *T) CString() string {
	start := r.Cursor
	for r.Byte() != 0 {
	}
	return string(r.Data[start : r.Cursor-1])
}
