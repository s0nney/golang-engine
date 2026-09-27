---
title: 'Review: Errors, Concurrency and Generics'
quiz:
  - question: |
      What does this program print?

      ```go
      package main

      import (
      	"errors"
      	"fmt"
      )

      var ErrNoCredits = errors.New("no credits")

      func charge(user string) error {
      	return fmt.Errorf("charge %s: %w", user, ErrNoCredits)
      }

      func main() {
      	err := charge("bob")
      	fmt.Println(err == ErrNoCredits, errors.Is(err, ErrNoCredits))
      }
      ```
    options:
      - text: '`true true`'
      - text: '`false true`'
        correct: true
      - text: '`false false`'
      - text: '`true false`'
    explanation: |
      Wrapping with `%w` creates a new error, so `==` no longer matches the
      sentinel. `errors.Is` unwraps the chain and finds `ErrNoCredits`.
  - question: |
      This program sometimes prints less than `100`. Why?

      ```go
      var total int
      var wg sync.WaitGroup
      for range 100 {
      	wg.Go(func() { total++ })
      }
      wg.Wait()
      fmt.Println(total)
      ```
    options:
      - text: '`wg.Wait` returns before all goroutines have finished'
      - text: Goroutines update `total` at the same time without a lock, which is a data race
        correct: true
      - text: '`for range 100` only runs 99 times'
    explanation: |
      `wg.Wait` does wait for every goroutine. The problem is that
      `total++` is a read-modify-write, and goroutines interleave. Protect
      `total` with a `sync.Mutex`, or send results over a channel.
  - question: 'Which is the right constraint for `func maxOf[T ???](a, b T) T` that uses `>`?'
    options:
      - text: '`any`'
      - text: '`comparable`'
      - text: '`cmp.Ordered`'
        correct: true
    explanation: |
      `>` needs an ordered type. `cmp.Ordered` covers integers, floats and
      strings. `comparable` only allows `==` and `!=`, and `any` allows
      neither.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strings"
    )

    var ErrInvalidPhone = errors.New("invalid phone number")

    func checkPhone(phone string) error {
    	if !strings.HasPrefix(phone, "+") {
    		return ErrInvalidPhone
    	}
    	return nil
    }

    // firstFailure runs check on every item and returns how many failed, plus
    // the first failure wrapped as "item <index>: <error>" (nil if none failed).
    func firstFailure(items []string, check func(string) error) (int, error) {
    	// ?
    	return 0, nil
    }

    func main() {
    	phones := []string{"+1-555-0100", "555-0199", "+44-20-7946-0000", "0800"}
    	n, err := firstFailure(phones, checkPhone)
    	fmt.Println(n, err, errors.Is(err, ErrInvalidPhone))
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"strings"
    )

    var ErrInvalidPhone = errors.New("invalid phone number")

    func checkPhone(phone string) error {
    	if !strings.HasPrefix(phone, "+") {
    		return ErrInvalidPhone
    	}
    	return nil
    }

    func firstFailure[T any](items []T, check func(T) error) (int, error) {
    	failed := 0
    	var first error
    	for i, item := range items {
    		if err := check(item); err != nil {
    			if first == nil {
    				first = fmt.Errorf("item %d: %w", i, err)
    			}
    			failed++
    		}
    	}
    	return failed, first
    }

    func main() {
    	phones := []string{"+1-555-0100", "555-0199", "+44-20-7946-0000", "0800"}
    	n, err := firstFailure(phones, checkPhone)
    	fmt.Println(n, err, errors.Is(err, ErrInvalidPhone))
    }
  tests: |
    package main

    import (
    	"errors"
    	"testing"
    )

    func TestFirstFailurePhones(t *testing.T) {
    	phones := []string{"+1-555-0100", "555-0199", "+44-20-7946-0000", "0800"}
    	n, err := firstFailure(phones, checkPhone)
    	if n != 2 {
    		t.Errorf("firstFailure(%q) count = %d, want 2", phones, n)
    	}
    	if err == nil || err.Error() != "item 1: invalid phone number" {
    		t.Errorf(`firstFailure(%q) error = %v, want "item 1: invalid phone number"`, phones, err)
    	}
    	if !errors.Is(err, ErrInvalidPhone) {
    		t.Errorf("errors.Is(err, ErrInvalidPhone) = false; wrap the error with %%w")
    	}

    	n, err = firstFailure([]string{"+1", "+2"}, checkPhone)
    	if n != 0 || err != nil {
    		t.Errorf(`firstFailure(["+1" "+2"]) = %d, %v; want 0, <nil>`, n, err)
    	}
    }

    var errTooBig = errors.New("too many segments")

    func TestFirstFailureGeneric(t *testing.T) {
    	segments := []int{1, 5, 2, 9, 4}
    	n, err := firstFailure(segments, func(s int) error {
    		if s > 3 {
    			return errTooBig
    		}
    		return nil
    	})
    	if n != 3 || err == nil || err.Error() != "item 1: too many segments" || !errors.Is(err, errTooBig) {
    		t.Errorf("firstFailure(%v, s > 3) = %d, %v; want 3, item 1: too many segments", segments, n, err)
    	}
    	if n, err := firstFailure([]int(nil), func(int) error { return errTooBig }); n != 0 || err != nil {
    		t.Errorf("firstFailure(nil) = %d, %v; want 0, <nil>", n, err)
    	}
    }
---

The last review covers the chapters that make Go feel like *Go*: errors as values, packages and tests, goroutines and channels, and generics.

## Quick recap

This final example ties them together. It checks several phone numbers concurrently, collects the errors through a channel, and uses a generic helper to count them:

```go
package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
)

var ErrInvalidPhone = errors.New("invalid phone number")

func validate(phone string) error {
	if !strings.HasPrefix(phone, "+") {
		return fmt.Errorf("validate %q: %w", phone, ErrInvalidPhone)
	}
	return nil
}

func count[T any](items []T, match func(T) bool) int {
	n := 0
	for _, item := range items {
		if match(item) {
			n++
		}
	}
	return n
}

func main() {
	phones := []string{"+1-555-0100", "555-0199", "+44-20-7946-0000", "0800"}

	results := make(chan error, len(phones))
	var wg sync.WaitGroup
	for _, p := range phones {
		wg.Go(func() {
			results <- validate(p)
		})
	}
	wg.Wait()
	close(results)

	var errs []error
	for err := range results {
		errs = append(errs, err)
	}

	invalid := count(errs, func(err error) bool { return errors.Is(err, ErrInvalidPhone) })
	fmt.Println(invalid, "invalid of", len(phones))

	var msgs []string
	for _, err := range errs {
		if err != nil {
			msgs = append(msgs, err.Error())
		}
	}
	slices.Sort(msgs) // goroutines finish in any order
	for _, m := range msgs {
		fmt.Println(m)
	}
}
```

```text
2 invalid of 4
validate "0800": invalid phone number
validate "555-0199": invalid phone number
```

Notice that the channel is **buffered** with room for every result, so no goroutine blocks while `main` is waiting in `wg.Wait()`, and that `close` lets the `range` loop end.

The key ideas from these chapters:

- Return errors, wrap them with `%w`, and inspect them with `errors.Is` and `errors.AsType`. Panic only for bugs.
- Exported names start with a capital letter. Run `go fmt`, `go vet` and `go test` before every commit, and `go fix` after upgrading Go.
- Tests live in `_test.go` files. Table-driven tests with `t.Run` are the Go way.
- `main` doesn't wait for goroutines. Use `sync.WaitGroup` or channels.
- Shared data needs a mutex; run `-race` to catch mistakes.
- Generics let you write one function for many types. Pick the loosest constraint that works.

## Your turn

One last exercise that mixes errors and generics. Complete `firstFailure`:

1. Call `check` on every item. Count how many return a non-nil error.
2. Remember only the **first** failure, wrapped with its index:
   `fmt.Errorf("item %d: %w", i, err)`. Using `%w` keeps `errors.Is` working.
3. Return the count and that wrapped error (`nil` if nothing failed).
4. Then make it **generic**: `func firstFailure[T any](items []T, check func(T) error) (int, error)`.
   **Submit** also calls it with a `[]int`.

**Run** should print `2 item 1: invalid phone number true`.

## You made it!

That's the end of **Learn Go**. Look back at how far you've come: you started by
printing one line of text, and you finished by building the Textio message report, a
real command-line tool that reads a file, counts and sorts words, handles errors
and takes flags. Along the way you learned variables, functions, loops, slices,
maps, structs, pointers, errors, packages, tests, goroutines and generics.

A few habits to keep:

- Write small programs often. The terminal version of the message report is a great
  thing to extend: add a `-min` flag, count emoji, or print a bar chart with
  `strings.Repeat`.
- Read error messages carefully, top to bottom, and trust the compiler.
- Run `go fmt`, `go vet` and `go test` before you call anything done.

## What's next

So far your programs have been a handful of functions and structs. As programs grow,
you need better ways to organize them. Next up is
[Learn Object-Oriented Programming in Go](/courses/learn-oop), where you'll use
methods, interfaces and composition to design types that fit together cleanly.
See you there, and welcome to Go!
