<!-- markdownlint-disable MD001 MD041 MD050 MD033 -->
[![go1.16+](https://img.shields.io/badge/Go-1.16--latest-blue?logo=go)](https://github.com/KEINOS/go-countline/blob/main/.github/workflows/version-tests.yaml "Supported versions")
[![Go Reference](https://pkg.go.dev/badge/github.com/KEINOS/go-countline.svg)](https://pkg.go.dev/github.com/KEINOS/go-countline#section-documentation "Read generated documentation of the app")

# go-countline

Go package "[go-countline](https://github.com/KEINOS/go-countline/cl)" does nothing more than **count the number of lines in a file**, but it tries to count as fast as possible.

> __Note__: Unlike the "`wc -l`" command, this package counts the last line that does not end in line breaks/line feeds (see the example below).

## Usage

```go
go get "github.com/KEINOS/go-countline"
```

```go
import "github.com/KEINOS/go-countline/cl"

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

        count, err := cl.CountLines(readerFile)
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

Benchmark of counting:

- 1 GiB of file size (72,323,529 lines)
- On Mac mini (Apple M4, 16 GB RAM, Tahoe 26.2)

```shellsession
$ go test -benchmem -count 10 -run=^$ -bench BenchmarkCountLines ./... > bench.txt && benchstat bench.txt
cpu: Apple M4
              │  bench.txt    │
              │    sec/op     │
CountLines-10   0.09183n ± 4%

              │  bench.txt │
              │    B/op    │
CountLines-10   1.000 ± 0%

              │  bench.txt │
              │ allocs/op  │
CountLines-10   0.000 ± 0%
```

```go
func BenchmarkCountLines(b *testing.B) {
    // 1 GiB size file
    pathFile := filepath.Join("testdata", "data_Giant.txt")

    expectNumLines := 72323529

    // Open file
    fileReader, err := os.Open(pathFile)
    if err != nil {
        b.Fatal(err)
    }

    b.Cleanup(func() {
        fileReader.Close()
    })

    b.ResetTimer() // Begin benchmark

    // Run function
    actualNumLines, err := cl.CountLines(fileReader)
    if err != nil {
        b.Fatal(err)
    }

    b.StopTimer() // End benchmark

    if expectNumLines != actualNumLines {
        b.Fatalf(
            "test %v failed: expect=%d, actual=%d",
            b.Name(), expectNumLines, actualNumLines,
        )
    }
}
```

<details><summary>bench.txt</summary>

```shellsession
$ cat bench.txt
goos: darwin
goarch: arm64
pkg: github.com/KEINOS/go-countline/cl
cpu: Apple M4
BenchmarkCountLines-10      1000000000           0.09415 ns/op         1 B/op         0 allocs/op
BenchmarkCountLines-10      1000000000           0.09049 ns/op         1 B/op         0 allocs/op
BenchmarkCountLines-10      1000000000           0.09107 ns/op         1 B/op         0 allocs/op
BenchmarkCountLines-10      1000000000           0.09030 ns/op         1 B/op         0 allocs/op
BenchmarkCountLines-10      1000000000           0.09347 ns/op         1 B/op         0 allocs/op
BenchmarkCountLines-10      1000000000           0.08913 ns/op         1 B/op         0 allocs/op
BenchmarkCountLines-10      1000000000           0.09019 ns/op         1 B/op         0 allocs/op
BenchmarkCountLines-10      1000000000           0.08834 ns/op         1 B/op         0 allocs/op
BenchmarkCountLines-10      1000000000           0.09008 ns/op         1 B/op         0 allocs/op
BenchmarkCountLines-10      1000000000           0.08829 ns/op         1 B/op         0 allocs/op
PASS
ok    github.com/KEINOS/go-countline/cl  9.566s
PASS
ok    github.com/KEINOS/go-countline/cl/spec  0.629s
```

</details>

- [See other alternative implementations](./cl/_alt)

## Contributing

### Statuses

[![Go 1.16~latest](https://github.com/KEINOS/go-countline/actions/workflows/version-tests.yaml/badge.svg)](https://github.com/KEINOS/go-countline/actions/workflows/version-tests.yaml)
[![Test on macOS/Win/Linux](https://github.com/KEINOS/go-countline/actions/workflows/platform-test.yaml/badge.svg)](https://github.com/KEINOS/go-countline/actions/workflows/platform-test.yaml)
[![golangci-lint](https://github.com/KEINOS/go-countline/actions/workflows/golangci-lint.yaml/badge.svg)](https://github.com/KEINOS/go-countline/actions/workflows/golangci-lint.yaml)

[![codecov](https://codecov.io/gh/KEINOS/go-countline/branch/main/graph/badge.svg?token=St2W66wHNQ)](https://codecov.io/gh/KEINOS/go-countline)
[![Go Report Card](https://goreportcard.com/badge/github.com/KEINOS/go-countline)](https://goreportcard.com/report/github.com/KEINOS/go-countline)
[![CodeQL](https://github.com/KEINOS/go-countline/actions/workflows/codeQL-analysis.yaml/badge.svg)](https://github.com/KEINOS/go-countline/actions/workflows/codeQL-analysis.yaml)

### Contribute

**If you have found a faster way** to count the number of lines in a file, feel free to contribute!

As long as the new function passes the test, it is merged. It then will be replaced to the main fucntion in the next release after the review by the contributors.

- [Issues](https://github.com/KEINOS/go-countline/issues): [![Issues](https://img.shields.io/github/issues/KEINOS/go-countline)](https://github.com/KEINOS/go-countline/issues)
  - Please provide a reproducible code snippet.
- Pull requests: [![Pull Requests](https://img.shields.io/github/issues-pr/KEINOS/go-countline)](https://github.com/KEINOS/go-countline/pulls)
  - Branch: `main`
  - **Any pull requests for the better is welcome!**
