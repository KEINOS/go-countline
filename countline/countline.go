// Package countline counts LF-delimited lines from an io.Reader.
//
// Unlike `wc -l`, it also counts a final line that does not end with a line
// feed.
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
		numRead, errRead := inputReader.Read(buf)
		if numRead > 0 {
			chunk := buf[:numRead]
			found := uint(bytes.Count(chunk, []byte{'\n'}))

			newCount, err := addChecked(count, found, maxInt)
			if err != nil {
				return 0, err
			}

			count = newCount
			hasData = true
			lastByte = chunk[len(chunk)-1]
		}

		if errRead != nil {
			if errRead == io.EOF {
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
