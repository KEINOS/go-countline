# Contribute

- Branch to PR: `main`
- Put new candidate implementations in `countline/_alt`.
- Add the candidate to `countline/_alt/alt_test.go`.
- Add the candidate to the benchmark list in `countline/countline_benchmark_test.go`.
- Keep the shared spec passing.
- Run the local checks before opening a pull request:

```shell
make test
```

For a smaller pass while iterating:

```shell
go test -race ./... ./countline/_alt ./countline/_gen
golangci-lint run ./... ./countline/_alt ./countline/_gen
```

## Adding new files

We only allow files explicitly enabled in `.gitignore` to be added to the repository.

Any new files or directories must be enabled in `.gitignore`.
