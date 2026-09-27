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

That's the course. Well done, and welcome to Go!
