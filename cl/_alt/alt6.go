//nolint:revive,stylecheck
package _alt

import (
	"bufio"
	"bytes"
	"io"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

// ----------------------------------------------------------------------------
//  CountLinesAlt6
// ----------------------------------------------------------------------------

// CountLinesAlt6 uses bufio.Reader and goroutines to count the number of lines.
//
//nolint:funlen,cyclop // only exceeds 4 lines(74/70), complexity of 1 cycle(11/19)
func CountLinesAlt6(inputReader io.Reader) (int, error) {
	// maxInt is the maximum possitive value of int on current system in uint.
	const maxInt = ^uint(0) >> 1
	// bufSize is the maximum size of the buffer.
	const bufSize = bufio.MaxScanTokenSize

	if inputReader == nil {
		return 0, errors.New("given reader is nil")
	}

	wg := new(sync.WaitGroup) //nolint:varnamelen
	count := uint64(0)
	bufReader := bufio.NewReader(inputReader)
	lastBuf := make([]byte, bufSize)
	numIte := 0

	for {
		numIte++
		buf := make([]byte, bufSize*numIte)

		numRead, err := bufReader.Read(buf) // loading chunk into the buffer
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

			// add only if "found" is less than maxInt
			if found < int(maxInt) {
				// count++ safely
				//
				//nolint:gosec // oveflow is checked above
				atomic.AddUint64(&count, uint64(found))
			}

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

	// Check overflow on 32bit systems
	if count > uint64(maxInt) {
		return 0, errors.New("number of lines exceeds the maximum value of int")
	}

	//nolint:gosec // oveflow is checked above
	return int(count), nil
}
