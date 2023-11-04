<h1 align="center">TLDeets</h1>

A small command-line tool to provide basic information about a top-level domain (TLD).

---

## Quick Start

### Using the Library

```bash
$ go get github.com/jakewilliami/tldeets  # TODO: get this working
```

```go
TODO: add usage examples
```

### Compiling the CLI

```bash
$ ./build.sh    # All-in-one build script
$ go run ./...  # Or run the project
$ go build -o ./tldeets cmd/tldeets/main.go  # Or build it into a binary
```

## Caveats

  - Currently only supports country code TLDs (ccTLDs), as this is my personal main usecase for the application.  It does not yet generic TLDs (gTLD).
  - Internationalised ccTLDs are not yet supported.
