package cl_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KEINOS/go-countline/cl"
	"github.com/stretchr/testify/require"
)

func ExampleCountLines() {
	for _, sample := range []struct {
		Input string
	}{
		{""},            // --> 0
		{"Hello"},       // --> 1
		{"Hello\n"},     // --> 1
		{"\n"},          // --> 1
		{"\n\n"},        // --> 2
		{"\nHello"},     // --> 2
		{"\nHello\n"},   // --> 2
		{"\n\nHello"},   // --> 3
		{"\n\nHello\n"}, // --> 3
	} {
		readerFile := strings.NewReader(sample.Input)

		count, err := cl.CountLines(readerFile)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%#v --> %v\n", sample.Input, count)
	}
	// Output:
	// "" --> 0
	// "Hello" --> 1
	// "Hello\n" --> 1
	// "\n" --> 1
	// "\n\n" --> 2
	// "\nHello" --> 2
	// "\nHello\n" --> 2
	// "\n\nHello" --> 3
	// "\n\nHello\n" --> 3
}

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
