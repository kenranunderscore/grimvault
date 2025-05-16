package stash

import (
	"testing"

	"github.com/kenranunderscore/grimvault/backend/golden"
)

func TestDecodeEmptyStashFile(t *testing.T) {
	t.Parallel()

	stash, err := ReadStash("../test_data/stashes/transfer_empty.gst")
	if err != nil {
		t.Fatalf("could not read stash: %v", err)
	}

	if tabCount := len(stash.Tabs); tabCount != 4 {
		t.Errorf("expected 4 tabs, got %d", tabCount)
	}

	for i := range stash.Tabs {
		if itemCount := len(stash.Tabs[i].Items); itemCount != 0 {
			t.Errorf("expected empty stash tab, found %d items", itemCount)
		}
	}
}

func TestDecodeNonEmptyStashFile(t *testing.T) {
	golden.Run(t, "../test_data/stashes", "*.gst", func(t *testing.T, file string) any {
		t.Parallel()
		stash, err := ReadStash(file)
		if err != nil {
			t.Fatalf("could not read stash: %v", err)
		}

		return stash
	})
}
