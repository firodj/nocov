# nocov

*golang tool for eliminating sections of code from coverage statistics*

## Description

This tool is used to post-process the cover profile files created by go test in order to eliminate blocks of code from the test coverage statistics.

## Usage

To see the tool in action you should instrument your code with comments of the form '//nocoverage'

```bash
go build
go test -coverprofile=c.out
sort -i c.out
./nocov c.out > c.out.modified
sort -i c.out.modified
go tool cover -html=c.out.modified
```

Or for development:

```bash
make
```
