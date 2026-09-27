---
title: go vet and staticcheck
quiz:
  - question: |
      What does `go vet` report for this line?

      ```go
      msg := fmt.Sprintf("%s: %v on %s", tx.Account, tx.Amount)
      ```
    options:
      - text: Nothing, it compiles
      - text: '`fmt.Sprintf format %s reads arg #3, but call has 2 args`'
        correct: true
      - text: A compile error
      - text: That `%v` should be `%d`
    explanation: |
      The compiler doesn't check format strings, but vet's `printf` check
      does. At run time the missing argument would show up as
      `%!s(MISSING)` in the output.
  - question: You run `go test` and it fails before running any test, printing a `go vet` message. Why?
    options:
      - text: '`go test` always runs every vet check'
      - text: '`go test` runs a small, high-confidence subset of vet checks (like `printf` and `errorsas`) and refuses to run tests if they fail'
        correct: true
      - text: The tests themselves call `go vet`
      - text: It's a bug in `go test`
    explanation: |
      The subset is chosen so that it almost never reports false positives.
      Run `go vet ./...` yourself for the full set of default checks.
exercise:
  starter: |
    package main

    import (
    	"encoding/json"
    	"errors"
    	"fmt"
    	"sync"
    )

    // LineError reports a bad line in a CSV import.
    type LineError struct {
    	Line int
    	Err  error
    }

    func (e *LineError) Error() string { return fmt.Sprintf("line %d: %v", e.Line, e.Err) }
    func (e *LineError) Unwrap() error { return e.Err }

    // Row is one imported transaction, as JSON: {"account":"rent","cents":-120000}.
    type Row struct {
    	Account string `json:account`
    	Cents   int64  `json:"cents"`
    }

    // Describe returns a summary like "rent: -120000 cents on 2026-09-01".
    func Describe(account string, cents int64, day string) string {
    	return fmt.Sprintf("%s: %d cents on %s", account, cents)
    }

    // Validate returns an error if cents is zero: Ledgerly doesn't record
    // empty transactions.
    func Validate(cents int64) error {
    	if cents == 0 {
    		fmt.Errorf("ledgerly: zero amount")
    	}
    	return nil
    }

    // LineOf returns the line number from a *LineError anywhere in err's
    // chain, or 0 if there isn't one.
    func LineOf(err error) int {
    	var le *LineError
    	if errors.As(err, le) {
    		return le.Line
    	}
    	return 0
    }

    // Counter counts imported rows. It's safe for concurrent use.
    type Counter struct {
    	mu sync.Mutex
    	n  int
    }

    // Inc adds one to the count.
    func (c Counter) Inc() {
    	c.mu.Lock()
    	defer c.mu.Unlock()
    	c.n++
    }

    // Value returns the current count.
    func (c *Counter) Value() int {
    	c.mu.Lock()
    	defer c.mu.Unlock()
    	return c.n
    }

    func main() {
    	b, _ := json.Marshal(Row{Account: "rent", Cents: -120000})
    	fmt.Println(string(b))
    	fmt.Println(Describe("rent", -120000, "2026-09-01"))
    	fmt.Println("Validate(0):", Validate(0))
    	err := fmt.Errorf("importing march.csv: %w", &LineError{Line: 7, Err: errors.New("bad amount")})
    	fmt.Println("LineOf:", LineOf(err))
    	var c Counter
    	for range 3 {
    		c.Inc()
    	}
    	fmt.Println("count:", c.Value())
    }
  solution: |
    package main

    import (
    	"encoding/json"
    	"errors"
    	"fmt"
    	"sync"
    )

    // LineError reports a bad line in a CSV import.
    type LineError struct {
    	Line int
    	Err  error
    }

    func (e *LineError) Error() string { return fmt.Sprintf("line %d: %v", e.Line, e.Err) }
    func (e *LineError) Unwrap() error { return e.Err }

    // Row is one imported transaction, as JSON: {"account":"rent","cents":-120000}.
    type Row struct {
    	Account string `json:"account"`
    	Cents   int64  `json:"cents"`
    }

    // Describe returns a summary like "rent: -120000 cents on 2026-09-01".
    func Describe(account string, cents int64, day string) string {
    	return fmt.Sprintf("%s: %d cents on %s", account, cents, day)
    }

    // Validate returns an error if cents is zero: Ledgerly doesn't record
    // empty transactions.
    func Validate(cents int64) error {
    	if cents == 0 {
    		return fmt.Errorf("ledgerly: zero amount")
    	}
    	return nil
    }

    // LineOf returns the line number from a *LineError anywhere in err's
    // chain, or 0 if there isn't one.
    func LineOf(err error) int {
    	var le *LineError
    	if errors.As(err, &le) {
    		return le.Line
    	}
    	return 0
    }

    // Counter counts imported rows. It's safe for concurrent use.
    type Counter struct {
    	mu sync.Mutex
    	n  int
    }

    // Inc adds one to the count.
    func (c *Counter) Inc() {
    	c.mu.Lock()
    	defer c.mu.Unlock()
    	c.n++
    }

    // Value returns the current count.
    func (c *Counter) Value() int {
    	c.mu.Lock()
    	defer c.mu.Unlock()
    	return c.n
    }

    func main() {
    	b, _ := json.Marshal(Row{Account: "rent", Cents: -120000})
    	fmt.Println(string(b))
    	fmt.Println(Describe("rent", -120000, "2026-09-01"))
    	fmt.Println("Validate(0):", Validate(0))
    	err := fmt.Errorf("importing march.csv: %w", &LineError{Line: 7, Err: errors.New("bad amount")})
    	fmt.Println("LineOf:", LineOf(err))
    	var c Counter
    	for range 3 {
    		c.Inc()
    	}
    	fmt.Println("count:", c.Value())
    }
  tests: |
    package main

    import (
    	"encoding/json"
    	"errors"
    	"fmt"
    	"sync"
    	"testing"
    )

    func TestRowJSON(t *testing.T) {
    	b, err := json.Marshal(Row{Account: "rent", Cents: -120000})
    	if err != nil {
    		t.Fatal(err)
    	}
    	if want := `{"account":"rent","cents":-120000}`; string(b) != want {
    		t.Errorf("json.Marshal(Row{...}) = %s, want %s (check the struct tags)", b, want)
    	}
    }

    func TestDescribe(t *testing.T) {
    	if got, want := Describe("rent", -120000, "2026-09-01"), "rent: -120000 cents on 2026-09-01"; got != want {
    		t.Errorf("Describe(\"rent\", -120000, \"2026-09-01\") = %q, want %q", got, want)
    	}
    }

    func TestValidate(t *testing.T) {
    	if err := Validate(0); err == nil {
    		t.Error("Validate(0) = nil, want an error")
    	}
    	if err := Validate(-5); err != nil {
    		t.Errorf("Validate(-5) = %v, want nil", err)
    	}
    }

    func TestLineOf(t *testing.T) {
    	err := fmt.Errorf("importing march.csv: %w", &LineError{Line: 7, Err: errors.New("bad amount")})
    	if got := LineOf(err); got != 7 {
    		t.Errorf("LineOf(wrapped *LineError{Line: 7}) = %d, want 7", got)
    	}
    	if got := LineOf(errors.New("disk full")); got != 0 {
    		t.Errorf("LineOf(errors.New(\"disk full\")) = %d, want 0", got)
    	}
    }

    func TestCounter(t *testing.T) {
    	var c Counter
    	var wg sync.WaitGroup
    	for range 50 {
    		wg.Go(c.Inc)
    	}
    	wg.Wait()
    	if got := c.Value(); got != 50 {
    		t.Errorf("after 50 concurrent Inc calls, Value() = %d, want 50", got)
    	}
    }
---

You've been running `go vet` since [Learn Go](/courses/learn-go/packages-and-modules/go-fmt-vet-and-fix). This lesson looks at what it actually checks, why `go test` runs only part of it, and what to reach for when vet isn't enough.

The compiler only rejects code that isn't valid Go. Plenty of valid Go is still obviously wrong: a format string with the wrong number of arguments, a mutex copied by value, an error created and then thrown away. **Static analysis** tools read your code without running it and flag those patterns.

## go vet

`go vet` ships with Go and runs a set of checks, called **analyzers**, over your packages:

```text
$ go vet ./...
./main.go:25:2: struct field tag `json:account` not compatible with reflect.StructTag.Get: bad syntax for struct tag value
./main.go:31:32: fmt.Sprintf format %s reads arg #3, but call has 2 args
./main.go:44:3: result of fmt.Errorf call not used
./main.go:58:9: Inc passes lock by value: ledgerly.Counter contains sync.Mutex
./main.go:83:5: second argument to errors.As must be a non-nil pointer to either a type that implements error, or to any interface type
```

Five real bugs, found in a fraction of a second, and every one of them compiles. Some of the default analyzers:

| Analyzer | Catches |
| --- | --- |
| `printf` | format verbs that don't match their arguments, missing or extra arguments |
| `copylocks` | a `sync.Mutex` (or a struct containing one) copied by value, so it doesn't lock anything |
| `structtag` | malformed struct tags like `json:account` (missing quotes), so the tag is silently ignored |
| `unusedresult` | calling `fmt.Errorf`, `fmt.Sprintf`, `slices.Sorted` and similar and ignoring the result |
| `errorsas` | passing a non-pointer to `errors.As`, which panics at run time |
| `tests` | malformed test, benchmark, fuzz and example names and signatures |
| `lostcancel` | a `context.WithCancel` whose cancel function is never called |
| `unreachable` | code after a `return` or `panic` |

`go tool vet help` lists them all, and `go tool vet help printf` explains one.

## go test runs vet for you

As part of building a test binary, `go test` runs a **high-confidence subset** of vet: `atomic`, `bools`, `buildtag`, `directive`, `errorsas`, `ifaceassert`, `nilfunc`, `printf`, `stdversion`, `stringintconv` and `tests`. If any of them report a problem, the tests don't run at all:

```text
$ go test ./ledgerly
# ledgerly
# [ledgerly]
./main.go:83:5: second argument to errors.As must be a non-nil pointer to either a type that implements error, or to any interface type
./main.go:31:32: fmt.Sprintf format %s reads arg #3, but call has 2 args
FAIL	ledgerly [build failed]
```

That subset was picked because it almost never flags correct code. The other checks, like `copylocks` and `unusedresult`, only run when you run `go vet` yourself, so put `go vet ./...` in CI.

## Your turn: clear the vet report

The code in the editor compiles, and **Run** even prints a few lines before it panics. It also contains one example of each of the first five bugs in this lesson's `go vet` output. Fix all five:

1. a struct tag that `encoding/json` silently ignores;
2. a `Sprintf` call with fewer arguments than verbs;
3. an error that's created and thrown away;
4. an `errors.As` call that would panic;
5. a mutex copied by value, so `Inc` never changes the real counter.

Press **Submit** before you change anything. The grader builds your code with `go test`, which runs its high-confidence vet checks first, so you'll see exactly the "tests don't run at all" failure described above. Once those two are fixed, the tests run and point at the other three, which only a full `go vet ./...` reports. Compare **Run**'s output before and after: every line of it is wrong at the start.

## staticcheck

vet is deliberately conservative. **staticcheck** (from honnef.co/go/tools, installed with `go install honnef.co/go/tools/cmd/staticcheck@latest`) is the most popular third-party analyzer and goes much further: over 150 checks for bugs, performance problems, simplifications and style. For example, vet doesn't notice this:

```go
func CleanAccount(name string) string {
	strings.TrimSpace(name) // result thrown away
	return strings.ToLower(name)
}
```

staticcheck does. Its check SA4017 flags calls to functions that have no side effects when their result is ignored. It also warns about uses of anything marked `Deprecated:` (SA1019), which is where the deprecation paragraphs from the Examples and Docs chapter pay off.

Two practical notes:

- **Keep it up to date.** staticcheck has to understand the Go version your code uses. A staticcheck built for an older Go fails with "export data version" errors on newer toolchains.
- **golangci-lint** bundles staticcheck, vet and dozens of other linters behind one command and one config file. Many teams run that in CI instead.

## Fixing, not silencing

When a tool flags something, the first assumption should be that it's right. The `copylocks` report above is a real bug: `Inc` has a value receiver, so it locks and increments a *copy* of the counter, and the real count never changes. `-race` might not even catch it, because each goroutine works on its own copy.

When a report really is a false positive, staticcheck supports `//lint:ignore SA4017 reason` comments. Always include the reason: the next reader deserves to know why the rule doesn't apply.

## Further reading

- [cmd/vet documentation](https://pkg.go.dev/cmd/vet)
- [staticcheck checks](https://staticcheck.dev/docs/checks/)
