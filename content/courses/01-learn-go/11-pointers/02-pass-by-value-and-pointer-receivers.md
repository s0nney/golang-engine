---
title: Pass by Value and Pointer Receivers
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import "fmt"

      type account struct {
      	credits int
      }

      func (a account) addValue(n int)    { a.credits += n }
      func (a *account) addPointer(n int) { a.credits += n }

      func main() {
      	acc := account{credits: 10}
      	acc.addValue(5)
      	acc.addPointer(1)
      	fmt.Println(acc.credits)
      }
      ```
    options:
      - text: '`16`'
      - text: '`15`'
      - text: '`11`'
        correct: true
      - text: '`10`'
    explanation: |
      `addValue` has a value receiver, so it changes a copy that is thrown
      away. `addPointer` has a pointer receiver, so it changes `acc` itself.
      Only the `+1` sticks: 11.
  - question: 'Why can you call `acc.addPointer(1)` when `acc` is an `account`, not a `*account`?'
    options:
      - text: Because Go automatically takes the address, turning it into `(&acc).addPointer(1)`
        correct: true
      - text: Because pointer receivers work on copies anyway
      - text: You can't; it doesn't compile
    explanation: |
      When a method has a pointer receiver and you call it on an addressable
      variable, Go inserts the `&` for you. That's why method calls look the
      same either way.
exercise:
  starter: |
    package main

    import "fmt"

    type account struct {
    	owner   string
    	credits int
    }

    // charge takes cost credits from the account and reports true.
    // If the account can't afford it, charge leaves the credits
    // alone and reports false.
    func (a account) charge(cost int) bool {
    	if cost > a.credits {
    		return false
    	}
    	a.credits -= cost
    	return true
    }

    func main() {
    	acc := account{owner: "alice", credits: 10}
    	fmt.Println(acc.charge(3), acc.credits)  // want true 7
    	fmt.Println(acc.charge(50), acc.credits) // want false 7
    }
  solution: |
    package main

    import "fmt"

    type account struct {
    	owner   string
    	credits int
    }

    func (a *account) charge(cost int) bool {
    	if cost > a.credits {
    		return false
    	}
    	a.credits -= cost
    	return true
    }

    func main() {
    	acc := account{owner: "alice", credits: 10}
    	fmt.Println(acc.charge(3), acc.credits)
    	fmt.Println(acc.charge(50), acc.credits)
    }
  tests: |
    package main

    import "testing"

    func TestCharge(t *testing.T) {
    	acc := account{owner: "alice", credits: 10}

    	steps := []struct {
    		cost        int
    		wantOK      bool
    		wantCredits int
    	}{
    		{3, true, 7},
    		{50, false, 7},
    		{7, true, 0},
    		{1, false, 0},
    		{0, true, 0},
    	}
    	for _, s := range steps {
    		before := acc.credits
    		ok := acc.charge(s.cost)
    		if ok != s.wantOK || acc.credits != s.wantCredits {
    			t.Fatalf("with %d credits, charge(%d) = %v leaving %d credits; want %v leaving %d",
    				before, s.cost, ok, acc.credits, s.wantOK, s.wantCredits)
    		}
    	}
    }
---

Here's a rule that explains a huge amount of Go's behaviour: **Go always passes arguments by value.** Every time you call a function, it gets a *copy* of each argument.

## Functions get copies

```go
package main

import "fmt"

func refund(credits int) {
	credits += 10
}

func main() {
	balance := 5
	refund(balance)
	fmt.Println(balance)
}
```

```text
5
```

`refund` added 10 to its own copy, which disappeared when the function returned. `balance` never changed.

Structs are copied the same way, every field of them.

## Pass a pointer to share

To let a function change the caller's variable, pass a **pointer** to it. The pointer itself is still copied, but the copy points at the same variable:

```go
package main

import "fmt"

func refund(credits *int) {
	*credits += 10
}

func main() {
	balance := 5
	refund(&balance)
	fmt.Println(balance)
}
```

```text
15
```

(Slices and maps already contain a pointer to their data internally, which is why functions can change their *elements* without you passing a pointer.)

## Pointer receivers

The same goes for methods. A **value receiver** gets a copy of the struct. A **pointer receiver**, written `(a *account)`, gets a pointer to the original, so the method can change it:

```go
package main

import "fmt"

type account struct {
	owner   string
	credits int
}

// Pointer receiver: can modify the account.
func (a *account) charge(cost int) {
	a.credits -= cost
}

// Only reads the account, but uses a pointer receiver
// too, to stay consistent with charge.
func (a *account) summary() string {
	return fmt.Sprintf("%s has %d credits", a.owner, a.credits)
}

func main() {
	acc := account{owner: "alice", credits: 50}
	acc.charge(3)
	acc.charge(2)
	fmt.Println(acc.summary())
}
```

```text
alice has 45 credits
```

Notice we wrote `acc.charge(3)`, not `(&acc).charge(3)`. When you call a pointer-receiver method on a variable, Go takes the address for you. It works the other way too: if you have a pointer, you can call value-receiver methods on it and Go follows the pointer automatically.

## The classic bug

Forgetting the `*` is one of the most common Go bugs. It compiles, runs and silently does nothing:

```go
func (a account) charge(cost int) {
	a.credits -= cost // modifies a copy. Oops!
}
```

If a method needs to modify its receiver, it **must** use a pointer receiver.

## Which should you use?

- Use a **pointer receiver** if the method changes the receiver.
- Use a **pointer receiver** if the struct is large, to avoid copying it on every call.
- Use a **pointer receiver** if the struct contains something that must not be copied, such as a `sync.Mutex` (you'll meet those in the concurrency chapter).
- Otherwise a value receiver is fine, especially for small types.

And be consistent: if any method on a type needs a pointer receiver, Go style is to give **all** its methods pointer receivers. Linters and code reviewers will nudge you towards that.

## Returning a pointer

Functions that create a struct often return a pointer to it. By convention these are called `newThing`:

```go
func newAccount(owner string) *account {
	return &account{owner: owner, credits: 10}
}
```

Returning the address of a local variable is perfectly safe in Go. The compiler notices the value outlives the function and keeps it alive. You never need to free memory yourself, because Go's **garbage collector** cleans up values that nothing points to any more.

## Your turn

Customers are getting free messages! `charge` reports success, but the credits never go down. Run the program to see it.

Fix `charge` so it really changes the account. You shouldn't need to touch `main` or the logic inside the method.

## Further reading

- [A Tour of Go: Pointer receivers](https://go.dev/tour/methods/4)
- [Learn Go with Tests: Pointers & errors](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/pointers-and-errors)
