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

	next := d.ReadUint()
	if next != 2 {
		t.Errorf("got: %d\n", next)
	}

	mainBlock := d.ReadBlock()
	fmt.Printf("block: %d\n", mainBlock)
	if mainBlock.result != 18 {
		t.Errorf("block got: %d\n", mainBlock.result)
	}

	version := d.ReadUint()
	fmt.Printf("version: %d\n", version)

	zero := d.ReadUintEx(false)
	if zero != 0 {
		t.Errorf("zero: got: %d\n", zero)
	}

	_, _ = d.ReadString()

	if version >= 5 {
		isExpansion := d.ReadBool()
		fmt.Printf("expansion: %t\n", isExpansion)
	}

	ntabs := d.ReadUint()
	fmt.Printf("ntabs: %d\n", ntabs)

	var tabs []StashTab
	for range ntabs {
		err, tab := d.ReadStashTab()
		if err != nil {
			t.Error(err)
		}
		tabs = append(tabs, *tab)
	}

	err = d.ReadBlockEnd(mainBlock)
	if err != nil {
		t.Error(err)
	}
}

func TestDecodeNonEmptyStashFile(t *testing.T) {
	path := "../test_data/transfer.gst"
	d, err := NewDecoder(path)
	if err != nil {
		t.Errorf("Decoder could not be created: %d\n", err)
	}

	next := d.ReadUint()
	if next != 2 {
		t.Errorf("got: %d\n", next)
	}

	mainBlock := d.ReadBlock()
	fmt.Printf("block: %d\n", mainBlock)
	if mainBlock.result != 18 {
		t.Errorf("block got: %d\n", mainBlock.result)
	}

	version := d.ReadUint()
	fmt.Printf("version: %d\n", version)

	zero := d.ReadUintEx(false)
	if zero != 0 {
		t.Errorf("zero: got: %d\n", zero)
	}

	_, _ = d.ReadString()

	if version >= 5 {
		isExpansion := d.ReadBool()
		fmt.Printf("expansion: %t\n", isExpansion)
	}

	ntabs := d.ReadUint()
	fmt.Printf("ntabs: %d\n", ntabs)

	var tabs []StashTab
	for range ntabs {
		err, tab := d.ReadStashTab()
		if err != nil {
			t.Error(err)
		}
		tabs = append(tabs, *tab)
	}

	err = d.ReadBlockEnd(mainBlock)
	if err != nil {
		t.Error(err)
	}
}
