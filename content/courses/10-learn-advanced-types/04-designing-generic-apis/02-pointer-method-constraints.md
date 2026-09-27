---
title: The Pointer-Method Constraint
quiz:
  - question: |
      What happens when you call `ParseBad[*Entry]([]string{"a=1"})`?

      ```go
      type Parser interface{ Parse(string) error }

      func ParseBad[T Parser](lines []string) []T {
      	out := make([]T, len(lines))
      	for i, l := range lines {
      		out[i].Parse(l)
      	}
      	return out
      }
      ```

      (`*Entry` has a pointer-receiver `Parse` method.)
    options:
      - text: It returns a slice with one parsed `*Entry`
      - text: It panics with a nil pointer dereference, because every `out[i]` is a nil `*Entry`
        correct: true
      - text: It doesn't compile, because `*Entry` doesn't satisfy `Parser`
      - text: It returns a slice of empty entries
    explanation: |
      `make([]T, n)` with `T = *Entry` makes `n` nil pointers. Nothing ever allocates an
      `Entry` for them to point at, so `Parse` writes through a nil pointer.
  - question: |
      Given

      ```go
      func ParseAll[T any, PT interface {
      	*T
      	Parse(string) error
      }](lines []string) ([]T, error)
      ```

      how do you call it for `Entry`?
    options:
      - text: '`ParseAll[Entry, *Entry](lines)` only; `PT` can''t be inferred'
      - text: '`ParseAll[Entry](lines)`; `PT` is inferred from its constraint `*T`'
        correct: true
      - text: '`ParseAll[*Entry](lines)`'
      - text: '`ParseAll(lines)`'
    explanation: |
      Once `T = Entry` is given, `PT`'s constraint has the single type term `*T`, so
      constraint type inference sets `PT = *Entry` and then checks it has `Parse`. The
      longer form works too, but nobody wants to type it.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strconv"
    	"strings"
    )

    // Entry is a Stash record, parsed from "key=size".
    type Entry struct {
    	Key  string
    	Size int
    }

    // Parse fills in e from a line like "logo.png=2048".
    func (e *Entry) Parse(line string) error {
    	k, v, ok := strings.Cut(line, "=")
    	if !ok {
    		return fmt.Errorf("missing '=' in %q", line)
    	}
    	n, err := strconv.Atoi(v)
    	if err != nil {
    		return err
    	}
    	e.Key, e.Size = k, n
    	return nil
    }

    // Tag is another parsable type: a line is just its name.
    type Tag struct{ Name string }

    func (t *Tag) Parse(line string) error {
    	if line == "" {
    		return fmt.Errorf("empty tag")
    	}
    	t.Name = line
    	return nil
    }

    // ParseAll parses each line into a T by calling Parse on a *T.
    // On failure it returns nil and an error of the form
    // "line N: <original error>" (N counts from 1), wrapping the original.
    func ParseAll[T any, PT interface {
    	*T
    	Parse(string) error
    }](lines []string) ([]T, error) {
    	// ?
    	return nil, nil
    }

    func main() {
    	entries, err := ParseAll[Entry]([]string{"logo.png=2048", "app.js=512"})
    	fmt.Println(entries, err)
    	_, err = ParseAll[Entry]([]string{"a=1", "oops"})
    	fmt.Println(err)
    	tags, err := ParseAll[Tag]([]string{"go", "types"})
    	fmt.Println(tags, err)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strconv"
    	"strings"
    )

    // Entry is a Stash record, parsed from "key=size".
    type Entry struct {
    	Key  string
    	Size int
    }

    // Parse fills in e from a line like "logo.png=2048".
    func (e *Entry) Parse(line string) error {
    	k, v, ok := strings.Cut(line, "=")
    	if !ok {
    		return fmt.Errorf("missing '=' in %q", line)
    	}
    	n, err := strconv.Atoi(v)
    	if err != nil {
    		return err
    	}
    	e.Key, e.Size = k, n
    	return nil
    }

    // Tag is another parsable type: a line is just its name.
    type Tag struct{ Name string }

    func (t *Tag) Parse(line string) error {
    	if line == "" {
    		return fmt.Errorf("empty tag")
    	}
    	t.Name = line
    	return nil
    }

    // ParseAll parses each line into a T by calling Parse on a *T.
    func ParseAll[T any, PT interface {
    	*T
    	Parse(string) error
    }](lines []string) ([]T, error) {
    	out := make([]T, len(lines))
    	for i, line := range lines {
    		if err := PT(&out[i]).Parse(line); err != nil {
    			return nil, fmt.Errorf("line %d: %w", i+1, err)
    		}
    	}
    	return out, nil
    }

    func main() {
    	entries, err := ParseAll[Entry]([]string{"logo.png=2048", "app.js=512"})
    	fmt.Println(entries, err)
    	_, err = ParseAll[Entry]([]string{"a=1", "oops"})
    	fmt.Println(err)
    	tags, err := ParseAll[Tag]([]string{"go", "types"})
    	fmt.Println(tags, err)
    }
  tests: |
    package main

    import (
    	"errors"
    	"strconv"
    	"testing"
    )

    func TestParseAllEntries(t *testing.T) {
    	got, err := ParseAll[Entry]([]string{"logo.png=2048", "app.js=512"})
    	if err != nil {
    		t.Fatalf("ParseAll[Entry] returned error %v", err)
    	}
    	want := []Entry{{"logo.png", 2048}, {"app.js", 512}}
    	if len(got) != len(want) {
    		t.Fatalf("ParseAll[Entry] = %v, want %v", got, want)
    	}
    	for i := range want {
    		if got[i] != want[i] {
    			t.Errorf("ParseAll[Entry][%d] = %v, want %v", i, got[i], want[i])
    		}
    	}
    }

    func TestParseAllTags(t *testing.T) {
    	got, err := ParseAll[Tag]([]string{"go", "types"})
    	if err != nil || len(got) != 2 || got[0].Name != "go" || got[1].Name != "types" {
    		t.Errorf("ParseAll[Tag]([go types]) = (%v, %v), want ([{go} {types}], nil)", got, err)
    	}
    	if got, err := ParseAll[Tag](nil); err != nil || len(got) != 0 {
    		t.Errorf("ParseAll[Tag](nil) = (%v, %v), want ([], nil)", got, err)
    	}
    }

    func TestParseAllErrors(t *testing.T) {
    	got, err := ParseAll[Entry]([]string{"a=1", "oops", "b=2"})
    	if err == nil {
    		t.Fatalf("ParseAll with a bad line 2 returned no error")
    	}
    	if got != nil {
    		t.Errorf("on error, ParseAll should return nil, got %v", got)
    	}
    	if want := `line 2: missing '=' in "oops"`; err.Error() != want {
    		t.Errorf("error = %q, want %q", err.Error(), want)
    	}
    	_, err = ParseAll[Entry]([]string{"a=x"})
    	if _, ok := errors.AsType[*strconv.NumError](err); !ok {
    		t.Errorf("error %v should wrap the original *strconv.NumError (use %%w)", err)
    	}
    }
---

Here's a very common wish: "parse a list of lines into values of *any* type that knows how to parse itself". The obvious attempt has a trap in it.

## The obvious attempt

```go
type Parser interface{ Parse(string) error }

func ParseBad[T Parser](lines []string) []T {
	out := make([]T, len(lines))
	for i, l := range lines {
		out[i].Parse(l)
	}
	return out
}
```

`Entry.Parse` has to modify the entry, so it has a **pointer receiver**. That means `*Entry` implements `Parser`, and `Entry` doesn't. So the caller must write `ParseBad[*Entry](lines)`, and now `out` is a slice of `*Entry`... all of them `nil`. The first `Parse` call dereferences nil and panics.

What we want is: `T` is the **value** type (so we can allocate `[]Entry`), and `*T` has the `Parse` method (so we can call it on `&out[i]`). A single type parameter can't say both things.

## Two type parameters, one tied to the other

```go
func ParseAll[T any, PT interface {
	*T
	Parse(string) error
}](lines []string) ([]T, error)
```

`PT`'s constraint says: "`PT` is exactly `*T`, and it has a `Parse` method". Inside the function, we allocate real `T` values and convert their addresses to `PT` to call the method:

```go
package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Entry struct {
	Key  string
	Size int
}

func (e *Entry) Parse(line string) error {
	k, v, ok := strings.Cut(line, "=")
	if !ok {
		return fmt.Errorf("missing '=' in %q", line)
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return err
	}
	e.Key, e.Size = k, n
	return nil
}

func ParseAll[T any, PT interface {
	*T
	Parse(string) error
}](lines []string) ([]T, error) {
	out := make([]T, len(lines))
	for i, line := range lines {
		if err := PT(&out[i]).Parse(line); err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
	}
	return out, nil
}

func main() {
	es, err := ParseAll[Entry]([]string{"a=1", "b=22"})
	fmt.Println(es, err)
	_, err = ParseAll[Entry]([]string{"a=1", "oops"})
	fmt.Println(err)
}
```

```
[{a 1} {b 22}] <nil>
line 2: missing '=' in "oops"
```

Three details make this work:

1. **`PT(&out[i])`**: `&out[i]` has type `*T`, and the constraint guarantees `PT`'s type set contains only `*T`, so the conversion is allowed.
2. **Inference**: callers write `ParseAll[Entry](lines)`. Given `T`, the `*T` term in `PT`'s constraint pins `PT = *Entry` (constraint type inference from the last chapter), and then the compiler checks that `*Entry` really has `Parse`.
3. **Real values**: `out` is a `[]Entry` of zero values, each with an address. No nil pointers, and one allocation for the whole slice instead of one per element.

## Where you'll see it

The pattern appears anywhere generic code needs to **create** a value and then call a **pointer method** on it: decoders (`UnmarshalText`, `Scan` for database rows), "reset" or "init" methods on pooled objects, and test helpers that build fixtures. It looks noisy the first time, so a short comment next to it helps readers:

```go
// PT is *T, and must have a Parse method. Callers only name T.
```

## Your turn

Complete `ParseAll` for Stash's line-based records. For each line, call `Parse` on a pointer to the matching element of the result. On failure, return `nil` and an error `"line N: <original error>"` that **wraps** the original (use `%w`), with `N` counting from 1. It must work for both `Entry` and `Tag`.
