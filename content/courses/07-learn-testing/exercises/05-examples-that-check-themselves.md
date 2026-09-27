---
title: Examples That Check Themselves
difficulty: easy
after: examples-and-docs
hints:
  - 'An example is a function with no parameters and no results whose body ends with a comment like `// Output:` followed by one comment line per line of output. Copy what **Run** prints into that comment.'
  - 'The name decides where the example appears in the docs: `ExampleSplit_remainder` is a second example for `Split` (the suffix after `_` must start with a lowercase letter), and `ExampleCents_String` documents the `String` method on `Cents`.'
exercise:
  starter: |
    package main

    import "fmt"

    // ExampleSplit is finished: use it as your model.
    func ExampleSplit() {
    	for _, part := range Split(900, 3) {
    		fmt.Println(part)
    	}
    	// Output:
    	// $3.00
    	// $3.00
    	// $3.00
    }

    // ExampleSplit_remainder should print the parts of Split(1000, 3),
    // one per line, and end with an // Output: comment.
    func ExampleSplit_remainder() {
    	// ?
    }

    // ExampleCents_String should print Cents(123456) and then Cents(-5),
    // one per line, and end with an // Output: comment.
    func ExampleCents_String() {
    	// ?
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    // String formats c as dollars, like "$12.34" or "-$0.05".
    func (c Cents) String() string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s$%d.%02d", sign, n/100, n%100)
    }

    // Split divides total into n parts (n >= 1) that differ by at most one
    // cent, with the larger parts first, so the parts always add up to total.
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

    func main() {
    	fmt.Println("== ExampleSplit")
    	ExampleSplit()
    	fmt.Println("== ExampleSplit_remainder")
    	ExampleSplit_remainder()
    	fmt.Println("== ExampleCents_String")
    	ExampleCents_String()
    }
  solution: |
    package main

    import "fmt"

    // ExampleSplit is finished: use it as your model.
    func ExampleSplit() {
    	for _, part := range Split(900, 3) {
    		fmt.Println(part)
    	}
    	// Output:
    	// $3.00
    	// $3.00
    	// $3.00
    }

    // ExampleSplit_remainder should print the parts of Split(1000, 3),
    // one per line, and end with an // Output: comment.
    func ExampleSplit_remainder() {
    	for _, part := range Split(1000, 3) {
    		fmt.Println(part)
    	}
    	// Output:
    	// $3.34
    	// $3.33
    	// $3.33
    }

    // ExampleCents_String should print Cents(123456) and then Cents(-5),
    // one per line, and end with an // Output: comment.
    func ExampleCents_String() {
    	fmt.Println(Cents(123456))
    	fmt.Println(Cents(-5))
    	// Output:
    	// $1234.56
    	// -$0.05
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    // String formats c as dollars, like "$12.34" or "-$0.05".
    func (c Cents) String() string {
    	sign, n := "", int64(c)
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s$%d.%02d", sign, n/100, n%100)
    }

    // Split divides total into n parts (n >= 1) that differ by at most one
    // cent, with the larger parts first, so the parts always add up to total.
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

    func main() {
    	fmt.Println("== ExampleSplit")
    	ExampleSplit()
    	fmt.Println("== ExampleSplit_remainder")
    	ExampleSplit_remainder()
    	fmt.Println("== ExampleCents_String")
    	ExampleCents_String()
    }
  tests: |
    package main

    import (
    	"go/doc"
    	"go/parser"
    	"go/token"
    	"io"
    	"os"
    	"strings"
    	"testing"
    )

    // capture runs f and returns what it printed to os.Stdout.
    func capture(t *testing.T, f func()) string {
    	t.Helper()
    	r, w, err := os.Pipe()
    	if err != nil {
    		t.Fatal(err)
    	}
    	stdout := os.Stdout
    	os.Stdout = w
    	defer func() { os.Stdout = stdout }()
    	f()
    	w.Close()
    	out, _ := io.ReadAll(r)
    	return string(out)
    }

    func TestExamples(t *testing.T) {
    	file, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, parser.ParseComments)
    	if err != nil {
    		t.Fatalf("parsing main.go: %v", err)
    	}
    	found := map[string]*doc.Example{}
    	for _, ex := range doc.Examples(file) {
    		found[ex.Name] = ex
    	}
    	for _, want := range []struct {
    		name   string
    		run    func()
    		output string
    	}{
    		{"Split", ExampleSplit, "$3.00\n$3.00\n$3.00"},
    		{"Split_remainder", ExampleSplit_remainder, "$3.34\n$3.33\n$3.33"},
    		{"Cents_String", ExampleCents_String, "$1234.56\n-$0.05"},
    	} {
    		ex := found[want.name]
    		fn := "Example" + want.name
    		if ex == nil {
    			t.Errorf("%s isn't an example function any more: keep its name, no parameters and no results", fn)
    			continue
    		}
    		if ex.Output == "" || ex.Unordered {
    			t.Errorf("%s has no // Output: comment at the end of its body, so go test would compile it but never run it", fn)
    			continue
    		}
    		printed := strings.TrimSpace(capture(t, want.run))
    		documented := strings.TrimSpace(ex.Output)
    		if printed != documented {
    			t.Errorf("%s prints:\n%s\nbut its // Output: comment says:\n%s", fn, printed, documented)
    		} else if printed != want.output {
    			t.Errorf("%s prints and documents:\n%s\nbut it should show:\n%s", fn, printed, want.output)
    		}
    	}
    }
---

In Go, an example function is documentation *and* a test. `go doc` and
pkg.go.dev show it next to the function it's named after, and `go test` runs
it and compares what it prints with its `// Output:` comment. If someone later
changes `Split` so it rounds the other way, the example fails instead of
quietly lying.

`ExampleSplit` is finished. Write the other two:

1. **`ExampleSplit_remainder`** prints the parts of `Split(1000, 3)`, one per
   line, and ends with an `// Output:` comment listing exactly what it prints.
2. **`ExampleCents_String`** prints `Cents(123456)`, then `Cents(-5)`, one per
   line, with an `// Output:` comment.

The grader parses `main.go` with `go/doc` (the same package `go test` and
`go doc` use to find examples), runs each example, and checks that it prints
what its comment claims and what the task asks for.

## Example

```go
func ExampleSplit() {
	for _, part := range Split(900, 3) {
		fmt.Println(part)
	}
	// Output:
	// $3.00
	// $3.00
	// $3.00
}
```

**Run** calls all three examples so you can see what they print. Because this
is one file, the examples live in `main.go`; in a real project they'd go in
`split_test.go` or `example_test.go`.

## Constraints

- Keep the function names exactly as given. The part after `Example` says what
  is being documented: `Split`, `Cents_String` (the method `Cents.String`), and
  `Split_remainder` (a second example for `Split` with the suffix
  `remainder`).
- Without an `// Output:` comment, `go test` compiles an example but never
  runs it. The grader treats that as a failure.
