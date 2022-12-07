package cl

import (
	"testing"

	"github.com/KEINOS/go-countline/cl/spec"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

// ============================================================================
//  Tests
// ============================================================================

func TestCountLines_golden(t *testing.T) {
	t.Parallel()

	spec.RunSpecTest(t, "CountLines", CountLines)
}

func TestCountLines_nil_input(t *testing.T) {
	t.Parallel()

	numLines, err := CountLines(nil)

	require.Error(t, err)
	require.Equal(t, 0, numLines, "returned number of lines should be 0 on error")
	require.Contains(t, err.Error(), "given reader is nil")
}

type DummyReader struct{}

func (r *DummyReader) Read(p []byte) (int, error) {
	return 0, errors.New("forced error")
}

func TestCountLines_io_read_fail(t *testing.T) {
	t.Parallel()

	dummyReader := &DummyReader{}

	numLines, err := CountLines(dummyReader)

	require.Error(t, err)
	require.Equal(t, 0, numLines, "returned number of lines should be 0 on error")
	require.Contains(t, err.Error(), "failed to read from reader")
	require.Contains(t, err.Error(), "forced error", "the error should contain the reason of the error")
}
