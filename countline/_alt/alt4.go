package alt

import (
	"bufio"
	"fmt"
	"io"
)

// CountLinesAlt4 uses atomic and goroutines to count the number of lines.
func CountLinesAlt4(inputReader io.Reader) (int, error) {
	if inputReader == nil {
		return 0, errNilReader
	}

	bufReader := bufio.NewReader(inputReader)
	count := 0

	for {
		_, isPrefix, err := bufReader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}

			return 0, fmt.Errorf("failed to read from reader: %w", err)
		}

		if isPrefix {
			continue
		}

		count++
	}

	return count, nil
}
