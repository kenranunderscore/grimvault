package arc

import (
	"testing"

	"github.com/kenranunderscore/grimvault/backend/golden"
)

func TestReadFile(t *testing.T) {
	golden.Run(t, "../test_data/arc", "*.arc", func(t *testing.T, arc string) any {
		t.Parallel()

		tags, err := ReadFile(arc)
		if err != nil {
			t.Fatalf("Could not parse '%s': %v", arc, err)
		}

		return tags
	})
}

func TestReadFileReturnsCorrectNumberOfTags(t *testing.T) {
	t.Parallel()

	tags, err := ReadFile("../test_data/arc/some.arc")
	if err != nil {
		t.Fatal(err)
	}

	ntags := len(tags)
	if ntags != 11097 {
		t.Errorf("expected 11 strings, got %d\n", ntags)
	}
}
