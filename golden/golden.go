package golden

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

var updateGoldenFiles = flag.Bool("update", false, "update golden files")

// A function that takes an input `file` and should return the value to be
// tested against the contents of the corresponding golden file.
//
// Its output is also used to initially create or update the golden files.
type GoldenTestFunc func(t *testing.T, file string) any

// For each file in directory `dir` that matches `glob`, run a sub-test calling
// `f` on said file and comparing its output with that of the corresponding
// golden file.
//
// Those golden files live in the "expected" directory in `dir`. Passing the
// `-update` flag to `go test` leads to initial creation or update of the golden
// files from the output of `f`.
func Run(t *testing.T, dir string, glob string, f GoldenTestFunc) {
	t.Parallel()

	if !flag.Parsed() {
		flag.Parse()
	}

	expectedDir := filepath.Join(dir, "expected")
	testFiles, err := filepath.Glob(filepath.Join(dir, glob))
	if err != nil {
		t.Fatalf("Failed reading test files: %v", err)
	}

	for _, file := range testFiles {
		base := filepath.Base(file)
		testName := base[:len(base)-len(filepath.Ext(base))]
		t.Run(testName, func(t *testing.T) {
			result := f(t, file)
			serialized, err := cbor.Marshal(result)
			if err != nil {
				t.Fatalf("Could not serialize result: %v", err)
			}

			goldenFile := filepath.Join(expectedDir, testName+".gold")
			if *updateGoldenFiles {
				if err := os.WriteFile(goldenFile, serialized, 0644); err != nil {
					t.Fatalf("Could not write golden file '%s': %v", goldenFile, err)
				}
				t.Logf("Updated golden file '%s'", goldenFile)
			} else {
				expected, err := os.ReadFile(goldenFile)
				if err != nil {
					t.Fatalf("Could not read golden file '%s': %v", goldenFile, err)
				}

				if !bytes.Equal(serialized, expected) {
					t.Errorf("Output mismatch")
				}
			}
		})
	}
}
