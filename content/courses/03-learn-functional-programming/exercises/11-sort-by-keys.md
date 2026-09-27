---
title: Sort by Keys
difficulty: medium
after: function-transformations
hints:
  - '`Desc(cmp)` returns a new comparator: `func(a, b T) int { return cmp(b, a) }`. Swapping the arguments flips the order.'
  - 'In `SortedBy`, first make a copy with `slices.Clone(items)` so the caller''s slice is untouched. Then build **one** comparator out of all of `cmps`: try each in turn and return the first non-zero answer, or `0` if they all say "equal".'
  - 'Sort the copy with `slices.SortStableFunc(copy, combined)`. The stable sort keeps items that compare equal in their original order.'
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"fmt"
    )

    type File struct {
    	Name  string
    	Ext   string
    	Pages int
    }

    func Desc[T any](cmp func(a, b T) int) func(a, b T) int {
    	return cmp
    }

    func SortedBy[T any](items []T, cmps ...func(a, b T) int) []T {
    	return items
    }

    func main() {
    	files := []File{
    		{"intro", "md", 4}, {"api", "html", 30}, {"faq", "md", 4},
    		{"guide", "html", 12}, {"notes", "txt", 1},
    	}
    	byExt := func(a, b File) int { return cmp.Compare(a.Ext, b.Ext) }
    	byPages := func(a, b File) int { return cmp.Compare(a.Pages, b.Pages) }
    	fmt.Println(SortedBy(files, byExt, Desc(byPages)))
    	// want: [{api html 30} {guide html 12} {intro md 4} {faq md 4} {notes txt 1}]
    	fmt.Println(files) // unchanged
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"slices"
    )

    type File struct {
    	Name  string
    	Ext   string
    	Pages int
    }

    func Desc[T any](cmp func(a, b T) int) func(a, b T) int {
    	return func(a, b T) int { return cmp(b, a) }
    }

    func SortedBy[T any](items []T, cmps ...func(a, b T) int) []T {
    	sorted := slices.Clone(items)
    	slices.SortStableFunc(sorted, func(a, b T) int {
    		for _, c := range cmps {
    			if r := c(a, b); r != 0 {
    				return r
    			}
    		}
    		return 0
    	})
    	return sorted
    }

    func main() {
    	files := []File{
    		{"intro", "md", 4}, {"api", "html", 30}, {"faq", "md", 4},
    		{"guide", "html", 12}, {"notes", "txt", 1},
    	}
    	byExt := func(a, b File) int { return cmp.Compare(a.Ext, b.Ext) }
    	byPages := func(a, b File) int { return cmp.Compare(a.Pages, b.Pages) }
    	fmt.Println(SortedBy(files, byExt, Desc(byPages)))
    	fmt.Println(files)
    }
  tests: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"slices"
    	"testing"
    )

    var (
    	byName  = func(a, b File) int { return cmp.Compare(a.Name, b.Name) }
    	byExt   = func(a, b File) int { return cmp.Compare(a.Ext, b.Ext) }
    	byPages = func(a, b File) int { return cmp.Compare(a.Pages, b.Pages) }
    )

    func library() []File {
    	return []File{
    		{"intro", "md", 4}, {"api", "html", 30}, {"faq", "md", 4},
    		{"guide", "html", 12}, {"notes", "txt", 1}, {"zeta", "md", 9},
    	}
    }

    func names(fs []File) string {
    	var out []string
    	for _, f := range fs {
    		out = append(out, f.Name)
    	}
    	return fmt.Sprint(out)
    }

    func TestDesc(t *testing.T) {
    	desc := Desc(cmp.Compare[int])
    	for _, tt := range []struct{ a, b, sign int }{{1, 2, 1}, {2, 1, -1}, {5, 5, 0}} {
    		got := desc(tt.a, tt.b)
    		if (got > 0) != (tt.sign > 0) || (got < 0) != (tt.sign < 0) {
    			t.Errorf("Desc(cmp.Compare)(%d, %d) = %d, want a result with sign %d", tt.a, tt.b, got, tt.sign)
    		}
    	}
    }

    func TestSortedBy(t *testing.T) {
    	tests := []struct {
    		name string
    		cmps []func(a, b File) int
    		want string
    	}{
    		{"by name", []func(a, b File) int{byName}, "[api faq guide intro notes zeta]"},
    		{"by name, descending", []func(a, b File) int{Desc(byName)}, "[zeta notes intro guide faq api]"},
    		{"by ext, then pages descending", []func(a, b File) int{byExt, Desc(byPages)}, "[api guide zeta intro faq notes]"},
    		{"by pages, then name", []func(a, b File) int{byPages, byName}, "[notes faq intro zeta guide api]"},
    		{"by ext only (stable)", []func(a, b File) int{byExt}, "[api guide intro faq zeta notes]"},
    		{"by pages only (stable)", []func(a, b File) int{byPages}, "[notes intro faq zeta guide api]"},
    		{"no comparators (original order)", nil, "[intro api faq guide notes zeta]"},
    	}
    	for _, tt := range tests {
    		if got := names(SortedBy(library(), tt.cmps...)); got != tt.want {
    			t.Errorf("SortedBy(library, %s) = %s, want %s", tt.name, got, tt.want)
    		}
    	}
    }

    func TestSortedByDoesNotMutate(t *testing.T) {
    	files := library()
    	orig := slices.Clone(files)
    	sorted := SortedBy(files, byName)
    	if !slices.Equal(files, orig) {
    		t.Errorf("SortedBy reordered its input: it is now %s, want %s", names(files), names(orig))
    	}
    	if len(sorted) > 0 {
    		sorted[0].Name = "CHANGED"
    		if files[0].Name == "CHANGED" || files[1].Name == "CHANGED" {
    			t.Errorf("changing the sorted result changed the input: return a copy")
    		}
    	}
    }

    func TestSortedByEmpty(t *testing.T) {
    	if got := SortedBy([]File{}, byName); len(got) != 0 {
    		t.Errorf("SortedBy(empty) = %v, want an empty slice", got)
    	}
    	if got := SortedBy[File](nil, byName); len(got) != 0 {
    		t.Errorf("SortedBy(nil) = %v, want an empty slice", got)
    	}
    }

    func TestSortedByOtherTypes(t *testing.T) {
    	words := []string{"bb", "a", "ccc", "dd", "e"}
    	byLen := func(a, b string) int { return cmp.Compare(len(a), len(b)) }
    	got := SortedBy(words, Desc(byLen), cmp.Compare[string])
    	want := []string{"ccc", "bb", "dd", "a", "e"}
    	if !slices.Equal(got, want) {
    		t.Errorf("SortedBy(%q, longest first, then alphabetical) = %q, want %q", words, got, want)
    	}
    }
---

Doc2Doc's file browser sorts documents by whatever columns the user clicks:
extension, then page count (biggest first), then name. Rather than write a new
sort for every combination, build them from small **comparators**, functions
`func(a, b T) int` that return a negative number if `a` comes first, positive
if `b` does, and `0` if they tie (like `cmp.Compare`).

Write two generic functions:

- `Desc(cmp)` returns a comparator that orders things the **opposite** way to `cmp`.
- `SortedBy(items, cmps...)` returns a **sorted copy** of `items`. It orders by
  the first comparator; when that says two items tie, the second comparator
  decides, and so on. Items that tie on every comparator (or all items, if
  there are no comparators) keep their original relative order.

`SortedBy` must be pure: it never reorders or modifies `items`, and the result
doesn't share memory with it.

## Example

```go
files := []File{
	{"intro", "md", 4}, {"api", "html", 30}, {"faq", "md", 4},
	{"guide", "html", 12}, {"notes", "txt", 1},
}
byExt := func(a, b File) int { return cmp.Compare(a.Ext, b.Ext) }
byPages := func(a, b File) int { return cmp.Compare(a.Pages, b.Pages) }

SortedBy(files, byExt, Desc(byPages))
// [{api html 30} {guide html 12} {intro md 4} {faq md 4} {notes txt 1}]
// intro and faq tie on both, so they stay in their original order.
```

## Constraints

- Up to a few thousand items; the `slices` package's sorts are plenty fast.
