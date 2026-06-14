package countline

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
//  Serial fallback path (generic io.Reader without random access)
// ----------------------------------------------------------------------------

// plainReader embeds only the io.Reader interface, so any ReaderAt/Size methods
// of the wrapped value are not promoted and the input is forced down the serial
// counting path.
type plainReader struct {
	io.Reader
}

func TestCountLines_serial_path(t *testing.T) {
	t.Parallel()

	reader := &plainReader{Reader: strings.NewReader("a\nb\nc")}

	numLines, err := CountLines(reader)

	require.NoError(t, err)
	require.Equal(t, 3, numLines, "serial path should count a final line without a line feed")
}

func TestCountLines_serial_overflow(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		input string
	}{
		{name: "newline_count_overflows", input: "\n"},
		{name: "final_fragment_overflows", input: "line without newline"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			reader := &plainReader{Reader: strings.NewReader(test.input)}

			numLines, err := countLines(reader, 0)

			require.Error(t, err)
			require.Equal(t, 0, numLines)
			require.Contains(t, err.Error(), "maximum value of int")
		})
	}
}

// readerAtLen has random access and a known size but, crucially, no Seek.
type readerAtLen interface {
	io.Reader
	io.ReaderAt
	Size() int64
	Len() int
}

// noSeekReader is a random-access reader that is NOT an io.Seeker, so it must
// not take the parallel path (which relies on Seek to consume the reader).
type noSeekReader struct {
	readerAtLen
}

// TestCountLines_non_seekable_random_reader verifies that a ReaderAt+Size+Len
// source WITHOUT Seek is counted by the serial path and is consumed to EOF,
// upholding the documented "consumes to EOF" contract for every input.
func TestCountLines_non_seekable_random_reader(t *testing.T) {
	t.Parallel()

	const lines = 800_000 // ~4.6 MiB: would be parallel-eligible if it had Seek.

	reader := &noSeekReader{readerAtLen: bytes.NewReader(bytes.Repeat([]byte("hello\n"), lines))}

	numLines, err := CountLines(reader)

	require.NoError(t, err)
	require.Equal(t, lines, numLines)
	require.Zero(t, reader.Len(), "non-seekable reader must be counted serially and consumed")
}

// ----------------------------------------------------------------------------
//  Parallel path (*os.File) - end-to-end with a real file above the threshold
// ----------------------------------------------------------------------------

// writeTempFile writes content to a fresh temp file and returns an open handle
// positioned at the start.
func writeTempFile(t *testing.T, content []byte) *os.File {
	t.Helper()

	path := filepath.Join(t.TempDir(), "sample.txt")
	require.NoError(t, os.WriteFile(path, content, 0o600))

	file, err := os.Open(path) //nolint:gosec // path is built from t.TempDir().
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, file.Close()) })

	return file
}

func TestCountLines_os_file_parallel(t *testing.T) {
	t.Parallel()

	const lines = 800_000 // "hello\n" * 800_000 = ~4.6 MiB, above parallelThreshold.

	file := writeTempFile(t, bytes.Repeat([]byte("hello\n"), lines))

	numLines, err := CountLines(file)

	require.NoError(t, err)
	require.Equal(t, lines, numLines)
}

func TestCountLines_os_file_parallel_final_fragment(t *testing.T) {
	t.Parallel()

	const lines = 800_000 // above parallelThreshold once combined with the tail.

	content := append(bytes.Repeat([]byte("hello\n"), lines), []byte("tail")...)

	file := writeTempFile(t, content)

	numLines, err := CountLines(file)

	require.NoError(t, err)
	require.Equal(t, lines+1, numLines, "the trailing fragment without a line feed must count")
}

func TestCountLines_os_file_parallel_overflow(t *testing.T) {
	t.Parallel()

	file := writeTempFile(t, bytes.Repeat([]byte("hello\n"), 800_000))

	numLines, err := countLines(file, 0)

	require.Error(t, err)
	require.Equal(t, 0, numLines)
	require.Contains(t, err.Error(), "maximum value of int")
}

// TestCountLines_bytes_reader_parallel exercises the parallel path via the
// io.ReaderAt+Size() branch (not *os.File) using a real *bytes.Reader, whose
// ReadAt faithfully honors the requested offset and length. The lines do not
// align to region boundaries ("hello\n" is 6 bytes), so this also guards the
// region-splitting arithmetic in countRegion.
func TestCountLines_bytes_reader_parallel(t *testing.T) {
	t.Parallel()

	const lines = 900_000 // ~5.1 MiB, above parallelThreshold.

	reader := bytes.NewReader(append(bytes.Repeat([]byte("hello\n"), lines), []byte("tail")...))

	numLines, err := CountLines(reader)

	require.NoError(t, err)
	require.Equal(t, lines+1, numLines, "trailing fragment via *bytes.Reader must count")
}

// TestCountLines_bytes_reader_partial verifies the parallel path counts only the
// UNREAD remainder of a partially-consumed reader, matching serial semantics
// (counting from the reader's current position, not absolute byte 0).
func TestCountLines_bytes_reader_partial(t *testing.T) {
	t.Parallel()

	const (
		total         = 900_000
		consumedLines = 100_000
	)

	reader := bytes.NewReader(bytes.Repeat([]byte("hello\n"), total))

	consumed, err := io.ReadFull(reader, make([]byte, consumedLines*len("hello\n")))
	require.NoError(t, err)
	require.Equal(t, consumedLines*len("hello\n"), consumed)

	numLines, err := CountLines(reader)

	require.NoError(t, err)
	require.Equal(t, total-consumedLines, numLines, "must count only the unread remainder")
}

func TestCountLines_os_file_partial(t *testing.T) {
	t.Parallel()

	const (
		total         = 900_000
		consumedLines = 100_000
	)

	file := writeTempFile(t, bytes.Repeat([]byte("hello\n"), total))

	consumed, err := io.ReadFull(file, make([]byte, consumedLines*len("hello\n")))
	require.NoError(t, err)
	require.Equal(t, consumedLines*len("hello\n"), consumed)

	numLines, err := CountLines(file)

	require.NoError(t, err)
	require.Equal(t, total-consumedLines, numLines, "must count only the unread remainder")
}

// TestCountLines_consumes_reader verifies the parallel path leaves the reader
// drained to EOF, matching the serial path (and the original implementation)
// rather than leaving the ReadAt-based reader at its original offset.
func TestCountLines_consumes_reader(t *testing.T) {
	t.Parallel()

	content := bytes.Repeat([]byte("hello\n"), 800_000) // ~4.6 MiB -> parallel.

	t.Run("bytes.Reader", func(t *testing.T) {
		t.Parallel()

		reader := bytes.NewReader(content)

		_, err := CountLines(reader)
		require.NoError(t, err)
		require.Zero(t, reader.Len(), "reader must be consumed to EOF after counting")
	})

	t.Run("os.File", func(t *testing.T) {
		t.Parallel()

		file := writeTempFile(t, content)

		_, err := CountLines(file)
		require.NoError(t, err)

		rest, err := io.ReadAll(file)
		require.NoError(t, err)
		require.Empty(t, rest, "file offset must be at EOF after counting")
	})
}

func TestCountLines_os_file_not_regular(t *testing.T) {
	t.Parallel()

	// os.DevNull is a character device, not a regular file, so it must fall
	// back to the serial path and report zero lines.
	file, err := os.Open(os.DevNull)
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, file.Close()) })

	numLines, err := CountLines(file)

	require.NoError(t, err)
	require.Equal(t, 0, numLines)
}

func TestCountLines_os_file_stat_error(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "closed.txt")
	require.NoError(t, os.WriteFile(path, []byte("x\n"), 0o600))

	file, err := os.Open(path) //nolint:gosec // path is built from t.TempDir().
	require.NoError(t, err)
	require.NoError(t, file.Close()) // Stat now fails -> serial fallback.

	numLines, err := CountLines(file)

	require.Error(t, err)
	require.Equal(t, 0, numLines)
}

// ----------------------------------------------------------------------------
//  Parallel path read failures (synthetic ReaderAt sources)
// ----------------------------------------------------------------------------

// failingReaderAt satisfies the random-access interface. It succeeds for the
// single trailing-byte probe but fails for the region reads, exercising the
// worker error path.
type failingReaderAt struct {
	size int64
}

func (f *failingReaderAt) Size() int64                    { return f.size }
func (f *failingReaderAt) Len() int                       { return int(f.size) }
func (f *failingReaderAt) Seek(int64, int) (int64, error) { return 0, nil }
func (f *failingReaderAt) Read([]byte) (int, error)       { panic("unused: parallel path uses ReadAt") }

func (f *failingReaderAt) ReadAt(buf []byte, offset int64) (int, error) {
	if offset == f.size-1 {
		buf[0] = 'x'

		return 1, nil
	}

	return 0, errForcedRead
}

func TestCountLines_parallel_region_read_fail(t *testing.T) {
	t.Parallel()

	numLines, err := CountLines(&failingReaderAt{size: parallelThreshold})

	require.Error(t, err)
	require.Equal(t, 0, numLines)
	require.Contains(t, err.Error(), "failed to read from reader")
}

// alwaysFailReaderAt fails on every ReadAt, including the trailing-byte probe.
type alwaysFailReaderAt struct {
	size int64
}

func (a *alwaysFailReaderAt) Size() int64                       { return a.size }
func (a *alwaysFailReaderAt) Len() int                          { return int(a.size) }
func (a *alwaysFailReaderAt) Seek(int64, int) (int64, error)    { return 0, nil }
func (a *alwaysFailReaderAt) Read([]byte) (int, error)          { panic("unused: parallel path uses ReadAt") }
func (a *alwaysFailReaderAt) ReadAt([]byte, int64) (int, error) { return 0, errForcedRead }

func TestCountLines_parallel_last_byte_read_fail(t *testing.T) {
	t.Parallel()

	numLines, err := CountLines(&alwaysFailReaderAt{size: parallelThreshold})

	require.Error(t, err)
	require.Equal(t, 0, numLines)
	require.Contains(t, err.Error(), "failed to read from reader")
}

// eofReaderAt returns data together with io.EOF in a single region read, which
// is how a worker reaching the true end of input terminates its loop.
type eofReaderAt struct {
	size int64
}

func (e *eofReaderAt) Size() int64                    { return e.size }
func (e *eofReaderAt) Len() int                       { return int(e.size) }
func (e *eofReaderAt) Seek(int64, int) (int64, error) { return 0, nil }
func (e *eofReaderAt) Read([]byte) (int, error)       { panic("unused: parallel path uses ReadAt") }

func (e *eofReaderAt) ReadAt(buf []byte, offset int64) (int, error) {
	if offset == e.size-1 {
		buf[0] = '\n'

		return 1, nil
	}

	return copy(buf, []byte("a\n")), io.EOF
}

func TestCountLines_parallel_region_eof(t *testing.T) {
	t.Parallel()

	numLines, err := CountLines(&eofReaderAt{size: parallelThreshold})

	require.NoError(t, err)
	// Each worker reads "a\n" then stops at EOF, so the total is one per worker.
	require.Equal(t, workerCount(parallelThreshold), numLines)
}

// constReaderAt returns a fixed byte for every read, used to drive the two
// overflow branches of sumCounts on the parallel path.
type constReaderAt struct {
	size int64
	fill byte
}

func (c *constReaderAt) Size() int64                    { return c.size }
func (c *constReaderAt) Len() int                       { return int(c.size) }
func (c *constReaderAt) Seek(int64, int) (int64, error) { return 0, nil }
func (c *constReaderAt) Read([]byte) (int, error)       { panic("unused: parallel path uses ReadAt") }

func (c *constReaderAt) ReadAt(buf []byte, _ int64) (int, error) {
	for idx := range buf {
		buf[idx] = c.fill
	}

	return len(buf), nil
}

// seekFailReaderAt counts successfully but fails to Seek, exercising the
// post-count consume step's error path.
type seekFailReaderAt struct {
	size int64
}

func (s *seekFailReaderAt) Size() int64                    { return s.size }
func (s *seekFailReaderAt) Len() int                       { return int(s.size) }
func (s *seekFailReaderAt) Seek(int64, int) (int64, error) { return 0, errForcedRead }
func (s *seekFailReaderAt) Read([]byte) (int, error)       { panic("unused: parallel path uses ReadAt") }

func (s *seekFailReaderAt) ReadAt(buf []byte, _ int64) (int, error) {
	for idx := range buf {
		buf[idx] = '\n'
	}

	return len(buf), nil
}

func TestCountLines_parallel_consume_seek_fail(t *testing.T) {
	t.Parallel()

	numLines, err := CountLines(&seekFailReaderAt{size: parallelThreshold})

	require.Error(t, err)
	require.Equal(t, 0, numLines)
	require.ErrorIs(t, err, ErrSeekFailed)
}

func TestCountLines_parallel_overflow_final_fragment(t *testing.T) {
	t.Parallel()

	// All 'a': zero line feeds, and a final fragment with no trailing '\n'
	// that cannot be represented when maxInt is 0.
	numLines, err := countLines(&constReaderAt{size: parallelThreshold, fill: 'a'}, 0)

	require.Error(t, err)
	require.Equal(t, 0, numLines)
	require.Contains(t, err.Error(), "maximum value of int")
}
