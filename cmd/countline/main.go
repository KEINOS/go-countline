//nolint:forbidigo,gochecknoglobals
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/KEINOS/go-countline/countline"
)

var msgHelp = `countline - Count the number of lines in a file.
Usage:
	countline [file]
`

var errInvalidArgs = errors.New("invalid number of arguments")

// osExit is a copy of os.Exit() to be able to mock it in tests.
var osExit = os.Exit

func main() {
	const lenArgs = 2 // program name and the file path

	if len(os.Args) != lenArgs {
		exitOnError(errInvalidArgs)
	}

	pathFile, err := filepath.Abs(filepath.Clean(os.Args[1]))
	exitOnError(err)

	//nolint:gosec // due to the nature of CLI, allow opening any file
	osFile, err := os.Open(pathFile)
	exitOnError(err)

	count, err := countline.CountLines(osFile)
	exitOnError(err)

	exitOnError(osFile.Close())

	fmt.Println(count)
}

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, msgHelp)

		fmt.Fprintln(os.Stderr, "error:", err.Error())

		osExit(1)
	}
}
