---
title: Breadcrumbs
difficulty: easy
after: recursion
hints:
  - 'Base case: if `s.Title == title`, the path is just `[]string{s.Title}`.'
  - 'Recursive case: ask each subsection in turn with `breadcrumbs(sub, title)`. The first one that answers `true` has found it, so return `s.Title` followed by the path it gave you.'
  - 'If no subsection finds it, return `nil, false`. The loop over `s.Subs` visits children in order, which is exactly the depth-first, first-match order you need.'
exercise:
  starter: |
    package main

    import "fmt"

    // Section is one section of a document, with nested subsections.
    type Section struct {
    	Title string
    	Subs  []Section
    }

    // breadcrumbs returns the titles on the path from s down to the first
    // section (depth first, in document order) whose Title is title, and true.
    // The path starts with s.Title and ends with title.
    // If no section matches, it returns nil and false.
    func breadcrumbs(s Section, title string) ([]string, bool) {
    	// Base case: does s itself match?
    	// Recursive case: try each subsection, and prepend s.Title to a match.
    	return nil, false
    }

    func main() {
    	doc := Section{"Manual", []Section{
    		{"Install", []Section{{"Linux", nil}, {"macOS", nil}}},
    		{"Usage", []Section{{"Convert", []Section{{"Flags", nil}}}}},
    	}}
    	fmt.Println(breadcrumbs(doc, "Flags")) // want: [Manual Usage Convert Flags] true
    }
  solution: |
    package main

    import "fmt"

    type Section struct {
    	Title string
    	Subs  []Section
    }

    func breadcrumbs(s Section, title string) ([]string, bool) {
    	if s.Title == title {
    		return []string{s.Title}, true
    	}
    	for _, sub := range s.Subs {
    		if path, ok := breadcrumbs(sub, title); ok {
    			return append([]string{s.Title}, path...), true
    		}
    	}
    	return nil, false
    }

    func main() {
    	doc := Section{"Manual", []Section{
    		{"Install", []Section{{"Linux", nil}, {"macOS", nil}}},
    		{"Usage", []Section{{"Convert", []Section{{"Flags", nil}}}}},
    	}}
    	fmt.Println(breadcrumbs(doc, "Flags"))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    var manual = Section{"Manual", []Section{
    	{"Install", []Section{{"Linux", nil}, {"macOS", nil}}},
    	{"Usage", []Section{
    		{"Convert", []Section{{"Flags", nil}, {"Examples", nil}}},
    		{"Examples", nil},
    	}},
    	{"FAQ", nil},
    }}

    func TestBreadcrumbs(t *testing.T) {
    	tests := []struct {
    		title  string
    		want   []string
    		wantOK bool
    	}{
    		{"Manual", []string{"Manual"}, true},
    		{"Install", []string{"Manual", "Install"}, true},
    		{"macOS", []string{"Manual", "Install", "macOS"}, true},
    		{"Flags", []string{"Manual", "Usage", "Convert", "Flags"}, true},
    		{"Examples", []string{"Manual", "Usage", "Convert", "Examples"}, true},
    		{"FAQ", []string{"Manual", "FAQ"}, true},
    		{"Windows", nil, false},
    		{"", nil, false},
    	}
    	for _, tt := range tests {
    		got, ok := breadcrumbs(manual, tt.title)
    		if !slices.Equal(got, tt.want) || ok != tt.wantOK {
    			t.Errorf("breadcrumbs(manual, %q) = %q, %v, want %q, %v", tt.title, got, ok, tt.want, tt.wantOK)
    		}
    	}
    }

    func TestBreadcrumbsLeaf(t *testing.T) {
    	leaf := Section{Title: "Notes"}
    	if got, ok := breadcrumbs(leaf, "Notes"); !slices.Equal(got, []string{"Notes"}) || !ok {
    		t.Errorf(`breadcrumbs({Notes}, "Notes") = %q, %v, want ["Notes"], true`, got, ok)
    	}
    	if got, ok := breadcrumbs(leaf, "Other"); got != nil || ok {
    		t.Errorf(`breadcrumbs({Notes}, "Other") = %q, %v, want nil, false`, got, ok)
    	}
    }

    func TestBreadcrumbsDeep(t *testing.T) {
    	s := Section{Title: "target"}
    	for range 500 {
    		s = Section{"level", []Section{{Title: "decoy"}, s}}
    	}
    	got, ok := breadcrumbs(s, "target")
    	if !ok || len(got) != 501 || got[0] != "level" || got[500] != "target" {
    		t.Errorf("breadcrumbs on a 501-level tree: got %d titles, ok=%v, want 501 titles from \"level\" to \"target\"", len(got), ok)
    	}
    }
---

Doc2Doc's HTML output shows **breadcrumbs** above each section, like
`Manual › Usage › Convert › Flags`, so readers always know where they are.
Documents nest sections inside sections to any depth, which makes this a job
for recursion.

Complete `breadcrumbs(s, title)`. It searches the tree rooted at `s` for the
first section whose `Title` equals `title`, visiting sections **depth first in
document order** (a section, then its first subsection's whole subtree, then the
second, and so on). It returns the list of titles from `s` down to that section,
and `true`. If nothing matches, it returns `nil, false`.

## Example

```go
doc := Section{"Manual", []Section{
	{"Install", []Section{{"Linux", nil}, {"macOS", nil}}},
	{"Usage", []Section{{"Convert", []Section{{"Flags", nil}}}}},
}}

breadcrumbs(doc, "Flags")   // [Manual Usage Convert Flags], true
breadcrumbs(doc, "Manual")  // [Manual], true
breadcrumbs(doc, "Windows") // nil, false
```

## Constraints

- Trees can be up to a few hundred levels deep.
- If several sections share the title, return the path to the first one found
  depth first.
