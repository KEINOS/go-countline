//nolint:revive,stylecheck
package _alt

import (
	"bufio"
	"io"

	"github.com/pkg/errors"
)

// CountLinesAlt4 uses atomic and goroutines to count the number of lines.
func CountLinesAlt4(inputReader io.Reader) (int, error) {
	if inputReader == nil {
		return 0, errors.New("given reader is nil")
	}

	bufReader := bufio.NewReader(inputReader)
	count := 0

	for {
		_, isPrefix, err := bufReader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}

			return 0, errors.Wrap(err, "failed to read from reader")
		}

		if isPrefix {
			continue
		}

		count++
	}

	return count, nil
}
