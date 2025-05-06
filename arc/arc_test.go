package arc

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

var updateGoldFiles = flag.Bool("update", false, "update golden files")

func TestReadFile(t *testing.T) {
	t.Parallel()
	flag.Parse()

	testData := "../test_data/arc"
	expectedDir := filepath.Join(testData, "expected")

	arcFiles, err := filepath.Glob(filepath.Join(testData, "*.arc"))
	if err != nil {
		t.Fatalf("Could not read test files: %v", err)
	}

	for _, arc := range arcFiles {
		base := filepath.Base(arc)
		name := base[:len(base)-len(filepath.Ext(base))]
		t.Run(name, func(t *testing.T) {
			tags, err := ReadFile(arc)
			if err != nil {
				t.Fatalf("Could not parse '%s': %v", arc, err)
			}

			content, err := cbor.Marshal(tags)
			if err != nil {
				t.Fatalf("Could not serialize tags: %v", err)
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
