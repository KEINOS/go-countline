# Workflow/Actions to Automate

- [codeQL-analysis.yaml](codeQL-analysis.yaml) (on push):
  - Runs CodeQL to scan Go code for security issues.
  - See the GitHub Docs guide for configuring code scanning.
- [golangci-lint.yaml](golangci-lint.yaml) (on push):
  - Runs static analysis via `golangci-lint run`.
  - For linter settings, see [.golangci.yml](../../.golangci.yml).
- [platform-test.yaml](platform-test.yaml) (on push):
  - Runs `go test -race ./...` on Ubuntu, macOS and Windows.
- [update-codecov.yaml](update-codecov.yaml) (on release):
  - Posts the coverage results to codecov.io.
- [version-tests.yaml](version-tests.yaml) (push):
  - Runs tests on the minimum supported Go version and latest Go.
