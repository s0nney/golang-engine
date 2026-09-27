---
title: Nil Is Not Nil
difficulty: easy
after: interfaces-in-depth
hints:
  - 'An interface value is `nil` only when **both** its type and its value are unset. `return check(e)` stores a `(*ValidationError, nil)` pair in the `error` result: it has a type, so `err != nil` is true.'
  - 'Keep the concrete pointer in a variable of its own type, test *that* for `nil`, and only then return it: `if ve := check(e); ve != nil { return ve }` followed by `return nil`.'
  - '`FirstInvalid` has the same bug in disguise: `var err error = check(e)` wraps the pointer before the `nil` check. Compare the `*ValidationError` itself.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // ValidationError lists everything wrong with one entry.
    type ValidationError struct {
    	Key      string
    	Problems []string
    }

    func (e *ValidationError) Error() string {
    	return fmt.Sprintf("invalid entry %q: %s", e.Key, strings.Join(e.Problems, "; "))
    }

    // Entry is one record in a Stash import batch.
    type Entry struct {
    	Key   string
    	Value string
    	TTL   int // seconds; 0 means "never expires"
    }

    // check returns nil if e is valid. This helper is correct: don't change it.
    func check(e Entry) *ValidationError {
    	var problems []string
    	if e.Key == "" {
    		problems = append(problems, "empty key")
    	}
    	if len(e.Value) > 64 {
    		problems = append(problems, "value longer than 64 bytes")
    	}
    	if e.TTL < 0 {
    		problems = append(problems, "negative TTL")
    	}
    	if problems == nil {
    		return nil
    	}
    	return &ValidationError{Key: e.Key, Problems: problems}
    }

    // Validate returns nil if e is valid, or a *ValidationError describing it.
    // BUG: it never returns nil, even for valid entries.
    func Validate(e Entry) error {
    	return check(e)
    }

    // FirstInvalid returns the error for the first invalid entry, or nil if
    // they're all valid.
    // BUG: it always reports the very first entry.
    func FirstInvalid(entries []Entry) error {
    	for _, e := range entries {
    		var err error = check(e)
    		if err != nil {
    			return err
    		}
    	}
    	return nil
    }

    func main() {
    	good := Entry{Key: "user:1", Value: "ada"}
    	bad := Entry{Key: "", TTL: -5}
    	fmt.Println("Validate(good) == nil:", Validate(good) == nil)   // want true
    	fmt.Println("Validate(bad):", Validate(bad))                   // want invalid entry "": empty key; negative TTL
    	fmt.Println("FirstInvalid:", FirstInvalid([]Entry{good, bad})) // want invalid entry "": ...
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    // ValidationError lists everything wrong with one entry.
    type ValidationError struct {
    	Key      string
    	Problems []string
    }

    func (e *ValidationError) Error() string {
    	return fmt.Sprintf("invalid entry %q: %s", e.Key, strings.Join(e.Problems, "; "))
    }

    // Entry is one record in a Stash import batch.
    type Entry struct {
    	Key   string
    	Value string
    	TTL   int // seconds; 0 means "never expires"
    }

    // check returns nil if e is valid. This helper is correct: don't change it.
    func check(e Entry) *ValidationError {
    	var problems []string
    	if e.Key == "" {
    		problems = append(problems, "empty key")
    	}
    	if len(e.Value) > 64 {
    		problems = append(problems, "value longer than 64 bytes")
    	}
    	if e.TTL < 0 {
    		problems = append(problems, "negative TTL")
    	}
    	if problems == nil {
    		return nil
    	}
    	return &ValidationError{Key: e.Key, Problems: problems}
    }

    // Validate returns nil if e is valid, or a *ValidationError describing it.
    func Validate(e Entry) error {
    	if ve := check(e); ve != nil {
    		return ve
    	}
    	return nil
    }

    // FirstInvalid returns the error for the first invalid entry, or nil if
    // they're all valid.
    func FirstInvalid(entries []Entry) error {
    	for _, e := range entries {
    		if ve := check(e); ve != nil {
    			return ve
    		}
    	}
    	return nil
    }

    func main() {
    	good := Entry{Key: "user:1", Value: "ada"}
    	bad := Entry{Key: "", TTL: -5}
    	fmt.Println("Validate(good) == nil:", Validate(good) == nil)
    	fmt.Println("Validate(bad):", Validate(bad))
    	fmt.Println("FirstInvalid:", FirstInvalid([]Entry{good, bad}))
    }
  tests: |
    package main

    import (
    	"errors"
    	"strings"
    	"testing"
    )

    func TestValidateValid(t *testing.T) {
    	for _, e := range []Entry{
    		{Key: "user:1", Value: "ada"},
    		{Key: "k", Value: "", TTL: 0},
    		{Key: "k", Value: strings.Repeat("x", 64), TTL: 3600},
    	} {
    		if err := Validate(e); err != nil {
    			t.Errorf("Validate(%+v) = %#v, want a nil error (is a nil *ValidationError hiding in the interface?)", e, err)
    		}
    	}
    }

    func TestValidateInvalid(t *testing.T) {
    	tests := []struct {
    		e    Entry
    		want string
    	}{
    		{Entry{Key: "", Value: "v"}, `invalid entry "": empty key`},
    		{Entry{Key: "", TTL: -5}, `invalid entry "": empty key; negative TTL`},
    		{Entry{Key: "big", Value: strings.Repeat("x", 65)}, `invalid entry "big": value longer than 64 bytes`},
    	}
    	for _, tt := range tests {
    		err := Validate(tt.e)
    		if err == nil {
    			t.Errorf("Validate(%+v) = nil, want an error", tt.e)
    			continue
    		}
    		if err.Error() != tt.want {
    			t.Errorf("Validate(%+v).Error() = %q, want %q", tt.e, err.Error(), tt.want)
    		}
    		if _, ok := errors.AsType[*ValidationError](err); !ok {
    			t.Errorf("Validate(%+v) returned a %T, want a *ValidationError", tt.e, err)
    		}
    	}
    }

    func TestFirstInvalid(t *testing.T) {
    	good1 := Entry{Key: "a", Value: "1"}
    	good2 := Entry{Key: "b", Value: "2"}
    	bad1 := Entry{Key: "c", TTL: -1}
    	bad2 := Entry{Key: ""}
    	if err := FirstInvalid(nil); err != nil {
    		t.Errorf("FirstInvalid(nil) = %#v, want nil", err)
    	}
    	if err := FirstInvalid([]Entry{good1, good2}); err != nil {
    		t.Errorf("FirstInvalid(two valid entries) = %#v, want nil (is a nil *ValidationError hiding in the interface?)", err)
    	}
    	err := FirstInvalid([]Entry{good1, bad1, good2, bad2})
    	ve, ok := errors.AsType[*ValidationError](err)
    	if !ok || ve == nil || ve.Key != "c" {
    		t.Errorf("FirstInvalid([good bad(c) good bad()]) = %#v, want the *ValidationError for key \"c\"", err)
    	}
    	err = FirstInvalid([]Entry{good1, good2, bad2})
    	if err == nil || err.Error() != `invalid entry "": empty key` {
    		t.Errorf("FirstInvalid([good good bad()]) = %v, want invalid entry \"\": empty key", err)
    	}
    }
---

Stash's importer validates every entry before storing it. The `check` helper
works fine and returns a `*ValidationError` (or `nil`). But the two public
functions built on it are broken: callers do the usual `if err != nil` and
every **valid** entry gets rejected too.

Fix `Validate` and `FirstInvalid` without changing `check`, the error type or
any signature:

- `Validate(e)` returns `nil` for a valid entry and the `*ValidationError` from
  `check` otherwise.
- `FirstInvalid(entries)` returns the error for the **first invalid** entry, or
  `nil` if all of them are valid (or there are none).

## Example

```go
good := Entry{Key: "user:1", Value: "ada"}
bad := Entry{Key: "", TTL: -5}

Validate(good) == nil            // true (currently false!)
Validate(bad)                    // invalid entry "": empty key; negative TTL
FirstInvalid([]Entry{good, bad}) // the error for bad, not for good
```

## Constraints

- The errors you return must still be `*ValidationError` values, so callers can
  use `errors.AsType[*ValidationError](err)`.
- Run the starter first: `Validate(good) == nil` prints `false`. Work out why
  before you change anything.
