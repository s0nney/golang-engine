---
title: 'Practice: No Sneaky Mutation'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    type Doc struct {
    	Title string
    	Tags  []string
    }

    // stripTrailing returns a copy of lines with trailing spaces and tabs
    // removed from each line. It must NOT modify lines.
    func stripTrailing(lines []string) []string {
    	for i, l := range lines {
    		lines[i] = strings.TrimRight(l, " \t")
    	}
    	return lines
    }

    // withTag returns a copy of d with tag added to the end of its Tags.
    // It must NOT modify d, and the result must not share its Tags with d.
    func withTag(d Doc, tag string) Doc {
    	d.Tags = append(d.Tags, tag)
    	return d
    }

    func main() {
    	original := []string{"# Title   ", "body\t"}
    	cleaned := stripTrailing(original)
    	fmt.Printf("cleaned:  %q\n", cleaned)
    	fmt.Printf("original: %q (should still have its spaces)\n", original)

    	base := Doc{Title: "Report", Tags: make([]string, 0, 4)}
    	draft := withTag(base, "draft")
    	final := withTag(base, "final")
    	fmt.Println("draft tags:", draft.Tags, "(should be [draft])")
    	fmt.Println("final tags:", final.Tags)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"slices"
    	"strings"
    )

    type Doc struct {
    	Title string
    	Tags  []string
    }

    // stripTrailing returns a copy of lines with trailing spaces and tabs
    // removed from each line. It must NOT modify lines.
    func stripTrailing(lines []string) []string {
    	out := make([]string, len(lines))
    	for i, l := range lines {
    		out[i] = strings.TrimRight(l, " \t")
    	}
    	return out
    }

    // withTag returns a copy of d with tag added to the end of its Tags.
    // It must NOT modify d, and the result must not share its Tags with d.
    func withTag(d Doc, tag string) Doc {
    	d.Tags = append(slices.Clone(d.Tags), tag)
    	return d
    }

    func main() {
    	original := []string{"# Title   ", "body\t"}
    	cleaned := stripTrailing(original)
    	fmt.Printf("cleaned:  %q\n", cleaned)
    	fmt.Printf("original: %q (should still have its spaces)\n", original)

    	base := Doc{Title: "Report", Tags: make([]string, 0, 4)}
    	draft := withTag(base, "draft")
    	final := withTag(base, "final")
    	fmt.Println("draft tags:", draft.Tags, "(should be [draft])")
    	fmt.Println("final tags:", final.Tags)
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestStripTrailing(t *testing.T) {
    	in := []string{"# Title   ", "body\t", "  indented  ", ""}
    	want := []string{"# Title", "body", "  indented", ""}
    	got := stripTrailing(in)
    	if !slices.Equal(got, want) {
    		t.Errorf("stripTrailing = %q, want %q", got, want)
    	}
    }

    func TestStripTrailingDoesNotMutate(t *testing.T) {
    	in := []string{"# Title   ", "body\t"}
    	stripTrailing(in)
    	if want := []string{"# Title   ", "body\t"}; !slices.Equal(in, want) {
    		t.Errorf("stripTrailing changed its input to %q; the original must stay %q", in, want)
    	}
    }

    func TestWithTag(t *testing.T) {
    	d := Doc{Title: "Notes", Tags: []string{"go"}}
    	got := withTag(d, "fp")
    	if want := []string{"go", "fp"}; !slices.Equal(got.Tags, want) {
    		t.Errorf("withTag(Tags %q, \"fp\").Tags = %q, want %q", []string{"go"}, got.Tags, want)
    	}
    	if got.Title != "Notes" {
    		t.Errorf("withTag changed the title to %q, want %q", got.Title, "Notes")
    	}
    	if want := []string{"go"}; !slices.Equal(d.Tags, want) {
    		t.Errorf("withTag modified the original doc's Tags to %q, want %q", d.Tags, want)
    	}
    }

    func TestWithTagDoesNotShareBackingArray(t *testing.T) {
    	base := Doc{Title: "Report", Tags: make([]string, 1, 8)}
    	base.Tags[0] = "q3"
    	draft := withTag(base, "draft")
    	final := withTag(base, "final")
    	if want := []string{"q3", "draft"}; !slices.Equal(draft.Tags, want) {
    		t.Errorf("draft.Tags = %q, want %q (did the two calls share a backing array?)", draft.Tags, want)
    	}
    	if want := []string{"q3", "final"}; !slices.Equal(final.Tags, want) {
    		t.Errorf("final.Tags = %q, want %q", final.Tags, want)
    	}
    	draft.Tags[0] = "changed"
    	if base.Tags[0] != "q3" {
    		t.Errorf("changing the result's Tags changed the original to %q; clone the slice", base.Tags[0])
    	}
    }
---

A teammate wrote two Doc2Doc helpers. Their output *looks* right, but both have hidden
side effects, the kind this chapter warned you about. Press **Run** and read the output
carefully: the original document changes, and one tag overwrites another.

## Your task

Fix both functions so they're pure:

- `stripTrailing(lines)` must return trimmed lines **without** changing the caller's
  slice. Build a new slice instead of writing into `lines`.
- `withTag(d, tag)` must return a doc with `tag` appended, **without** changing `d`, and
  the result must not share its `Tags` backing array with `d`. Otherwise two calls on the
  same base doc can overwrite each other's tags when the slice has spare capacity:

```go
base := Doc{Tags: make([]string, 0, 4)}
draft := withTag(base, "draft")
final := withTag(base, "final")
// buggy version: draft.Tags is now [final]!
```

## Tips

- `make([]string, len(lines))` gives you a fresh slice to fill.
- `slices.Clone` copies a slice into a new backing array. Clone *before* you append.
- `d` is already a copy of the caller's struct, because structs are passed by value. The
  problem is only the slice *inside* it.

The tests check the results **and** check that the originals are untouched afterwards.
