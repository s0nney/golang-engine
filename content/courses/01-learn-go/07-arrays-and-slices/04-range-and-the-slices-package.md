---
title: Range and the slices Package
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	costs := []int{1, 2, 3}
      	for _, c := range costs {
      		c *= 10
      	}
      	fmt.Println(costs)
      }
      ```
    options:
      - text: '`[10 20 30]`'
      - text: '`[1 2 3]`'
        correct: true
      - text: '`[0 0 0]`'
      - text: It doesn't compile, because `c` isn't used
    explanation: |
      The range value `c` is a *copy* of each element. Multiplying the copy
      doesn't change the slice. To modify elements, use the index:
      `for i := range costs { costs[i] *= 10 }`.
  - question: What does `slices.Contains([]string{"US", "GB"}, "gb")` return?
    options:
      - text: '`true`'
      - text: '`false`'
        correct: true
      - text: '`1`'
    explanation: |
      `slices.Contains` compares with `==`, and string comparison is
      case-sensitive. `"gb"` is not equal to `"GB"`, so the result is `false`.
exercise:
  starter: |
    package main

    import "fmt"

    // failedIndexes returns the positions in statuses that are "failed",
    // in order.
    func failedIndexes(statuses []string) []int {
    	var failed []int
    	// ?
    	return failed
    }

    func main() {
    	statuses := []string{"delivered", "failed", "delivered", "failed", "queued"}
    	fmt.Println(failedIndexes(statuses)) // want [1 3]
    }
  solution: |
    package main

    import "fmt"

    func failedIndexes(statuses []string) []int {
    	var failed []int
    	for i, s := range statuses {
    		if s == "failed" {
    			failed = append(failed, i)
    		}
    	}
    	return failed
    }

    func main() {
    	statuses := []string{"delivered", "failed", "delivered", "failed", "queued"}
    	fmt.Println(failedIndexes(statuses))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestFailedIndexes(t *testing.T) {
    	tests := []struct {
    		statuses []string
    		want     []int
    	}{
    		{[]string{"delivered", "failed", "delivered", "failed", "queued"}, []int{1, 3}},
    		{[]string{"failed"}, []int{0}},
    		{[]string{"delivered", "queued"}, nil},
    		{nil, nil},
    		{[]string{"failed", "failed", "Failed"}, []int{0, 1}},
    	}
    	for _, tt := range tests {
    		got := failedIndexes(tt.statuses)
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("failedIndexes(%q) = %v, want %v", tt.statuses, got, tt.want)
    		}
    	}
    }
---

You'll loop over slices constantly. Go gives you a special form of `for` for it, and the standard library's `slices` package handles the most common jobs so you don't have to write the loops yourself.

## `for ... range`

`range` over a slice gives you each **index** and **value** in turn:

```go
package main

import "fmt"

func main() {
	recipients := []string{"Alice", "Bob", "Carol"}
	for i, name := range recipients {
		fmt.Println(i, name)
	}
}
```

```text
0 Alice
1 Bob
2 Carol
```

If you only need the values, discard the index with `_`. If you only need the index, leave off the second variable:

```go
for _, name := range recipients { ... } // values only
for i := range recipients { ... }       // indexes only
```

## The value is a copy

The value variable is a **copy** of the element. Changing it doesn't change the slice:

```go
prices := []float64{0.01, 0.02}
for _, p := range prices {
	p *= 2 // changes the copy only
}
// prices is still [0.01 0.02]
```

To modify elements in place, use the index:

```go
package main

import "fmt"

func main() {
	prices := []float64{0.01, 0.02}
	for i := range prices {
		prices[i] *= 2
	}
	fmt.Println(prices)
}
```

```text
[0.02 0.04]
```

## The `slices` package

Loads of everyday list jobs, like searching, sorting and finding the biggest element, are already written for you in the standard library's `slices` package:

```go
package main

import (
	"fmt"
	"slices"
)

func main() {
	lengths := []int{42, 160, 7, 88}

	fmt.Println(slices.Contains(lengths, 7))
	fmt.Println(slices.Index(lengths, 160))
	fmt.Println(slices.Max(lengths))

	slices.Sort(lengths)
	fmt.Println(lengths)
}
```

```text
true
1
160
[7 42 88 160]
```

| Function | What it does |
|----------|--------------|
| `slices.Contains(s, v)` | Reports whether `v` is in `s` |
| `slices.Index(s, v)` | Index of the first `v`, or `-1` if it's missing |
| `slices.Min(s)`, `slices.Max(s)` | Smallest or largest element (panics on an empty slice) |
| `slices.Sort(s)` | Sorts `s` in place, smallest first |
| `slices.Reverse(s)` | Reverses `s` in place |
| `slices.Equal(a, b)` | Reports whether two slices have the same elements |
| `slices.Clone(s)` | Returns an independent copy |

`slices.Sort` changes the slice you pass in (the "in place" part), which works because slices share their underlying array with the caller.

## Comparing slices

You can't compare two slices with `==` (only against `nil`). Use `slices.Equal`:

```go
a := []string{"hi"}
b := []string{"hi"}
// fmt.Println(a == b) // error: slice can only be compared to nil
fmt.Println(slices.Equal(a, b)) // true
```

## Removing elements

`slices.Delete` removes a range of elements and returns the shortened slice. Just like `append`, you must use the result:

```go
package main

import (
	"fmt"
	"slices"
)

func main() {
	queue := []string{"a", "b", "c", "d"}
	queue = slices.Delete(queue, 1, 3) // remove indexes 1 and 2
	fmt.Println(queue)
}
```

```text
[a d]
```

When you find yourself writing a loop to search or sort a slice, check the `slices` package first. The code will be shorter, clearer and well tested.

## Your turn

Textio's retry service needs to know which messages in a batch failed. Complete `failedIndexes` so it returns the **index** of every status that is exactly `"failed"`, in order.

Use `for ... range` to get both the index and the value, and `append` to build the result. (Remember to assign `append`'s result back!)

## Further reading

- [Go by Example: Range over Built-in Types](https://gobyexample.com/range-over-built-in-types)
- [Go by Example: Sorting](https://gobyexample.com/sorting)
