---
title: Writing Iterators
quiz:
  - question: Which loop fits an `iter.Seq2[int, string]`?
    options:
      - text: '`for line := range seq.Values() { ... }`'
      - text: '`for seq.Next() { ... }`'
      - text: '`for i := 0; i < len(seq); i++ { ... }`'
      - text: '`for n, line := range seq { ... }`'
        correct: true
    explanation: |
      `iter.Seq2[K, V]` yields pairs, so you range over it with two variables, just
      like ranging over a slice or a map. You can also write `for n := range seq`
      to take only the first value of each pair.
  - question: |
      By convention, what should a method named `All` on a collection type return?
    options:
      - text: An iterator over every element, such as `iter.Seq[T]` or `iter.Seq2[K, V]`
        correct: true
      - text: A slice copy of every element
      - text: The number of elements
    explanation: |
      The standard library uses `All` for "iterate over everything", as in
      `slices.All` and `maps.All`, and `Backward`, `Keys` and `Values` for other
      orders and views. Following the convention makes your types feel familiar.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"iter"
    )

    // Headings yields the 1-based line number and title of every line that
    // starts with "# ". It must stop as soon as yield returns false.
    func Headings(doc string) iter.Seq2[int, string] {
    	return func(yield func(int, string) bool) {
    		// ?
    	}
    }

    func main() {
    	doc := "# Intro\nWelcome.\n\n# Setup\nRun go.\n# Usage\n"
    	for n, title := range Headings(doc) {
    		fmt.Printf("line %d: %s\n", n, title)
    	}
    	for _, title := range Headings(doc) {
    		fmt.Println("first heading:", title)
    		break
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"iter"
    	"strings"
    )

    func Headings(doc string) iter.Seq2[int, string] {
    	return func(yield func(int, string) bool) {
    		n := 0
    		for line := range strings.Lines(doc) {
    			n++
    			title, ok := strings.CutPrefix(strings.TrimRight(line, "\n"), "# ")
    			if !ok {
    				continue
    			}
    			if !yield(n, title) {
    				return
    			}
    		}
    	}
    }

    func main() {
    	doc := "# Intro\nWelcome.\n\n# Setup\nRun go.\n# Usage\n"
    	for n, title := range Headings(doc) {
    		fmt.Printf("line %d: %s\n", n, title)
    	}
    	for _, title := range Headings(doc) {
    		fmt.Println("first heading:", title)
    		break
    	}
    }
  tests: |
    package main

    import (
    	"fmt"
    	"testing"
    )

    func collect(doc string) []string {
    	var got []string
    	for n, title := range Headings(doc) {
    		got = append(got, fmt.Sprintf("%d:%s", n, title))
    	}
    	return got
    }

    func TestHeadings(t *testing.T) {
    	for _, tt := range []struct {
    		doc  string
    		want string
    	}{
    		{"# Intro\nWelcome.\n\n# Setup\nRun go.\n# Usage\n", "[1:Intro 4:Setup 6:Usage]"},
    		{"no headings here\n## not level one\n#nospace\n", "[]"},
    		{"text\n# Last line without newline", "[2:Last line without newline]"},
    		{"", "[]"},
    		{"# A\n# B\n", "[1:A 2:B]"},
    	} {
    		got := collect(tt.doc)
    		if fmt.Sprint(got) != tt.want && !(len(got) == 0 && tt.want == "[]") {
    			t.Errorf("Headings(%q) yielded %v, want %s", tt.doc, got, tt.want)
    		}
    	}
    }

    func TestHeadingsStopsEarly(t *testing.T) {
    	defer func() {
    		if r := recover(); r != nil {
    			t.Fatalf("Headings kept yielding after the loop broke: %v", r)
    		}
    	}()
    	var got []string
    	for n, title := range Headings("# One\n# Two\n# Three\n") {
    		got = append(got, fmt.Sprintf("%d:%s", n, title))
    		if len(got) == 2 {
    			break
    		}
    	}
    	if fmt.Sprint(got) != "[1:One 2:Two]" {
    		t.Errorf("breaking after two headings gave %v, want [1:One 2:Two]", got)
    	}
    }
---

Now let's write iterators for Doc2Doc's own types, and meet `iter.Seq`'s two-valued
sibling.

## iter.Seq2: pairs

```go
type Seq2[K, V any] func(yield func(K, V) bool)
```

`Seq2` yields two values at a time, like the index and value of a slice or the key
and value of a map. Doc2Doc's error messages need line numbers, so here's an iterator
that yields `(lineNumber, line)` pairs:

```go
package main

import (
	"fmt"
	"iter"
	"strings"
)

func NumberedLines(doc string) iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		n := 1
		for line := range strings.Lines(doc) {
			if !yield(n, strings.TrimSuffix(line, "\n")) {
				return
			}
			n++
		}
	}
}

func main() {
	doc := "# Title\nsome text   \nmore text\n"
	for n, line := range NumberedLines(doc) {
		if strings.HasSuffix(line, " ") {
			fmt.Printf("line %d: trailing whitespace\n", n)
		}
	}
}
```

```text
line 2: trailing whitespace
```

`NumberedLines` is itself built on top of another iterator, `strings.Lines`.
Iterators compose nicely.

## Iterator methods on your types

Collection types conventionally offer iterator methods named after what they yield:
`All`, `Backward`, `Keys`, `Values`. Here's a Doc2Doc section tree with an `All`
method that walks every section, depth first:

```go
package main

import (
	"fmt"
	"iter"
)

type Section struct {
	Title string
	Subs  []*Section
}

// All yields every section in the tree with its depth.
func (s *Section) All() iter.Seq2[int, *Section] {
	return func(yield func(int, *Section) bool) {
		s.walk(0, yield)
	}
}

func (s *Section) walk(depth int, yield func(int, *Section) bool) bool {
	if !yield(depth, s) {
		return false
	}
	for _, sub := range s.Subs {
		if !sub.walk(depth+1, yield) {
			return false
		}
	}
	return true
}

func main() {
	doc := &Section{Title: "Guide", Subs: []*Section{
		{Title: "Install", Subs: []*Section{{Title: "Linux"}, {Title: "macOS"}}},
		{Title: "Usage"},
	}}
	for depth, s := range doc.All() {
		fmt.Printf("%*s%s\n", depth*2, "", s.Title)
	}
}
```

```text
Guide
  Install
    Linux
    macOS
  Usage
```

This is recursion from chapter 4 hiding behind a simple `for` loop. The caller has no
idea there's a tree walk going on. Notice that `walk` returns a `bool`: if the loop
body breaks, `yield` returns `false`, and that `false` must travel all the way back
up the recursion so every level stops.

## Iterators vs returning a slice

Why not just return a `[]*Section`?

- **Laziness.** An iterator does work only as the loop asks for it. Break after the
  first match, and the rest of the tree is never visited.
- **No allocation.** No slice is built to hold every element.
- **Encapsulation.** Callers can't modify your internal slice through the iterator.

A slice is still better when the caller needs random access, `len`, or to loop over
the data several times. And if they want a slice anyway, `slices.Collect(seq)` makes
one.

## Your turn

Complete `Headings(doc)`, an `iter.Seq2[int, string]` that yields the **1-based line number** and the **title** of every line starting with `# ` (a level-one Markdown heading; `## ` and `#nospace` don't count).

- `strings.Lines(doc)` iterates over the lines, each still ending in `"\n"` (except possibly the last), so trim that off before looking at the line.
- `strings.CutPrefix(line, "# ")` gives you the title and whether the prefix was there.
- Respect early termination: the moment `yield` returns `false`, return. The tests `break` out of a loop and fail if your iterator keeps going.

## Further reading

- [The Go Blog: Range Over Function Types](https://go.dev/blog/range-functions)
