package main

import (
	"fmt"
	"github.com/kenranunderscore/grimvault/database"
)

func main() {
	path := "./test_data/some.arz"
	res, err := database.GetEntries(path)
	if err != nil {
		panic(err)
	}
	fmt.Printf("loaded %d entries from %s\n", len(res), path)
}
