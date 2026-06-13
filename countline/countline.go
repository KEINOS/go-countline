package countline

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

// ErrNilReader is returned when a nil reader is passed to CountLines.
var ErrNilReader = errors.New("given reader is nil")

// ErrLineCountOverflow is returned when the line count exceeds the maximum int value.
var ErrLineCountOverflow = errors.New("number of lines exceeds the maximum value of int")

// ErrReadFailed is returned when reading from the reader fails.
var ErrReadFailed = errors.New("failed to read from reader")

// CountLines counts LF-delimited lines from inputReader.
//
// Unlike wc -l, it also counts a final line that does not end with '\n'.
func CountLines(inputReader io.Reader) (int, error) {
	// maxInt is the maximum positive value of int on the current system.
	const maxInt = ^uint(0) >> 1

	return countLines(inputReader, maxInt)
}

func countLines(inputReader io.Reader, maxInt uint) (int, error) {
	if inputReader == nil {
		return 0, ErrNilReader
	}

	buf := make([]byte, bufio.MaxScanTokenSize)
	count := uint(0)
	hasData := false
	lastByte := byte('\n')

	for {
		numRead, err := inputReader.Read(buf)
		if numRead > 0 {
			chunk := buf[:numRead]

			found := uint(bytes.Count(chunk, []byte{'\n'}))

			err := addLineBreaks(&count, found, maxInt)
			if err != nil {
				return 0, err
			}

			hasData = true
			lastByte = chunk[len(chunk)-1]
		}

		if err != nil {
			if err == io.EOF {
				break
			}

			return 0, fmt.Errorf("%w: %w", ErrReadFailed, err)
		}
	}

	if hasData && lastByte != '\n' {
		err := addFinalLine(&count, maxInt)
		if err != nil {
			return 0, err
		}
	}

	return int(count), nil
}

func addLineBreaks(count *uint, found, maxInt uint) error {
	if found > maxInt || *count > maxInt-found {
		return ErrLineCountOverflow
	}

	*count += found

	return nil
}

func addFinalLine(count *uint, maxInt uint) error {
	if *count == maxInt {
		return ErrLineCountOverflow
	}

	*count++

	return nil
}
