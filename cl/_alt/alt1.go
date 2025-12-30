package alt

import (
	"io"

	"github.com/pkg/errors"
	"golang.org/x/text/transform"
)

// ----------------------------------------------------------------------------
//  CountLinesAlt1
// ----------------------------------------------------------------------------

// CountLinesAlt1 uses a Transformer to count the number of lines in a file.
func CountLinesAlt1(inputReader io.Reader) (int, error) {
	if inputReader == nil {
		return 0, errors.New("given reader is nil")
	}

	bufReader := new(LineCounterAlt1)
	transformer := transform.NewReader(inputReader, bufReader)

	_, err := io.ReadAll(transformer)
	if err != nil {
		return 0, errors.Wrap(err, "failed to read the file")
	}

	count := bufReader.Count
	if bufReader.HasFragments {
		count++
	}

	return count, nil
}

// LineCounterAlt1 is a Transformer implementation to count lines.
type LineCounterAlt1 struct {
	LenRead      int
	Count        int
	HasFragments bool
}

// Transform is the implementation of the Transformer interface.
//
//nolint:nonamedreturns // named returns are used for clarity to match the interface
func (lc *LineCounterAlt1) Transform(_, src []byte, _ bool) (nDst, nSrc int, err error) {
	readBytes := 0

	for _, value := range src {
		readBytes++

		if value == '\n' {
			lc.Count++
			lc.HasFragments = false

			continue
		}

		lc.HasFragments = true
	}

	lc.LenRead += readBytes

	return readBytes, readBytes, nil
}

// Reset resets the internal state of LineCounterAlt1.
func (lc *LineCounterAlt1) Reset() {
	lc.LenRead = 0
	lc.Count = 0
	lc.HasFragments = false
}
