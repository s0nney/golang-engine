---
title: Comma-ok and delete
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      func main() {
      	credits := map[string]int{"alice": 0}
      	a, okA := credits["alice"]
      	b, okB := credits["bob"]
      	fmt.Println(a, okA, b, okB)
      }
      ```
    options:
      - text: '`0 true 0 false`'
        correct: true
      - text: '`0 false 0 false`'
      - text: '`0 true nil false`'
      - text: It panics, because `"bob"` isn't in the map
    explanation: |
      Both lookups return the value `0`, but for different reasons. Alice is
      in the map with a balance of 0, so `okA` is `true`. Bob isn't in the
      map at all, so `b` is the zero value and `okB` is `false`.
  - question: What happens if you call `delete(m, "zed")` and `"zed"` isn't in `m`?
    options:
      - text: It panics
      - text: It returns an error
      - text: Nothing, it's a no-op
        correct: true
    explanation: |
      Deleting a key that doesn't exist is safe and does nothing. (It's even
      safe on a nil map.)
exercise:
  starter: |
    package main

    import "fmt"

    // totalCredits adds up the credits of every user in users.
    // Users that aren't in the credits map are returned in missing,
    // in the order they appear in users.
    func totalCredits(credits map[string]int, users []string) (total int, missing []string) {
    	for _, u := range users {
    		total += credits[u] // ? what about users that don't exist?
    	}
    	return total, missing
    }

    func main() {
    	credits := map[string]int{"alice": 50, "bob": 0}
    	total, missing := totalCredits(credits, []string{"alice", "bob", "mallory"})
    	fmt.Println(total, missing) // want 50 [mallory]
    }
  solution: |
    package main

    import "fmt"

    func totalCredits(credits map[string]int, users []string) (total int, missing []string) {
    	for _, u := range users {
    		c, ok := credits[u]
    		if !ok {
    			missing = append(missing, u)
    			continue
    		}
    		total += c
    	}
    	return total, missing
    }

    func main() {
    	credits := map[string]int{"alice": 50, "bob": 0}
    	total, missing := totalCredits(credits, []string{"alice", "bob", "mallory"})
    	fmt.Println(total, missing)
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func TestTotalCredits(t *testing.T) {
    	credits := map[string]int{"alice": 50, "bob": 0, "carol": 25}
    	tests := []struct {
    		users       []string
    		wantTotal   int
    		wantMissing []string
    	}{
    		{[]string{"alice", "bob", "mallory"}, 50, []string{"mallory"}},
    		{[]string{"alice", "carol"}, 75, nil},
    		{[]string{"bob"}, 0, nil},
    		{[]string{"zed", "alice", "yan"}, 50, []string{"zed", "yan"}},
    		{nil, 0, nil},
    	}
    	for _, tt := range tests {
    		total, missing := totalCredits(credits, tt.users)
    		if total != tt.wantTotal || !slices.Equal(missing, tt.wantMissing) {
    			t.Errorf("totalCredits(credits, %q) = %d, %q; want %d, %q",
    				tt.users, total, missing, tt.wantTotal, tt.wantMissing)
    		}
    	}
    }
---

A missing key gives you the zero value. That's convenient, but sometimes you need to know whether a key is really there. Is Bob a customer with zero credits, or not a customer at all?

## The comma-ok idiom

A map lookup can return a second value: a `bool` that says whether the key was found. By convention it's named `ok`:

```go
package main

import "fmt"

func main() {
	credits := map[string]int{
		"alice": 50,
		"bob":   0,
	}

	c, ok := credits["bob"]
	fmt.Println(c, ok)

	c, ok = credits["mallory"]
	fmt.Println(c, ok)
}
```

```text
0 true
0 false
```

Both lookups return `0`, but `ok` tells them apart. This pattern is called the **comma-ok idiom**, and you'll see it in several places in Go.

It pairs perfectly with `if`'s init statement, keeping `ok` scoped to the check:

```go
package main

import "fmt"

func main() {
	phoneNumbers := map[string]string{"alice": "+1-555-0100"}

	if number, ok := phoneNumbers["bob"]; ok {
		fmt.Println("sending to", number)
	} else {
		fmt.Println("bob has no phone number on file")
	}
}
```

```text
bob has no phone number on file
```

If you only care whether the key exists, discard the value: `_, ok := m[key]`.

## Removing keys with `delete`

The built-in `delete` removes a key and its value:

```go
package main

import "fmt"

func main() {
	subscribers := map[string]string{
		"alice": "+1-555-0100",
		"bob":   "+1-555-0199",
	}

	delete(subscribers, "bob") // Bob replied STOP
	delete(subscribers, "zed") // not there: does nothing

	fmt.Println(subscribers, len(subscribers))
}
```

```text
map[alice:+1-555-0100] 1
```

Deleting a key that isn't there is a no-op. No panic, no error.

To remove *every* key at once, use the built-in `clear`:

```go
clear(subscribers) // now empty, len 0
```

## Maps are references

Like slices, a map value is a small handle pointing at the real data. When you pass a map to a function, the function gets a copy of the handle, but it points at the **same** map. Changes made inside the function are visible outside:

```go
package main

import "fmt"

func charge(credits map[string]int, user string, cost int) {
	credits[user] -= cost
}

func main() {
	credits := map[string]int{"alice": 50}
	charge(credits, "alice", 5)
	fmt.Println(credits["alice"])
}
```

```text
45
```

Assigning a map to another variable does the same. Both names refer to one map:

```go
backup := credits
backup["alice"] = 0 // credits["alice"] is now 0 too!
```

If you need an independent copy, use `maps.Clone`, which you'll meet in the next lesson.

## Putting it together

Here's a small function that uses comma-ok as a guard clause:

```go
func sendCode(phoneNumbers map[string]string, user string) string {
	number, ok := phoneNumbers[user]
	if !ok {
		return "unknown user " + user
	}
	return "code sent to " + number
}
```

Missing data is handled first, and the happy path stays on the left.

## Your turn

Complete `totalCredits`. It should add up the credits of every user in `users`, and also report which users aren't customers at all.

The tricky part: Bob has `0` credits, but he *is* a customer, so he must not appear in `missing`. A plain `credits[u]` lookup can't tell Bob apart from a stranger. Use the comma-ok idiom.

## Further reading

- [A Tour of Go: Mutating Maps](https://go.dev/tour/moretypes/22)
