---
title: Building and Installing Programs
quiz:
  - question: 'Ledgerly''s module has a library package at the root and a program in `cmd/ledgerly`. What does `go build ./...` produce?'
    options:
      - text: One binary per package
      - text: A binary for `cmd/ledgerly` only
      - text: Nothing on disk. With several packages, it compiles them all and discards the results, which is a quick "does everything build?" check
        correct: true
      - text: A `.a` library file for each package
    explanation: |
      `go build` writes an executable only when it builds a single `main`
      package. With a pattern matching several packages, it just checks
      they compile. Use `go build -o` with one package to get a binary.
  - question: 'Where does `go install ./cmd/ledgerly` put the `ledgerly` binary?'
    options:
      - text: In the current directory
      - text: In `$GOBIN`, or `$GOPATH/bin` (usually `~/go/bin`) if `GOBIN` isn't set
        correct: true
      - text: In `/usr/local/bin`
      - text: Next to `main.go`
    explanation: |
      `go install` builds the program and copies it to the install
      directory. Add that directory to your `PATH` and you can run the
      program by name from anywhere. `go env GOBIN GOPATH` shows the
      settings.
---

[Learn Go](/courses/learn-go/introduction/compiling-and-running) introduced `go run` and `go build`. Ledgerly is now a library with a command-line program on top, so here's how the `go` command handles a module with several packages.

## Layout

The usual layout puts the library at the module root and each program under `cmd/`:

```text
ledgerly/
├── go.mod              module ledgerly
├── amount.go           package ledgerly
├── amount_test.go
├── testdata/
└── cmd/
    └── ledgerly/
        └── main.go     package main, imports "ledgerly"
```

The program stays thin. It parses arguments and calls the library, so almost all the logic lives in code your tests already cover.

## run, build, install

```text
$ go run ./cmd/ledgerly balance -account rent   # compile to a temp dir and run
$ go build -o bin/ledgerly ./cmd/ledgerly       # write the binary to bin/
$ go install ./cmd/ledgerly                     # build and copy to the install dir
$ go build ./...                                # does everything compile?
```

- `go run` takes a package, not only a file. `go run .` builds every file in the package, while `go run main.go` builds just that one, which breaks as soon as `main` is split across files. Arguments after the package go to your program. If the program exits with a non-zero status, `go run` prints `exit status 3` and exits with status 1 itself, so use a built binary when the exact code matters.
- `go build` on one `main` package writes a binary named after the directory, or wherever `-o` says. On several packages it writes nothing. It only checks that they compile.
- `go install` puts the binary in `$GOBIN`, or `$GOPATH/bin` (by default `~/go/bin`). Put that directory on your `PATH`.

`go install` also takes a version, and that's how you installed tools earlier in the course:

```text
$ go install golang.org/x/perf/cmd/benchstat@latest
```

To pin a tool's version for a whole team, record it in `go.mod` instead, with `go get -tool honnef.co/go/tools/cmd/staticcheck`, and run it with `go tool staticcheck ./...`. Everyone then gets the same version.

## Build-time settings

Two useful tricks for release builds:

```text
$ GOOS=windows GOARCH=amd64 go build -o ledgerly.exe ./cmd/ledgerly
$ go build -ldflags "-X main.version=v1.2.0" ./cmd/ledgerly
```

The first **cross-compiles**. Go can build for another operating system and CPU from your machine, no extra tools needed. The second sets a package-level string variable at link time. `var version = "dev"` in `main` becomes `"v1.2.0"` in the release binary, so `ledgerly -version` can report it.

`go version -m ./ledgerly` prints the Go version and module information embedded in any Go binary. That's handy when someone asks "which build are you running?".

## In CI

A typical CI job for Ledgerly is a chain of commands whose exit statuses decide the result:

```text
test -z "$(gofmt -l .)" && go vet ./... && go build ./... && go test -race ./...
```

Any command that fails stops the chain and turns the job red.
