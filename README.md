<!-- markdownlint-disable MD001 MD041 MD050 MD033 -->
[![Go 1.26+](https://img.shields.io/badge/Go-1.26%2B-blue?logo=go)](https://github.com/KEINOS/go-countline/blob/main/.github/workflows/version-tests.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/KEINOS/go-countline.svg)](https://pkg.go.dev/github.com/KEINOS/go-countline)

# go-countline

Go package
"[go-countline](https://github.com/KEINOS/go-countline/cl)" does one thing:
**count lines in an `io.Reader` quickly**.

Unlike `wc -l`, `go-countline` counts a final line even when the input does not
end with a line feed.

## Usage

```shell
go get "github.com/KEINOS/go-countline"
```

```go
import "github.com/KEINOS/go-countline/countline"

func ExampleCountLines() {
    for _, sample := range []struct {
        Input string
    }{
        {""},            // --> 0
        {"Hello"},       // --> 1
        {"Hello\n"},     // --> 1
        {"\n"},          // --> 1
        {"\n\n"},        // --> 2
        {"\nHello"},     // --> 2
        {"\nHello\n"},   // --> 2
        {"\n\nHello"},   // --> 3
        {"\n\nHello\n"}, // --> 3
    } {
        readerFile := strings.NewReader(sample.Input)

        count, err := countline.CountLines(readerFile)
        if err != nil {
            log.Fatal(err)
        }

        fmt.Printf("%#v --> %v\n", sample.Input, count)
    }
    // Output:
    // "" --> 0
    // "Hello" --> 1
    // "Hello\n" --> 1
    // "\n" --> 1
    // "\n\n" --> 2
    // "\nHello" --> 2
    // "\nHello\n" --> 2
    // "\n\nHello" --> 3
    // "\n\nHello\n" --> 3
}
```

## Benchmark Status

Performance test on Apple M4, 16 GB RAM:

### File I/O (real-world)

Measures speed with file open and read operations:

| File Size | Speed | Time |
| :-- | :-- | :-- |
| 1 KiB | 42 MB/s | 24 μs |
| 1 MiB | 6.5 GB/s | 160 μs |
| 10 MiB | 9.5 GB/s | 1.1 ms |
| 50 MiB | 10.3 GB/s | 5.1 ms |
| 100 MiB | 10.7 GB/s | 9.8 ms |
| **1 GiB** | **10.5 GB/s** | **102 ms** |

### In-Memory (fast path)

Measures speed with data already in memory:

| File Size | Speed |
| :-- | :-- |
| **1 GiB** | **~20 GB/s** |

> **Note**: File I/O is limited by disk speed. In-memory is faster because no
> disk access. Use file I/O results for real-world expectations.

- [See other alternative implementations](./countline/_alt)

## CLI

Install the command-line wrapper:

```shell
go install "github.com/KEINOS/go-countline/cmd/countline@latest"
```

Run it with one file path:

```shell
countline ./path/to/file.txt
```

## Contributing

### Statuses

[![Go 1.26+](https://github.com/KEINOS/go-countline/actions/workflows/version-tests.yaml/badge.svg)](https://github.com/KEINOS/go-countline/actions/workflows/version-tests.yaml)
[![Test on macOS/Win/Linux](https://github.com/KEINOS/go-countline/actions/workflows/platform-test.yaml/badge.svg)](https://github.com/KEINOS/go-countline/actions/workflows/platform-test.yaml)
[![golangci-lint](https://github.com/KEINOS/go-countline/actions/workflows/golangci-lint.yaml/badge.svg)](https://github.com/KEINOS/go-countline/actions/workflows/golangci-lint.yaml)

[![codecov](https://codecov.io/gh/KEINOS/go-countline/branch/main/graph/badge.svg?token=St2W66wHNQ)](https://codecov.io/gh/KEINOS/go-countline)
[![Go Report Card](https://goreportcard.com/badge/github.com/KEINOS/go-countline)](https://goreportcard.com/report/github.com/KEINOS/go-countline)
[![CodeQL](https://github.com/KEINOS/go-countline/actions/workflows/codeQL-analysis.yaml/badge.svg)](https://github.com/KEINOS/go-countline/actions/workflows/codeQL-analysis.yaml)

### Contribute

**Found a faster way** to count lines? Contributions are welcome.

Alternative implementations live in [`countline/_alt`](./countline/_alt). If an alternative
passes the shared spec and benchmarks faster, it can replace the main
implementation in a later release after review.

- [Issues](https://github.com/KEINOS/go-countline/issues): [![Issues](https://img.shields.io/github/issues/KEINOS/go-countline)](https://github.com/KEINOS/go-countline/issues)
  - Please provide a reproducible code snippet.
- Pull requests: [![Pull Requests](https://img.shields.io/github/issues-pr/KEINOS/go-countline)](https://github.com/KEINOS/go-countline/pulls)
  - Branch: `main`
  - **Any pull requests for the better is welcome!**
