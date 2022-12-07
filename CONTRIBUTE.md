# Contribute

- Branch to PR: `main`
- Place of the new function: `cl/_alt` directory
  - New alternative implementations must be set in the `cl/_alt` directory.
    - See the other implementations for how as examples.
- Minimum requirements to be auto-merged:
  - New function is set in the `cl/_alt` directory.
  - Only two files are edited:
    - `cl/_alt/alt_test.go` and `cl/_alt/alt*.go` (your new implementation).
  - Passes all test with:
    - `go test -race ./...`
  - Passes static analysis with:
    - `golangci-lint run`