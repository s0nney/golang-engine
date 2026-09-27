---
title: The split Command
difficulty: easy
after: terminal-git-and-go
hints:
  - 'Create your own `flag.NewFlagSet("split", flag.ContinueOnError)`, call `fs.SetOutput(stderr)` so its messages go to the right writer, define `ways := fs.Int("ways", 2, "...")`, then `fs.Parse(args)`. If `Parse` returns an error it has already printed it: just return 2.'
  - 'After parsing, `fs.Args()` holds what''s left. Anything other than exactly one argument is a usage error (exit 2). A bad amount or `-ways` below 1 is an error in otherwise well-formed input (exit 1). Write every error message to `stderr` with `fmt.Fprintln(stderr, ...)`.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"io"
    	"os"
    	"strings"
    )

    // run is the whole split command, minus the os package:
    //
    //	split [-ways N] CENTS
    //
    // It prints the parts of Split(CENTS, N) to stdout, one per line, and
    // returns the exit status: 0 on success, 1 for a bad amount or a -ways
    // below 1, and 2 for a usage error (unknown flag, wrong number of
    // arguments). Error messages go to stderr.
    func run(args []string, stdout, stderr io.Writer) int {
    	// 1. Make a flag.FlagSet named "split" with flag.ContinueOnError,
    	//    send its output to stderr, and define -ways (default 2).
    	// 2. Parse args. On error, return 2.
    	// 3. Exactly one argument must remain, or print the usage line and return 2.
    	// 4. Parse it with strconv.ParseInt; check ways >= 1. On error, return 1.
    	// 5. Print each part of Split to stdout and return 0.
    	return 0
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    func (c Cents) String() string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s$%d.%02d", sign, n/100, n%100)
    }

    // Split divides total into n parts (n >= 1) that differ by at most one
    // cent, with the larger parts first.
    func Split(total Cents, n int) []Cents {
    	parts := make([]Cents, n)
    	base := total / Cents(n)
    	if total%Cents(n) < 0 {
    		base-- // round down, not toward zero, for negative totals
    	}
    	extra := int(total - base*Cents(n))
    	for i := range parts {
    		parts[i] = base
    		if i < extra {
    			parts[i]++
    		}
    	}
    	return parts
    }

    // main would normally be os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)).
    // Here it tries a few command lines so Run shows something.
    func main() {
    	for _, args := range [][]string{
    		{"1000"},
    		{"-ways", "3", "1000"},
    		{},
    		{"-ways", "0", "1000"},
    		{"ten"},
    	} {
    		fmt.Println("$ split", strings.Join(args, " "))
    		code := run(args, os.Stdout, os.Stdout)
    		fmt.Printf("(exit %d)\n\n", code)
    	}
    }
  solution: |
    package main

    import (
    	"flag"
    	"fmt"
    	"io"
    	"os"
    	"strconv"
    	"strings"
    )

    // run is the whole split command, minus the os package:
    //
    //	split [-ways N] CENTS
    //
    // It prints the parts of Split(CENTS, N) to stdout, one per line, and
    // returns the exit status: 0 on success, 1 for a bad amount or a -ways
    // below 1, and 2 for a usage error (unknown flag, wrong number of
    // arguments). Error messages go to stderr.
    func run(args []string, stdout, stderr io.Writer) int {
    	fs := flag.NewFlagSet("split", flag.ContinueOnError)
    	fs.SetOutput(stderr)
    	ways := fs.Int("ways", 2, "number of parts")
    	if err := fs.Parse(args); err != nil {
    		return 2
    	}
    	if fs.NArg() != 1 {
    		fmt.Fprintln(stderr, "usage: split [-ways N] CENTS")
    		return 2
    	}
    	total, err := strconv.ParseInt(fs.Arg(0), 10, 64)
    	if err != nil {
    		fmt.Fprintf(stderr, "split: bad amount %q\n", fs.Arg(0))
    		return 1
    	}
    	if *ways < 1 {
    		fmt.Fprintf(stderr, "split: -ways must be at least 1, got %d\n", *ways)
    		return 1
    	}
    	for _, part := range Split(Cents(total), *ways) {
    		fmt.Fprintln(stdout, part)
    	}
    	return 0
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    func (c Cents) String() string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s$%d.%02d", sign, n/100, n%100)
    }

    // Split divides total into n parts (n >= 1) that differ by at most one
    // cent, with the larger parts first.
    func Split(total Cents, n int) []Cents {
    	parts := make([]Cents, n)
    	base := total / Cents(n)
    	if total%Cents(n) < 0 {
    		base-- // round down, not toward zero, for negative totals
    	}
    	extra := int(total - base*Cents(n))
    	for i := range parts {
    		parts[i] = base
    		if i < extra {
    			parts[i]++
    		}
    	}
    	return parts
    }

    // main would normally be os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)).
    // Here it tries a few command lines so Run shows something.
    func main() {
    	for _, args := range [][]string{
    		{"1000"},
    		{"-ways", "3", "1000"},
    		{},
    		{"-ways", "0", "1000"},
    		{"ten"},
    	} {
    		fmt.Println("$ split", strings.Join(args, " "))
    		code := run(args, os.Stdout, os.Stdout)
    		fmt.Printf("(exit %d)\n\n", code)
    	}
    }
  tests: |
    package main

    import (
    	"bytes"
    	"strings"
    	"testing"
    )

    func TestRun(t *testing.T) {
    	tests := []struct {
    		name      string
    		args      []string
    		code      int
    		stdout    string
    		stderrHas string // "" means stderr must be empty
    	}{
    		{"default two ways", []string{"1000"}, 0, "$5.00\n$5.00\n", ""},
    		{"three ways", []string{"-ways", "3", "1000"}, 0, "$3.34\n$3.33\n$3.33\n", ""},
    		{"equals form", []string{"-ways=4", "10"}, 0, "$0.03\n$0.03\n$0.02\n$0.02\n", ""},
    		{"negative after --", []string{"-ways", "1", "--", "-250"}, 0, "-$2.50\n", ""},
    		{"negative three ways", []string{"-ways=3", "--", "-1000"}, 0, "-$3.33\n-$3.33\n-$3.34\n", ""},
    		{"no arguments", []string{}, 2, "", "usage"},
    		{"two amounts", []string{"100", "200"}, 2, "", "usage"},
    		{"unknown flag", []string{"-people", "3", "100"}, 2, "", "people"},
    		{"flag after amount", []string{"100", "-ways", "3"}, 2, "", "usage"},
    		{"bad amount", []string{"ten"}, 1, "", "ten"},
    		{"dollars not cents", []string{"12.50"}, 1, "", "12.50"},
    		{"zero ways", []string{"-ways", "0", "100"}, 1, "", "ways"},
    		{"negative ways", []string{"-ways", "-2", "100"}, 1, "", "ways"},
    	}
    	for _, tt := range tests {
    		t.Run(tt.name, func(t *testing.T) {
    			var stdout, stderr bytes.Buffer
    			code := run(tt.args, &stdout, &stderr)
    			cmd := "split " + strings.Join(tt.args, " ")
    			if code != tt.code {
    				t.Errorf("%s: exit status %d, want %d (stderr: %q)", cmd, code, tt.code, stderr.String())
    			}
    			if stdout.String() != tt.stdout {
    				t.Errorf("%s: stdout = %q, want %q", cmd, stdout.String(), tt.stdout)
    			}
    			if tt.stderrHas == "" && stderr.Len() > 0 {
    				t.Errorf("%s: stderr = %q, want nothing on success", cmd, stderr.String())
    			}
    			if tt.stderrHas != "" && !strings.Contains(stderr.String(), tt.stderrHas) {
    				t.Errorf("%s: stderr = %q, want a message mentioning %q", cmd, stderr.String(), tt.stderrHas)
    			}
    		})
    	}
    }

    // run must not keep state between calls (for example in a global FlagSet).
    func TestRunTwice(t *testing.T) {
    	var out1, out2, errs bytes.Buffer
    	run([]string{"-ways", "3", "300"}, &out1, &errs)
    	run([]string{"300"}, &out2, &errs)
    	if out2.String() != "$1.50\n$1.50\n" {
    		t.Errorf("split -ways 3 300, then split 300: second stdout = %q, want %q (does -ways leak between calls?)", out2.String(), "$1.50\n$1.50\n")
    	}
    }
---

Ledgerly is getting a command-line tool, starting with `split`, which shares
a bill:

```text
$ split -ways 3 1000
$3.34
$3.33
$3.33
```

`main` itself can't be tested: it reads `os.Args`, writes straight to the
terminal and ends with `os.Exit`. So everything lives in `run`, which takes
the arguments and the two output streams, and **returns** the exit status.

Complete `run(args, stdout, stderr)` for the command `split [-ways N] CENTS`:

- `-ways N` (default `2`) is the number of parts. `CENTS` is an integer
  amount in cents, which may be negative.
- On success, print each part of `Split(CENTS, N)` on its own line to
  `stdout` and return `0`. Nothing goes to `stderr`.
- **Usage errors** return `2` with a message on `stderr`: an unknown flag (the
  `flag` package prints that one for you), or anything other than exactly one
  argument after the flags. Include the word `usage` in your message, e.g.
  `usage: split [-ways N] CENTS`.
- **Bad values** return `1` with a message on `stderr` that mentions the
  problem: an amount that isn't an integer (mention the amount), or a `-ways`
  below 1 (mention `-ways`).

## Examples

```text
split 1000             → stdout "$5.00\n$5.00\n", exit 0
split -ways=4 10       → stdout "$0.03\n$0.03\n$0.02\n$0.02\n", exit 0
split                  → stderr "usage: ...", exit 2
split 100 -ways 3      → exit 2 (flags must come first; -ways is then a 2nd argument)
split ten              → stderr "split: bad amount \"ten\"", exit 1
split -- -250          → stdout "-$1.25\n-$1.25\n", exit 0
```

That last one is a classic gotcha: without `--`, the flag package would read
`-250` as an (unknown) flag. `--` means "no more flags after this", and
`FlagSet.Parse` handles it for you.

**Run** tries a handful of command lines and prints both streams with the
exit status.

## Constraints

- Use your own `flag.NewFlagSet` with `flag.ContinueOnError`, not the global
  `flag.Parse`: the global one exits the whole program on a bad flag and keeps
  its values between calls.
