package stash

import (
	"fmt"
	"testing"
)

func TestDecodeEmptyStashFile(t *testing.T) {
	path := "../test_data/transfer_empty.gst"
	d, err := NewDecoder(path)
	if err != nil {
		t.Errorf("Decoder could not be created: %d\n", err)
	}

	next := d.ReadUInt()
	if next != 2 {
		t.Errorf("got: %d\n", next)
	}

	block := d.ReadBlock()
	fmt.Printf("block: %d\n", block)
	if block.result != 18 {
		t.Errorf("block got: %d\n", block.result)
	}

	version := d.ReadUInt()
	fmt.Printf("version: %d\n", version)
	if version > 5 {
		t.Errorf("version: got: %d\n", version)
	}

	zero := d.ReadUIntEx(false)
	if zero != 0 {
		t.Errorf("zero: got: %d\n", zero)
	}

	_, _ = d.ReadString()

	isExpansion := d.ReadBool()
	fmt.Printf("expansion: %t\n", isExpansion)

	ntabs := d.ReadUInt()
	fmt.Printf("ntabs: %d\n", ntabs)

	err = d.ReadBlockEnd(block)
	if err != nil {
		t.Error(err)
	}
	t.Error("done")
}
