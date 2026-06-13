package alt

import (
	"bufio"
	"fmt"
	"io"
)

// CountLinesAlt3 is the 3rd attempt to count the number of lines in a file using
// bufio.Reader without goroutines.
func CountLinesAlt3(inputReader io.Reader) (int, error) {
	if inputReader == nil {
		return 0, errNilReader
	}

	bufReader := bufio.NewReader(inputReader)

	buf := make([]byte, bufio.MaxScanTokenSize)
	hasFragment := false
	count := 0

	countLF := func(b []byte) int {
		count2 := 0

		for _, c := range b {
			hasFragment = true

			if c == '\n' {
				count2++

				hasFragment = false
			}
		}

		return count2
	}

	for {
		numRead, err := bufReader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}

			return 0, fmt.Errorf("failed to read from reader: %w", err)
		}

		if numRead > 0 {
			count += countLF(buf[:numRead])
		}
	}

	if hasFragment {
		count++
	}

	return count, nil
}
