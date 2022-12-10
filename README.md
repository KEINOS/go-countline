<!-- markdownlint-disable MD001 MD041 MD050 -->
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

Benchmark of counting 1 GiB of file size (72,323,529 lines) on MacBook Pro (Retina, 13-inch, Early 2015, 2.7 GHz Intel Core i5, 4 core).

```shellsession
$ go test -benchmem -count 10 -run=^$ -bench BenchmarkCountLines ./... > bench.txt && benchstat bench.txt
name          time/op
CountLines-4  0.39ns ±19%

name          alloc/op
CountLines-4   1.00B ± 0%

name          allocs/op
CountLines-4    0.00
```

```shellsession
$ cat bench.txt
goos: darwin
goarch: amd64
pkg: github.com/KEINOS/go-countline/cl
cpu: Intel(R) Core(TM) i5-5257U CPU @ 2.70GHz
BenchmarkCountLines-4           1000000000               0.4294 ns/op          1 B/op          0 allocs/op
BenchmarkCountLines-4           1000000000               0.4659 ns/op          1 B/op          0 allocs/op
BenchmarkCountLines-4           1000000000               0.3811 ns/op          1 B/op          0 allocs/op
BenchmarkCountLines-4           1000000000               0.3696 ns/op          1 B/op          0 allocs/op
BenchmarkCountLines-4           1000000000               0.3672 ns/op          1 B/op          0 allocs/op
BenchmarkCountLines-4           1000000000               0.3888 ns/op          1 B/op          0 allocs/op
BenchmarkCountLines-4           1000000000               0.4071 ns/op          1 B/op          0 allocs/op
BenchmarkCountLines-4           1000000000               0.3875 ns/op          1 B/op          0 allocs/op
BenchmarkCountLines-4           1000000000               0.3604 ns/op          1 B/op          0 allocs/op
BenchmarkCountLines-4           1000000000               0.3613 ns/op          1 B/op          0 allocs/op
PASS
ok      github.com/KEINOS/go-countline/cl       85.368s
PASS
ok      github.com/KEINOS/go-countline/cl/spec  0.275s
```

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
