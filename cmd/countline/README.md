# Countline Command

## Install

```shell
go install "github.com/KEINOS/go-countline/cmd/countline@latest"
```

## Usage

```shell
countline ./path/to/file.txt
```

The command prints the line count to standard output. Like the library, it
counts a final line even when the file does not end with `\n`.
