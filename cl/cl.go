package cl

import (
	"bufio"
	"bytes"
	"io"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

// CountLines counts the number of lines that contains a line break (LF) in a file.
func CountLines(inputReader io.Reader) (int, error) {
	if inputReader == nil {
		return 0, errors.New("given reader is nil")
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

		numRead, err := bufReader.Read(buf) // loading chunk into buffer
		if err != nil {
			if err == io.EOF {
				break
			}

			return 0, errors.Wrap(err, "failed to read from reader")
		}

		task := buf[:numRead]
		lastBuf = task

		wg.Add(1)

		go func() {
			found := bytes.Count(task, []byte{'\n'})

			atomic.AddUint64(&count, uint64(found)) // count++ safely

			wg.Done()
		}()
	}

	wg.Wait()

	// Detect the file ends without a line break and count up if so.
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
		atomic.AddUint64(&count, 1) // count++ safely
	}

	return int(count), nil
}
