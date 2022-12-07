# Alternate implementations of the CountLines function

This directory contains alternate implementations of the CountLines function.

New implementations must be placed in this directory first and Pull-Requested.
After benchmarking on code-review, they may be swapped into the main function.

## Create a file

- File name: `altN.go` where `N` is the next available number. e.g. `alt4.go`

```go
// Replace the `N` in `CountLinesAltN` with the latest number of
// implementations. e.g. `CountLinesAlt4`
func CountLinesAltN(r io.Reader) (int, error) {
    // Your implementation here
}
```

## Add the function to the list

- File name: `alt_test.go`

```diff
func TestCountLines_specs(t *testing.T) {
    for _, targetFunc := range []struct {
        name string
        fn   func(io.Reader) (int, error)
    }{
        // Add the alternate implementations here.
        {"CountLinesAlt1", CountLinesAlt1},
        {"CountLinesAlt2", CountLinesAlt2},
        {"CountLinesAlt3", CountLinesAlt3},
+       {"CountLinesAlt4", CountLinesAlt4},
    } {
        t.Run(targetFunc.name, func(t *testing.T) {
            spec.RunSpecTest(t, targetFunc.name, targetFunc.fn)
        })
    }
}
```

## Regulations

You need the following to be covered in your implementation:

- Pass all the test cases in `cl/spec/spec.go`.
- Pass the test with 100% coverage.
- Pass the lint and static analysis checks of [golangci-lint](https://golangci-lint.run/).
  - Run: `golangci-lint run`
  - For the rules, see: [../../.golangci.yml](../../.golangci.yml)
