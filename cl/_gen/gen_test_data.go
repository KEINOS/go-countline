// This creates a file of 1 GiB in size with lines of random values.
//
//nolint:gochecknoglobals,forbidigo
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

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
			return errors.Wrap(err, "failed to generate file")
		}
	}

	return nil
}

// forceFailWraite forces the writer to fail if true. It is only used for testing as
// a dependency injection.
var forceFailWraite = false

//nolint:cyclop,funlen // acceptable complexity and length for this function
func genFile(sizeFile int, pathFile string) (retErr error) {
	pathFile = filepath.Clean(pathFile)

	fmt.Printf("  - %s ...\r", pathFile)

	fileP, err := os.Create(pathFile)
	if err != nil {
		return errors.Wrap(err, "failed to open/create file")
	}

	defer func() {
		err := fileP.Close()
		if err != nil {
			if retErr == nil {
				retErr = errors.Wrap(err, "failed to close file")
			} else {
				retErr = errors.Wrap(retErr, "failed to close file")
			}
		}
	}()

	bufP := bufio.NewWriter(fileP)

	defer func() {
		err := bufP.Flush()
		if err != nil {
			if retErr == nil {
				retErr = errors.Wrap(err, "failed to flush buffer")
			} else {
				retErr = errors.Wrap(retErr, "failed to flush buffer")
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
				err = errors.New("forced error")
			}

			return errors.Wrap(err, "failed to write line")
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
