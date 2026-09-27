---
title: go fmt, go vet and go fix
quiz:
  - question: |
      `go vet` reports a problem with this line. What is it?

      ```go
      fmt.Printf("%d messages sent to %s\n", "Alice")
      ```
    options:
      - text: '`%d` is given a string, and there is no argument for `%s`'
        correct: true
      - text: '`Printf` should be `Println`'
      - text: The format string needs single quotes
      - text: Nothing; `go vet` only checks formatting
    explanation: |
      The code compiles, but it's clearly wrong: `"Alice"` is matched to
      `%d`, and `%s` has nothing to print. `go vet` checks `Printf`-style
      calls and catches exactly this kind of bug.
  - question: What does `go fix` do in Go 1.26 and later?
    options:
      - text: Fixes every bug in your program automatically
      - text: Formats your code with tabs and aligned comments
      - text: Rewrites code to use newer, more modern Go idioms and APIs
        correct: true
      - text: Updates your dependencies to their latest versions
    explanation: |
      `go fix` runs "modernizers" that upgrade old patterns, such as turning
      `for i := 0; i < n; i++` into `for i := range n`, or `interface{}`
      into `any`. It doesn't find bugs (that's closer to `go vet`) or
      format code (that's `go fmt`).
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strconv"
    )

    // receipt describes a batch of sent messages, e.g. "sent 3 messages for $0.05".
    func receipt(sent int, dollars float64) string {
    	return fmt.Sprintf("sent %s messages for $%d", sent, dollars)
    }

    // codeText builds a verification text, e.g. "Your code is 4821".
    func codeText(code int) string {
    	return "Your code is " + string(rune(code))
    }

    func main() {
    	fmt.Println(receipt(3, 0.05))
    	fmt.Println(codeText(4821))
    	_ = strconv.Itoa // strconv is here for you to use
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strconv"
    )

    func receipt(sent int, dollars float64) string {
    	return fmt.Sprintf("sent %d messages for $%.2f", sent, dollars)
    }

    func codeText(code int) string {
    	return "Your code is " + strconv.Itoa(code)
    }

    func main() {
    	fmt.Println(receipt(3, 0.05))
    	fmt.Println(codeText(4821))
    }
  tests: |
    package main

    import "testing"

    func TestReceipt(t *testing.T) {
    	for _, tc := range []struct {
    		sent    int
    		dollars float64
    		want    string
    	}{
    		{3, 0.05, "sent 3 messages for $0.05"},
    		{120, 1.2, "sent 120 messages for $1.20"},
    		{0, 0, "sent 0 messages for $0.00"},
    	} {
    		if got := receipt(tc.sent, tc.dollars); got != tc.want {
    			t.Errorf("receipt(%d, %v) = %q, want %q", tc.sent, tc.dollars, got, tc.want)
    		}
    	}
    }

    func TestCodeText(t *testing.T) {
    	for _, tc := range []struct {
    		code int
    		want string
    	}{
    		{4821, "Your code is 4821"},
    		{65, "Your code is 65"},
    		{7, "Your code is 7"},
    	} {
    		if got := codeText(tc.code); got != tc.want {
    			t.Errorf("codeText(%d) = %q, want %q", tc.code, got, tc.want)
    		}
    	}
    }
---

Go ships with a toolbox built into the `go` command. Three tools keep your code tidy, correct and modern: `go fmt`, `go vet` and `go fix`.

## `go fmt`: one true style

Go has **one** official code style, and a tool that applies it. `go fmt` rewrites your files with standard indentation (tabs), spacing and alignment:

```go
// Before
func cost(length int)float64{
if length>160{return 0.02}
    return 0.01
}
```

```go
// After go fmt
func cost(length int) float64 {
	if length > 160 {
		return 0.02
	}
	return 0.01
}
```

```text
$ go fmt ./...
```

`./...` means "this folder and every folder below it". The command prints the names of the files it changed.

Because everyone uses the same formatter, all Go code looks the same, and teams never argue about brace placement. Most editors run it every time you save. (Under the hood, `go fmt` runs the `gofmt` tool.)

## `go vet`: find suspicious code

The compiler catches code that's *invalid*. `go vet` catches code that's valid but **probably wrong**:

```go
package main

import "fmt"

func main() {
	sent := 3
	fmt.Printf("sent %s messages\n", sent)
}
```

```text
$ go vet .
main.go:7:19: fmt.Printf format %s has arg sent of wrong type int
```

Some things `go vet` checks:

- `Printf` verbs that don't match their arguments.
- Copying a value that must not be copied, like a `sync.Mutex`.
- `string(n)` conversions from an integer, which give a character, not digits.
- Unreachable code, suspicious comparisons, and more.

`go test` runs a subset of `go vet` checks automatically. Run the full `go vet ./...` before you commit.

## `go fix`: modernise your code

Go keeps improving, and old code keeps working. But code written years ago doesn't benefit from newer, cleaner idioms. Since Go 1.26, `go fix` runs a suite of **modernizers** that rewrite your code to use them:

```go
// Before
for i := 0; i < len(recipients); i++ {
	var data interface{} = recipients[i]
	fmt.Println(data)
}
```

```text
$ go fix ./...
```

```go
// After
for i := range recipients {
	var data any = recipients[i]
	fmt.Println(data)
}
```

Some of the modernizers you'll recognise from this course:

| Modernizer | Rewrites |
|------------|----------|
| `rangeint` | Three-part `for` loops into `for i := range n` (or `range s` for a slice) |
| `any` | `interface{}` into `any` |
| `minmax` | `if`/`else` comparisons into `min` and `max` |
| `forvar` | Removes redundant `v := v` loop variable copies |
| `slicescontains` | Search loops into `slices.Contains` |
| `newexpr` | Temporary variables into `new(expr)` |
| `errorsastype` | `errors.As` into `errors.AsType` |
| `waitgroupgo` | `wg.Add(1)` / `go` / `wg.Done()` into `wg.Go` |

Each fix is designed to keep the code's behaviour exactly the same. Use `go fix -diff ./...` to preview the changes without applying them. `go tool fix help` lists every modernizer.

The fixes respect the `go` version in your `go.mod`, so `go fix` won't introduce a feature your module says it doesn't support.

## A good habit

Before committing any Go code:

```text
$ go fmt ./...
$ go vet ./...
$ go test ./...
```

And every so often, especially after upgrading Go, run `go fix ./...` and review the diff. You'll learn about `go test` in the next chapter.

## Your turn

This code compiles, but `go vet` would complain about it, and the output is garbage.
Press **Run** to see how `Printf`-style verbs report mistakes at runtime, like
`%!s(int=3)`. (**Submit** runs `go test`, which runs `go vet`'s `printf` check first.)

Fix both functions:

1. In `receipt`, use `%d` for the count and `%.2f` for the dollars (two decimal
   places), so `receipt(3, 0.05)` returns `sent 3 messages for $0.05`.
2. In `codeText`, converting a number with `string(rune(code))` gives you a
   *character*, not digits. Use `strconv.Itoa(code)` instead, and delete the
   `_ = strconv.Itoa` line in `main`.

## Further reading

- [go fmt your code (Go blog)](https://go.dev/blog/gofmt)
- [cmd/vet documentation](https://pkg.go.dev/cmd/vet)
