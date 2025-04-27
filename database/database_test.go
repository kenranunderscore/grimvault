package database

import (
	"testing"
)

func TestGetItemDeclarations(t *testing.T) {
	res, err := GetItemDeclarations("../test_data/some.arz")
	if err != nil || len(res) == 0 {
		t.Error(err)
	}

	t.Fail()
}
