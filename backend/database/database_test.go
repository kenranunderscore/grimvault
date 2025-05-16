package database

import (
	"testing"

	"github.com/kenranunderscore/grimvault/backend/golden"
)

func TestAllDatabasesGolden(t *testing.T) {
	golden.Run(t, "../test_data/arz", "*.arz", func(t *testing.T, arz string) any {
		t.Parallel()

		entries, err := GetEntries(arz)
		if err != nil {
			t.Fatalf("Could not parse '%s': %v", arz, err)
		}

		return entries
	})
}
