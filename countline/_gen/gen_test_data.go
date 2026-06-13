// This creates a file of 1 GiB in size with lines of random values.
//
//nolint:gochecknoglobals,forbidigo
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var errForcedWrite = errors.New("forced error")

const KiB = 1024
const MiB = 1024 * KiB
const GiB = 1024 * MiB

// DataSizes defines the data sizes to generate.
//
//nolint:mnd
var DataSizes = []struct {
	Name string
	Size int
}{
	{Name: "Tiny", Size: 1 * KiB},    // 1KiB
	{Name: "Small", Size: 1 * MiB},   // 1MiB
	{Name: "Medium", Size: 10 * MiB}, // 10MiB
	{Name: "Large", Size: 50 * MiB},  // 50MiB
	{Name: "Huge", Size: 100 * MiB},  // 100MiB
	{Name: "Giant", Size: 1 * GiB},   // 1GiB
}

func main() {
	fmt.Printf("Generating consistent data file:\n")

	err := genFiles(filepath.Join(".", "testdata"))
	exitOnError(err)
}

var OsExit = os.Exit

func exitOnError(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		OsExit(1)
	}
}

func genFiles(pathDirBase string) error {
	for _, data := range DataSizes {
		dataSize := data.Size
		nameFile := "data_" + data.Name + ".txt"
		pathFile := filepath.Join(pathDirBase, nameFile)

		// Skip if file already exists
		infoFile, err := os.Stat(pathFile)
		if err == nil && infoFile.Size() >= int64(dataSize) {
			fmt.Printf("  - %s ... OK (exits)\n", pathFile)

			continue
		}

		err = genFile(dataSize, pathFile)
		if err != nil {
			return fmt.Errorf("failed to generate file: %w", err)
		}
	}

	return nil
}

// forceFailWraite forces the writer to fail if true. It is only used for testing as
// a dependency injection.
var forceFailWraite = false

type bufferedWriter interface {
	io.Writer
	Flush() error
}

var createFile = func(name string) (io.WriteCloser, error) {
	//nolint:gosec // path is intentionally provided by the generator caller.
	return os.Create(name)
}

var newBufferedWriter = func(writer io.Writer) bufferedWriter {
	return bufio.NewWriter(writer)
}

//nolint:cyclop,funlen // acceptable complexity and length for this function
func genFile(sizeFile int, pathFile string) (retErr error) {
	pathFile = filepath.Clean(pathFile)

	fmt.Printf("  - %s ...\r", pathFile)

	fileP, err := createFile(pathFile)
	if err != nil {
		return fmt.Errorf("failed to open/create file: %w", err)
	}

	defer func() {
		err := fileP.Close()
		if err != nil {
			if retErr == nil {
				retErr = fmt.Errorf("failed to close file: %w", err)
			} else {
				retErr = fmt.Errorf("failed to close file: %w", retErr)
			}
		}
	}()

	bufP := newBufferedWriter(fileP)

	defer func() {
		err := bufP.Flush()
		if err != nil {
			if retErr == nil {
				retErr = fmt.Errorf("failed to flush buffer: %w", err)
			} else {
				retErr = fmt.Errorf("failed to flush buffer: %w", retErr)
			}
		}
	}()

	totalSize := int64(0)
	countLine := 0

	for {
		countLine++

		written, err := fmt.Fprintf(bufP, "line: %d\n", countLine)
		if err != nil || forceFailWraite {
			if forceFailWraite {
				err = errForcedWrite
			}

			return fmt.Errorf("failed to write line: %w", err)
		}

		totalSize += int64(written)

		if totalSize >= int64(sizeFile) {
			fmt.Printf("  - %s, size: %d, line: %d ... OK\n", pathFile, totalSize, countLine)

			break
		}

		if totalSize%1024 == 0 && !IsCI() {
			fmt.Printf("  - %s, size: %d, line: %d\r", pathFile, totalSize, countLine)
		}
	}

	return retErr
}

var pathDockerEnv = filepath.Join("/", ".dockerenv")

// IsCI returns true if the current process is running inside a CI environment. Such as Github Actions or Docker.
func IsCI() bool {
	return IsGithubActions() || IsDocker()
}

// IsDocker returns true if the current process is running inside a Docker container.
func IsDocker() bool {
	_, err := os.Stat(pathDockerEnv)

	return err == nil
}

// IsGithubActions returns true if the current process is running inside a Github Actions.
func IsGithubActions() bool {
	return os.Getenv("GITHUB_ACTIONS") != ""
}
