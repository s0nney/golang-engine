---
title: Footnote Numbers
difficulty: easy
after: closures
hints:
  - 'The state (a map from key to number, and the next number to hand out) must live **inside** `footnoter`, declared before the `return func(...)`. The returned closure captures it.'
  - 'Inside the closure: if the key is already in the map, return its number. Otherwise store the next number for it, bump the counter, and return it.'
  - 'Don''t use package-level variables: then every numbering would share one counter, and a second document would start at 4 instead of 1.'
exercise:
  starter: |
    package main

    import "fmt"

    // footnoter returns a function that numbers footnote keys in the order
    // they are first seen: the first new key gets 1, the next new key 2, and
    // so on. Asking again for a key that was already seen returns its
    // existing number. Each footnoter numbers on its own.
    func footnoter() func(key string) int {
    	// Declare the closure's state here (before returning the function).
    	return func(key string) int {
    		// ?
    		return 0
    	}
    }

    func main() {
    	note := footnoter()
    	fmt.Println(note("rfc"), note("spec"), note("rfc"), note("blog")) // want: 1 2 1 3
    }
  solution: |
    package main

    import "fmt"

    func footnoter() func(key string) int {
    	numbers := map[string]int{}
    	return func(key string) int {
    		if n, ok := numbers[key]; ok {
    			return n
    		}
    		n := len(numbers) + 1
    		numbers[key] = n
    		return n
    	}
    }

    func main() {
    	note := footnoter()
    	fmt.Println(note("rfc"), note("spec"), note("rfc"), note("blog"))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func numberAll(note func(string) int, keys []string) []int {
    	var got []int
    	for _, k := range keys {
    		got = append(got, note(k))
    	}
    	return got
    }

    func TestFootnoter(t *testing.T) {
    	tests := []struct {
    		keys []string
    		want []int
    	}{
    		{[]string{"rfc", "spec", "rfc", "blog"}, []int{1, 2, 1, 3}},
    		{[]string{"a", "b", "c"}, []int{1, 2, 3}},
    		{[]string{"x", "x", "x"}, []int{1, 1, 1}},
    		{[]string{"b", "a", "b", "a", "c", "a"}, []int{1, 2, 1, 2, 3, 2}},
    		{[]string{"", "empty", ""}, []int{1, 2, 1}},
    		{[]string{"Case", "case"}, []int{1, 2}},
    	}
    	for _, tt := range tests {
    		if got := numberAll(footnoter(), tt.keys); !slices.Equal(got, tt.want) {
    			t.Errorf("a new footnoter given %q returned %v, want %v", tt.keys, got, tt.want)
    		}
    	}
    }

    func TestFootnotersAreIndependent(t *testing.T) {
    	doc1, doc2 := footnoter(), footnoter()
    	doc1("a")
    	doc1("b")
    	doc1("c")
    	if got := doc2("z"); got != 1 {
    		t.Errorf("second footnoter's first key got %d, want 1: each footnoter must have its own state", got)
    	}
    	if got := doc2("a"); got != 2 {
    		t.Errorf("second footnoter: key \"a\" (seen only by the first footnoter) got %d, want 2", got)
    	}
    	if got := doc1("d"); got != 4 {
    		t.Errorf("first footnoter's 4th new key got %d after using the second one, want 4", got)
    	}
    }

    func TestFootnoterManyKeys(t *testing.T) {
    	note := footnoter()
    	for i := range 1000 {
    		key := string(rune('A'+i%26)) + string(rune('a'+i/26))
    		if got := note(key); got != i+1 {
    			t.Fatalf("new key #%d (%q) got %d, want %d", i+1, key, got, i+1)
    		}
    	}
    }
---

Doc2Doc turns inline citations like `[^rfc]` into numbered footnotes. The
first source cited becomes footnote 1, the next *new* one becomes 2, and
citing an earlier source again reuses its number.

Complete `footnoter()`. It returns a **closure** that takes a footnote key and
returns its number:

- a key seen for the first time gets the next number, starting at 1;
- a key that was seen before gets the same number as last time;
- every call to `footnoter()` starts a fresh numbering, so two documents
  converted side by side don't share numbers.

## Example

```go
note := footnoter()
note("rfc")  // 1
note("spec") // 2
note("rfc")  // 1 (seen before)
note("blog") // 3

other := footnoter()
other("blog") // 1 (a fresh numbering)
```

## Constraints

- Keys are compared exactly (`"Case"` and `"case"` are different keys; `""` is a valid key).
- No package-level variables.
