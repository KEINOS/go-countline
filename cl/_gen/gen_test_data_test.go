package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/zenizh/go-capturer"
)

//nolint:paralleltest // do not parallelize due to temporary changing the global variable
func Test_main(t *testing.T) {
	// Prepare to change directory
	pathReturn, err := os.Getwd()
	require.NoError(t, err, "failed to get current working directory")

	defer require.NoError(t, os.Chdir(pathReturn))

	// Create test data directory under temp dir
	pathDirTemp := t.TempDir()

	pathDirTempData := filepath.Join(pathDirTemp, "testdata")
	require.NoError(t, os.MkdirAll(pathDirTempData, 0o755), "failed to create temp directory")

	// Chenge directory to the temp dir
	require.NoError(t, os.Chdir(pathDirTemp))

	// Mock testdata and os.Exit function
	oldDataSizes := DataSizes
	oldOsExit := OsExit

	defer func() {
		DataSizes = oldDataSizes
		OsExit = oldOsExit
	}()

	OsExit = func(code int) {
		panic("panic insted of os.Exit")
	}

	DataSizes = []struct {
		Name string
		Size int
	}{
		{Name: "Dummy1", Size: 32},
		{Name: "Dummy2", Size: 1024 * 1024},
	}

	// Test
	require.NotPanics(t, func() {
		main()
	})

	require.FileExists(t, filepath.Join(pathDirTempData, "data_Dummy1.txt"), "test data not generated")
	require.FileExists(t, filepath.Join(pathDirTempData, "data_Dummy2.txt"), "test data not generated")

	// Re-run test and use generated files
	require.NotPanics(t, func() {
		main()
	})
}

//nolint:paralleltest // do not parallelize due to temporary changing the function variable
func Test_exitOnError(t *testing.T) {
	// Backup and defer restore
	oldOsExit := OsExit
	defer func() {
		OsExit = oldOsExit
	}()

	capturedStatus := 0

	// Mock the os.Exit function
	OsExit = func(code int) {
		capturedStatus = code
	}

	out := capturer.CaptureStderr(func() {
		exitOnError(errors.New("forced error"))
	})

	require.Equal(t, 1, capturedStatus, "it should exit with status 1")
	require.Contains(t, out, "forced error", "it should print the error message to STDERR")
}

//nolint:paralleltest // do not parallelize due to temporary changing global variables
func Test_genFiles_fail_generate_file(t *testing.T) {
	// Mock the bufio.Writer to fail
	forceFailWraite = true
	defer func() {
		forceFailWraite = false
	}()

	// Mock testdata and os.Exit function
	oldDataSizes := DataSizes
	defer func() {
		DataSizes = oldDataSizes
	}()

	DataSizes = []struct {
		Name string
		Size int
	}{
		{Name: "Dummy1", Size: 32},
	}

	err := genFiles(t.TempDir())

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write line",
		"it should contain the error reason if failed to writer")
}

func Test_genFile(t *testing.T) {
	t.Parallel()

	pathFileTemp := filepath.Join(t.TempDir(), "test_"+t.Name()+".txt")

	// Generate a file with 16 bytes in size
	err := genFile(16, pathFileTemp)

	require.NoError(t, err, "failed to generate file")
	require.FileExists(t, pathFileTemp, "file not generated")

	// Test content
	expect := []byte("line: 1\nline: 2\n")
	actual, err := os.ReadFile(pathFileTemp)

	require.NoError(t, err, "failed to read generated file")
	require.Equal(t, string(expect), string(actual), "generated file content mismatch")
}

func Test_genFile_file_is_dir(t *testing.T) {
	t.Parallel()

	err := genFile(16, t.TempDir())

	require.Error(t, err, "it should fail if the path is a directory")
	require.Contains(t, err.Error(), "failed to open/create file", "it should contain the error reason")
}
