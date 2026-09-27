---
title: os.Args and the flag Package
quiz:
  - question: |
      `report` defines a bool flag `-csv` and a string flag `-account`.
      What does `fs.Args()` return after this call?

      ```go
      fs.Parse([]string{"-account", "rent", "march.csv", "-csv"})
      ```
    options:
      - text: '`["march.csv"]`, with `-csv` set to true'
      - text: '`["march.csv", "-csv"]`, with `-csv` still false'
        correct: true
      - text: An error, because `-csv` comes after a file name
      - text: '`["rent", "march.csv", "-csv"]`'
    explanation: |
      The flag package stops parsing at the first argument that isn't a
      flag. Everything from `march.csv` on is a positional argument, even
      if it starts with `-`. Put flags before file names.
  - question: A program calls `flag.Parse()` and the user passes `-limit five` for an `int` flag. What happens?
    options:
      - text: '`-limit` is silently set to 0'
      - text: The program panics
      - text: '`flag.Parse` prints the error and the usage message to standard error, then exits with status 2'
        correct: true
      - text: '`flag.Parse` returns an error for `main` to handle'
    explanation: |
      The top-level `flag` functions use `flag.ExitOnError`. That's
      convenient in `main`, but impossible to test, which is why testable
      code builds its own `flag.FlagSet` with `flag.ContinueOnError`.
exercise:
  starter: |
    package main

    import (
    	"flag"
    	"fmt"
    	"io"
    )

    // Config holds the settings for `ledgerly report`.
    type Config struct {
    	Account string   // -account: required
    	Limit   int      // -limit: rows to show, default 20, must be positive
    	CSV     bool     // -csv: print CSV instead of a table
    	Files   []string // the arguments after the flags: at least one
    }

    // parseArgs parses the arguments after `ledgerly report` into a Config.
    func parseArgs(args []string) (Config, error) {
    	var cfg Config
    	fs := flag.NewFlagSet("report", flag.ContinueOnError)
    	fs.SetOutput(io.Discard) // return parse errors instead of printing them
    	// ? define -account, -limit and -csv, parse args, then validate cfg
    	_ = fs
    	return cfg, nil
    }

    func main() {
    	for _, args := range [][]string{
    		{"-account", "rent", "march.csv"},
    		{"-account=rent", "-limit", "5", "-csv", "march.csv", "april.csv"},
    		{"-account", "rent", "march.csv", "-csv"},
    		{"-limit", "five", "march.csv"},
    		{"march.csv"},
    	} {
    		cfg, err := parseArgs(args)
    		fmt.Printf("%q\n  => %+v, err: %v\n", args, cfg, err)
    	}
    }
  solution: |
    package main

    import (
    	"errors"
    	"flag"
    	"fmt"
    	"io"
    )

    // Config holds the settings for `ledgerly report`.
    type Config struct {
    	Account string   // -account: required
    	Limit   int      // -limit: rows to show, default 20, must be positive
    	CSV     bool     // -csv: print CSV instead of a table
    	Files   []string // the arguments after the flags: at least one
    }

    // parseArgs parses the arguments after `ledgerly report` into a Config.
    func parseArgs(args []string) (Config, error) {
    	var cfg Config
    	fs := flag.NewFlagSet("report", flag.ContinueOnError)
    	fs.SetOutput(io.Discard)
    	fs.StringVar(&cfg.Account, "account", "", "account to report on (required)")
    	fs.IntVar(&cfg.Limit, "limit", 20, "number of rows to show")
    	fs.BoolVar(&cfg.CSV, "csv", false, "print CSV instead of a table")
    	if err := fs.Parse(args); err != nil {
    		return Config{}, err
    	}
    	cfg.Files = fs.Args()
    	switch {
    	case cfg.Account == "":
    		return Config{}, errors.New("-account is required")
    	case cfg.Limit <= 0:
    		return Config{}, fmt.Errorf("-limit must be positive, got %d", cfg.Limit)
    	case len(cfg.Files) == 0:
    		return Config{}, errors.New("no files to report on")
    	}
    	return cfg, nil
    }

    func main() {
    	for _, args := range [][]string{
    		{"-account", "rent", "march.csv"},
    		{"-account=rent", "-limit", "5", "-csv", "march.csv", "april.csv"},
    		{"-account", "rent", "march.csv", "-csv"},
    		{"-limit", "five", "march.csv"},
    		{"march.csv"},
    	} {
    		cfg, err := parseArgs(args)
    		fmt.Printf("%q\n  => %+v, err: %v\n", args, cfg, err)
    	}
    }
  tests: |
    package main

    import (
    	"errors"
    	"flag"
    	"slices"
    	"testing"
    )

    func TestParseArgsValid(t *testing.T) {
    	for _, tc := range []struct {
    		args []string
    		want Config
    	}{
    		{[]string{"-account", "rent", "march.csv"},
    			Config{Account: "rent", Limit: 20, Files: []string{"march.csv"}}},
    		{[]string{"-account=rent", "-limit", "5", "-csv", "march.csv", "april.csv"},
    			Config{Account: "rent", Limit: 5, CSV: true, Files: []string{"march.csv", "april.csv"}}},
    		{[]string{"--account", "cash", "-limit=1", "-csv=false", "q1.csv"},
    			Config{Account: "cash", Limit: 1, Files: []string{"q1.csv"}}},
    		{[]string{"-csv", "-account", "rent", "march.csv", "-limit", "3"},
    			Config{Account: "rent", Limit: 20, CSV: true, Files: []string{"march.csv", "-limit", "3"}}},
    		{[]string{"-account", "rent", "--", "-weird-name.csv"},
    			Config{Account: "rent", Limit: 20, Files: []string{"-weird-name.csv"}}},
    	} {
    		got, err := parseArgs(tc.args)
    		if err != nil {
    			t.Errorf("parseArgs(%q) returned error %v, want %+v", tc.args, err, tc.want)
    			continue
    		}
    		if got.Account != tc.want.Account || got.Limit != tc.want.Limit || got.CSV != tc.want.CSV || !slices.Equal(got.Files, tc.want.Files) {
    			t.Errorf("parseArgs(%q) = %+v, want %+v", tc.args, got, tc.want)
    		}
    	}
    }

    func TestParseArgsErrors(t *testing.T) {
    	for _, tc := range []struct {
    		args []string
    		why  string
    	}{
    		{[]string{"march.csv"}, "-account is missing"},
    		{[]string{"-account", "", "march.csv"}, "-account is empty"},
    		{[]string{"-account", "rent"}, "there are no files"},
    		{[]string{"-account", "rent", "-limit", "0", "march.csv"}, "-limit is 0"},
    		{[]string{"-account", "rent", "-limit", "-4", "march.csv"}, "-limit is negative"},
    		{[]string{"-account", "rent", "-limit", "five", "march.csv"}, "-limit isn't a number"},
    		{[]string{"-account", "rent", "-verbose", "march.csv"}, "-verbose isn't a flag"},
    		{[]string{"-account"}, "-account has no value"},
    	} {
    		if cfg, err := parseArgs(tc.args); err == nil {
    			t.Errorf("parseArgs(%q) = %+v, nil; want an error because %s", tc.args, cfg, tc.why)
    		}
    	}
    }

    func TestParseArgsHelp(t *testing.T) {
    	_, err := parseArgs([]string{"-h"})
    	if !errors.Is(err, flag.ErrHelp) {
    		t.Errorf("parseArgs([\"-h\"]) error = %v, want flag.ErrHelp (return the error from Parse, or wrap it with %%w)", err)
    	}
    }
---

Ledgerly needs a command-line interface: `ledgerly report -account rent march.csv`. Before you build one, here's how a Go program sees its arguments, and how to parse them in a way you can test.

## os.Args

`os.Args` is a `[]string` holding the command line, already split into words by the shell:

```go
fmt.Printf("%q\n", os.Args)
```

```text
$ ledgerly report -account "petty cash" march.csv
["ledgerly" "report" "-account" "petty cash" "march.csv"]
```

`os.Args[0]` is the program's name (or path). The arguments start at `os.Args[1]`. The shell has already removed the quotes, so `"petty cash"` arrives as one word. For a subcommand such as `report`, look at `os.Args[1]` yourself, then parse the rest.

## The flag package

Picking flags out of `os.Args` by hand gets messy quickly. The `flag` package does it for you:

```go
limit := flag.Int("limit", 20, "number of rows to show")
csv := flag.Bool("csv", false, "print CSV instead of a table")
flag.Parse()
files := flag.Args() // the arguments left after the flags
```

Each definition returns a pointer that `Parse` fills in. The `flag.StringVar`/`IntVar`/`BoolVar` forms write into a variable you already have, such as a struct field. The package accepts `-limit 5`, `-limit=5` and `--limit=5`. `-h` or `-help` prints a generated usage message:

```text
Usage of ledgerly:
  -csv
    	print CSV instead of a table
  -limit int
    	number of rows to show (default 20)
```

Three gotchas:

- **Parsing stops at the first non-flag argument.** In `report -account rent march.csv -csv`, the `-csv` is a file name as far as `flag` is concerned. A lone `--` also ends the flags, so `-- -weird.csv` passes a file whose name starts with a dash.
- **Bool flags need `=` to be false.** `-csv=false` works. `-csv false` sets `-csv` to true and treats `false` as a file name.
- **The global functions exit for you.** A bad value makes `flag.Parse` print the error and usage, then call `os.Exit(2)`.

## FlagSets you can test

That last point is what makes the globals awkward in tests. A test can't survive `os.Exit`, and global flags can only be defined once per process. The fix is a `flag.FlagSet`, a private set of flags that parses any slice you give it:

```go
fs := flag.NewFlagSet("report", flag.ContinueOnError)
fs.SetOutput(io.Discard)
fs.StringVar(&cfg.Account, "account", "", "account to report on")
if err := fs.Parse(args); err != nil {
	return Config{}, err
}
cfg.Files = fs.Args()
```

With `ContinueOnError`, `Parse` returns the error instead of exiting. For `-h` it returns `flag.ErrHelp`. `SetOutput(io.Discard)` stops it printing usage text, which is useful in a function whose caller decides what to print. Now argument parsing is an ordinary function from `[]string` to a `Config` and an error, and a table test can throw any command line at it.

## Your turn

Complete `parseArgs`, which parses the arguments after `ledgerly report` into a `Config`:

- `-account` (string) is **required** and must not be empty.
- `-limit` (int) defaults to 20 and must be **positive**.
- `-csv` (bool) defaults to false.
- The remaining arguments are the files, and there must be **at least one**.

Return an error for anything else, including unknown flags and values that don't parse. For `-h`, return an error that `errors.Is` matches with `flag.ErrHelp` (returning `Parse`'s error does that). The grader tries both flag syntaxes, flags after a file name, `--`, and every kind of mistake. **Run** prints what your function returns for a few command lines.
