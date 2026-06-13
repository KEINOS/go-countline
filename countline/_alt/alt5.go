package alt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

const bufSizeDefault = 1024

// CountLinesAlt5 uses bufio.Scanner to count the number of lines.
func CountLinesAlt5(inputReader io.Reader) (int, error) {
	return countLinesAlt5(inputReader, bufSizeDefault)
}

func countLinesAlt5(inputReader io.Reader, bufSize int) (int, error) {
	if inputReader == nil {
		return 0, errNilReader
	}

	bufScanner := bufio.NewScanner(inputReader)

	buf := make([]byte, bufSize)
	bufScanner.Buffer(buf, bufSize)

	countLine := 0

	for bufScanner.Scan() {
		countLine++
	}

	err := bufScanner.Err()
	if err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return countLinesAlt5(inputReader, bufSize*bufSizeDefault)
		}

		return 0, fmt.Errorf("failed to scan reader: %w", err)
	}

	return countLine, nil
}
