package main

import (
	"fmt"
	"github.com/kenranunderscore/grimvault/stash"
)

func main() {
	decoder, err := stash.NewDecoder("test_data/transfer.gst")
	if err != nil {
		panic(err)
	}

	fmt.Printf("The stash: %d\n", decoder.ReadUInt())
}
