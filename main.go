package main

import (
	"fmt"

	"github.com/kenranunderscore/grimvault/backend/stash"
)

func main() {
	st, err := stash.ReadStash("./backend/test_data/stashes/transfer.gst")
	if err != nil {
		panic(err)
	}
	fmt.Printf("got %d items in tab\n", len(st.Tabs[2].Items))
	for _, item := range st.Tabs[2].Items {
		fmt.Printf("%s\n", item.Pretty())
	}
}
