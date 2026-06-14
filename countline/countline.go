// Package countline counts LF-delimited lines from an io.Reader.
//
// Unlike `wc -l`, it also counts a final line that does not end with a line
// feed.
//
// For large random-access sources (*os.File, *bytes.Reader, *strings.Reader)
// the work is counted concurrently across goroutines; any other reader is
// counted with a serial streaming fallback.
package countline

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
)

// ErrNilReader is returned when a nil reader is passed to CountLines.
var ErrNilReader = errors.New("given reader is nil")

// ErrLineCountOverflow is returned when the line count exceeds the maximum int value.
var ErrLineCountOverflow = errors.New("number of lines exceeds the maximum value of int")

// ErrReadFailed is returned when reading from the reader fails.
var ErrReadFailed = errors.New("failed to read from reader")

// ErrSeekFailed is returned when advancing the reader to EOF after a successful
// parallel count fails. It is only reachable with a broken custom seeker; the
// supported types (*os.File, *bytes.Reader, *strings.Reader) never trigger it.
var ErrSeekFailed = errors.New("failed to seek reader")

// lineFeed is the separator counted by bytes.Count. Kept as a package-level
// value so the hot loop does not re-allocate it on every chunk.
//
//nolint:gochecknoglobals // intentional shared constant slice for the hot path.
var lineFeed = []byte{'\n'}

const (
	// parallelThreshold is the smallest input for which the concurrent path
	// pays off. Below it, the extra Stat, the trailing-byte probe and the
	// goroutine coordination cost more than they save, so counting stays
	// serial. The crossover measured around a few MiB on an Apple M4.
	//
	// It is kept >= minBytesPerWorker so countParallel always gets at least
	// one worker's worth of data.
	parallelThreshold = 4 << 20 // 4 MiB

	// minBytesPerWorker caps the number of workers so each one has enough work
	// to amortize its goroutine and buffer overhead.
	minBytesPerWorker = 1 << 20 // 1 MiB

	// readChunkSize is how much each worker reads per ReadAt call. 256 KiB is
	// a sweet spot: small enough that the copy-then-count stays hot in cache,
	// large enough to keep the pread syscall count low for *os.File.
	readChunkSize = 256 << 10 // 256 KiB
)

// CountLines counts LF-delimited lines from inputReader, reading from the
// reader's current position and consuming it to EOF.
//
// Unlike wc -l, it also counts a final line that does not end with '\n'.
//
// When the source supports random access with a known size (for example
// *os.File, *bytes.Reader or *strings.Reader) and enough remains unread, the
// work is split across goroutines and counted concurrently. Smaller inputs and
// any other io.Reader are counted with a serial stream. Both paths return the
// same count.
func CountLines(inputReader io.Reader) (int, error) {
	// maxInt is the maximum positive value of int on the current system.
	const maxInt = ^uint(0) >> 1

	return countLines(inputReader, maxInt)
}

func countLines(inputReader io.Reader, maxInt uint) (int, error) {
	if inputReader == nil {
		return 0, ErrNilReader
	}

	if reader, start, size, ok := randomAccessSource(inputReader); ok && size-start >= parallelThreshold {
		return countParallel(reader, start, size, maxInt)
	}

	return countSerial(inputReader, maxInt)
}

// randomReader is a source the parallel path can both read at arbitrary offsets
// (concurrently) and advance. Requiring io.Seeker guarantees the reader can be
// consumed to EOF afterwards, matching the serial path; any source that lacks
// Seek falls back to serial counting instead.
type randomReader interface {
	io.ReaderAt
	io.Seeker
}

// randomAccessSource reports whether reader qualifies for the parallel path and
// exposes its current position (start) and total size. The parallel path counts
// only the not-yet-read region [start,size), matching the serial path which
// reads from the reader's current position.
//
//nolint:ireturn // returns the ReaderAt+Seeker capability shared by *os.File, *bytes.Reader and *strings.Reader.
func randomAccessSource(reader io.Reader) (randomReader, int64, int64, bool) {
	if file, ok := reader.(*os.File); ok {
		info, err := file.Stat()
		if err != nil || !info.Mode().IsRegular() {
			return nil, 0, 0, false
		}

		// A successfully-Stat'd regular file always supports Seek, so the
		// current offset is retrieved without error (blank assignment).
		offset, _ := file.Seek(0, io.SeekCurrent)

		return file, offset, info.Size(), true
	}

	// *bytes.Reader and *strings.Reader satisfy this interface. Len reports the
	// unread remainder, so the already-consumed prefix is Size()-Len().
	type readerAtSizer interface {
		randomReader
		Size() int64
		Len() int
	}

	if sized, ok := reader.(readerAtSizer); ok {
		total := sized.Size()

		return sized, total - int64(sized.Len()), total, true
	}

	return nil, 0, 0, false
}

// countParallel splits the unread region [start,size) into one sub-region per
// worker and counts line feeds in each concurrently. Counting '\n' is
// associative, so the split points never change the total.
// The caller (countLines) guarantees size-start >= parallelThreshold.
func countParallel(reader randomReader, start, size int64, maxInt uint) (int, error) {
	lastByte, err := readLastByte(reader, size)
	if err != nil {
		return 0, err
	}

	span := size - start
	workers := workerCount(span)
	counts := make([]uint64, workers)
	errs := make([]error, workers)
	region := span / int64(workers)

	waitGroup := new(sync.WaitGroup)

	for idx := range workers {
		regionStart := start + int64(idx)*region
		regionEnd := regionStart + region

		if idx == workers-1 {
			regionEnd = size
		}

		waitGroup.Add(1)

		go func(idx int, regionStart, regionEnd int64) {
			defer waitGroup.Done()

			counts[idx], errs[idx] = countRegion(reader, regionStart, regionEnd)
		}(idx, regionStart, regionEnd)
	}

	waitGroup.Wait()

	count, err := sumCounts(counts, errs, lastByte, maxInt)
	if err != nil {
		return 0, err
	}

	// ReadAt does not move the reader's offset, but the serial path consumes its
	// input to EOF. Match that by advancing the reader to the end so callers see
	// the same drained reader regardless of which path ran. Seek is guaranteed
	// by the randomReader interface; a failure (only reachable with a broken
	// custom seeker) is surfaced rather than silently leaving it unconsumed.
	_, errSeek := reader.Seek(size, io.SeekStart)
	if errSeek != nil {
		return 0, fmt.Errorf("%w: %w", ErrSeekFailed, errSeek)
	}

	return count, nil
}

// workerCount returns how many goroutines to use, capped so each handles at
// least minBytesPerWorker bytes. Because countParallel only runs when the span
// is >= parallelThreshold (which is >= minBytesPerWorker), the result is always
// at least one.
func workerCount(span int64) int {
	return boundedWorkerCount(span, runtime.GOMAXPROCS(0))
}

// boundedWorkerCount caps procs so each worker handles at least
// minBytesPerWorker bytes. It is split out from workerCount so both the clamped
// and unclamped branches can be unit-tested without depending on the host CPU
// count — otherwise the clamp only runs when GOMAXPROCS exceeds the span's
// worker cap, leaving that branch uncovered on low-core CI runners.
func boundedWorkerCount(span int64, procs int) int {
	// Clamp in int64 so the final value is always small (<= procs) before the
	// conversion to int, avoiding any truncation on 32-bit platforms.
	workers := int64(procs)

	if capBySize := span / minBytesPerWorker; capBySize < workers {
		workers = capBySize
	}

	return int(workers)
}

// sumCounts adds the per-region counts, applies the final-line rule, and
// guards against exceeding maxInt.
func sumCounts(counts []uint64, errs []error, lastByte byte, maxInt uint) (int, error) {
	maxCount := uint64(maxInt)
	total := uint64(0)

	for idx, count := range counts {
		if errs[idx] != nil {
			return 0, errs[idx]
		}

		if count > maxCount-total {
			return 0, ErrLineCountOverflow
		}

		total += count
	}

	// A final line with content but no trailing line feed still counts.
	if lastByte != '\n' {
		if total >= maxCount {
			return 0, ErrLineCountOverflow
		}

		total++
	}

	return int(total), nil
}

// readLastByte returns the final byte of the input, used to decide whether the
// last line lacks a trailing line feed. The caller guarantees size >= 1 (the
// unread span is >= parallelThreshold), so the size-1 offset is always valid.
func readLastByte(readerAt io.ReaderAt, size int64) (byte, error) {
	var buf [1]byte

	_, err := readerAt.ReadAt(buf[:], size-1)
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, fmt.Errorf("%w: %w", ErrReadFailed, err)
	}

	return buf[0], nil
}

// countRegion counts line feeds in [start,end) by reading the region in
// readChunkSize pieces.
func countRegion(readerAt io.ReaderAt, start, end int64) (uint64, error) {
	buf := make([]byte, readChunkSize)
	count := uint64(0)
	offset := start

	for offset < end {
		want := int64(len(buf))
		if remaining := end - offset; remaining < want {
			want = remaining
		}

		numRead, err := readerAt.ReadAt(buf[:want], offset)
		if numRead > 0 {
			//nolint:gosec // bytes.Count never returns a negative value.
			count += uint64(bytes.Count(buf[:numRead], lineFeed))
			offset += int64(numRead)
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return 0, fmt.Errorf("%w: %w", ErrReadFailed, err)
		}
	}

	return count, nil
}

// countSerial counts line feeds by streaming the reader sequentially. It is the
// fallback for sources without random access (pipes, network connections, ...).
func countSerial(inputReader io.Reader, maxInt uint) (int, error) {
	buf := make([]byte, bufio.MaxScanTokenSize)
	count := uint(0)
	hasData := false
	lastByte := byte('\n')

	for {
		numRead, errRead := inputReader.Read(buf)
		if numRead > 0 {
			chunk := buf[:numRead]
			found := uint(bytes.Count(chunk, lineFeed))

			newCount, err := addChecked(count, found, maxInt)
			if err != nil {
				return 0, err
			}

			count = newCount
			hasData = true
			lastByte = chunk[len(chunk)-1]
		}

		if errRead != nil {
			if errors.Is(errRead, io.EOF) {
				break
			}

			return 0, fmt.Errorf("%w: %w", ErrReadFailed, errRead)
		}
	}

	// A final line with content but no trailing line feed still counts.
	if hasData && lastByte != '\n' {
		if count == maxInt {
			return 0, ErrLineCountOverflow
		}

		count++
	}

	return int(count), nil
}

// addChecked returns count+found, or ErrLineCountOverflow if the sum would
// exceed maxInt.
func addChecked(count, found, maxInt uint) (uint, error) {
	if found > maxInt || count > maxInt-found {
		return 0, ErrLineCountOverflow
	}

	return count + found, nil
}
