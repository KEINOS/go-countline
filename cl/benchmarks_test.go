// To keep the repository size small, most of the test data is left uncommitted.
// You must generate it yourself before testing and benchmarking.
//
// ```shellsession
// $ # From the root of the repository
// $ go generate ./...
// ...
// ```
//
//go:generate go run ./_gen
package cl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KEINOS/go-countline/cl/_alt"
)

// targetFuncions is a map of functions to be tested.
// We are using a map to avoid the order of the tests.
//
//nolint:gochecknoglobals
var targetFuncions = map[string]struct {
	fn func(io.Reader) (int, error)
}{
	// Current implementation
	"CountLinesCurr": {CountLines},
	// Alternate implementation. See alt_test.go.
	"CountLinesAlt1": {_alt.CountLinesAlt1},
	"CountLinesAlt2": {_alt.CountLinesAlt2},
	"CountLinesAlt3": {_alt.CountLinesAlt3},
	"CountLinesAlt4": {_alt.CountLinesAlt4},
}

// targetDatas is a list of files under `cl/testdata/` directory to be tested.
// We are using a map to avoid the order of the tests.
//
// These files are created via `go generate ...` command. See `cl/_gen/gen.go`.
//
//nolint:gochecknoglobals
var targetDatas = map[string]struct {
	nameFile string
	typeSize string
	sizeFile int
	numLine  int
}{
	"Tiny":   {nameFile: "data_Tiny.txt", typeSize: "medium", sizeFile: 1032, numLine: 114},
	"Small":  {nameFile: "data_Small.txt", typeSize: "medium", sizeFile: 1048578, numLine: 88307},
	"Medium": {nameFile: "data_Medium.txt", typeSize: "medium", sizeFile: 10485767, numLine: 815144},
	"Large":  {nameFile: "data_Large.txt", typeSize: "large", sizeFile: 52428802, numLine: 3824279},
	"Huge":   {nameFile: "data_Huge.txt", typeSize: "large", sizeFile: 104857612, numLine: 7569194},
	"Giant":  {nameFile: "data_Giant.txt", typeSize: "giant", sizeFile: 1073741832, numLine: 72323529},
}

// ============================================================================
//  Benchmarks
// ============================================================================

// Benchmark of 1GiB file.
func Benchmark_giant(b *testing.B) {
	nameFile := "data_Giant.txt"
	sizeFile := 1073741832
	expectNumLine := 72323529

	for nameFunc, targetFunc := range targetFuncions {
		pathFile := filepath.Join("testdata", nameFile)
		nameTest := fmt.Sprintf("size-%s_%s_%s", readableSize(sizeFile), "Gigantic", nameFunc)

		for i := 0; i < b.N; i++ {
			b.Run(nameTest, func(b *testing.B) {
				runBench(b, expectNumLine, pathFile, targetFunc.fn)
			})
		}
	}
}

// Benchmark of light weight size files (Tiny, Small, Medium).
func Benchmark_light(b *testing.B) {
	for _, data := range targetDatas {
		nameData := strings.TrimSuffix(strings.TrimPrefix(data.nameFile, "data_"), ".txt")

		for nameFunc, targetFunc := range targetFuncions {
			if data.typeSize != "medium" {
				continue
			}

			pathFile := filepath.Join("testdata", data.nameFile)
			nameTest := fmt.Sprintf("size-%s_%s_%s", readableSize(data.sizeFile), nameData, nameFunc)

			for i := 0; i < b.N; i++ {
				b.Run(nameTest, func(b *testing.B) {
					expectNumLine := data.numLine
					runBench(b, expectNumLine, pathFile, targetFunc.fn)
				})
			}
		}
	}
}

// Benchmark of heavy weight size files (Large, Huge, Giant).
func Benchmark_heavy(b *testing.B) {
	for _, data := range targetDatas {
		nameData := strings.TrimSuffix(strings.TrimPrefix(data.nameFile, "data_"), ".txt")

		for nameFunc, targetFunc := range targetFuncions {
			if data.typeSize != "large" {
				continue
			}

			pathFile := filepath.Join("testdata", data.nameFile)
			nameTest := fmt.Sprintf("size-%s_%s_%s", readableSize(data.sizeFile), nameData, nameFunc)

			for i := 0; i < b.N; i++ {
				b.Run(nameTest, func(b *testing.B) {
					expectNumLine := data.numLine
					runBench(b, expectNumLine, pathFile, targetFunc.fn)
				})
			}
		}
	}
}

// ----------------------------------------------------------------------------
//  Helper functions
// ----------------------------------------------------------------------------

//nolint:varnamelen // "fn" is short for the scope of its usage but allow it
func runBench(b *testing.B, expectNumLines int, pathFile string, fn func(io.Reader) (int, error)) {
	b.Helper()

	fileReader, err := os.Open(pathFile)
	if err != nil {
		b.Fatal(err)
	}
	defer fileReader.Close()

	b.ResetTimer() // Begin benchmark

	countLines, err := fn(fileReader)
	if err != nil {
		b.Fatal(err)
	}

	b.StopTimer() // End benchmark

	expectLineCount := expectNumLines
	actualLineCount := countLines

	if expectLineCount != actualLineCount {
		b.Fatalf(
			"test %v failed: expect=%d, actual=%d",
			b.Name(), expectLineCount, actualLineCount,
		)
	}
}

func readableSize(value int) string {
	switch {
	case value >= 1024*1024*1024:
		return fmt.Sprintf("%dGiB", value/(1024*1024*1024))
	case value >= 1024*1024:
		return fmt.Sprintf("%dMiB", value/(1024*1024))
	case value >= 1024:
		return fmt.Sprintf("%dKiB", value/1024)
	default:
		return fmt.Sprintf("%dByte", value)
	}
}
