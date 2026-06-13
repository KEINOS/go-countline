// ============================================================================
//
//	Alternate implementations of CountLines function
//
// ============================================================================
//
//	This file contains the alternate implementations of CountLines().
//	We benchmark them to see which one is the fastest.
//
//	Note that all implementations MUST pass the test for specifications.
//	See the "Spec Tests" section below.
package alt

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/KEINOS/go-countline/countline/spec"
	"github.com/stretchr/testify/require"
)

// ============================================================================
//  Tests
// ============================================================================
//  Specification tests for alternate implementations of CountLines().

//nolint:paralleltest
func TestCountLines_specs(t *testing.T) {
	for _, targetFunc := range []struct {
		name string
		fn   func(io.Reader) (int, error)
	}{
		// Add the alternate implementations here.
		{"CountLinesAlt1", CountLinesAlt1},
		{"CountLinesAlt2", CountLinesAlt2},
		{"CountLinesAlt3", CountLinesAlt3},
		{"CountLinesAlt4", CountLinesAlt4},
		{"CountLinesAlt5", CountLinesAlt5},
		{"CountLinesAlt6", CountLinesAlt6},
	} {
		t.Run(targetFunc.name, func(t *testing.T) {
			spec.RunSpecTest(t, targetFunc.name, targetFunc.fn)
		})

		t.Run(targetFunc.name+"_nil_input", func(t *testing.T) {
			numLines, err := targetFunc.fn(nil)

			require.Error(t, err, "should return an error on nil input")
			require.Equal(t, 0, numLines, "returned number of lines should be 0 on error")
		})

		t.Run(targetFunc.name+"_io_read_fail", func(t *testing.T) {
			dummyReader := &DummyReader{}

			numLines, err := targetFunc.fn(dummyReader)

			require.Error(t, err, "it should return an error on io.Reader read failure")
			require.Equal(t, 0, numLines, "returned number of lines should be 0 on error")
			require.Contains(t, err.Error(), "forced error", "the returned error should contain the reason of the error")
		})

		t.Run(targetFunc.name+"_zero_padded", func(t *testing.T) {
			// Create a dummy reader with zero-padded/capped bytes
			dummyReader := bytes.NewReader(make([]byte, 1024))

			_, err := targetFunc.fn(dummyReader)

			require.NoError(t, err, "it should not return an error on zero padded/empty capped byte slice input")
		})
	}
}

func TestCountLines_overflow(t *testing.T) {
	t.Parallel()

	for _, targetFunc := range []struct {
		name string
		fn   func(io.Reader, uint) (int, error)
	}{
		{"CountLinesAlt2", countLinesAlt2},
		{"CountLinesAlt6", countLinesAlt6},
	} {
		t.Run(targetFunc.name, func(t *testing.T) {
			t.Parallel()

			numLines, err := targetFunc.fn(strings.NewReader("line without newline"), 0)

			require.Error(t, err, "should return an error on line count overflow")
			require.Equal(t, 0, numLines, "returned number of lines should be 0 on error")
			require.Contains(t, err.Error(), "maximum value of int")
		})
	}
}

// DummyReader is a dummy io.Reader that always returns errForcedRead on Read().
type DummyReader struct{}

// Read implements io.Reader interface. Always returns errForcedRead.
func (r *DummyReader) Read(_ []byte) (int, error) {
	return 0, errForcedRead
}
