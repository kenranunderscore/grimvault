package database

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

var updateGoldFiles = flag.Bool("update", false, "update golden files")

func TestAllDatabasesGolden(t *testing.T) {
	t.Parallel()
	flag.Parse()
	testData := "../test_data/arz"
	expectedDir := filepath.Join(testData, "expected")

	arzFiles, err := filepath.Glob(filepath.Join(testData, "*.arz"))
	if err != nil {
		t.Fatalf("Could not read test files: %v", err)
	}

	for _, arz := range arzFiles {
		base := filepath.Base(arz)
		name := base[:len(base)-len(filepath.Ext(base))]
		t.Run(name, func(t *testing.T) {
			entries, err := GetEntries(arz)
			if err != nil {
				t.Fatalf("Could not parse '%s': %v", arz, err)
			}

			content, err := cbor.Marshal(entries)
			if err != nil {
				t.Fatalf("Could not serialize entries: %v", err)
			}

			goldFile := filepath.Join(expectedDir, name+".gold")
			if *updateGoldFiles {
				if err := os.WriteFile(goldFile, content, 0o644); err != nil {
					t.Fatalf("Could not write gold file '%s': %v", goldFile, err)
				}
				t.Logf("Updated gold file '%s'", goldFile)
				return
			}

			expected, err := os.ReadFile(goldFile)
			if err != nil {
				t.Fatalf("Could not read gold file '%s': %v", goldFile, err)
			}

			if !bytes.Equal(content, expected) {
				t.Errorf("Output mismatch")
			}
		})
	}
}
