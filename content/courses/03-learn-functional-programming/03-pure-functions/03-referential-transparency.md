---
title: Referential Transparency
quiz:
  - question: What does it mean for an expression to be *referentially transparent*?
    options:
      - text: It uses only pointers and references
      - text: Its source code is visible to other packages
      - text: It never allocates memory
      - text: It can be replaced by its value without changing the program's behaviour
        correct: true
    explanation: |
      If `wordCount("a b c")` is referentially transparent, you could swap every call
      for the literal `3` and the program would behave exactly the same.
  - question: Why is it safe to cache (memoize) the result of a pure function?
    options:
      - text: Because the same input always gives the same output and calling it has no side effects to skip
        correct: true
      - text: Because pure functions are always slow
      - text: Because Go caches every function call automatically
    explanation: |
      Caching is only correct if skipping the call changes nothing. For a pure
      function the stored answer is exactly what the call would have returned, and
      no side effect is lost.
---

An expression is **referentially transparent** if you can replace it with its value
and the program behaves exactly the same.

```go
n := wordCount("the quick brown fox")
```

If `wordCount` is pure, you could replace the call with `4` and nothing would change.
The call and its result are interchangeable. That's all the fancy name means.

Now compare:

```go
line := readLine(os.Stdin)
```

You *can't* replace that call with a fixed string, because each call reads a
different line. And replacing it would also skip the side effect of consuming input.
`readLine` isn't referentially transparent.

## Why you should care

Referential transparency is what makes code easy to reason about. When you see
`slugify(title)` twice in a function, you *know* both calls give the same answer. You
can refactor freely:

```go
// Before
if slugify(title) != "" {
	save(slugify(title))
}

// After: safe only because slugify is pure
if s := slugify(title); s != "" {
	save(s)
}
```

If `slugify` secretly incremented a counter or read a config file, that refactor
could change behaviour.

## Memoization

Because a pure function's answer never changes, you can **cache** it. This is called
*memoization*. Doc2Doc renders the same Markdown snippets again and again (headers,
footers, boilerplate), so caching them saves time:

```go
package main

import "fmt"

var calls int

func render(s string) string {
	calls++ // only here so we can see how often it runs
	return "<p>" + s + "</p>"
}

func memoize(f func(string) string) func(string) string {
	cache := map[string]string{}
	return func(s string) string {
		if v, ok := cache[s]; ok {
			return v
		}
		v := f(s)
		cache[s] = v
		return v
	}
}

func main() {
	fast := memoize(render)
	for range 3 {
		fmt.Println(fast("footer"))
	}
	fmt.Println("render ran", calls, "time(s)")
}
```

```text
<p>footer</p>
<p>footer</p>
<p>footer</p>
render ran 1 time(s)
```

`memoize` is a higher-order function: it takes a function and returns a faster
function with the same signature. (The `calls` counter is a side effect we added just
to peek inside. Real code wouldn't have it.)

Two cautions for Go:

- This cache isn't safe for concurrent use. If several goroutines share `fast`, guard
  the map with a `sync.Mutex` or use a `sync.Map`.
- The cache grows forever. For unbounded inputs you'd want a size limit.

Memoizing an *impure* function is a bug waiting to happen. Cache `readFile(path)` and
you'll keep serving the old contents after the file changes.
