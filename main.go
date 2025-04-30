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
	_, err := arc.ReadFile("./test_data/some.arc")
	if err != nil {
		fmt.Println("read arc file!")
	}
	// for range res {
	// 	// fmt.Printf("  stats: %d\n", len(x.Stats))
	// }
	// fmt.Printf("loaded %d entries from %s\n", len(res), path)
}
