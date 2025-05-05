package arc

import (
	"testing"
)

func TestReadFile(t *testing.T) {
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
