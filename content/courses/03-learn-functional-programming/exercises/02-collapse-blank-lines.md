---
title: Collapse Blank Lines
difficulty: easy
after: pure-functions
hints:
  - 'A line is blank when `strings.TrimSpace(line) == ""`.'
  - 'Build a brand new slice with `append` instead of shifting elements around inside `lines`. Keep a boolean that remembers whether the last line you **kept** was blank.'
  - 'Blank lines are kept as `""`, whatever whitespace they had. Blank lines at the very start and end are dropped: skip them while the result is still empty, and trim a trailing `""` off the result at the end.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // collapseBlankLines returns a NEW slice where:
    //   - every run of blank lines (empty or only whitespace) becomes a single ""
    //   - blank lines at the start and end are removed entirely
    //   - all other lines are copied unchanged
    //
    // It must be pure: it must NOT modify lines.
    func collapseBlankLines(lines []string) []string {
    	// Loop over lines and append what you keep to a new slice.
    	_ = strings.TrimSpace
    	return lines
    }

    func main() {
    	doc := []string{"", "# Title", "", "  ", "", "Body text.", "\t", "End.", ""}
    	fmt.Printf("%q\n", collapseBlankLines(doc)) // want: ["# Title" "" "Body text." "" "End."]
    	fmt.Printf("%q\n", doc)                     // unchanged
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    func collapseBlankLines(lines []string) []string {
    	out := []string{}
    	lastBlank := false
    	for _, line := range lines {
    		if strings.TrimSpace(line) == "" {
    			if len(out) > 0 && !lastBlank {
    				out = append(out, "")
    			}
    			lastBlank = true
    			continue
    		}
    		out = append(out, line)
    		lastBlank = false
    	}
    	if len(out) > 0 && out[len(out)-1] == "" {
    		out = out[:len(out)-1]
    	}
    	return out
    }

    func main() {
    	doc := []string{"", "# Title", "", "  ", "", "Body text.", "\t", "End.", ""}
    	fmt.Printf("%q\n", collapseBlankLines(doc))
    	fmt.Printf("%q\n", doc)
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestCollapseBlankLines(t *testing.T) {
    	tests := []struct {
    		lines []string
    		want  []string
    	}{
    		{[]string{"", "# Title", "", "  ", "", "Body text.", "\t", "End.", ""}, []string{"# Title", "", "Body text.", "", "End."}},
    		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}},
    		{[]string{"a", "", "b"}, []string{"a", "", "b"}},
    		{[]string{"a", " ", "\t", "  \t ", "b"}, []string{"a", "", "b"}},
    		{[]string{"", "", "only"}, []string{"only"}},
    		{[]string{"only", "", "  "}, []string{"only"}},
    		{[]string{"  indented", "", "keep  trailing  "}, []string{"  indented", "", "keep  trailing  "}},
    		{[]string{"", " ", "\t"}, []string{}},
    		{[]string{}, []string{}},
    		{nil, []string{}},
    	}
    	for _, tt := range tests {
    		got := collapseBlankLines(tt.lines)
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("collapseBlankLines(%q) = %q, want %q", tt.lines, got, tt.want)
    		}
    	}
    }

    func TestCollapseBlankLinesDoesNotMutate(t *testing.T) {
    	lines := []string{"", "a", "", "", "b", " ", "c", ""}
    	orig := slices.Clone(lines)
    	collapseBlankLines(lines)
    	if !slices.Equal(lines, orig) {
    		t.Errorf("collapseBlankLines modified its input: it is now %q, want %q", lines, orig)
    	}
    }

    func TestCollapseBlankLinesNoSharedMemory(t *testing.T) {
    	lines := []string{"a", "b", "c"}
    	got := collapseBlankLines(lines)
    	if len(got) > 0 {
    		got[0] = "CHANGED"
    	}
    	if lines[0] != "a" {
    		t.Errorf("changing the result also changed the input (%q): return a new slice, not lines itself", lines)
    	}
    }

    func TestCollapseBlankLinesDeterministic(t *testing.T) {
    	lines := []string{"x", "", "", "y"}
    	a, b := collapseBlankLines(lines), collapseBlankLines(lines)
    	if !slices.Equal(a, b) {
    		t.Errorf("two calls with the same input gave %q and %q: a pure function always returns the same result", a, b)
    	}
    }
---

Documents pasted into Doc2Doc are full of stray blank lines. Before converting,
Doc2Doc tidies them up, and because the original is still shown in the "before"
pane, the tidy-up must be **pure**: it returns a new slice and leaves its
input alone.

Complete `collapseBlankLines(lines)`. A line is **blank** if it's empty or holds
only whitespace. The result:

- replaces every run of one or more blank lines between content with a single `""`;
- drops blank lines at the very start and the very end;
- copies every other line unchanged (including its own leading or trailing spaces).

It must not modify `lines`, and the result must not share memory with it.
For an input with no content lines, return an empty slice (`nil` is fine).

## Example

```go
doc := []string{"", "# Title", "", "  ", "", "Body text.", "\t", "End.", ""}
collapseBlankLines(doc) // ["# Title" "" "Body text." "" "End."]
doc                     // still ["" "# Title" "" "  " "" "Body text." "\t" "End." ""]
```

## Constraints

- `lines` holds up to a few thousand lines.
- No global variables and no printing: the function's only output is its return value.
