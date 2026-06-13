package alt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

// CountLinesAlt2 uses bufio.Reader and goroutines to count the number of lines.
func CountLinesAlt2(inputReader io.Reader) (int, error) {
	const maxInt = ^uint(0) >> 1

	return countLinesAlt2(inputReader, maxInt)
}

//nolint:cyclop,funlen // complexity and length are inherited from the original benchmark implementation.
func countLinesAlt2(inputReader io.Reader, maxInt uint) (int, error) {
	if inputReader == nil {
		return 0, errNilReader
	}

	wg := new(sync.WaitGroup) //nolint:varnamelen
	bufSize := bufio.MaxScanTokenSize
	count := uint64(0)
	bufReader := bufio.NewReader(inputReader)
	lastBuf := make([]byte, bufSize)
	numIte := 0

	for {
		numIte++
		buf := make([]byte, bufSize*numIte)

		numRead, err := bufReader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}

			return 0, fmt.Errorf("failed to read from reader: %w", err)
		}

		task := buf[:numRead]
		lastBuf = task

		wg.Go(func() {
			found := bytes.Count(task, []byte{'\n'})

			if found < int(maxInt) {
				//nolint:gosec // overflow is checked above
				atomic.AddUint64(&count, uint64(found))
			}
		})
	}

	wg.Wait()

	lenLastBuf := len(lastBuf)
	hasFragment := false

	for i := lenLastBuf; i > 0; i-- {
		tmpChar := lastBuf[i-1]
		if tmpChar == '\x00' {
			continue
		}

		if tmpChar == '\n' {
			break
		}

		hasFragment = true
	}

	if hasFragment {
		atomic.AddUint64(&count, 1)
	}

	if count > uint64(maxInt) {
		return 0, errLineCountOverflow
	}

	return int(count), nil
}
