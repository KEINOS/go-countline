package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/zenizh/go-capturer"
)

//nolint:paralleltest // do not parallelize due to temporary changing global variables
func Test_main(t *testing.T) {
	oldOsArgs := os.Args
	oldOsExit := osExit

	defer func() {
		os.Args = oldOsArgs
		osExit = oldOsExit
	}()

	// Mock os.Exit() to capture the exit code
	capturedCode := 0
	osExit = func(code int) {
		capturedCode = code
	}

	pathData := filepath.Join("..", "..", "cl", "testdata", "data_Giant.txt")
	expect := "72323529"

	os.Args = []string{t.Name(), pathData}

	out := capturer.CaptureOutput(func() {
		main()
	})

	require.Contains(t, out, expect, "output should contain the number of lines")
	require.Equal(t, 0, capturedCode, "exit code should be 0")
}

//nolint:paralleltest // do not parallelize due to temporary changing global variables
func Test_main_missing_args(t *testing.T) {
	oldOsArgs := os.Args
	oldOsExit := osExit

	defer func() {
		os.Args = oldOsArgs
		osExit = oldOsExit
	}()

	// Mock os.Exit() to capture the exit code and panic instead of exiting
	capturedCode := 0
	osExit = func(code int) {
		capturedCode = code

		panic("forced panic")
	}

	os.Args = []string{t.Name()}

	out := capturer.CaptureStderr(func() {
		require.Panics(t, func() {
			main()
		})
	})

	require.Contains(t, out, "Count the number of lines in a file",
		"STDERR should contain the help message on error")
	require.Contains(t, out, "error: invalid number of arguments",
		"STDERR should contain the error reason")
	require.Equal(t, 1, capturedCode, "exit code should be 1 on error")
}

//nolint:paralleltest // do not parallelize due to temporary changing global variables
func TestExitOnError(t *testing.T) {
	oldOsExit := osExit

	defer func() {
		osExit = oldOsExit
	}()

	// Mock os.Exit() to capture the exit code
	capturedCode := 0
	osExit = func(code int) {
		capturedCode = code
	}

	out := capturer.CaptureStderr(func() {
		ExitOnError(errors.New("test error"))
	})

	require.Equal(t, 1, capturedCode, "exit code should be 1")
	require.Contains(t, out, "error: test error", "error reason should be printed to STDERR")
	require.Contains(t, out, "Usage:", "help should be printed on error")
}
