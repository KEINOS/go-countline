# Alternate implementations of the CountLines function

This directory contains alternate implementations of the CountLines function.

New implementations must be placed in this directory first and Pull-Requested.
After benchmarking on code-review, they may be swapped into the main function.

## Files to change

You need to create and edit the following 3 files:

1. Create a new file for your new implementation.
2. Add your function to the `TestCountLines_specs` function in [`alt_test.go`](./alt_test.go).
3. Add your function to the `targetFuncions` variable in [`../../benchmarks_test.go`](../../benchmarks_test.go).

### Create a file

- File name: `altN.go` where `N` is the next available number. e.g. `alt4.go`

```go
// Replace the `N` in `CountLinesAltN` with the latest number of
// implementations. e.g. `CountLinesAlt4`
func CountLinesAltN(r io.Reader) (int, error) {
    // Your implementation here
}
```

### Add the function to the test list

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

### Add the function to the benchmark list

- File name: `../../benchmarks_test.go`

```diff
var targetFuncions = map[string]struct {
    fn func(io.Reader) (int, error)
}{
    // Current implementation
    "CountLinesCurr": {CountLines},
    // Alternate implementation. See alt_test.go.
    "CountLinesAlt1": {_alt.CountLinesAlt1},
    "CountLinesAlt2": {_alt.CountLinesAlt2},
    "CountLinesAlt3": {_alt.CountLinesAlt3},
+   "CountLinesAlt4": {_alt.CountLinesAlt4},
}

```

## Regulations

You need the following to be covered in your implementation:

- **Pass the tests** in all cases of `cl/spec/spec.go`.
- **Pass the lint and static analysis checks** of [golangci-lint](https://golangci-lint.run/).
  - Run: `golangci-lint run`
  - For the rules, see: [../../.golangci.yml](../../.golangci.yml)
- **Keep the code coverage** up to 100%.

Use the [`make` command](https://en.wikipedia.org/wiki/Make_(software)) for convenience.

```shellsession
$ # Runs unit tests, lints, and cove coverage.
$ make test
...

$ make benchmark
...
```
