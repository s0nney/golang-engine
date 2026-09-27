---
title: Testing Pure Functions
quiz:
  - question: Why are pure functions easy to test?
    options:
      - text: Go runs pure functions in a special test mode
      - text: You just call them with inputs and compare the outputs, with no setup, mocks or clean-up
        correct: true
      - text: Pure functions can't contain bugs
      - text: They don't need to be compiled
    explanation: |
      A pure function's result depends only on its arguments, and it doesn't touch
      files, clocks or globals. So a test is just "input in, expected output out".
  - question: |
      You want to test `func Greeting() string`, which returns "Good morning" or "Good evening" based on `time.Now()`. What's the most functional fix?
    options:
      - text: Only run the test in the morning
      - text: Set the computer's clock from the test
      - text: Change it to `func Greeting(now time.Time) string` so the test can pass a fixed time
        correct: true
      - text: Skip testing it
    explanation: |
      Passing the time in as a parameter makes `Greeting` pure. Production code
      passes `time.Now()`, and tests pass whatever time they need.
---

The biggest practical payoff of pure functions is testing. A pure function needs no
temporary files, no fake clocks and no mock servers. You call it and check the answer.

## A table-driven test

Here's a pure Doc2Doc helper that turns a heading into a URL-friendly anchor:

```go
package doc2doc

import "strings"

// Slug turns "Getting Started!" into "getting-started".
func Slug(heading string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(heading) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		case !dash && b.Len() > 0:
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}
```

It uses a `strings.Builder` and mutates local variables, but from the outside it's
pure: same heading in, same slug out, nothing else touched.

Testing it is a table of inputs and expected outputs:

```go
package doc2doc

import "testing"

func TestSlug(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Getting Started", "getting-started"},
		{"  Hello,   World!  ", "hello-world"},
		{"Go 1.27 Release", "go-1-27-release"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := Slug(tt.in); got != tt.want {
				t.Errorf("Slug(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
```

```text
$ go test
PASS
ok  	doc2doc	0.002s
```

Each case is one line, so adding a new edge case costs almost nothing. Because
`Slug` is pure, the subtests could even run in parallel with `t.Parallel()` with no
risk of them interfering with each other.

## Testing the impure edges

Now compare testing the *impure* shell that reads a file and writes the result. You'd
need `t.TempDir()`, you'd write fixture files, read the output back and check
permissions. It's doable, but slower and noisier.

That's why the "functional core, imperative shell" design pays off. Push as much logic
as you can into pure functions like `Slug`, `Wrap` and `NumberHeadings`, and test them
heavily with tables. Keep the shell so thin that a couple of integration tests cover
it.

## Injecting impurity

When a function needs the time, randomness or configuration, pass it in:

```go
// Hard to test: depends on the real clock.
func Footer() string {
	return "Generated " + time.Now().Format(time.DateOnly)
}

// Easy to test: the caller supplies the time.
func Footer(now time.Time) string {
	return "Generated " + now.Format(time.DateOnly)
}
```

The same trick works for functions: accept a `func() time.Time` or a
`func(string) ([]byte, error)` parameter, and tests can pass in a fake. Functions
as values make dependency injection cheap in Go, with no interfaces or frameworks
needed.
