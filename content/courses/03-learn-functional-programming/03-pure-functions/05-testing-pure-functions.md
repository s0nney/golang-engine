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
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    	"time"
    )

    // wordsPerMinute is global configuration. Pure functions shouldn't read it.
    var wordsPerMinute = 200

    // readingTime returns how many minutes it takes to read doc at wpm words
    // per minute, rounded UP, and never less than 1.
    func readingTime(doc string, wpm int) int {
    	words := len(strings.Fields(doc))
    	return words / wordsPerMinute // BUG: ignores wpm, rounds down
    }

    // header returns the line Doc2Doc puts above every document, e.g.
    //
    //	Updated 2026-09-27 · 3 min read
    //
    // using now for the date (in time.DateOnly format) and readingTime(doc, wpm).
    func header(doc string, wpm int, now time.Time) string {
    	return fmt.Sprintf("Updated %s · %d min read",
    		time.Now().Format(time.DateOnly), readingTime(doc, wpm)) // BUG: ignores now
    }

    func main() {
    	doc := strings.Repeat("word ", 450)
    	day := time.Date(2026, time.September, 27, 9, 0, 0, 0, time.UTC)
    	fmt.Println(header(doc, 200, day)) // should print: Updated 2026-09-27 · 3 min read
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    	"time"
    )

    // wordsPerMinute is global configuration. Pure functions shouldn't read it.
    var wordsPerMinute = 200

    // readingTime returns how many minutes it takes to read doc at wpm words
    // per minute, rounded UP, and never less than 1.
    func readingTime(doc string, wpm int) int {
    	words := len(strings.Fields(doc))
    	return max(1, (words+wpm-1)/wpm)
    }

    // header returns the line Doc2Doc puts above every document, e.g.
    //
    //	Updated 2026-09-27 · 3 min read
    //
    // using now for the date (in time.DateOnly format) and readingTime(doc, wpm).
    func header(doc string, wpm int, now time.Time) string {
    	return fmt.Sprintf("Updated %s · %d min read",
    		now.Format(time.DateOnly), readingTime(doc, wpm))
    }

    func main() {
    	doc := strings.Repeat("word ", 450)
    	day := time.Date(2026, time.September, 27, 9, 0, 0, 0, time.UTC)
    	fmt.Println(header(doc, 200, day))
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    	"time"
    )

    func TestReadingTime(t *testing.T) {
    	wordsPerMinute = 1 // pure functions must ignore this
    	defer func() { wordsPerMinute = 200 }()
    	for _, tt := range []struct{ words, wpm, want int }{
    		{0, 200, 1},
    		{1, 200, 1},
    		{200, 200, 1},
    		{201, 200, 2},
    		{450, 200, 3},
    		{450, 100, 5},
    		{1000, 250, 4},
    	} {
    		doc := strings.Repeat("word ", tt.words)
    		if got := readingTime(doc, tt.wpm); got != tt.want {
    			t.Errorf("readingTime(<%d words>, %d) = %d, want %d", tt.words, tt.wpm, got, tt.want)
    		}
    	}
    }

    func TestHeader(t *testing.T) {
    	for _, tt := range []struct {
    		words int
    		wpm   int
    		now   time.Time
    		want  string
    	}{
    		{450, 200, time.Date(2026, time.September, 27, 9, 0, 0, 0, time.UTC), "Updated 2026-09-27 · 3 min read"},
    		{10, 200, time.Date(1999, time.December, 31, 23, 59, 0, 0, time.UTC), "Updated 1999-12-31 · 1 min read"},
    		{600, 120, time.Date(2030, time.January, 2, 0, 0, 0, 0, time.UTC), "Updated 2030-01-02 · 5 min read"},
    	} {
    		doc := strings.Repeat("word ", tt.words)
    		if got := header(doc, tt.wpm, tt.now); got != tt.want {
    			t.Errorf("header(<%d words>, %d, %s) = %q, want %q", tt.words, tt.wpm, tt.now.Format(time.DateOnly), got, tt.want)
    		}
    		if a, b := header(doc, tt.wpm, tt.now), header(doc, tt.wpm, tt.now); a != b {
    			t.Errorf("header is not pure: two identical calls returned %q and %q", a, b)
    		}
    	}
    }
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

## Your turn

Doc2Doc's header line (`Updated 2026-09-27 · 3 min read`) was written in a hurry, and it's impure in two ways: `readingTime` reads a global setting, and `header` asks the real clock for the date. The hidden tests call them with fixed inputs and expect fixed outputs, so neither can sneak a peek at the outside world.

1. Fix `readingTime(doc, wpm)` to use its `wpm` parameter (not the global), rounding **up**, and never returning less than 1 minute.
2. Fix `header(doc, wpm, now)` to use the `now` it was given.

Integer division rounds down; `(words + wpm - 1) / wpm` rounds up. The `max` built-in handles the minimum.
