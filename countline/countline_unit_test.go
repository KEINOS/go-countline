package countline

import (
	"errors"
	"strings"
	"testing"

	"github.com/KEINOS/go-countline/countline/spec"
	"github.com/stretchr/testify/require"
)

var errForcedRead = errors.New("forced error")

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

func (r *DummyReader) Read(_ []byte) (int, error) {
	return 0, errForcedRead
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

func TestCountLines_trailing_nul_is_content(t *testing.T) {
	t.Parallel()

	numLines, err := CountLines(strings.NewReader("first\n\x00"))

	require.NoError(t, err)
	require.Equal(t, 2, numLines, "a trailing NUL byte is file content, not buffer padding")
}

func Test_countLines_overflow(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		input string
	}{
		{
			name:  "newline_count_overflows",
			input: "\n",
		},
		{
			name:  "final_fragment_overflows",
			input: "line without newline",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			numLines, err := countLines(strings.NewReader(test.input), 0)

			require.Error(t, err)
			require.Equal(t, 0, numLines, "returned number of lines should be 0 on error")
			require.Contains(t, err.Error(), "maximum value of int")
		})
	}
}
