package main

import (
	"fmt"

	"github.com/kenranunderscore/grimvault/arc"
)

func main() {
	// path := "./test_data/some.arz"
	// path := "/home/void/Documents/GrimDawn/database/database.arz"
	// path := "/home/void/Documents/GrimDawn/resources/Text_EN.arc"
	// res, err := database.GetEntries(path)
	// if err != nil {
	// 	panic(err)
	// }
	tags, err := arc.ReadFile("./test_data/arc/some.arc")
	if err != nil {
		panic(fmt.Errorf("could not read arc file: %v", err))
	}
	fmt.Printf("read arc file!, got %d tags\n", len(tags))
	// for range res {
	// 	// fmt.Printf("  stats: %d\n", len(x.Stats))
	// }
	// fmt.Printf("loaded %d entries from %s\n", len(res), path)
}
