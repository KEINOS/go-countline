# Contribute

- Branch to PR: `main`
- Put new candidate implementations in `cl/_alt`.
- Add the candidate to `cl/_alt/alt_test.go`.
- Add the candidate to the benchmark list in `cl/cl_benchmark_test.go`.
- Keep the shared spec passing.
- Run the local checks before opening a pull request:

```shell
make test
```

For a smaller pass while iterating:

```shell
go test -race ./... ./cl/_alt ./cl/_gen
golangci-lint run ./... ./cl/_alt ./cl/_gen
```

## Adding new files

We only allow files explicitly enabled in `.gitignore` to be added to the repository.

Any new files or directories must be enabled in `.gitignore`.
