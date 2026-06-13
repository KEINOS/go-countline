package alt

import "errors"

var (
	errNilReader        = errors.New("given reader is nil")
	errLineCountOverflow = errors.New("number of lines exceeds the maximum value of int")
	errForcedRead       = errors.New("forced error")
)
