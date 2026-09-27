---
title: 'Practice: Walking the Section Tree'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    type Section struct {
    	Title string
    	Body  string
    	Subs  []Section
    }

    // wordCount returns the number of words in s.Body plus the words in
    // every nested subsection, at any depth.
    func wordCount(s Section) int {
    	// ?
    	return 0
    }

    // depth returns how many levels deep the section tree goes.
    // A section with no subsections has depth 1.
    func depth(s Section) int {
    	// ?
    	return 0
    }

    // outline returns every title in the tree, depth first, indented by
    // two spaces per level below s.
    func outline(s Section) []string {
    	// ?
    	return nil
    }

    func main() {
    	guide := Section{Title: "Guide", Body: "Welcome to Doc2Doc", Subs: []Section{
    		{Title: "Install", Body: "Pick your platform", Subs: []Section{
    			{Title: "Linux", Body: "Use go install"},
    			{Title: "macOS", Body: "Use Homebrew"},
    		}},
    		{Title: "Usage", Body: "Run doc2doc on a file"},
    	}}
    	fmt.Println("words:", wordCount(guide)) // 16
    	fmt.Println("depth:", depth(guide))     // 3
    	fmt.Println(strings.Join(outline(guide), "\n"))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    type Section struct {
    	Title string
    	Body  string
    	Subs  []Section
    }

    // wordCount returns the number of words in s.Body plus the words in
    // every nested subsection, at any depth.
    func wordCount(s Section) int {
    	n := len(strings.Fields(s.Body))
    	for _, sub := range s.Subs {
    		n += wordCount(sub)
    	}
    	return n
    }

    // depth returns how many levels deep the section tree goes.
    // A section with no subsections has depth 1.
    func depth(s Section) int {
    	deepest := 0
    	for _, sub := range s.Subs {
    		deepest = max(deepest, depth(sub))
    	}
    	return 1 + deepest
    }

    // outline returns every title in the tree, depth first, indented by
    // two spaces per level below s.
    func outline(s Section) []string {
    	lines := []string{s.Title}
    	for _, sub := range s.Subs {
    		for _, line := range outline(sub) {
    			lines = append(lines, "  "+line)
    		}
    	}
    	return lines
    }

    func main() {
    	guide := Section{Title: "Guide", Body: "Welcome to Doc2Doc", Subs: []Section{
    		{Title: "Install", Body: "Pick your platform", Subs: []Section{
    			{Title: "Linux", Body: "Use go install"},
    			{Title: "macOS", Body: "Use Homebrew"},
    		}},
    		{Title: "Usage", Body: "Run doc2doc on a file"},
    	}}
    	fmt.Println("words:", wordCount(guide)) // 16
    	fmt.Println("depth:", depth(guide))     // 3
    	fmt.Println(strings.Join(outline(guide), "\n"))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func testGuide() Section {
    	return Section{Title: "Guide", Body: "Welcome to Doc2Doc", Subs: []Section{
    		{Title: "Install", Body: "Pick your platform", Subs: []Section{
    			{Title: "Linux", Body: "Use go install"},
    			{Title: "macOS", Body: "Use Homebrew", Subs: []Section{
    				{Title: "Apple silicon", Body: "Same steps"},
    			}},
    		}},
    		{Title: "Usage", Body: "Run doc2doc on a file"},
    	}}
    }

    func TestWordCount(t *testing.T) {
    	if got := wordCount(Section{Title: "Empty"}); got != 0 {
    		t.Errorf("wordCount(empty section) = %d, want 0", got)
    	}
    	if got := wordCount(Section{Body: "one two three"}); got != 3 {
    		t.Errorf("wordCount(leaf with 3 words) = %d, want 3", got)
    	}
    	if got := wordCount(testGuide()); got != 18 {
    		t.Errorf("wordCount(guide) = %d, want 18 (count every nested level)", got)
    	}
    }

    func TestDepth(t *testing.T) {
    	if got := depth(Section{Title: "Leaf"}); got != 1 {
    		t.Errorf("depth(leaf) = %d, want 1", got)
    	}
    	if got := depth(testGuide()); got != 4 {
    		t.Errorf("depth(guide) = %d, want 4", got)
    	}
    }

    func TestOutline(t *testing.T) {
    	want := []string{
    		"Guide",
    		"  Install",
    		"    Linux",
    		"    macOS",
    		"      Apple silicon",
    		"  Usage",
    	}
    	got := outline(testGuide())
    	if !slices.Equal(got, want) {
    		t.Errorf("outline(guide) =\n%q\nwant\n%q", got, want)
    	}
    	if got := outline(Section{Title: "Solo"}); !slices.Equal(got, []string{"Solo"}) {
    		t.Errorf("outline(leaf) = %q, want [\"Solo\"]", got)
    	}
    }
---

Doc2Doc stores a document as a tree of sections. Each section has a title, some body
text and any number of subsections, which can have subsections of their own. You don't
know how deep it goes, so recursion is the right tool.

## Your task

Complete three recursive functions:

- `wordCount(s)` returns the words in `s.Body` **plus** the words in every nested
  subsection.
- `depth(s)` returns how many levels the tree has. A section with no subsections has
  depth 1.
- `outline(s)` returns every title, depth first, indented by two spaces per level:

```text
Guide
  Install
    Linux
    macOS
  Usage
```

## Tips

- For each function ask: "If the recursive call already works for each subsection, how
  do I combine those answers with this section?"
- The base case can hide inside a loop. When `Subs` is empty, the loop runs zero times.
- For `outline`, you can prefix each line returned by the recursive call with two
  spaces. Each level adds its own two spaces, so deeper sections get more.
- The built-in `max` is handy for `depth`.
