---
title: Stopping Early
quiz:
  - question: |
      What does this print?

      ```go
      func count() iter.Seq[int] {
          return func(yield func(int) bool) {
              for i := 1; ; i++ {
                  fmt.Print("make", i, " ")
                  if !yield(i) {
                      return
                  }
              }
          }
      }

      func main() {
          for n := range count() {
              if n == 2 {
                  break
              }
          }
      }
      ```
    options:
      - text: '`make1 make2 `'
        correct: true
      - text: '`make1 `'
      - text: '`make1 make2 make3 `'
      - text: It loops forever
    explanation: |
      The iterator makes 1 and yields it; the body continues. It makes 2 and yields
      it; the body hits `break`, so `yield` returns `false` and the iterator returns.
      Even an infinite iterator stops cleanly.
  - question: What happens if an iterator keeps calling `yield` after it returned `false`?
    options:
      - text: The extra values are silently dropped
      - text: The loop body runs again
      - text: 'The program panics with "range function continued iteration after function for loop body returned false"'
        correct: true
      - text: Nothing, it's allowed
    explanation: |
      The runtime detects this bug and panics. Every iterator must check `yield`'s
      result and stop as soon as it's `false`.
---

The `bool` that `yield` returns is how the loop talks back to the iterator. It says
"keep going" (`true`) or "I'm done" (`false`). Getting this right is the one rule
every iterator author must follow.

## When yield returns false

`yield` returns `false` when the loop body leaves the loop early, by `break`,
`return`, `goto` to a label outside the loop, or a panic. Once that happens, the
iterator **must stop calling `yield`** and return.

Here's what happens if an iterator ignores it:

```go
package main

import (
	"fmt"
	"iter"
)

func sloppy() iter.Seq[string] {
	return func(yield func(string) bool) {
		yield("intro.md")
		yield("setup.md") // ignores the result: bug!
	}
}

func main() {
	for name := range sloppy() {
		fmt.Println(name)
		break
	}
}
```

```text
intro.md
panic: runtime error: range function continued iteration after function for loop body returned false
```

The runtime catches the mistake immediately rather than running your loop body after
you asked it to stop. The fix is always the same pattern:

```go
if !yield(v) {
	return
}
```

## Infinite iterators

Because the loop can stop the iterator, an iterator can be **infinite**. Doc2Doc
generates footnote labels `a`, `b`, ..., `z`, `aa`, `ab`, and so on, and nobody knows
in advance how many a document needs:

```go
package main

import (
	"fmt"
	"iter"
)

func Labels() iter.Seq[string] {
	return func(yield func(string) bool) {
		for n := 0; ; n++ {
			label := ""
			for i := n; ; i = i/26 - 1 {
				label = string(rune('a'+i%26)) + label
				if i < 26 {
					break
				}
			}
			if !yield(label) {
				return
			}
		}
	}
}

func main() {
	var got []string
	for l := range Labels() {
		got = append(got, l)
		if len(got) == 3 {
			break
		}
	}
	fmt.Println(got)

	i := 0
	for l := range Labels() {
		if i == 27 {
			fmt.Println("label 28 is", l)
			break
		}
		i++
	}
}
```

```text
[a b c]
label 28 is ab
```

The `for n := 0; ; n++` loop would run forever on its own. `yield` returning `false`
is its only exit, and that's perfectly fine.

## Cleaning up

Since the iterator is an ordinary function that simply returns when the loop stops,
`defer` works for clean-up. An iterator over lines of a file can open the file, `defer`
closing it, and yield lines. However the caller's loop ends, the file gets closed:

```go
func FileLines(path string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		f, err := os.Open(path)
		if err != nil {
			yield("", err)
			return
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if !yield(sc.Text(), nil) {
				return // the deferred Close still runs
			}
		}
		if err := sc.Err(); err != nil {
			yield("", err)
		}
	}
}
```

Yielding `(value, error)` pairs through `iter.Seq2` is the common way to report
errors from an iterator.

## Pull iterators

Occasionally you need to *ask* for values one at a time instead, for example to walk
two sequences side by side. `iter.Pull` converts a push iterator into a `next`
function plus a `stop` function:

```go
next, stop := iter.Pull(Labels())
defer stop()
a, _ := next() // "a"
b, _ := next() // "b"
```

Always call `stop` when you're done (a `defer` is easiest) so the iterator can clean
up. Most of the time, a plain `for range` is all you need.
