---
title: 'Your Turn: Zero-Allocation Formatting'
quiz:
  - question: |
      Why does this benchmark report allocations even when `AppendCents`
      itself never allocates?

      ```go
      for b.Loop() {
          buf := AppendCents(nil, c)
          _ = buf
      }
      ```
    options:
      - text: '`b.Loop` allocates on every iteration'
      - text: Appending to a `nil` slice has to allocate a new backing array every time
        correct: true
      - text: '`_ = buf` copies the slice'
      - text: It doesn't; the benchmark reports 0 allocs/op
    explanation: |
      `AppendCents` avoids allocating only when `dst` has spare capacity.
      Create the buffer once, before the loop, and reuse it with
      `buf = AppendCents(buf[:0], c)`.
exercise:
  starter: |
    package main

    import (
    	"flag"
    	"fmt"
    	"strconv"
    	"testing"
    )

    // AppendCents appends c, formatted exactly like c.String(), to dst and
    // returns the extended slice. It must not allocate when dst has enough
    // spare capacity.
    func AppendCents(dst []byte, c Cents) []byte {
    	// Correct, but String allocates several times on every call.
    	return append(dst, c.String()...)
    }

    // BenchmarkAppendCents measures AppendCents formatting Cents(123456789)
    // into a reused buffer.
    func BenchmarkAppendCents(b *testing.B) {
    	// ?
    }

    // ---- Ledgerly code ----

    // Cents is an amount of money in cents. $12.34 is Cents(1234).
    type Cents int64

    // String formats c like "$1,234.56" or "-$0.05".
    func (c Cents) String() string {
    	sign := ""
    	n := int64(c)
    	if n < 0 {
    		sign = "-"
    		n = -n
    	}
    	return fmt.Sprintf("%s$%s.%02d", sign, groupThousands(n/100), n%100)
    }

    // groupThousands formats n with a comma between each group of three digits.
    func groupThousands(n int64) string {
    	s := strconv.FormatInt(n, 10)
    	for i := len(s) - 3; i > 0; i -= 3 {
    		s = s[:i] + "," + s[i:]
    	}
    	return s
    }

    func main() {
    	var buf []byte
    	for _, c := range []Cents{5, -1234, 100000, 123456789} {
    		buf = AppendCents(buf[:0], c)
    		fmt.Printf("AppendCents(%d) = %s\n", int64(c), buf)
    	}

    	buf = make([]byte, 0, 64)
    	allocs := testing.AllocsPerRun(100, func() {
    		buf = AppendCents(buf[:0], -123456789)
    	})
    	fmt.Println("allocations per AppendCents call:", allocs)

    	testing.Init()
    	flag.Set("test.benchtime", "200ms") // keep Run quick
    	r := testing.Benchmark(BenchmarkAppendCents)
    	fmt.Println("BenchmarkAppendCents:", r.String(), r.MemString())
    }
  solution: |
    package main

    import (
    	"flag"
    	"fmt"
    	"strconv"
    	"testing"
    )

    // AppendCents appends c, formatted exactly like c.String(), to dst and
    // returns the extended slice. It must not allocate when dst has enough
    // spare capacity.
    func AppendCents(dst []byte, c Cents) []byte {
    	n := int64(c)
    	if n < 0 {
    		dst = append(dst, '-')
    		n = -n
    	}
    	dst = append(dst, '$')

    	// Format the dollars into dst, then insert commas by working backwards
    	// through a small stack array.
    	var digits [20]byte
    	d := strconv.AppendInt(digits[:0], n/100, 10)
    	for i, ch := range d {
    		if i > 0 && (len(d)-i)%3 == 0 {
    			dst = append(dst, ',')
    		}
    		dst = append(dst, ch)
    	}
    	cents := n % 100
    	return append(dst, '.', byte('0'+cents/10), byte('0'+cents%10))
    }

    // BenchmarkAppendCents measures AppendCents formatting Cents(123456789)
    // into a reused buffer.
    func BenchmarkAppendCents(b *testing.B) {
    	buf := make([]byte, 0, 64)
    	c := Cents(123456789)
    	for b.Loop() {
    		buf = AppendCents(buf[:0], c)
    	}
    }

    // ---- Ledgerly code ----

    // Cents is an amount of money in cents. $12.34 is Cents(1234).
    type Cents int64

    // String formats c like "$1,234.56" or "-$0.05".
    func (c Cents) String() string {
    	sign := ""
    	n := int64(c)
    	if n < 0 {
    		sign = "-"
    		n = -n
    	}
    	return fmt.Sprintf("%s$%s.%02d", sign, groupThousands(n/100), n%100)
    }

    // groupThousands formats n with a comma between each group of three digits.
    func groupThousands(n int64) string {
    	s := strconv.FormatInt(n, 10)
    	for i := len(s) - 3; i > 0; i -= 3 {
    		s = s[:i] + "," + s[i:]
    	}
    	return s
    }

    func main() {
    	var buf []byte
    	for _, c := range []Cents{5, -1234, 100000, 123456789} {
    		buf = AppendCents(buf[:0], c)
    		fmt.Printf("AppendCents(%d) = %s\n", int64(c), buf)
    	}

    	buf = make([]byte, 0, 64)
    	allocs := testing.AllocsPerRun(100, func() {
    		buf = AppendCents(buf[:0], -123456789)
    	})
    	fmt.Println("allocations per AppendCents call:", allocs)

    	testing.Init()
    	flag.Set("test.benchtime", "200ms") // keep Run quick
    	r := testing.Benchmark(BenchmarkAppendCents)
    	fmt.Println("BenchmarkAppendCents:", r.String(), r.MemString())
    }
  tests: |
    package main

    import (
    	"flag"
    	"go/ast"
    	"go/parser"
    	"go/token"
    	"strings"
    	"testing"
    )

    func TestAppendCentsFormat(t *testing.T) {
    	for _, c := range []Cents{0, 5, 70, 99, 100, 1234, 99999, 100000, 12345678, 123456789, -5, -100, -1234, -100000000, 922337203685477} {
    		want := c.String()
    		got := string(AppendCents(nil, c))
    		if got != want {
    			t.Errorf("AppendCents(nil, %d) = %q, want %q (the same as String)", int64(c), got, want)
    		}
    		prefixed := string(AppendCents([]byte("balance: "), c))
    		if prefixed != "balance: "+want {
    			t.Errorf("AppendCents([]byte(\"balance: \"), %d) = %q, want %q (append to dst, don't replace it)", int64(c), prefixed, "balance: "+want)
    		}
    	}
    }

    func TestAppendCentsAllocs(t *testing.T) {
    	for _, c := range []Cents{5, -1234, 123456789, -100000000} {
    		buf := make([]byte, 0, 64)
    		allocs := testing.AllocsPerRun(100, func() {
    			buf = AppendCents(buf[:0], c)
    		})
    		if allocs != 0 {
    			t.Errorf("AppendCents(buf[:0], %d) with spare capacity makes %v allocations per call, want 0", int64(c), allocs)
    		}
    	}
    }

    // findBenchmark parses main.go and returns the BenchmarkAppendCents declaration.
    func findBenchmark(t *testing.T) *ast.FuncDecl {
    	t.Helper()
    	f, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, 0)
    	if err != nil {
    		t.Fatalf("parsing main.go: %v", err)
    	}
    	for _, d := range f.Decls {
    		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == "BenchmarkAppendCents" {
    			return fn
    		}
    	}
    	t.Fatal("BenchmarkAppendCents not found in main.go")
    	return nil
    }

    func TestBenchmarkShape(t *testing.T) {
    	fn := findBenchmark(t)
    	var loop, loopBodyCalls bool
    	ast.Inspect(fn.Body, func(n ast.Node) bool {
    		fs, ok := n.(*ast.ForStmt)
    		if !ok || fs.Cond == nil {
    			return true
    		}
    		if call, ok := fs.Cond.(*ast.CallExpr); ok {
    			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Loop" {
    				loop = true
    				ast.Inspect(fs.Body, func(n ast.Node) bool {
    					if call, ok := n.(*ast.CallExpr); ok {
    						if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "AppendCents" {
    							loopBodyCalls = true
    						}
    					}
    					return true
    				})
    			}
    		}
    		return true
    	})
    	if !loop {
    		t.Fatal("BenchmarkAppendCents has no `for b.Loop() { ... }` loop")
    	}
    	if !loopBodyCalls {
    		t.Fatal("BenchmarkAppendCents's b.Loop() loop doesn't call AppendCents")
    	}
    }

    func TestBenchmarkRuns(t *testing.T) {
    	findBenchmark(t)
    	if err := flag.Set("test.benchtime", "1000x"); err != nil {
    		t.Fatal(err)
    	}
    	r := testing.Benchmark(BenchmarkAppendCents)
    	if r.N == 0 {
    		t.Fatal("BenchmarkAppendCents didn't run any iterations")
    	}
    	if a := r.AllocsPerOp(); a != 0 {
    		t.Errorf("BenchmarkAppendCents reports %d allocs/op, want 0: create the buffer once, before the loop, and reuse it with buf = AppendCents(buf[:0], c)", a)
    	}
    	if !strings.Contains(r.String(), "ns/op") {
    		t.Errorf("unexpected benchmark result %q", r.String())
    	}
    }
---

Ledgerly's CSV export formats millions of amounts, and a profile shows `Cents.String` near the top: `fmt.Sprintf` plus the string surgery in `groupThousands` allocate several times per call. You'll write a zero-allocation version, and a benchmark to prove it.

## Your task

1. **Rewrite `AppendCents`** so it appends `c` to `dst`, formatted exactly like `c.String()` (`"$1,234.56"`, `"-$0.05"`), **without allocating** when `dst` has spare capacity. The starter version is correct but calls `String()`, which allocates up to six times.

   Build the output with `append` and `strconv.AppendInt`. For the thousands separators, one approach: format the dollars into a small stack array (`var digits [20]byte`, then `strconv.AppendInt(digits[:0], dollars, 10)`), then copy the digits into `dst`, adding a `,` before each digit whose distance from the end is a multiple of three. Arrays declared like that stay on the stack.

2. **Write `BenchmarkAppendCents`** using `for b.Loop()`. It should format `Cents(123456789)` into a buffer that's created once, before the loop, and reused with `buf = AppendCents(buf[:0], c)`.

The grader checks your output against `String()` for many amounts, uses `testing.AllocsPerRun` to require zero allocations, checks that your benchmark has a `b.Loop()` loop calling `AppendCents`, and runs it with `testing.Benchmark` to confirm it reports 0 allocs/op.

**Run** prints some formatted amounts, the allocations per call, and your benchmark's result. Try it before and after your rewrite. Normally you'd run the benchmark with `go test -bench AppendCents -benchmem`; the exercise calls `testing.Benchmark` from `main` so you can see it here.
