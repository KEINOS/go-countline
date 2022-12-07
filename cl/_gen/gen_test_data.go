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

// DataSizes defines the data sizes to generate.
//
//nolint:gomnd
var DataSizes = []struct {
	Name string
	Size int
}{
	{Name: "Tiny", Size: 1024},                // 1KiB
	{Name: "Small", Size: 1024 * 1024},        // 1MiB
	{Name: "Medium", Size: 1024 * 1024 * 10},  // 10MiB
	{Name: "Large", Size: 1024 * 1024 * 50},   // 50MiB
	{Name: "Huge", Size: 1024 * 1024 * 100},   // 100MiB
	{Name: "Giant", Size: 1024 * 1024 * 1024}, // 1GiB
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
		if infoFile, err := os.Stat(pathFile); err == nil && infoFile.Size() >= int64(dataSize) {
			fmt.Printf("  - %s ... OK (exits)\n", pathFile)

			continue
		}

		if err := genFile(dataSize, pathFile); err != nil {
			return errors.Wrap(err, "failed to generate file")
		}
	}

	return nil
}

// forceFailWraite forces the writer to fail if true. It is only used for testing as
// a dependency injection.
var forceFailWraite = false

func genFile(sizeFile int, pathFile string) error {
	fmt.Printf("  - %s ...\r", pathFile)

	fileP, err := os.Create(pathFile)
	if err != nil {
		return errors.Wrap(err, "failed to open/create file")
	}

	defer fileP.Close()

	bufP := bufio.NewWriter(fileP)
	defer bufP.Flush()

	totalSize := int64(0)
	countLine := 0

	for {
		countLine++

		written, err := bufP.WriteString(fmt.Sprintf("line: %d\n", countLine))
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

	return nil
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
