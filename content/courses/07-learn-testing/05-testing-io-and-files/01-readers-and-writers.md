---
title: Readers and Writers in Tests
quiz:
  - question: Which signature makes a CSV importer easiest to test?
    options:
      - text: '`func ImportCSV(path string) ([]Transaction, error)`'
      - text: '`func ImportCSV(f *os.File) ([]Transaction, error)`'
      - text: '`func ImportCSV(r io.Reader) ([]Transaction, error)`'
        correct: true
      - text: '`func ImportCSV() ([]Transaction, error)` reading from stdin'
    explanation: |
      An `io.Reader` can be a file, a network response, stdin, or a
      `strings.Reader` built from a string literal in the test. The path
      and `*os.File` versions force every test to create a real file.
  - question: |
      What does this test prove that a `strings.Reader` test wouldn't?

      ```go
      txns, err := ImportCSV(iotest.OneByteReader(strings.NewReader(input)))
      ```
    options:
      - text: That the importer is fast
      - text: That the importer copes with `Read` returning fewer bytes than asked for, as network connections often do
        correct: true
      - text: That the importer closes the reader
      - text: That the input is valid UTF-8
    explanation: |
      `Read` may return any number of bytes up to `len(p)`. Code that
      assumes one `Read` fills the buffer works on small strings and
      breaks on real connections. `OneByteReader` returns one byte per
      call and exposes that bug.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"io"
    	"strconv"
    	"strings"
    	"testing/iotest"
    )

    // SumCents adds up amounts in cents, one per line, ignoring blank lines.
    // It fails if a line isn't a whole number, or if reading r fails.
    func SumCents(r io.Reader) (Cents, error) {
    	buf := make([]byte, 4096)
    	n, err := r.Read(buf)
    	if err != nil {
    		return 0, err
    	}
    	var total Cents
    	for _, line := range strings.Split(string(buf[:n]), "\n") {
    		if line == "" {
    			continue
    		}
    		c, err := strconv.ParseInt(line, 10, 64)
    		if err != nil {
    			return 0, fmt.Errorf("bad amount %q", line)
    		}
    		total += Cents(c)
    	}
    	return total, nil
    }

    // ---- Ledgerly code ----

    // Cents is an amount of money in cents.
    type Cents int64

    func main() {
    	const input = "-120000\n5000\n\n-1500\n"
    	for _, tc := range []struct {
    		name string
    		r    io.Reader
    	}{
    		{"strings.Reader", strings.NewReader(input)},
    		{"OneByteReader", iotest.OneByteReader(strings.NewReader(input))},
    		{"DataErrReader", iotest.DataErrReader(strings.NewReader(input))},
    		{"TimeoutReader", iotest.TimeoutReader(strings.NewReader(input))},
    		{"ErrReader", iotest.ErrReader(errors.New("disk on fire"))},
    	} {
    		total, err := SumCents(tc.r)
    		fmt.Printf("%-15s total=%d err=%v\n", tc.name, total, err)
    	}
    }
  solution: |
    package main

    import (
    	"bufio"
    	"errors"
    	"fmt"
    	"io"
    	"strconv"
    	"strings"
    	"testing/iotest"
    )

    // SumCents adds up amounts in cents, one per line, ignoring blank lines.
    // It fails if a line isn't a whole number, or if reading r fails.
    func SumCents(r io.Reader) (Cents, error) {
    	var total Cents
    	sc := bufio.NewScanner(r)
    	for line := 1; sc.Scan(); line++ {
    		text := strings.TrimSpace(sc.Text())
    		if text == "" {
    			continue
    		}
    		c, err := strconv.ParseInt(text, 10, 64)
    		if err != nil {
    			return 0, fmt.Errorf("line %d: bad amount %q", line, text)
    		}
    		total += Cents(c)
    	}
    	if err := sc.Err(); err != nil {
    		return 0, fmt.Errorf("reading amounts: %w", err)
    	}
    	return total, nil
    }

    // ---- Ledgerly code ----

    // Cents is an amount of money in cents.
    type Cents int64

    func main() {
    	const input = "-120000\n5000\n\n-1500\n"
    	for _, tc := range []struct {
    		name string
    		r    io.Reader
    	}{
    		{"strings.Reader", strings.NewReader(input)},
    		{"OneByteReader", iotest.OneByteReader(strings.NewReader(input))},
    		{"DataErrReader", iotest.DataErrReader(strings.NewReader(input))},
    		{"TimeoutReader", iotest.TimeoutReader(strings.NewReader(input))},
    		{"ErrReader", iotest.ErrReader(errors.New("disk on fire"))},
    	} {
    		total, err := SumCents(tc.r)
    		fmt.Printf("%-15s total=%d err=%v\n", tc.name, total, err)
    	}
    }
  tests: |
    package main

    import (
    	"errors"
    	"io"
    	"strings"
    	"testing"
    	"testing/iotest"
    )

    const input = "-120000\n5000\n\n-1500\n"

    func TestSumCentsWellBehaved(t *testing.T) {
    	for _, tc := range []struct {
    		in   string
    		want Cents
    	}{
    		{input, -116500},
    		{"", 0},
    		{"42", 42},
    		{"7\r\n8\r\n", 15},
    	} {
    		got, err := SumCents(strings.NewReader(tc.in))
    		if err != nil || got != tc.want {
    			t.Errorf("SumCents(%q) = %d, %v; want %d, nil", tc.in, got, err, tc.want)
    		}
    	}
    }

    func TestSumCentsLargeInput(t *testing.T) {
    	in := strings.Repeat("100\n", 5000) // 20,000 bytes
    	got, err := SumCents(strings.NewReader(in))
    	if err != nil || got != 500000 {
    		t.Errorf("SumCents(5000 lines of \"100\") = %d, %v; want 500000, nil (does one Read get everything?)", got, err)
    	}
    }

    func TestSumCentsAwkwardReaders(t *testing.T) {
    	for _, tc := range []struct {
    		name string
    		wrap func(io.Reader) io.Reader
    		hint string
    	}{
    		{"iotest.OneByteReader", iotest.OneByteReader, "a single Read may return just one byte"},
    		{"iotest.HalfReader", iotest.HalfReader, "a single Read may return only part of the data"},
    		{"iotest.DataErrReader", iotest.DataErrReader, "the last Read can return data and io.EOF together"},
    	} {
    		got, err := SumCents(tc.wrap(strings.NewReader(input)))
    		if err != nil || got != -116500 {
    			t.Errorf("SumCents(%s(%q)) = %d, %v; want -116500, nil (%s)", tc.name, input, got, err, tc.hint)
    		}
    	}
    }

    func TestSumCentsReadErrors(t *testing.T) {
    	boom := errors.New("disk on fire")
    	_, err := SumCents(iotest.ErrReader(boom))
    	if !errors.Is(err, boom) {
    		t.Errorf("SumCents(iotest.ErrReader(boom)) error = %v; want an error wrapping boom (use %%w)", err)
    	} else if err.Error() == boom.Error() {
    		t.Errorf("SumCents(iotest.ErrReader(boom)) error = %q; wrap it with some context, e.g. fmt.Errorf(\"reading amounts: %%w\", err)", err)
    	}

    	got, err := SumCents(iotest.TimeoutReader(strings.NewReader(input)))
    	if !errors.Is(err, iotest.ErrTimeout) {
    		t.Errorf("SumCents(iotest.TimeoutReader(...)) = %d, %v; the second Read fails, so want an error wrapping iotest.ErrTimeout, not a partial total", got, err)
    	}
    }

    func TestSumCentsBadLine(t *testing.T) {
    	_, err := SumCents(strings.NewReader("5\n12.50\n"))
    	if err == nil || !strings.Contains(err.Error(), "line 2") {
    		t.Errorf("SumCents(%q) error = %v; want an error that mentions \"line 2\"", "5\n12.50\n", err)
    	}
    }
---

Ledgerly's CSV import and statement printing do I/O. The previous chapter's advice applies directly: don't let the function open files or write to the terminal itself. Take the most general interface that does the job, and in Go that's nearly always `io.Reader` or `io.Writer`.

## Accept an io.Reader

```go
// ImportCSV reads "date,account,amount" rows after a header line.
func ImportCSV(r io.Reader) ([]Transaction, error)
```

Production passes an `*os.File` or an HTTP response body. Tests pass a string:

```go
func TestImportCSV(t *testing.T) {
	in := strings.NewReader("date,account,amount\n" +
		"2026-03-01,rent,-1200.00\n" +
		"2026-03-02,salary,3500\n")

	txns, err := ImportCSV(in)
	if err != nil {
		t.Fatalf("ImportCSV: %v", err)
	}
	if len(txns) != 2 {
		t.Fatalf("got %d transactions, want 2", len(txns))
	}
	if txns[0].Amount != -120000 {
		t.Errorf("txns[0].Amount = %v, want -$1200.00", txns[0].Amount)
	}
}
```

The input sits right there in the test, where a reader can see it. No files, no cleanup, no paths. `bytes.NewReader` does the same for a `[]byte`.

## Accept an io.Writer

The mirror image: output goes to an `io.Writer`, and the test hands in a `bytes.Buffer` (or `strings.Builder`):

```go
func WriteStatement(w io.Writer, txns []Transaction) error

func TestWriteStatement(t *testing.T) {
	var buf bytes.Buffer
	err := WriteStatement(&buf, []Transaction{{Date: "2026-03-01", Account: "rent", Amount: -120000}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "-$1200.00") {
		t.Errorf("statement = %q, want it to contain -$1200.00", buf.String())
	}
}
```

`main` calls `WriteStatement(os.Stdout, txns)`. Pass `&buf`, not `buf`: `Write` has a pointer receiver on `bytes.Buffer`.

## Misbehaving readers: testing/iotest

Real readers are messier than `strings.Reader`. The `testing/iotest` package wraps a reader to misbehave on purpose:

- `iotest.OneByteReader(r)` returns one byte per `Read`. It catches code that assumes a single `Read` gets everything.
- `iotest.HalfReader(r)` returns half of what was asked for.
- `iotest.DataErrReader(r)` returns the final data *together with* `io.EOF`, instead of in a separate call. It catches code that throws away data when `err != nil`.
- `iotest.ErrReader(err)` fails every `Read` with `err`. It's the easiest way to test your error path.
- `iotest.TimeoutReader(r)` fails the second `Read` with `iotest.ErrTimeout`.

```go
func TestImportCSVReadError(t *testing.T) {
	boom := errors.New("disk on fire")
	_, err := ImportCSV(iotest.ErrReader(boom))
	if !errors.Is(err, boom) {
		t.Errorf("ImportCSV error = %v, want it to wrap %v", err, boom)
	}
}
```

This checks two things at once: the importer doesn't swallow read errors, and it wraps them with `%w` so callers can inspect them.

## Misbehaving writers

There's no ready-made failing writer in `iotest` (only `TruncateWriter`, which silently drops bytes), so write one. It's a tiny stub:

```go
// failWriter accepts n writes, then fails every write after that.
type failWriter struct{ n int }

func (w *failWriter) Write(p []byte) (int, error) {
	if w.n == 0 {
		return 0, errors.New("disk full")
	}
	w.n--
	return len(p), nil
}

func TestWriteStatementDiskFull(t *testing.T) {
	err := WriteStatement(&failWriter{n: 1}, twoTransactions)
	if err == nil {
		t.Fatal("WriteStatement succeeded on a full disk, want an error")
	}
}
```

Many functions forget to check `Fprintf`'s error. That doesn't matter for `os.Stdout` in a toy, but a statement written to a full disk should fail loudly, not produce half a file.

## Readers you can inspect

Sometimes you want to see what code *did* with a reader or writer. `iotest.NewReadLogger(prefix, r)` and `iotest.NewWriteLogger(prefix, w)` log every call through the `log` package, which is handy when debugging a protocol. And `io.MultiWriter(&buf, os.Stdout)` copies output to a buffer while still showing it, if you want to watch a test run.

## Your turn: survive awkward readers

`SumCents` in the editor adds up one amount (in cents) per line. It passes a test with `strings.NewReader`, and it's still badly broken. It calls `Read` once and assumes that returns everything. It treats an `err` that arrives *with* data as a reason to drop the data. It also returns read errors with no context.

Rewrite it so it works with **any** `io.Reader`:

- Read line by line with a `bufio.Scanner`, which loops over `Read` for you and handles data that arrives together with `io.EOF`. (Reading everything with `io.ReadAll` first also works.)
- Trim spaces from each line (this also removes the `\r` of Windows line endings) and skip blank lines.
- For a line that isn't a whole number, return an error that mentions its line number, like `line 2: bad amount "12.50"`.
- If reading fails, return that error wrapped with context using `%w`, and no partial total. Check `sc.Err()` after the loop.

The grader feeds your function through `iotest.OneByteReader`, `HalfReader`, `DataErrReader`, `TimeoutReader` and `ErrReader`, plus an input bigger than the starter's buffer. **Run** shows what each reader does to the current code.
