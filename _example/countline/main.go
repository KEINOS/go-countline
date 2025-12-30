//nolint:forbidigo,gochecknoglobals
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/KEINOS/go-countline/cl"
	"github.com/pkg/errors"
)

var msgHelp = `cl - Count the number of lines in a file.
Usage:
	cl [file]
`

// osExit is a copy of os.Exit() to be able to mock it in tests.
var osExit = os.Exit

func main() {
	const lenArgs = 2 // program name and the file path

	if len(os.Args) != lenArgs {
		ExitOnError(errors.New("invalid number of arguments"))
	}

	pathFile := os.Args[1]

	osFile, err := os.Open(filepath.Clean(pathFile))
	ExitOnError(err)

	count, err := cl.CountLines(osFile)
	ExitOnError(err)

	fmt.Println(count)
}

func ExitOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, msgHelp)
		fmt.Fprintln(os.Stderr, "error:", err.Error())

		osExit(1)
	}
}
