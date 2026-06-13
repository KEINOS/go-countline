//go:generate go run ./_gen

// To keep the repository size small, most of the test data is left uncommitted.
// Generate it before benchmarking, from the repository root: `go generate ./...`.

package countline

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	alt "github.com/KEINOS/go-countline/countline/_alt"
	"github.com/stretchr/testify/require"
)

// targetFunctions is a map of functions to be tested.
// We are using a map to avoid the order of the tests.
//
//nolint:gochecknoglobals
var targetFunctions = map[string]struct {
	fn func(io.Reader) (int, error)
}{
	// Current implementation
	"CountLinesCurr": {CountLines},
	// Alternate implementation. See alt_test.go.
	"CountLinesAlt1": {alt.CountLinesAlt1},
	"CountLinesAlt2": {alt.CountLinesAlt2},
	"CountLinesAlt3": {alt.CountLinesAlt3},
	"CountLinesAlt4": {alt.CountLinesAlt4},
	"CountLinesAlt5": {alt.CountLinesAlt5},
	"CountLinesAlt6": {alt.CountLinesAlt6},
}

const (
	MediumSize = "medium"
	LargeSize  = "large"
	GiantSize  = "giant"
)

// targetDatas is a list of files under `countline/testdata/` directory to be
// tested. We are using a map to avoid the order of the tests.
//
// These files are created via the `go generate ./...` command. See
// `countline/_gen/gen_test_data.go`.
//
//nolint:gochecknoglobals
var targetDatas = map[string]struct {
	nameFile string
	typeSize string
	sizeFile int // in bytes
	numLine  int
}{
	"Tiny":   {nameFile: "data_Tiny.txt", typeSize: MediumSize, sizeFile: 1032, numLine: 114},
	"Small":  {nameFile: "data_Small.txt", typeSize: MediumSize, sizeFile: 1048578, numLine: 88307},
	"Medium": {nameFile: "data_Medium.txt", typeSize: MediumSize, sizeFile: 10485767, numLine: 815144},
	"Large":  {nameFile: "data_Large.txt", typeSize: LargeSize, sizeFile: 52428802, numLine: 3824279},
	"Huge":   {nameFile: "data_Huge.txt", typeSize: LargeSize, sizeFile: 104857612, numLine: 7569194},
	"Giant":  {nameFile: "data_Giant.txt", typeSize: GiantSize, sizeFile: 1073741832, numLine: 72323529},
}

// ============================================================================
//  Benchmarks
// ============================================================================

// Benchmark of light weight size files (Tiny, Small, Medium).
func Benchmark_light(b *testing.B) {
	benchEachImplementation(b, MediumSize)
}

// Benchmark of heavy weight size files (Large, Huge).
func Benchmark_heavy(b *testing.B) {
	benchEachImplementation(b, LargeSize)
}

// Benchmark of the 1 GiB file.
//
// Note: This benchmark is heavy and takes a long time (a minute on average per
// implementation) to run.
func Benchmark_giant(b *testing.B) {
	benchEachImplementation(b, GiantSize)
}

// benchEachImplementation benchmarks every registered implementation against
// each test data file of the given size type.
func benchEachImplementation(b *testing.B, typeSize string) {
	b.Helper()

	for _, data := range targetDatas {
		if data.typeSize != typeSize {
			continue
		}

		nameData := strings.TrimSuffix(strings.TrimPrefix(data.nameFile, "data_"), ".txt")
		pathFile := filepath.Join("testdata", data.nameFile)

		// since targetFunctions is a map, the order of the tests is random.
		for nameFunc, targetFunc := range targetFunctions {
			nameTest := fmt.Sprintf("size-%s_%s_%s", readableSize(data.sizeFile), nameData, nameFunc)

			b.Run(nameTest, func(b *testing.B) {
				runBench(b, data.numLine, int64(data.sizeFile), pathFile, targetFunc.fn)
			})
		}
	}
}

// ----------------------------------------------------------------------------
//  Real-world file I/O performance benchmark
// ----------------------------------------------------------------------------

// Benchmark including file I/O (real-world scenario).
func BenchmarkCountLines_IO(b *testing.B) {
	for sizeType, data := range targetDatas {
		benchFile := filepath.Clean(filepath.Join("testdata", data.nameFile))

		title := fmt.Sprintf("%s_%s", sizeType, readableSize(data.sizeFile))

		b.Run(title, func(b *testing.B) {
			fi, err := os.Stat(benchFile)
			if err != nil {
				b.Fatalf("stat failed: %v", err)
			}

			b.SetBytes(fi.Size())

			b.ResetTimer()

			for b.Loop() {
				filePtr, err := os.Open(benchFile)
				if err != nil {
					b.Fatalf("open failed: %v", err)
				}

				_, err = CountLines(filePtr)
				if err != nil {
					b.Fatalf("count failed: %v", err)
				}

				err = filePtr.Close()
				if err != nil {
					b.Fatalf("close failed: %v", err)
				}
			}
		})
	}
}

// Benchmark algorithm only (memory-resident data).
func BenchmarkCountLines_Memory(b *testing.B) {
	benchFile := filepath.Clean(filepath.Join("testdata", "data_Giant.txt"))

	data, err := os.ReadFile(benchFile)
	if err != nil {
		b.Fatalf("read failed: %v", err)
	}

	size := int64(len(data))
	b.SetBytes(size)

	b.ResetTimer()

	for b.Loop() {
		r := bytes.NewReader(data)

		_, err := CountLines(r)
		if err != nil {
			b.Fatalf("count failed: %v", err)
		}
	}
}

// ----------------------------------------------------------------------------
//  Helper functions
// ----------------------------------------------------------------------------

// runBench measures fn over a real file (open + count + close) and reports
// throughput in bytes per second. It uses b.Loop so the benchmark converges on
// -benchtime instead of running forever.
//
//nolint:varnamelen // "fn" is short for the scope of its usage but allow it
func runBench(b *testing.B, expectNumLines int, sizeFile int64, pathFile string, fn func(io.Reader) (int, error)) {
	b.Helper()

	cleanPath := filepath.Clean(pathFile)

	b.SetBytes(sizeFile)
	b.ReportAllocs()

	for b.Loop() {
		fileReader, err := os.Open(cleanPath)
		if err != nil {
			b.Fatal(err)
		}

		actual, err := fn(fileReader)
		if err != nil {
			b.Fatal(err)
		}

		require.NoError(b, fileReader.Close())

		if actual != expectNumLines {
			b.Fatalf("test %v failed: expect=%d, actual=%d", b.Name(), expectNumLines, actual)
		}
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
