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
