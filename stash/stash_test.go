package stash

import (
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
}
