<!-- markdownlint-disable MD001 MD041 MD050 MD033 -->
[![Go 1.26+](https://img.shields.io/badge/Go-1.26%2B-blue?logo=go)](https://github.com/KEINOS/go-countline/blob/main/.github/workflows/version-tests.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/KEINOS/go-countline.svg)](https://pkg.go.dev/github.com/KEINOS/go-countline)

# go-countline

**Blazing-fast line counting for Go.** `go-countline` does one thing — count the lines in an `io.Reader` — and does it at memory speed: a 1 GiB buffer in about 13 ms (~85 GB/s). For large files and in-memory readers it counts concurrently across CPU cores; small inputs and streaming readers use a serial fallback.

On a 1 GiB file it counts lines about **32× faster than `wc -l`** (≈26 ms vs ≈820 ms on an Apple M4). Verify it yourself with `make bench_vs_wc`.

Unlike `wc -l`, it also counts the final line when the input does not end with a line feed.

## Usage

### As a CLI

Install the command-line wrapper:

```shell
go install "github.com/KEINOS/go-countline/cmd/countline@latest"
```

Run it with one file path:

```shell
countline ./path/to/file.txt
```

### As a package

Add it to your module:

```shell
go get "github.com/KEINOS/go-countline"
```

Then pass any `io.Reader` to `CountLines`:

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

## Performance

Counting itself is cheap (a SIMD `bytes.Count`); the work is bound by memory bandwidth. For inputs of 4 MiB or more that support random access, the input is split into regions counted in parallel across CPU cores, so throughput scales past a single core's bandwidth. Smaller inputs and non-seekable readers (pipes, sockets) use a serial stream.

Measured on Apple M4 (10 cores), 16 GB RAM, Go 1.26.

### Versus `wc -l`

Counting the 1 GiB test file (`data_Giant.txt`, warm in the page cache) as a command-line tool, via `hyperfine --warmup 3 --runs 10`:

| Tool | Time (mean) | |
| :-- | :-- | :-- |
| **`countline`** | **26 ms** | _baseline_ |
| `wc -l` | 823 ms | **≈32× slower** |

Both report the same 72,323,529 lines. `wc` here is the BSD build shipped with macOS; GNU `wc` on Linux uses a faster newline scan, so expect a smaller gap there. Reproduce with `make bench_vs_wc`.

### Throughput (file already in the OS page cache)

The benchmark (`BenchmarkCountLines_IO`, `-count=6` medians) re-reads the same file in a loop, so the kernel serves it from RAM. These numbers measure the counting work plus syscall overhead — **not** cold-disk read speed. Reproduce with `make bench`.

| File Size | Time | Throughput | |
| :-- | :-- | :-- | :-- |
| 1 KiB | 12 μs | 89 MB/s | _serial; dominated by `open()` overhead_ |
| 1 MiB | 53 μs | 20 GB/s | _serial; fits in CPU cache_ |
| 10 MiB | 0.34 ms | 31 GB/s | _parallel_ |
| 50 MiB | 1.4 ms | 37 GB/s | _parallel_ |
| 100 MiB | 2.6 ms | 41 GB/s | _parallel_ |
| **1 GiB** | **23 ms** | **~47 GB/s** | _parallel_ |

### In-memory (`bytes.Reader`, no syscalls)

| File Size | Time | Throughput |
| :-- | :-- | :-- |
| **1 GiB** | **13 ms** | **~85 GB/s** |

> **Note**: A first read from cold storage is bound by your disk, not by these figures — expect your SSD/NVMe sequential read speed for uncached files. These results show how little overhead the counting itself adds once the bytes are available.

- [See other alternative implementations](./countline/_alt)

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

Alternative implementations live in [`countline/_alt`](./countline/_alt). If an alternative passes the shared spec and benchmarks faster, it can replace the main implementation in a later release after review.

- [Issues](https://github.com/KEINOS/go-countline/issues): [![Issues](https://img.shields.io/github/issues/KEINOS/go-countline)](https://github.com/KEINOS/go-countline/issues)
  - Please provide a reproducible code snippet.
- Pull requests: [![Pull Requests](https://img.shields.io/github/issues-pr/KEINOS/go-countline)](https://github.com/KEINOS/go-countline/pulls)
  - Branch: `main`
  - **Any pull request that makes it better is welcome!**
