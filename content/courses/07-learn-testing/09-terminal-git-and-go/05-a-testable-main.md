---
title: A Testable main
quiz:
  - question: |
      What's wrong with this `main`?

      ```go
      func main() {
          f, err := os.Create("report.csv")
          if err != nil {
              log.Fatal(err)
          }
          defer f.Close()
          if err := writeReport(f); err != nil {
              fmt.Fprintln(os.Stderr, err)
              os.Exit(1)
          }
      }
      ```
    options:
      - text: Nothing
      - text: '`log.Fatal` writes to standard output'
      - text: When `writeReport` fails, `os.Exit` ends the program immediately and the deferred `f.Close()` never runs
        correct: true
      - text: '`os.Exit(1)` should be `os.Exit(0)`'
    explanation: |
      `os.Exit` doesn't run deferred calls. Putting the work in a `run`
      function that returns a status, and calling `os.Exit` only in a
      one-line `main`, lets every defer inside `run` finish first.
  - question: Why does `run` take `stdout` and `stderr` as `io.Writer` parameters instead of using `os.Stdout` and `os.Stderr` directly?
    options:
      - text: Writing to `os.Stdout` is slow
      - text: So tests can pass in buffers and check exactly what went to each stream
        correct: true
      - text: '`os.Stdout` isn''t available in a `package main`'
      - text: Because `fmt.Println` can't write to `os.Stdout`
    explanation: |
      It's the same dependency injection you used for test doubles and for
      readers and writers. A `bytes.Buffer` stands in for the terminal, so
      a test can check the output, the errors and the exit status of one
      whole run of the program.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"io"
    	"os"
    	"strconv"
    	"strings"
    )

    // run is the whole ledgerly program. It reads transactions from stdin,
    // writes results to stdout and errors to stderr, and returns the exit
    // status: 0 for success, 1 for bad input, 2 for a usage mistake.
    func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
    	// ?
    	return 0
    }

    // ---- Ledgerly code ----

    const usage = "usage: ledgerly sum [-account NAME] < transactions.csv"

    // Cents is an amount of money in cents.
    type Cents int64

    // String formats c like "$1,234.56" or "-$0.05".
    func (c Cents) String() string {
    	sign, u := "", uint64(c)
    	if c < 0 {
    		sign, u = "-", -u
    	}
    	s := strconv.FormatUint(u/100, 10)
    	for i := len(s) - 3; i > 0; i -= 3 {
    		s = s[:i] + "," + s[i:]
    	}
    	return fmt.Sprintf("%s$%s.%02d", sign, s, u%100)
    }

    func main() {
    	// A real program ends with just:
    	//   os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
    	// Here we call run with a few command lines instead.
    	input := "rent,-120000\ncash,5000\nrent,-1500\n"
    	for _, args := range [][]string{
    		{"sum"},
    		{"sum", "-account", "rent"},
    		{},
    		{"sum", "-verbose"},
    	} {
    		fmt.Printf("$ ledgerly %s\n", strings.Join(args, " "))
    		code := run(args, strings.NewReader(input), os.Stdout, os.Stdout)
    		fmt.Printf("(exit status %d)\n\n", code)
    	}
    	fmt.Println("$ ledgerly sum  (with a bad line 2)")
    	code := run([]string{"sum"}, strings.NewReader("rent,-120000\ncash,fifty\n"), os.Stdout, os.Stdout)
    	fmt.Printf("(exit status %d)\n", code)
    }
  solution: |
    package main

    import (
    	"bufio"
    	"errors"
    	"flag"
    	"fmt"
    	"io"
    	"os"
    	"strconv"
    	"strings"
    )

    // run is the whole ledgerly program. It reads transactions from stdin,
    // writes results to stdout and errors to stderr, and returns the exit
    // status: 0 for success, 1 for bad input, 2 for a usage mistake.
    func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
    	if len(args) == 0 {
    		fmt.Fprintln(stderr, usage)
    		return 2
    	}
    	if args[0] != "sum" {
    		fmt.Fprintf(stderr, "ledgerly: unknown command %q\n%s\n", args[0], usage)
    		return 2
    	}
    	fs := flag.NewFlagSet("sum", flag.ContinueOnError)
    	fs.SetOutput(stderr)
    	account := fs.String("account", "", "only add up this account")
    	if err := fs.Parse(args[1:]); err != nil {
    		if errors.Is(err, flag.ErrHelp) {
    			return 0
    		}
    		return 2
    	}
    	if fs.NArg() > 0 {
    		fmt.Fprintf(stderr, "ledgerly: unexpected argument %q\n%s\n", fs.Arg(0), usage)
    		return 2
    	}

    	var total Cents
    	sc := bufio.NewScanner(stdin)
    	for line := 1; sc.Scan(); line++ {
    		text := strings.TrimSpace(sc.Text())
    		if text == "" {
    			continue
    		}
    		name, amount, ok := strings.Cut(text, ",")
    		n, err := strconv.ParseInt(amount, 10, 64)
    		if !ok || err != nil {
    			fmt.Fprintf(stderr, "ledgerly: line %d: bad record %q\n", line, text)
    			return 1
    		}
    		if *account == "" || name == *account {
    			total += Cents(n)
    		}
    	}
    	if err := sc.Err(); err != nil {
    		fmt.Fprintf(stderr, "ledgerly: reading input: %v\n", err)
    		return 1
    	}
    	fmt.Fprintln(stdout, total)
    	return 0
    }

    // ---- Ledgerly code ----

    const usage = "usage: ledgerly sum [-account NAME] < transactions.csv"

    // Cents is an amount of money in cents.
    type Cents int64

    // String formats c like "$1,234.56" or "-$0.05".
    func (c Cents) String() string {
    	sign, u := "", uint64(c)
    	if c < 0 {
    		sign, u = "-", -u
    	}
    	s := strconv.FormatUint(u/100, 10)
    	for i := len(s) - 3; i > 0; i -= 3 {
    		s = s[:i] + "," + s[i:]
    	}
    	return fmt.Sprintf("%s$%s.%02d", sign, s, u%100)
    }

    func main() {
    	// A real program ends with just:
    	//   os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
    	// Here we call run with a few command lines instead.
    	input := "rent,-120000\ncash,5000\nrent,-1500\n"
    	for _, args := range [][]string{
    		{"sum"},
    		{"sum", "-account", "rent"},
    		{},
    		{"sum", "-verbose"},
    	} {
    		fmt.Printf("$ ledgerly %s\n", strings.Join(args, " "))
    		code := run(args, strings.NewReader(input), os.Stdout, os.Stdout)
    		fmt.Printf("(exit status %d)\n\n", code)
    	}
    	fmt.Println("$ ledgerly sum  (with a bad line 2)")
    	code := run([]string{"sum"}, strings.NewReader("rent,-120000\ncash,fifty\n"), os.Stdout, os.Stdout)
    	fmt.Printf("(exit status %d)\n", code)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"strings"
    	"testing"
    )

    const march = "rent,-120000\ncash,5000\nrent,-1500\n"

    func runCase(args []string, input string) (code int, stdout, stderr string) {
    	var out, errOut bytes.Buffer
    	code = run(args, strings.NewReader(input), &out, &errOut)
    	return code, out.String(), errOut.String()
    }

    func TestRunSuccess(t *testing.T) {
    	for _, tc := range []struct {
    		args  []string
    		input string
    		want  string
    	}{
    		{[]string{"sum"}, march, "-$1,165.00\n"},
    		{[]string{"sum", "-account", "rent"}, march, "-$1,215.00\n"},
    		{[]string{"sum", "-account=cash"}, march, "$50.00\n"},
    		{[]string{"sum", "-account", "savings"}, march, "$0.00\n"},
    		{[]string{"sum"}, "", "$0.00\n"},
    		{[]string{"sum"}, "\nrent,-120000\n\n  cash,5000  \n", "-$1,150.00\n"},
    	} {
    		code, out, errOut := runCase(tc.args, tc.input)
    		if code != 0 || out != tc.want {
    			t.Errorf("ledgerly %q with input %q: exit status %d, stdout %q; want 0, %q", tc.args, tc.input, code, out, tc.want)
    		}
    		if errOut != "" {
    			t.Errorf("ledgerly %q succeeded but wrote to stderr: %q (successful runs should keep stderr empty)", tc.args, errOut)
    		}
    	}
    }

    func TestRunErrors(t *testing.T) {
    	for _, tc := range []struct {
    		args     []string
    		input    string
    		wantCode int
    		why      string
    	}{
    		{[]string{}, march, 2, "there's no command"},
    		{[]string{"report"}, march, 2, "report isn't a command"},
    		{[]string{"sum", "-verbose"}, march, 2, "-verbose isn't a flag"},
    		{[]string{"sum", "-account"}, march, 2, "-account has no value"},
    		{[]string{"sum", "march.csv"}, march, 2, "sum takes no arguments (input comes from stdin)"},
    		{[]string{"sum"}, "rent,-120000\ncash,fifty\n", 1, "line 2 has a bad amount"},
    		{[]string{"sum"}, "rent,-120000\ncash\n", 1, "line 2 has no amount"},
    		{[]string{"sum", "-account", "rent"}, "rent,-120000\ncash,1,2\n", 1, "line 2 has too many fields, even though it's not a rent line"},
    	} {
    		code, out, errOut := runCase(tc.args, tc.input)
    		if code != tc.wantCode {
    			t.Errorf("ledgerly %q with input %q: exit status %d, want %d because %s", tc.args, tc.input, code, tc.wantCode, tc.why)
    		}
    		if out != "" {
    			t.Errorf("ledgerly %q with input %q failed but wrote %q to stdout; errors belong on stderr, and a failed run shouldn't print a partial total", tc.args, tc.input, out)
    		}
    		if errOut == "" {
    			t.Errorf("ledgerly %q with input %q: stderr is empty; say what went wrong (%s)", tc.args, tc.input, tc.why)
    		}
    	}
    }

    func TestRunBadLineNumber(t *testing.T) {
    	_, _, errOut := runCase([]string{"sum"}, "rent,-120000\ncash,5000\ncash,fifty\n")
    	if !strings.Contains(errOut, "line 3") {
    		t.Errorf("for a bad third line, stderr = %q; it should mention \"line 3\"", errOut)
    	}
    }

    func TestRunHelp(t *testing.T) {
    	code, out, errOut := runCase([]string{"sum", "-h"}, march)
    	if code != 0 {
    		t.Errorf("ledgerly sum -h: exit status %d, want 0 (asking for help isn't an error)", code)
    	}
    	if !strings.Contains(errOut, "-account") || out != "" {
    		t.Errorf("ledgerly sum -h: stdout %q, stderr %q; want the flag usage, mentioning -account, on stderr", out, errOut)
    	}
    }
---

A command-line program's behaviour is more than its output. It's the output on stdout, the messages on stderr, and the **exit status**. Scripts and CI depend on all three. So they deserve tests, and a plain `main` makes that impossible: it takes no arguments, returns nothing, and may call `os.Exit` in the middle of a test run.

## Move everything into run

The standard fix is a `main` that does nothing except wire the real world into a function that does everything:

```go
func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// parse args, do the work, print, and return 0, 1 or 2
}
```

All of the program's inputs and outputs are now parameters, which you know from the test doubles and I/O chapters. A test calls `run` with a slice of arguments, a `strings.Reader` for input and two `bytes.Buffer`s:

```go
var out, errOut bytes.Buffer
code := run([]string{"sum", "-account", "rent"}, strings.NewReader(march), &out, &errOut)
if code != 0 || out.String() != "-$1,215.00\n" {
	t.Errorf("ledgerly sum -account rent = %d, %q; want 0, %q", code, out.String(), "-$1,215.00\n")
}
```

No subprocesses, no temp files, and it runs in microseconds. One more benefit: `os.Exit` skips deferred calls, so keeping it out of `run` means every `defer` inside `run` (closing files, flushing a `bufio.Writer`) runs before the program exits.

## Exit status conventions

Pick statuses that scripts can act on, and document them:

| Status | Meaning | Ledgerly example |
| --- | --- | --- |
| 0 | success | printed the total |
| 1 | the program ran but failed | line 3 of the input isn't a valid record |
| 2 | usage error | unknown command, unknown flag, missing argument |

That matches the `flag` package, which exits with 2 on a bad flag, and `go test`, which exits with 1 when a test fails. Two more habits of good CLI tools: results go to stdout and **everything else** goes to stderr, and a failed run prints **no partial result** on stdout. `ledgerly sum > total.txt` shouldn't leave a wrong total in the file.

## Your turn

Complete `run` for `ledgerly sum`, which adds up transactions read from `stdin`, one `account,cents` record per line:

1. With no arguments, print `usage` to stderr and return 2. If the first argument isn't `sum`, print an error and the usage to stderr and return 2.
2. Parse the rest with a `flag.FlagSet` that has one string flag, `-account`. Send its output to `stderr` with `fs.SetOutput(stderr)`. A parse error returns 2, except `flag.ErrHelp` (from `-h`), which returns 0 because asking for help isn't a mistake. Any argument left after the flags is a usage error, since input comes from stdin.
3. Read `stdin` line by line with a `bufio.Scanner`. Trim spaces and skip blank lines. Split each line at the comma with `strings.Cut` and parse the amount with `strconv.ParseInt`. A malformed line prints `ledgerly: line N: bad record "..."` to stderr and returns 1. Check every line, even ones for other accounts.
4. Add up the amounts (all of them, or only `-account`'s), print the total as a `Cents` followed by a newline to stdout, and return 0.

The grader calls `run` with buffers and checks the exit status and both streams for each case. **Run** shows a few runs, with stdout and stderr both going to the screen.
