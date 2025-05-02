package arc

import (
	"testing"
)

func TestReadFile(t *testing.T) {
	tags, err := ReadFile("../test_data/some.arc")
	if err != nil {
		t.Error(err)
	}
	ntags := len(tags)
	if ntags != 11097 {
		t.Errorf("expected 11 strings, got %d\n", ntags)
	}
}
