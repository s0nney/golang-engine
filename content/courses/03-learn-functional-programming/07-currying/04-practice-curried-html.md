---
title: 'Practice: A Curried HTML Builder'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"slices"
    )

    // element is a curried HTML builder. element(tag)(class)(content) returns
    // <tag class="class">content</tag>, or <tag>content</tag> when class is "".
    func element(tag string) func(class string) func(content string) string {
    	return func(class string) func(string) string {
    		return func(content string) string {
    			// ?
    			return ""
    		}
    	}
    }

    // longerThan returns a check that reports whether a line is longer than
    // limit bytes, ready to pass to functions like slices.IndexFunc.
    func longerThan(limit int) func(line string) bool {
    	return func(line string) bool {
    		// ?
    		return false
    	}
    }

    func main() {
    	span := element("span")
    	warning := span("warning")
    	fmt.Println(warning("Back up your files")) // <span class="warning">Back up your files</span>
    	fmt.Println(element("p")("")("Plain"))     // <p>Plain</p>

    	lines := []string{"short", "this line is definitely too long", "ok"}
    	fmt.Println(slices.IndexFunc(lines, longerThan(20))) // 1
    }
  solution: |
    package main

    import (
    	"fmt"
    	"slices"
    )

    // element is a curried HTML builder. element(tag)(class)(content) returns
    // <tag class="class">content</tag>, or <tag>content</tag> when class is "".
    func element(tag string) func(class string) func(content string) string {
    	return func(class string) func(string) string {
    		return func(content string) string {
    			if class == "" {
    				return fmt.Sprintf("<%s>%s</%s>", tag, content, tag)
    			}
    			return fmt.Sprintf(`<%s class="%s">%s</%s>`, tag, class, content, tag)
    		}
    	}
    }

    // longerThan returns a check that reports whether a line is longer than
    // limit bytes, ready to pass to functions like slices.IndexFunc.
    func longerThan(limit int) func(line string) bool {
    	return func(line string) bool {
    		return len(line) > limit
    	}
    }

    func main() {
    	span := element("span")
    	warning := span("warning")
    	fmt.Println(warning("Back up your files")) // <span class="warning">Back up your files</span>
    	fmt.Println(element("p")("")("Plain"))     // <p>Plain</p>

    	lines := []string{"short", "this line is definitely too long", "ok"}
    	fmt.Println(slices.IndexFunc(lines, longerThan(20))) // 1
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestElement(t *testing.T) {
    	for _, tt := range []struct {
    		tag, class, content, want string
    	}{
    		{"span", "warning", "Careful", `<span class="warning">Careful</span>`},
    		{"p", "intro", "Welcome to Doc2Doc", `<p class="intro">Welcome to Doc2Doc</p>`},
    		{"em", "", "wow", `<em>wow</em>`},
    		{"h1", "title", "", `<h1 class="title"></h1>`},
    	} {
    		if got := element(tt.tag)(tt.class)(tt.content); got != tt.want {
    			t.Errorf("element(%q)(%q)(%q) = %q, want %q", tt.tag, tt.class, tt.content, got, tt.want)
    		}
    	}
    }

    func TestElementPartial(t *testing.T) {
    	note := element("div")("note")
    	if got, want := note("first"), `<div class="note">first</div>`; got != want {
    		t.Errorf("note(%q) = %q, want %q", "first", got, want)
    	}
    	if got, want := note("second"), `<div class="note">second</div>`; got != want {
    		t.Errorf("reusing note: note(%q) = %q, want %q", "second", got, want)
    	}
    }

    func TestLongerThan(t *testing.T) {
    	check := longerThan(5)
    	for _, tt := range []struct {
    		line string
    		want bool
    	}{{"abcdef", true}, {"abcde", false}, {"", false}} {
    		if got := check(tt.line); got != tt.want {
    			t.Errorf("longerThan(5)(%q) = %v, want %v", tt.line, got, tt.want)
    		}
    	}
    	lines := []string{"tiny", "a much longer line", "mid-sized"}
    	if got := slices.IndexFunc(lines, longerThan(8)); got != 1 {
    		t.Errorf("slices.IndexFunc(%q, longerThan(8)) = %d, want 1", lines, got)
    	}
    }
---

Doc2Doc's HTML exporter wraps text in elements all day long. A curried builder lets it
fix the tag and class once, then reuse the result for every paragraph.

## Your task

1. Finish `element(tag)(class)(content)`. It returns
   `<tag class="class">content</tag>`, or `<tag>content</tag>` when the class is empty.
   Each stage must be reusable:

```go
note := element("div")("note")
note("first")  // <div class="note">first</div>
note("second") // <div class="note">second</div>
```

2. Finish `longerThan(limit)`. It returns a `func(string) bool` that reports whether a
   line is longer than `limit` bytes. This is the everyday Go use of currying: shaping
   a function so it fits an API such as `slices.IndexFunc`.

## Tips

- The innermost function can see `tag`, `class` and `content`, because each closure
  captures the parameters of the functions around it.
- A raw string (backticks) saves you escaping the quotes in `class="..."`.
