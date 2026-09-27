---
title: The Terminal and Exit Codes
quiz:
  - question: |
      Your CI script runs this line, and `go vet` reports a problem. What happens next?

      ```text
      go vet ./... && go test ./...
      ```
    options:
      - text: '`go test` runs anyway, and the script''s result is whatever the tests return'
      - text: '`go test` doesn''t run, because `&&` only runs the second command when the first exits with status 0'
        correct: true
      - text: The shell retries `go vet` until it passes
      - text: '`go test` runs first, because commands on one line run right to left'
    explanation: |
      Every command finishes with an exit status. 0 means success, and
      anything else means failure. `a && b` runs `b` only if `a` succeeded,
      so the line as a whole fails with vet's status, and the CI job fails.
  - question: A Go program ends with an unrecovered `panic`. What exit status does the process report?
    options:
      - text: '0'
      - text: '1'
      - text: '2'
        correct: true
      - text: It depends on the panic message
    explanation: |
      The Go runtime prints the panic and stack trace to standard error and
      exits with status 2. Returning from `main` normally gives 0, and
      `os.Exit(n)` gives `n`.
---

Every command in this course, from `go test` and `go vet` to `benchstat` and `go tool pprof`, runs in a terminal. You've been typing them for eight chapters. This lesson covers the terminal ideas behind them. They matter most when commands run without you, in CI.

## Where you are

A shell always has a **current directory**. Relative paths and Go's package patterns start from it:

```text
$ pwd
/home/you/ledgerly
$ ls
amount.go  amount_test.go  cmd  go.mod  testdata
$ cd cmd/ledgerly     # into a subdirectory
$ cd ../..            # back up two levels
$ go test ./...       # . is here, ./... is here and everything below
```

## Two output streams

A program has two output streams. **Standard output** (stdout) is for results. **Standard error** (stderr) is for diagnostics: errors, progress, warnings. They both appear in your terminal, but the shell can send them to different places:

```text
$ go test -run '^$' -bench . ./... > bench.txt     # stdout into a file
$ ledgerly import march.csv 2> errors.log          # stderr into a file
$ go test -v ./... 2>&1 | grep FAIL                # both, piped into grep
```

`>` redirects stdout to a file, `2>` redirects stderr, and `|` pipes one program's stdout into the next one's input. That's why well-behaved programs print their results to stdout and their complaints to stderr. `ledgerly balance > balance.txt` should still show you an error on screen.

## Exit codes

When a program finishes, it hands the shell a number: its **exit status**. `0` means success, and any other value means failure. The shell keeps the last one in `$?`:

```text
$ go test ./...
ok  	ledgerly	0.004s
$ echo $?
0
$ go test ./...
--- FAIL: TestParseAmount (0.00s)
FAIL
$ echo $?
1
```

This is how CI knows your tests failed: not by reading `FAIL`, but from the exit status. The same goes for `gofmt -l`, which always exits 0, so CI checks its output instead (`test -z "$(gofmt -l .)"`).

In a Go program:

- returning from `main` exits with 0;
- `os.Exit(n)` exits immediately with `n`, **without running deferred functions**;
- an unrecovered panic exits with 2;
- `log.Fatal` prints a message and calls `os.Exit(1)`.

Unix convention uses 2 for "you called me wrong" (a bad flag or a missing argument), which is exactly what the `flag` package does, as you'll see later in this chapter.

## Chaining commands

The shell uses exit statuses to chain commands:

```text
go vet ./... && go test ./...        # test only if vet passed
go build ./... || echo "build broke" # run the second only on failure
```

A CI job is essentially one long `&&` chain. The first non-zero status stops it and turns the build red.
