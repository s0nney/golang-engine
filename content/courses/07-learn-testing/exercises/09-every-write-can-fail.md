---
title: Every Write Can Fail
difficulty: medium
after: testing-io-and-files
hints:
  - 'Start with the happy path: export three transactions into a `bytes.Buffer` and compare `buf.String()` with the exact CSV you expect, header included. That alone catches bugs in the format and a forgotten `Flush`.'
  - 'For the error paths, write a tiny writer type with a `room int` field: its `Write` copies at most `room` bytes into a buffer, lowers `room`, and returns `ErrDiskFull` (with the number of bytes it did accept) whenever it couldn''t take everything.'
  - 'Now loop `for room := range len(want)`: a disk with room for 0, 1, 2, ... bytes fills up at every possible point of the output, so every write call the exporter makes gets its turn to fail. Each time, `export` must return an error for which `errors.Is(err, ErrDiskFull)` is true. Do the same for an export of no transactions, where the header is the only write.'
exercise:
  starter: |
    package main

    import (
    	"bufio"
    	"errors"
    	"fmt"
    	"io"
    	"time"
    )

    // checkExport tests an implementation of the CSV export contract. It
    // returns nil if export is correct, or an error describing a problem.
    func checkExport(export func(io.Writer, []Txn) error) error {
    	return nil
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    type Txn struct {
    	Date    time.Time
    	Account string
    	Amount  Cents
    }

    var ErrDiskFull = errors.New("disk full")

    // Export writes txns to w as CSV: a header line "date,account,cents",
    // then one line per transaction like "2026-03-01,rent,-120000", each
    // ending in "\n". It returns the first error from writing, or nil.
    func Export(w io.Writer, txns []Txn) error {
    	bw := bufio.NewWriter(w)
    	fmt.Fprintln(bw, "date,account,cents")
    	for _, t := range txns {
    		fmt.Fprintf(bw, "%s,%s,%d\n", t.Date.Format(time.DateOnly), t.Account, t.Amount)
    	}
    	return bw.Flush() // a bufio.Writer remembers its first error
    }

    // sloppyExport forgets that writes can fail.
    func sloppyExport(w io.Writer, txns []Txn) error {
    	fmt.Fprintln(w, "date,account,cents")
    	for _, t := range txns {
    		fmt.Fprintf(w, "%s,%s,%d\n", t.Date.Format(time.DateOnly), t.Account, t.Amount)
    	}
    	return nil
    }

    func main() {
    	fmt.Println("Export:      ", checkExport(Export))
    	fmt.Println("sloppyExport:", checkExport(sloppyExport))
    }
  solution: |
    package main

    import (
    	"bufio"
    	"bytes"
    	"errors"
    	"fmt"
    	"io"
    	"time"
    )

    // diskWriter is a fake disk with room for a limited number of bytes.
    type diskWriter struct {
    	room int
    	buf  bytes.Buffer
    }

    func (d *diskWriter) Write(p []byte) (int, error) {
    	n := min(len(p), d.room)
    	d.buf.Write(p[:n])
    	d.room -= n
    	if n < len(p) {
    		return n, ErrDiskFull
    	}
    	return n, nil
    }

    // checkExport tests an implementation of the CSV export contract. It
    // returns nil if export is correct, or an error describing a problem.
    func checkExport(export func(io.Writer, []Txn) error) error {
    	txns := []Txn{
    		{time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), "rent", -120000},
    		{time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), "cash", 5000},
    		{time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC), "groceries", -8734},
    	}
    	const want = "date,account,cents\n" +
    		"2026-03-01,rent,-120000\n" +
    		"2026-03-02,cash,5000\n" +
    		"2026-03-15,groceries,-8734\n"

    	var buf bytes.Buffer
    	if err := export(&buf, txns); err != nil {
    		return fmt.Errorf("export to a bytes.Buffer returned %v, want nil", err)
    	}
    	if buf.String() != want {
    		return fmt.Errorf("export wrote:\n%s\nwant:\n%s", buf.String(), want)
    	}

    	buf.Reset()
    	if err := export(&buf, nil); err != nil || buf.String() != "date,account,cents\n" {
    		return fmt.Errorf("export of no transactions = %q, %v; want just the header", buf.String(), err)
    	}

    	for room := range len(want) {
    		d := &diskWriter{room: room}
    		if err := export(d, txns); !errors.Is(err, ErrDiskFull) {
    			return fmt.Errorf("export to a disk with room for %d bytes (it wrote %q) returned %v, want ErrDiskFull", room, d.buf.String(), err)
    		}
    	}
    	for room := range len("date,account,cents\n") {
    		d := &diskWriter{room: room}
    		if err := export(d, nil); !errors.Is(err, ErrDiskFull) {
    			return fmt.Errorf("export of no transactions to a disk with room for %d bytes returned %v, want ErrDiskFull", room, err)
    		}
    	}
    	return nil
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    type Txn struct {
    	Date    time.Time
    	Account string
    	Amount  Cents
    }

    var ErrDiskFull = errors.New("disk full")

    // Export writes txns to w as CSV: a header line "date,account,cents",
    // then one line per transaction like "2026-03-01,rent,-120000", each
    // ending in "\n". It returns the first error from writing, or nil.
    func Export(w io.Writer, txns []Txn) error {
    	bw := bufio.NewWriter(w)
    	fmt.Fprintln(bw, "date,account,cents")
    	for _, t := range txns {
    		fmt.Fprintf(bw, "%s,%s,%d\n", t.Date.Format(time.DateOnly), t.Account, t.Amount)
    	}
    	return bw.Flush() // a bufio.Writer remembers its first error
    }

    // sloppyExport forgets that writes can fail.
    func sloppyExport(w io.Writer, txns []Txn) error {
    	fmt.Fprintln(w, "date,account,cents")
    	for _, t := range txns {
    		fmt.Fprintf(w, "%s,%s,%d\n", t.Date.Format(time.DateOnly), t.Account, t.Amount)
    	}
    	return nil
    }

    func main() {
    	fmt.Println("Export:      ", checkExport(Export))
    	fmt.Println("sloppyExport:", checkExport(sloppyExport))
    }
  tests: |
    package main

    import (
    	"bufio"
    	"errors"
    	"fmt"
    	"io"
    	"strings"
    	"testing"
    	"time"
    )

    func line(t Txn) string {
    	return fmt.Sprintf("%s,%s,%d\n", t.Date.Format(time.DateOnly), t.Account, t.Amount)
    }

    func run(export func(io.Writer, []Txn) error) (err error, panicked string) {
    	defer func() {
    		if p := recover(); p != nil {
    			panicked = fmt.Sprint(p)
    		}
    	}()
    	return checkExport(export), ""
    }

    // oneWrite builds the whole file in memory and writes it once.
    func oneWrite(w io.Writer, txns []Txn) error {
    	var b strings.Builder
    	b.WriteString("date,account,cents\n")
    	for _, t := range txns {
    		b.WriteString(line(t))
    	}
    	_, err := io.WriteString(w, b.String())
    	if err != nil {
    		return fmt.Errorf("export: %w", err)
    	}
    	return nil
    }

    // perLine writes line by line, checking every error.
    func perLine(w io.Writer, txns []Txn) error {
    	if _, err := io.WriteString(w, "date,account,cents\n"); err != nil {
    		return err
    	}
    	for _, t := range txns {
    		if _, err := io.WriteString(w, line(t)); err != nil {
    			return err
    		}
    	}
    	return nil
    }

    func TestCheckAcceptsCorrectExports(t *testing.T) {
    	for name, export := range map[string]func(io.Writer, []Txn) error{
    		"Export (buffered)":                  Export,
    		"an export that writes once":         oneWrite,
    		"an export that writes line by line": perLine,
    	} {
    		if err, p := run(export); p != "" {
    			t.Errorf("checkExport panicked on %s: %s", name, p)
    		} else if err != nil {
    			t.Errorf("checkExport rejected %s, which is correct: %v", name, err)
    		}
    	}
    }

    func TestCheckCatchesBugs(t *testing.T) {
    	bugs := []struct {
    		bug    string
    		export func(io.Writer, []Txn) error
    	}{
    		{"ignores every write error", sloppyExport},
    		{"ignores the error from writing the header", func(w io.Writer, txns []Txn) error {
    			io.WriteString(w, "date,account,cents\n")
    			for _, t := range txns {
    				if _, err := io.WriteString(w, line(t)); err != nil {
    					return err
    				}
    			}
    			return nil
    		}},
    		{"ignores the error from writing the last line", func(w io.Writer, txns []Txn) error {
    			if _, err := io.WriteString(w, "date,account,cents\n"); err != nil {
    				return err
    			}
    			for i, t := range txns {
    				if _, err := io.WriteString(w, line(t)); err != nil && i < len(txns)-1 {
    					return err
    				}
    			}
    			return nil
    		}},
    		{"forgets to Flush its bufio.Writer", func(w io.Writer, txns []Txn) error {
    			bw := bufio.NewWriter(w)
    			fmt.Fprintln(bw, "date,account,cents")
    			for _, t := range txns {
    				bw.WriteString(line(t))
    			}
    			return nil
    		}},
    		{"flushes but ignores Flush's error", func(w io.Writer, txns []Txn) error {
    			bw := bufio.NewWriter(w)
    			fmt.Fprintln(bw, "date,account,cents")
    			for _, t := range txns {
    				bw.WriteString(line(t))
    			}
    			bw.Flush()
    			return nil
    		}},
    		{"replaces the write error with its own, unwrapped one", func(w io.Writer, txns []Txn) error {
    			if err := perLine(w, txns); err != nil {
    				return errors.New("export failed")
    			}
    			return nil
    		}},
    		{"writes amounts as dollars (-1200.00)", func(w io.Writer, txns []Txn) error {
    			io.WriteString(w, "date,account,cents\n")
    			for _, t := range txns {
    				s := fmt.Sprintf("%s,%s,%.2f\n", t.Date.Format(time.DateOnly), t.Account, float64(t.Amount)/100)
    				if _, err := io.WriteString(w, s); err != nil {
    					return err
    				}
    			}
    			return nil
    		}},
    		{"leaves out the header", func(w io.Writer, txns []Txn) error {
    			for _, t := range txns {
    				if _, err := io.WriteString(w, line(t)); err != nil {
    					return err
    				}
    			}
    			return nil
    		}},
    		{"skips the header when there are no transactions", func(w io.Writer, txns []Txn) error {
    			if len(txns) == 0 {
    				return nil
    			}
    			return perLine(w, txns)
    		}},
    	}
    	for _, b := range bugs {
    		if err, p := run(b.export); p != "" {
    			t.Errorf("checkExport panicked on an export that %s: %s", b.bug, p)
    		} else if err == nil {
    			t.Errorf("checkExport returned nil for an export that %s", b.bug)
    		}
    	}
    }
---

Ledgerly's CSV export has a bug report: "the disk filled up during a backup
and Ledgerly said it succeeded". Writes fail in real life (full disks, closed
pipes, dropped network connections), and an exporter that ignores those
errors silently produces truncated files.

Write `checkExport(export)`, a test for **any** implementation of the export
contract. It returns `nil` if `export` behaves, or an error saying what's
wrong.

## The contract

`export(w, txns)` writes CSV to `w`:

```text
date,account,cents
2026-03-01,rent,-120000
2026-03-02,cash,5000
```

A header line, then one line per transaction (date as `YYYY-MM-DD`, account,
amount in cents), each ending in `\n`. With no transactions it writes just the
header. It returns `nil` on success, or an error wrapping the **first error**
the writer returned. Implementations may write the output in one call, line
by line, or through a `bufio.Writer`; your check mustn't care which.

## What to check

1. **The output.** Export a few transactions (you choose them) into a
   `bytes.Buffer` and compare with the exact text you expect. Check the empty
   case too.
2. **Every write error.** Use a fake writer that fails partway through. The
   grader defines `ErrDiskFull` for your fake to return; `export` must return
   an error for which `errors.Is(err, ErrDiskFull)` holds, **wherever** the
   disk fills up, and whether or not there are any transactions.

## How you're graded

Your `checkExport` must return `nil` for three correct exporters (buffered,
single write, line by line) and an error for **nine** broken ones, without
panicking.

**Run** checks the real `Export` and `sloppyExport`, which ignores every
error.

## Constraints

- A fake that fails on "the Nth call to `Write`" depends on how the exporter
  splits up its output. A fake that runs out of **bytes** doesn't.
