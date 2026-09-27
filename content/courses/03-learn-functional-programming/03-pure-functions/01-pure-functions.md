---
title: Pure Functions
quiz:
  - question: Which function is pure?
    options:
      - text: '`func wrap(s, tag string) string { return "<" + tag + ">" + s + "</" + tag + ">" }`'
        correct: true
      - text: '`func stamp(s string) string { return time.Now().Format(time.DateOnly) + " " + s }`'
      - text: '`func log(s string) string { fmt.Println(s); return s }`'
      - text: '`func next() int { counter++; return counter }`'
    explanation: |
      `wrap` depends only on its arguments and changes nothing outside itself.
      `stamp` reads the clock, `log` prints and `next` modifies a package-level
      variable, so none of those are pure.
  - question: |
      Is this function pure?

      ```go
      var tabWidth = 4

      func expandTabs(s string) string {
          return strings.ReplaceAll(s, "\t", strings.Repeat(" ", tabWidth))
      }
      ```
    options:
      - text: No, because its result depends on a package-level variable that can change
        correct: true
      - text: Yes, because it doesn't modify anything
      - text: No, because `strings.ReplaceAll` is impure
    explanation: |
      A pure function's output must depend *only* on its inputs. If some other code
      sets `tabWidth = 8`, the same call returns a different result. Pass
      `tabWidth` in as a parameter to make it pure.
---

A **pure function** follows two rules:

1. **Same input, same output.** Its result depends only on its arguments, never on
   the clock, a random number, a file, or a variable outside the function.
2. **No side effects.** It doesn't change anything the caller can see: no printing,
   no writing files, no modifying globals or the arguments passed in.

Pure functions are like maths functions. `len("doc")` is 3 today, tomorrow and on
every machine on Earth.

## Pure vs impure in Doc2Doc

```go
// Pure: depends only on its inputs.
func wordCount(doc string) int {
	return len(strings.Fields(doc))
}

// Impure: reads the clock, so the output changes every day.
func header(title string) string {
	return title + " (" + time.Now().Format(time.DateOnly) + ")"
}

// Impure: depends on a package-level variable.
var lineWidth = 80

func tooLong(line string) bool {
	return len(line) > lineWidth
}
```

You can often turn an impure function into a pure one by **passing in** whatever it
was reaching out for:

```go
package main

import (
	"fmt"
	"time"
)

func header(title string, now time.Time) string {
	return title + " (" + now.Format(time.DateOnly) + ")"
}

func tooLong(line string, width int) bool {
	return len(line) > width
}

func main() {
	day := time.Date(2026, time.March, 14, 9, 0, 0, 0, time.UTC)
	fmt.Println(header("Release Notes", day))
	fmt.Println(tooLong("short line", 80))
	fmt.Println(tooLong("short line", 5))
}
```

```text
Release Notes (2026-03-14)
false
true
```

Now `header` is pure. The *caller* decides what time it is. In production it passes
`time.Now()`, and in a test it passes a fixed date and gets a predictable answer.

## Why purity matters

- **Predictable.** You can understand a pure function by reading it and nothing else.
- **Testable.** No mocks, no setup, no clean-up. Call it and compare the result.
- **Safe to run concurrently.** A pure function touches no shared state, so a
  hundred goroutines can call it at once without locks.
- **Cacheable.** Since the answer never changes for the same input, you can store
  it and reuse it.

## Programs still need effects

A program with *no* side effects is useless: it can't read your document or save
the result. The goal isn't zero side effects. It's to keep them **at the edges**.
Doc2Doc reads a file (impure), runs a big pile of pure transformations, then writes
the output (impure). The interesting logic lives in the pure middle, where it's easy
to test.
