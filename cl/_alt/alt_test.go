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
	"testing"

	"github.com/KEINOS/go-countline/cl/spec"
	"github.com/pkg/errors"
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
			dummyReader := &DummyReader{msg: "forced error"}

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

// DummyReader is a dummy io.Reader that returns an error on Read().
type DummyReader struct {
	msg string
}

// Read implements io.Reader interface. This method always returns an error with the msg field.
func (r *DummyReader) Read(_ []byte) (int, error) {
	return 0, errors.New(r.msg)
}
