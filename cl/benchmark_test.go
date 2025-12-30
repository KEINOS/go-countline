package cl_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KEINOS/go-countline/cl"
	"github.com/stretchr/testify/require"
)

func BenchmarkCountLines(b *testing.B) {
	// 1 GiB size file
	pathFile := filepath.Clean(filepath.Join("testdata", "data_Giant.txt"))

	expectNumLines := 72323529

	// Open file
	fileReader, err := os.Open(pathFile)
	if err != nil {
		b.Fatal(err)
	}

	b.Cleanup(func() {
		require.NoError(b, fileReader.Close())
	})

	b.ResetTimer() // Begin benchmark

	// Run function
	actualNumLines, err := cl.CountLines(fileReader)
	if err != nil {
		b.Fatal(err)
	}

	b.StopTimer() // End benchmark

	if expectNumLines != actualNumLines {
		b.Fatalf(
			"test %v failed: expect=%d, actual=%d",
			b.Name(), expectNumLines, actualNumLines,
		)
	}
}
