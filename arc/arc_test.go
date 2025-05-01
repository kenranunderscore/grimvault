package arc

import (
	"testing"
)

func TestReadFile(t *testing.T) {
	nstrings, err := ReadFile("../test_data/some.arc")
	if err != nil {
		t.Error(err)
	}
	if nstrings != 11 {
		t.Errorf("expected 11 strings, got %d\n", nstrings)
	}
}
