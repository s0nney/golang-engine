---
title: 'Practice: Finding Headings'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // isHeading reports whether line is a Markdown heading: one to six '#'
    // characters followed by a space, like "# Intro" or "### Setup".
    func isHeading(line string) bool {
    	// ?
    	return false
    }

    // hasHeading reports whether any line is a heading.
    // Hint: slices.ContainsFunc.
    func hasHeading(lines []string) bool {
    	// ?
    	return false
    }

    // headingTitles returns the text of every heading, without the '#'s
    // and the space, in document order.
    func headingTitles(lines []string) []string {
    	// ?
    	return nil
    }

    func main() {
    	doc := "# Doc2Doc\nConverts files.\n## Install\nRun go install.\n#hashtag"
    	lines := strings.Split(doc, "\n")
    	fmt.Println(isHeading(lines[0]), isHeading(lines[4]))
    	fmt.Println(hasHeading(lines))
    	fmt.Printf("%q\n", headingTitles(lines))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"slices"
    	"strings"
    )

    func isHeading(line string) bool {
    	hashes := len(line) - len(strings.TrimLeft(line, "#"))
    	return hashes >= 1 && hashes <= 6 && strings.HasPrefix(line[hashes:], " ")
    }

    func hasHeading(lines []string) bool {
    	return slices.ContainsFunc(lines, isHeading)
    }

    func headingTitles(lines []string) []string {
    	var titles []string
    	for _, line := range lines {
    		if isHeading(line) {
    			titles = append(titles, strings.TrimSpace(strings.TrimLeft(line, "#")))
    		}
    	}
    	return titles
    }

    func main() {
    	doc := "# Doc2Doc\nConverts files.\n## Install\nRun go install.\n#hashtag"
    	lines := strings.Split(doc, "\n")
    	fmt.Println(isHeading(lines[0]), isHeading(lines[4]))
    	fmt.Println(hasHeading(lines))
    	fmt.Printf("%q\n", headingTitles(lines))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestIsHeading(t *testing.T) {
    	for _, tt := range []struct {
    		line string
    		want bool
    	}{
    		{"# Intro", true},
    		{"### Setup", true},
    		{"###### Tiny", true},
    		{"####### Too deep", false},
    		{"#hashtag", false},
    		{"Not a heading", false},
    		{"", false},
    		{"#", false},
    	} {
    		if got := isHeading(tt.line); got != tt.want {
    			t.Errorf("isHeading(%q) = %v, want %v", tt.line, got, tt.want)
    		}
    	}
    }

    func TestHasHeading(t *testing.T) {
    	for _, tt := range []struct {
    		lines []string
    		want  bool
    	}{
    		{[]string{"text", "## Usage"}, true},
    		{[]string{"text", "#tag"}, false},
    		{nil, false},
    	} {
    		if got := hasHeading(tt.lines); got != tt.want {
    			t.Errorf("hasHeading(%q) = %v, want %v", tt.lines, got, tt.want)
    		}
    	}
    }

    func TestHeadingTitles(t *testing.T) {
    	lines := []string{"# Doc2Doc", "text", "## Install", "#tag", "### Linux "}
    	want := []string{"Doc2Doc", "Install", "Linux"}
    	if got := headingTitles(lines); !slices.Equal(got, want) {
    		t.Errorf("headingTitles(%q) = %q, want %q", lines, got, want)
    	}
    	if got := headingTitles([]string{"no headings here"}); len(got) != 0 {
    		t.Errorf("headingTitles with no headings = %q, want an empty result", got)
    	}
    }
---

Time to write some code! Doc2Doc needs to know which lines of a Markdown document are
headings, so it can build a table of contents later.

## Your task

Complete the three functions in the editor:

1. `isHeading(line)` reports whether a line is a heading: **one to six** `#` characters
   followed by a space. `"## Install"` is a heading; `"#hashtag"` and
   `"####### Too deep"` are not.
2. `hasHeading(lines)` reports whether *any* line is a heading. Write it declaratively:
   one call to `slices.ContainsFunc`, passing `isHeading` itself as the argument.
3. `headingTitles(lines)` returns the text of each heading without the `#`s and the
   surrounding space, in order.

```go
headingTitles([]string{"# Doc2Doc", "text", "## Install"})
// ["Doc2Doc" "Install"]
```

## Tips

- `strings.TrimLeft(line, "#")` removes all leading `#`s. Comparing its length with
  the original tells you how many there were.
- Build `hasHeading` and `headingTitles` on top of `isHeading` so the rule lives in
  exactly one place. That's the declarative habit: small, well-named pieces that you
  combine.
- You'll need to add `"slices"` to the imports.

Press **Run** to see what your code prints, then **Submit** to run the tests.
