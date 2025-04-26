package database

import (
	"encoding/binary"
	"fmt"
	"os"
)

type reader struct {
	data   *[]byte
	cursor uint
}

func newReader(file string) (error, *reader) {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("could not read %s: %w", file, err), nil
	}

	return nil, &reader{&bytes, 0}
}

func (d *reader) getBytes(count uint) []byte {
	bytes := (*d.data)[d.cursor : d.cursor+count]
	d.cursor += count
	return bytes
}

func (d *reader) readUint16() uint16 {
	bytes := d.getBytes(2)
	return binary.LittleEndian.Uint16(bytes)
}

func (d *reader) readUint32() uint32 {
	bytes := d.getBytes(4)
	return binary.LittleEndian.Uint32(bytes)
}

func (d *reader) readString() string {
	length := uint(d.readUint32())
	remaining := uint(len(*d.data)) - d.cursor
	if length > remaining || length > 10*1024 {
		fmt.Printf("  encountered strange length: %d, remaining: %d\n", length, remaining)
		return ""
	}
	return string(d.getBytes(length))
}

func (d *reader) getStringTable(start uint, byteCount uint) []string {
	end := start + byteCount
	var strings []string
	d.cursor = start
	for d.cursor < end {
		n := d.readUint32()
		for range n {
			strings = append(strings, d.readString())
		}
	}
	return strings
}

func GetItemDeclarations(file string) (error, []string) {
	err, reader := newReader(file)
	if err != nil {
		return err, nil
	}

	tag := reader.readUint16()
	if tag != 2 {
		return fmt.Errorf("unexpected tag: %d", tag), nil
	}

	version := reader.readUint16()
	if version != 3 {
		return fmt.Errorf("unsupported version: %d", version), nil
	}

	_ = reader.readUint32()
	_ = reader.readUint32()
	_ = reader.readUint32()
	stringTableStart := reader.readUint32()
	stringTableSize := reader.readUint32()
	return nil, reader.getStringTable(uint(stringTableStart), uint(stringTableSize))
}
