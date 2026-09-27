---
title: Benchmark the Importer
difficulty: medium
after: benchmarks
hints:
  - 'Everything before `for b.Loop()` is setup and isn''t timed: build the statement there, once, and tell the benchmark its size with `b.SetBytes(int64(len(data)))`. Call `b.ReportAllocs()` there too.'
  - 'A `bytes.Reader` is used up once `importCSV` has read it to the end. If you create it before the loop, every iteration after the first imports an empty file, which is very fast and completely meaningless. Create a fresh `bytes.NewReader(data)` **inside** the loop.'
  - 'A benchmark is still a test: if `importCSV` returns an error, or fewer than 1,000 transactions, stop with `b.Fatal`/`b.Fatalf` rather than timing garbage.'
exercise:
  starter: |
    package main

    import (
    	"bytes"
    	"encoding/csv"
    	"flag"
    	"fmt"
    	"io"
    	"strconv"
    	"testing"
    )

    // BenchmarkImport measures importCSV on a 1,000-line statement.
    func BenchmarkImport(b *testing.B) {
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    type Txn struct {
    	Account string
    	Amount  Cents
    }

    // importCSV is a variable so the grader can swap in its own importer.
    var importCSV = ImportCSV

    // ImportCSV reads "account,cents" records from r.
    func ImportCSV(r io.Reader) ([]Txn, error) {
    	cr := csv.NewReader(r)
    	cr.FieldsPerRecord = 2
    	cr.ReuseRecord = true
    	var txns []Txn
    	for {
    		rec, err := cr.Read()
    		if err == io.EOF {
    			return txns, nil
    		}
    		if err != nil {
    			return nil, err
    		}
    		n, err := strconv.ParseInt(rec[1], 10, 64)
    		if err != nil {
    			return nil, fmt.Errorf("line %d: %w", len(txns)+1, err)
    		}
    		txns = append(txns, Txn{rec[0], Cents(n)})
    	}
    }

    var statementsMade int

    // makeStatement returns a CSV statement with n transactions.
    func makeStatement(n int) []byte {
    	statementsMade++
    	var buf bytes.Buffer
    	accounts := []string{"rent", "groceries", "cash", "fuel"}
    	for i := range n {
    		fmt.Fprintf(&buf, "%s,%d\n", accounts[i%len(accounts)], (i*7919)%100000-50000)
    	}
    	return buf.Bytes()
    }

    func main() {
    	testing.Init()
    	flag.Set("test.benchtime", "200ms") // keep Run quick
    	r := testing.Benchmark(BenchmarkImport)
    	fmt.Printf("BenchmarkImport  %s  %s\n", r, r.MemString())
    	fmt.Println("statements built:", statementsMade)
    }
  solution: |
    package main

    import (
    	"bytes"
    	"encoding/csv"
    	"flag"
    	"fmt"
    	"io"
    	"strconv"
    	"testing"
    )

    // BenchmarkImport measures importCSV on a 1,000-line statement.
    func BenchmarkImport(b *testing.B) {
    	data := makeStatement(1000)
    	b.SetBytes(int64(len(data)))
    	b.ReportAllocs()
    	for b.Loop() {
    		txns, err := importCSV(bytes.NewReader(data))
    		if err != nil {
    			b.Fatal(err)
    		}
    		if len(txns) != 1000 {
    			b.Fatalf("imported %d transactions, want 1000", len(txns))
    		}
    	}
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    type Txn struct {
    	Account string
    	Amount  Cents
    }

    // importCSV is a variable so the grader can swap in its own importer.
    var importCSV = ImportCSV

    // ImportCSV reads "account,cents" records from r.
    func ImportCSV(r io.Reader) ([]Txn, error) {
    	cr := csv.NewReader(r)
    	cr.FieldsPerRecord = 2
    	cr.ReuseRecord = true
    	var txns []Txn
    	for {
    		rec, err := cr.Read()
    		if err == io.EOF {
    			return txns, nil
    		}
    		if err != nil {
    			return nil, err
    		}
    		n, err := strconv.ParseInt(rec[1], 10, 64)
    		if err != nil {
    			return nil, fmt.Errorf("line %d: %w", len(txns)+1, err)
    		}
    		txns = append(txns, Txn{rec[0], Cents(n)})
    	}
    }

    var statementsMade int

    // makeStatement returns a CSV statement with n transactions.
    func makeStatement(n int) []byte {
    	statementsMade++
    	var buf bytes.Buffer
    	accounts := []string{"rent", "groceries", "cash", "fuel"}
    	for i := range n {
    		fmt.Fprintf(&buf, "%s,%d\n", accounts[i%len(accounts)], (i*7919)%100000-50000)
    	}
    	return buf.Bytes()
    }

    func main() {
    	testing.Init()
    	flag.Set("test.benchtime", "200ms") // keep Run quick
    	r := testing.Benchmark(BenchmarkImport)
    	fmt.Printf("BenchmarkImport  %s  %s\n", r, r.MemString())
    	fmt.Println("statements built:", statementsMade)
    }
  tests: |
    package main

    import (
    	"errors"
    	"flag"
    	"go/ast"
    	"go/parser"
    	"go/token"
    	"io"
    	"testing"
    )

    // bench runs BenchmarkImport with importer swapped in for importCSV.
    func bench(importer func(io.Reader) ([]Txn, error)) testing.BenchmarkResult {
    	flag.Set("test.benchtime", "100ms")
    	importCSV = importer
    	defer func() { importCSV = ImportCSV }()
    	statementsMade = 0
    	return testing.Benchmark(BenchmarkImport)
    }

    func TestBenchmarkMeasuresRealWork(t *testing.T) {
    	calls, short := 0, 0
    	r := bench(func(rd io.Reader) ([]Txn, error) {
    		calls++
    		txns, err := ImportCSV(rd)
    		if len(txns) != 1000 {
    			short++
    		}
    		return txns, err
    	})
    	if r.N == 0 {
    		t.Fatal("BenchmarkImport failed with the real importer (did it call b.Fatal?)")
    	}
    	if calls < r.N {
    		t.Fatalf("BenchmarkImport reported %d iterations but called importCSV %d times: call importCSV once per b.Loop() iteration", r.N, calls)
    	}
    	if short > 0 {
    		t.Errorf("%d of %d importCSV calls got fewer than 1,000 lines: create a new bytes.Reader inside the loop, a used one is empty", short, calls)
    	}
    	if statementsMade != 1 {
    		t.Errorf("makeStatement was called %d times, want exactly once: build the input before the b.Loop() loop (and use b.Loop, not b.N)", statementsMade)
    	}
    	if want := int64(len(makeStatement(1000))); r.Bytes != want {
    		t.Errorf("benchmark set %d bytes per op, want %d (call b.SetBytes with the statement's length)", r.Bytes, want)
    	}
    }

    func TestBenchmarkFailsOnErrors(t *testing.T) {
    	r := bench(func(io.Reader) ([]Txn, error) { return nil, errors.New("line 1: broken") })
    	if r.N != 0 {
    		t.Error("with an importer that always returns an error, BenchmarkImport still ran: stop with b.Fatal when importCSV fails")
    	}
    	r = bench(func(rd io.Reader) ([]Txn, error) {
    		txns, err := ImportCSV(rd)
    		return txns[:min(len(txns), 10)], err
    	})
    	if r.N != 0 {
    		t.Error("with an importer that returns only 10 of the 1,000 transactions, BenchmarkImport still ran: check the count and b.Fatalf if it's wrong")
    	}
    }

    func TestBenchmarkSource(t *testing.T) {
    	f, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, 0)
    	if err != nil {
    		t.Fatalf("parsing main.go: %v", err)
    	}
    	calls := map[string]bool{}
    	for _, d := range f.Decls {
    		fn, ok := d.(*ast.FuncDecl)
    		if !ok || fn.Name.Name != "BenchmarkImport" {
    			continue
    		}
    		ast.Inspect(fn, func(n ast.Node) bool {
    			if sel, ok := n.(*ast.SelectorExpr); ok {
    				calls[sel.Sel.Name] = true
    			}
    			return true
    		})
    	}
    	if !calls["Loop"] {
    		t.Error("BenchmarkImport should loop with for b.Loop()")
    	}
    	if !calls["ReportAllocs"] {
    		t.Error("BenchmarkImport should call b.ReportAllocs() so allocations show up without -benchmem")
    	}
    }
---

Someone says Ledgerly's CSV importer is slow. Before anyone optimizes it,
you need a benchmark you can trust. Benchmarks are easy to write and just as
easy to get subtly wrong, so that they measure nothing, or the wrong thing.

Write `BenchmarkImport(b)`. It must:

1. Build the input **once**, before the timed loop, with `makeStatement(1000)`
   (a 1,000-line CSV statement as a `[]byte`).
2. Call `b.SetBytes` with the input's length, so the result shows MB/s, and
   `b.ReportAllocs()`, so it shows allocations.
3. Loop with `for b.Loop()`, calling **`importCSV`** (the variable, so the
   grader can swap the importer) on a reader over the data each time.
4. Stop with `b.Fatal`/`b.Fatalf` if `importCSV` returns an error or doesn't
   return exactly 1,000 transactions.

## Example

```text
BenchmarkImport      3412      63804 ns/op   196.41 MB/s     79055 B/op   1021 allocs/op
statements built: 1
```

(Your numbers will differ.) If you ever see something like `36000 MB/s` and
2 allocs/op, be suspicious: that's what importing an already-empty reader
looks like. Normally you'd run it with
`go test -bench Import`. **Run** calls `testing.Benchmark` from `main` with a
short benchmark time instead.

## How you're graded

The grader runs your benchmark with instrumented importers and checks that
`makeStatement` ran once, that `importCSV` ran at least once per iteration
with a full 1,000-line input every time, that `SetBytes` matches, and that the
benchmark fails when the importer returns an error or too few transactions.

## Constraints

- Setup inside the loop gets timed. Setup outside the loop with `b.N` style
  benchmarks runs once per ramp-up round, not once overall: use `b.Loop()`.
