package database

import (
	"testing"
)

func TestGetItemDeclarations(t *testing.T) {
	res, err := GetEntries("../test_data/some.arz")
	// res, err := GetItemDeclarations("/home/void/Documents/GrimDawn/gdx1/database/GDX1.arz")
	// res, err := GetItemDeclarations("/home/void/Documents/GrimDawn/database/database.arz")
	if err != nil || len(res) == 0 {
		t.Error(err)
	}
}
