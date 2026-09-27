---
title: Nil Pointers
quiz:
  - question: |
      What happens when this program runs?

      ```go
      package main

      import "fmt"

      type user struct {
      	name string
      }

      func main() {
      	var u *user
      	fmt.Println(u == nil)
      	fmt.Println(u.name)
      }
      ```
    options:
      - text: It prints `true` and then an empty line
      - text: It prints `true` and then panics
        correct: true
      - text: It doesn't compile
      - text: It prints `false` and then an empty line
    explanation: |
      `u` is a nil pointer, so the first line prints `true`. Reading
      `u.name` means following the pointer, and there's nothing to follow,
      so the program panics with a nil pointer dereference.
  - question: Which is the safest way to write a function that takes a `*user` that might be `nil`?
    options:
      - text: Check `if u == nil` at the top and return early
        correct: true
      - text: Read `u.name` and hope for the best
      - text: Use `*u` instead of `u.name`, because it's safer
    explanation: |
      A guard clause that handles `nil` first keeps the rest of the function
      safe. `*u` is also a dereference, so it panics just the same.
exercise:
  starter: |
    package main

    import "fmt"

    type user struct {
    	name    string
    	credits int
    }

    // totalCredits adds up the credits of every user, skipping nil pointers.
    func totalCredits(users []*user) int {
    	total := 0
    	for _, u := range users {
    		total += u.credits
    	}
    	return total
    }

    func main() {
    	users := []*user{{name: "alice", credits: 12}, nil, {name: "bob", credits: 3}}
    	fmt.Println(totalCredits(users))
    }
  solution: |
    package main

    import "fmt"

    type user struct {
    	name    string
    	credits int
    }

    func totalCredits(users []*user) int {
    	total := 0
    	for _, u := range users {
    		if u == nil {
    			continue
    		}
    		total += u.credits
    	}
    	return total
    }

    func main() {
    	users := []*user{{name: "alice", credits: 12}, nil, {name: "bob", credits: 3}}
    	fmt.Println(totalCredits(users))
    }
  tests: |
    package main

    import "testing"

    func TestTotalCredits(t *testing.T) {
    	for _, tc := range []struct {
    		name  string
    		users []*user
    		want  int
    	}{
    		{"no nils", []*user{{"alice", 12}, {"bob", 3}}, 15},
    		{"nil in the middle", []*user{{"alice", 12}, nil, {"bob", 3}}, 15},
    		{"only nils", []*user{nil, nil}, 0},
    		{"nil slice", nil, 0},
    	} {
    		t.Run(tc.name, func(t *testing.T) {
    			defer func() {
    				if r := recover(); r != nil {
    					t.Fatalf("totalCredits panicked: %v (did you check for nil?)", r)
    				}
    			}()
    			if got := totalCredits(tc.users); got != tc.want {
    				t.Errorf("totalCredits = %d, want %d", got, tc.want)
    			}
    		})
    	}
    }
---

A pointer can point at nothing. Its value is then `nil`. Following a nil pointer is the most famous crash in Go, and in plenty of other languages too. The inventor of the null reference, Tony Hoare, called it his "billion-dollar mistake".

## The crash

```go
package main

import "fmt"

type user struct {
	name    string
	credits int
}

func main() {
	var u *user // nil: doesn't point at any user
	fmt.Println("about to read credits")
	fmt.Println(u.credits)
	fmt.Println("never printed")
}
```

```text
about to read credits
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x10 pc=0x...]

goroutine 1 [running]:
main.main()
	/tmp/main.go:13 +0x...
exit status 2
```

A **panic** stops the program immediately and prints what went wrong plus a **stack trace**: the chain of function calls that led to the crash, with file names and line numbers. Line 13 is `fmt.Println(u.credits)`. Always look for the first line that points into *your* code.

## Where nil pointers come from

- A pointer variable declared without a value: `var u *user`.
- A struct field of pointer type that was never set.
- A function that returns `nil` to mean "not found":

```go
func findUser(users map[string]*user, name string) *user {
	return users[name] // nil if name isn't in the map
}
```

## Check before you follow

The fix is almost always a guard clause:

```go
package main

import "fmt"

type user struct {
	name    string
	credits int
}

func describe(u *user) string {
	if u == nil {
		return "no such user"
	}
	return fmt.Sprintf("%s has %d credits", u.name, u.credits)
}

func main() {
	users := map[string]*user{
		"alice": {name: "alice", credits: 12},
	}
	fmt.Println(describe(users["alice"]))
	fmt.Println(describe(users["bob"]))
}
```

```text
alice has 12 credits
no such user
```

(Inside a map literal of pointers, you can write `{name: ...}` and Go fills in the `&user` for you.)

## Methods on nil pointers

Here's a surprise: calling a pointer-receiver method on a nil pointer does **not** crash by itself. The method runs with a nil receiver. It only crashes if the method then follows the pointer. A method can even handle `nil` gracefully:

```go
package main

import "fmt"

type user struct {
	name string
}

func (u *user) displayName() string {
	if u == nil {
		return "(unknown)"
	}
	return u.name
}

func main() {
	var u *user
	fmt.Println(u.displayName())
}
```

```text
(unknown)
```

Don't rely on this everywhere, but it's a neat trick for types where "nothing" has a sensible behaviour.

## Prefer values when you can

Every pointer is a potential nil. So don't use pointers just because you can. If a function only needs to *read* a small struct, pass it by value: a value can never be nil. Use pointers when you need to share or modify, or when "no value" is a meaningful answer.

## Your turn

Textio's user list can contain `nil` entries for accounts that were deleted.
Press **Run** and read the panic: which line does the stack trace point at?

Fix `totalCredits` with a guard clause so it skips `nil` users instead of
crashing. **Run** should then print `15`.

## Further reading

- [A Tour of Go: Pointers to structs](https://go.dev/tour/moretypes/4)
