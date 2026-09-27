---
title: Recursing Through Nested Data
quiz:
  - question: |
      What does this print?

      ```go
      func depth(v any) int {
          list, ok := v.([]any)
          if !ok {
              return 0
          }
          deepest := 0
          for _, item := range list {
              deepest = max(deepest, depth(item))
          }
          return 1 + deepest
      }

      func main() {
          fmt.Println(depth([]any{"a", []any{"b", []any{}}}))
      }
      ```
    options:
      - text: '`2`'
      - text: '`3`'
        correct: true
      - text: '`1`'
      - text: '`0`'
    explanation: |
      The outer list is depth 1 plus its deepest child. `"a"` is 0. The middle list
      is 1 plus its deepest child, and the empty list inside it is `1 + 0 = 1`, so
      the middle list is 2. The outer list is `1 + 2 = 3`.
  - question: 'In the `toc` function from this lesson, why is `prefix` passed down as a parameter instead of kept in a package-level variable?'
    options:
      - text: Package-level variables can't hold strings
      - text: Each call gets its own prefix, so the function stays pure and sibling sections can't clobber each other's numbers
        correct: true
      - text: It makes the function faster
    explanation: |
      Passing state down through parameters gives each stack frame its own copy.
      A shared global would need careful resetting on the way back up, and it
      would break if two goroutines built a table of contents at the same time.
exercise:
  starter: |
    package main

    import "fmt"

    func join(prefix, key string) string {
    	if prefix == "" {
    		return key
    	}
    	return prefix + "." + key
    }

    // leafPaths lists every leaf value inside v as "path=value", where path
    // joins the keys and indexes that lead to it with dots (use join). Visit
    // map keys in sorted order and slice items in index order. Anything that
    // isn't a map[string]any or []any is a leaf; format it with fmt's %v.
    func leafPaths(v any, prefix string) []string {
    	// ?
    	return nil
    }

    func main() {
    	frontMatter := map[string]any{
    		"title": "Quarterly Report",
    		"meta": map[string]any{
    			"draft": false,
    			"tags":  []any{"finance", "q3"},
    		},
    	}
    	// should print meta.draft=false, meta.tags.0=finance, meta.tags.1=q3,
    	// title=Quarterly Report, one per line
    	for _, p := range leafPaths(frontMatter, "") {
    		fmt.Println(p)
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"maps"
    	"slices"
    	"strconv"
    )

    func join(prefix, key string) string {
    	if prefix == "" {
    		return key
    	}
    	return prefix + "." + key
    }

    func leafPaths(v any, prefix string) []string {
    	switch v := v.(type) {
    	case map[string]any:
    		var out []string
    		for _, k := range slices.Sorted(maps.Keys(v)) {
    			out = append(out, leafPaths(v[k], join(prefix, k))...)
    		}
    		return out
    	case []any:
    		var out []string
    		for i, item := range v {
    			out = append(out, leafPaths(item, join(prefix, strconv.Itoa(i)))...)
    		}
    		return out
    	default:
    		return []string{fmt.Sprintf("%s=%v", prefix, v)}
    	}
    }

    func main() {
    	frontMatter := map[string]any{
    		"title": "Quarterly Report",
    		"meta": map[string]any{
    			"draft": false,
    			"tags":  []any{"finance", "q3"},
    		},
    	}
    	for _, p := range leafPaths(frontMatter, "") {
    		fmt.Println(p)
    	}
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestLeafPaths(t *testing.T) {
    	for _, tt := range []struct {
    		name   string
    		v      any
    		prefix string
    		want   []string
    	}{
    		{"a single leaf", "hello", "greeting", []string{"greeting=hello"}},
    		{"flat map, sorted keys", map[string]any{"b": 2, "a": 1, "c": true}, "", []string{"a=1", "b=2", "c=true"}},
    		{"list", []any{"x", "y"}, "tags", []string{"tags.0=x", "tags.1=y"}},
    		{"nested", map[string]any{
    			"title": "Report",
    			"meta":  map[string]any{"draft": false, "tags": []any{"go", "fp"}},
    		}, "", []string{"meta.draft=false", "meta.tags.0=go", "meta.tags.1=fp", "title=Report"}},
    		{"lists of maps", []any{map[string]any{"n": 1}, map[string]any{"n": 2, "m": nil}}, "items", []string{"items.0.n=1", "items.1.m=<nil>", "items.1.n=2"}},
    		{"empty containers have no leaves", map[string]any{"a": []any{}, "b": map[string]any{}}, "", nil},
    	} {
    		got := leafPaths(tt.v, tt.prefix)
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("%s: leafPaths(%v, %q) = %q, want %q", tt.name, tt.v, tt.prefix, got, tt.want)
    		}
    	}
    }
---

Trees aren't only folders. Plenty of everyday data is nested: sections inside
sections, JSON objects inside JSON objects, lists inside lists. Whenever the nesting
depth isn't fixed, recursion is the natural tool.

## Doc2Doc: a numbered table of contents

A document is a list of sections, and each section can have subsections:

```go
type Section struct {
	Title string
	Subs  []Section
}
```

Doc2Doc should produce a numbered table of contents, like `1.`, `1.1.`, `1.2.`, `2.`
and so on. The number of each section depends on its parent's number, so you **pass
the parent's prefix down** as a parameter:

```go
package main

import (
	"fmt"
	"strings"
)

type Section struct {
	Title string
	Subs  []Section
}

func toc(sections []Section, prefix string, depth int) []string {
	var lines []string
	for i, s := range sections {
		num := fmt.Sprintf("%s%d.", prefix, i+1)
		line := strings.Repeat("  ", depth) + num + " " + s.Title
		lines = append(lines, line)
		lines = append(lines, toc(s.Subs, num, depth+1)...)
	}
	return lines
}

func main() {
	doc := []Section{
		{Title: "Introduction"},
		{Title: "Installation", Subs: []Section{
			{Title: "Linux"},
			{Title: "macOS", Subs: []Section{
				{Title: "Homebrew"},
			}},
		}},
		{Title: "Usage"},
	}
	for _, line := range toc(doc, "", 0) {
		fmt.Println(line)
	}
}
```

```text
1. Introduction
2. Installation
  2.1. Linux
  2.2. macOS
    2.2.1. Homebrew
3. Usage
```

A few things to notice:

- `toc` is **pure**. It returns lines instead of printing them, so it's easy to test.
- The base case is an empty `Subs` slice, where the loop does nothing and returns `nil`.
  Appending a nil slice with `...` is fine and adds nothing.
- Information flows **down** through parameters (`prefix`, `depth`) and **up**
  through return values (`lines`).

## Nested JSON

When Doc2Doc reads front matter or config as JSON without a fixed struct, you get
`map[string]any` and `[]any` nested to any depth. A recursive function with a type
switch walks it:

```go
func countStrings(v any) int {
	switch v := v.(type) {
	case string:
		return 1
	case []any:
		n := 0
		for _, item := range v {
			n += countStrings(item)
		}
		return n
	case map[string]any:
		n := 0
		for _, item := range v {
			n += countStrings(item)
		}
		return n
	default: // numbers, booleans, nil
		return 0
	}
}
```

Each `case` handles one shape of data. The leaf cases (`string` and `default`) are the
base cases, and the container cases recurse.

## Recursion and immutability

Notice that none of these functions modify their input. They read the nested data
and build a *new* result. That's the functional sweet spot: recursive data, recursive
functions and no mutation. They're easy to test with small hand-written inputs like
the one in `main`.

## Your turn

When front matter looks wrong, Doc2Doc's `--debug` flag prints every value in it with its full path, like `meta.tags.1=q3`. Complete `leafPaths(v, prefix)`:

- A `map[string]any` recurses into each value, visiting keys in **sorted** order (`slices.Sorted(maps.Keys(m))`), with the key added to the path.
- A `[]any` recurses into each item in order, with the index added to the path (`strconv.Itoa(i)`).
- Anything else is a leaf: return a one-element slice holding `path=value`, formatted with `%v`.

Use the provided `join(prefix, key)` to add a path segment; it leaves out the dot when the prefix is empty. Empty maps and lists have no leaves, so they contribute nothing. Add the imports you need.
