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
//  CountLinesAlt2
// ----------------------------------------------------------------------------

// CountLinesAlt2 uses bufio.Reader and goroutines to count the number of lines.
func CountLinesAlt2(inputReader io.Reader) (int, error) {
	if inputReader == nil {
		return 0, errors.New("given reader is nil")
	}

	const double = 2

	bufSizeBase := 1024
	count := uint64(0)
	wg := new(sync.WaitGroup) //nolint:varnamelen
	bufReader := bufio.NewReader(inputReader)
	lastBuf := []byte{}
	numIte := 0

	for {
		numIte++
		buf := make([]byte, bufSizeBase*(numIte*double))

		numRead, err := bufReader.Read(buf)
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
