package stash

import (
	"fmt"
	"testing"
)

func TestStashLoading(t *testing.T) {
	path := "../test_data/transfer.gst"
	decoder, err := NewDecoder(path)
	if err != nil {
		t.Errorf("Decoder could not be created: %d\n", err)
	}

	next := decoder.ReadUInt()
	if next != 2 {
		t.Errorf("got: %d\n", next)
	}

	block := decoder.ReadBlockStart()
	fmt.Printf("block: %d\n", block)
	if block.result != 18 {
		t.Errorf("block got: %d\n", block.result)
	}

	version := decoder.ReadUInt()
	fmt.Printf("version: %d\n", version)
	if version != 5 {
		t.Errorf("version: got: %d\n", version)
	}

	zero := decoder.ReadUIntEx(false)
	if zero != 0 {
		t.Errorf("zero: got: %d\n", zero)
	}
}
