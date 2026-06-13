package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"errors"
	"github.com/stretchr/testify/require"
	"github.com/zenizh/go-capturer"
)

var (
	errForcedClose = errors.New("forced close error")
	errForcedFlush = errors.New("forced flush error")
	errForcedExit  = errors.New("forced error")
)

//nolint:paralleltest // do not parallelize due to temporary changing the global variable
func Test_main(t *testing.T) {
	// Mock the path to the Docker environment file. Due to cover the case when
	// the tests are running in Docker.
	if IsDocker() {
		oldPathDockerEnv := pathDockerEnv

		defer func() {
			pathDockerEnv = oldPathDockerEnv
		}()

		pathDockerEnv = "/foo/bar"
	}

	// Create test data directory under temp dir
	pathDirTemp := t.TempDir()

	pathDirTempData := filepath.Join(pathDirTemp, "testdata")
	//nolint:gosec // high perm is not an issue in test code
	require.NoError(t, os.MkdirAll(pathDirTempData, 0o755), "failed to create temp directory")

	// Chenge directory to the temp dir
	t.Chdir(pathDirTemp)

	// Mock testdata and os.Exit function
	oldDataSizes := DataSizes
	oldOsExit := OsExit

	defer func() {
		DataSizes = oldDataSizes
		OsExit = oldOsExit
	}()

	OsExit = func(_ int) {
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
		exitOnError(errForcedExit)
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

	pathFileTemp := filepath.Clean(filepath.Join(t.TempDir(), "test_"+t.Name()+".txt"))

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

//nolint:paralleltest // do not parallelize due to temporary changing function variables
func Test_genFile_close_error(t *testing.T) {
	oldCreateFile := createFile
	oldNewBufferedWriter := newBufferedWriter

	defer func() {
		createFile = oldCreateFile
		newBufferedWriter = oldNewBufferedWriter
	}()

	createFile = func(_ string) (io.WriteCloser, error) {
		return &fakeWriteCloser{
			closeErr: errForcedClose,
			writeErr: nil,
		}, nil
	}
	newBufferedWriter = func(writer io.Writer) bufferedWriter {
		return oldNewBufferedWriter(writer)
	}

	err := genFile(16, filepath.Join(t.TempDir(), "close_error.txt"))

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to close file")
	require.Contains(t, err.Error(), "forced close error")
}

//nolint:paralleltest // do not parallelize due to temporary changing function variables
func Test_genFile_flush_error(t *testing.T) {
	oldCreateFile := createFile
	oldNewBufferedWriter := newBufferedWriter

	defer func() {
		createFile = oldCreateFile
		newBufferedWriter = oldNewBufferedWriter
	}()

	createFile = func(_ string) (io.WriteCloser, error) {
		return &fakeWriteCloser{
			closeErr: nil,
			writeErr: nil,
		}, nil
	}
	newBufferedWriter = func(_ io.Writer) bufferedWriter {
		return &fakeBufferedWriter{
			flushErr: errForcedFlush,
			writeErr: nil,
		}
	}

	err := genFile(16, filepath.Join(t.TempDir(), "flush_error.txt"))

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to flush buffer")
	require.Contains(t, err.Error(), "forced flush error")
}

//nolint:paralleltest // do not parallelize due to temporary changing function variables
func Test_genFile_flush_error_with_existing_error(t *testing.T) {
	oldCreateFile := createFile
	oldNewBufferedWriter := newBufferedWriter

	defer func() {
		createFile = oldCreateFile
		newBufferedWriter = oldNewBufferedWriter
	}()

	createFile = func(_ string) (io.WriteCloser, error) {
		return &fakeWriteCloser{
			closeErr: nil,
			writeErr: nil,
		}, nil
	}
	newBufferedWriter = func(_ io.Writer) bufferedWriter {
		return &fakeBufferedWriter{
			flushErr: errForcedFlush,
			writeErr: errForcedWrite,
		}
	}

	err := genFile(16, filepath.Join(t.TempDir(), "flush_after_write_error.txt"))

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to flush buffer")
	require.Contains(t, err.Error(), "failed to write line")
}

//nolint:paralleltest // do not parallelize due to temporary changing function variables
func Test_genFile_close_error_with_existing_error(t *testing.T) {
	oldCreateFile := createFile
	oldNewBufferedWriter := newBufferedWriter

	defer func() {
		createFile = oldCreateFile
		newBufferedWriter = oldNewBufferedWriter
	}()

	createFile = func(_ string) (io.WriteCloser, error) {
		return &fakeWriteCloser{
			closeErr: errForcedClose,
			writeErr: nil,
		}, nil
	}
	newBufferedWriter = func(_ io.Writer) bufferedWriter {
		return &fakeBufferedWriter{
			flushErr: errForcedFlush,
			writeErr: nil,
		}
	}

	err := genFile(16, filepath.Join(t.TempDir(), "close_after_flush_error.txt"))

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to close file")
	require.Contains(t, err.Error(), "failed to flush buffer")
}

//nolint:paralleltest // do not parallelize due to temporary changing the global variable
func TestIsDocker(t *testing.T) {
	oldPathDockerEnv := pathDockerEnv

	defer func() {
		pathDockerEnv = oldPathDockerEnv
	}()

	pathDummy := filepath.Join(t.TempDir(), "docker_env_dummy")
	require.NoError(t, os.WriteFile(pathDummy, []byte{}, os.ModeTemporary))

	// Mock the path to the Docker environment file
	pathDockerEnv = pathDummy

	// Test in Docker
	require.True(t, IsDocker(), "it should return true if running in Docker")
}

type fakeWriteCloser struct {
	closeErr error
	writeErr error
}

func (w *fakeWriteCloser) Write(data []byte) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}

	return len(data), nil
}

func (w *fakeWriteCloser) Close() error {
	return w.closeErr
}

type fakeBufferedWriter struct {
	flushErr error
	writeErr error
}

func (w *fakeBufferedWriter) Write(data []byte) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}

	return len(data), nil
}

func (w *fakeBufferedWriter) Flush() error {
	return w.flushErr
}
