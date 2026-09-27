---
title: First Repeat
difficulty: easy
after: type-sets-and-constraints
hints:
  - '`any` lets you accept every type, but it doesn''t let you use `==` or make a map key from `T`. The `comparable` constraint does: `func FirstRepeat[T comparable](xs []T) (T, bool)`.'
  - 'Remember what you''ve seen in a `map[T]bool` (or `map[T]struct{}`). The first element that is already in the map is the answer.'
  - 'When nothing repeats, return the zero value of `T`: declare `var zero T` and return `zero, false`.'
exercise:
  starter: |
    package main

    import "fmt"

    // FirstRepeat returns the first element of xs that is equal to an earlier
    // element, and true. If every element is distinct, it returns the zero
    // value of T and false.
    //
    // Fix the constraint first: with `any` you can't compare two T values.
    func FirstRepeat[T any](xs []T) (T, bool) {
    	// 1. Keep a set of the values you've already seen.
    	// 2. Return the first value that's already in the set.
    	var zero T
    	return zero, false
    }

    func main() {
    	fmt.Println(FirstRepeat([]string{"k1", "k2", "k3", "k2", "k1"})) // want k2 true
    	fmt.Println(FirstRepeat([]int{1, 2, 3}))                         // want 0 false
    }
  solution: |
    package main

    import "fmt"

    // FirstRepeat returns the first element of xs that is equal to an earlier
    // element, and true. If every element is distinct, it returns the zero
    // value of T and false.
    func FirstRepeat[T comparable](xs []T) (T, bool) {
    	seen := make(map[T]struct{}, len(xs))
    	for _, x := range xs {
    		if _, ok := seen[x]; ok {
    			return x, true
    		}
    		seen[x] = struct{}{}
    	}
    	var zero T
    	return zero, false
    }

    func main() {
    	fmt.Println(FirstRepeat([]string{"k1", "k2", "k3", "k2", "k1"}))
    	fmt.Println(FirstRepeat([]int{1, 2, 3}))
    }
  tests: |
    package main

    import (
    	"testing"
    	"time"
    )

    type Key string

    type Point struct{ X, Y int }

    func TestFirstRepeatStrings(t *testing.T) {
    	tests := []struct {
    		in     []string
    		want   string
    		wantOK bool
    	}{
    		{nil, "", false},
    		{[]string{"a"}, "", false},
    		{[]string{"a", "a"}, "a", true},
    		{[]string{"k1", "k2", "k3", "k2", "k1"}, "k2", true},
    		{[]string{"x", "y", "z", "z", "x"}, "z", true},
    		{[]string{"", "", "q"}, "", true},
    	}
    	for _, tt := range tests {
    		got, ok := FirstRepeat(tt.in)
    		if got != tt.want || ok != tt.wantOK {
    			t.Errorf("FirstRepeat(%q) = %q, %v, want %q, %v", tt.in, got, ok, tt.want, tt.wantOK)
    		}
    	}
    }

    func TestFirstRepeatOtherTypes(t *testing.T) {
    	if got, ok := FirstRepeat([]int{3, 1, 4, 1, 5}); got != 1 || !ok {
    		t.Errorf("FirstRepeat([3 1 4 1 5]) = %v, %v, want 1, true", got, ok)
    	}
    	if got, ok := FirstRepeat([]int{3, 1, 4}); got != 0 || ok {
    		t.Errorf("FirstRepeat([3 1 4]) = %v, %v, want 0, false", got, ok)
    	}
    	if got, ok := FirstRepeat([]Key{"user:1", "user:2", "user:1"}); got != "user:1" || !ok {
    		t.Errorf("FirstRepeat([]Key{user:1 user:2 user:1}) = %q, %v, want user:1, true", got, ok)
    	}
    	pts := []Point{{1, 2}, {2, 1}, {0, 0}, {2, 1}}
    	if got, ok := FirstRepeat(pts); got != (Point{2, 1}) || !ok {
    		t.Errorf("FirstRepeat(%v) = %v, %v, want {2 1}, true", pts, got, ok)
    	}
    	if got, ok := FirstRepeat([]Point{{1, 2}, {2, 1}}); got != (Point{}) || ok {
    		t.Errorf("FirstRepeat([{1 2} {2 1}]) = %v, %v, want {0 0}, false", got, ok)
    	}
    	arrs := [][2]string{{"a", "b"}, {"b", "a"}, {"a", "b"}}
    	if got, ok := FirstRepeat(arrs); got != [2]string{"a", "b"} || !ok {
    		t.Errorf("FirstRepeat(%q) = %q, %v, want [a b], true", arrs, got, ok)
    	}
    }

    func TestFirstRepeatPointersCompareByIdentity(t *testing.T) {
    	a, b := &Point{1, 1}, &Point{1, 1}
    	if got, ok := FirstRepeat([]*Point{a, b}); got != nil || ok {
    		t.Errorf("FirstRepeat([a b]) with two different pointers to equal Points = %v, %v, want nil, false", got, ok)
    	}
    	if got, ok := FirstRepeat([]*Point{a, b, a}); got != a || !ok {
    		t.Errorf("FirstRepeat([a b a]) = %p, %v, want a (%p), true", got, ok, a)
    	}
    }

    func TestFirstRepeatAny(t *testing.T) {
    	// Since Go 1.20, any satisfies comparable. 1 and "1" are different values.
    	in := []any{1, "1", 1.0, 1}
    	if got, ok := FirstRepeat(in); got != any(1) || !ok {
    		t.Errorf("FirstRepeat(%#v) = %#v, %v, want 1, true", in, got, ok)
    	}
    }

    func TestFirstRepeatLarge(t *testing.T) {
    	n := 100_000
    	in := make([]int, n)
    	for i := range in {
    		in[i] = i * 3
    	}
    	in[n-1] = in[n-2]
    	start := time.Now()
    	got, ok := FirstRepeat(in)
    	if got != in[n-2] || !ok {
    		t.Errorf("FirstRepeat(100,000 ints) = %v, %v, want %v, true", got, ok, in[n-2])
    	}
    	if d := time.Since(start); d > time.Second {
    		t.Errorf("FirstRepeat(100,000 ints) took %v: remember what you've seen in a map instead of comparing every pair", d)
    	}
    }
---

Stash's bulk importer rejects a batch that contains a duplicate key, and the
error message should name the **first** repeat so the user can find it.

Write `FirstRepeat(xs)`. It returns the first element of `xs` that is equal to
an element earlier in the slice, and `true`. "First" means the repeat that
shows up earliest while reading left to right. If all elements are distinct,
return the zero value of the element type and `false`.

The function must work for every element type that supports `==`: strings,
numbers, your own named types, structs, arrays, pointers and even `any`.
The starter's constraint is `any`, which is too loose. Fix it.

## Examples

```go
FirstRepeat([]string{"k1", "k2", "k3", "k2", "k1"})  // "k2", true
FirstRepeat([]Point{{1, 2}, {2, 1}, {0, 0}, {2, 1}}) // {2 1}, true
FirstRepeat([]int{1, 2, 3})                          // 0, false
```

## Constraints

- Up to 100,000 elements. Comparing every pair is too slow: aim for O(n).
- Pointers are equal only when they point at the same variable, just like `==`.
