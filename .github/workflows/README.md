# Workflow/Actions to Automate

- [codeQL-analysis.yml](codeQL-analysis.yml) (on push):
  - Runs [CodeQL](https://codeql.github.com/docs/codeql-overview/about-codeql/) action to check known-vulnerability and scan security in Go code.
  - See: [Configuring code scanning](https://docs.github.com/en/free-pro-team@latest/github/finding-security-vulnerabilities-and-errors-in-your-code/configuring-code-scanning#changing-the-languages-that-are-analyzed) @ docs.github.com
- [golangci-lint.yaml](golangci-lint.yaml) (on push):
  - Runs static analysis via `golangci-lint run`.
  - The contents of the test see: [.golangci.yml](../../.golangci.yml)
- [platform-test.yaml](platform-test.yaml) (on push):
  - Runs `go test -race ./...` on Ubuntu, macOS and Windows.
- [update-codecov.yaml](update-codecov.yaml) (on release):
  - Posts the coverage results to codecov.io.
- [version-tests.yaml](version-tests.yaml) (push):
  - Runs `go test ./...` on vaious Go versions. Minimum supported Go version to the latest.
