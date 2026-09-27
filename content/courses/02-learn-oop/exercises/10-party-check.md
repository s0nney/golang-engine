---
title: Party Check
difficulty: medium
after: polymorphism
hints:
  - '`HeroError` needs `Error() string` and `Unwrap() error` (pointer receivers). `Unwrap` returning the sentinel is what lets `errors.Is(err, ErrFallen)` see through it.'
  - 'Beware the nil-interface gotcha: if `checkHero` stores its result in a `var e *HeroError` and returns `e`, a healthy hero gives an `error` that is **not** `nil` (it holds a nil `*HeroError`). Return the literal `nil` on the happy path. `errors.Join` of zero errors (or only `nil`s) is `nil`, which makes `checkParty` short.'
  - 'For `benched`, walk the error tree with a type switch whose cases are types **and** interfaces: `case *HeroError:` gives you the name; `case interface{ Unwrap() []error }:` (what `errors.Join` returns) means recurse into every child; `case interface{ Unwrap() error }:` (what `fmt.Errorf` with `%w` returns) means recurse into the one wrapped error.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var (
    	ErrFallen    = errors.New("has fallen")
    	ErrExhausted = errors.New("is out of stamina")
    )

    type Hero struct {
    	Name    string
    	HP      int
    	Stamina int
    }

    type HeroError struct {
    	Hero string
    	Err  error
    }

    func (e *HeroError) Error() string {
    	return ""
    }

    func (e *HeroError) Unwrap() error {
    	return nil
    }

    func checkHero(h Hero) error {
    	return nil
    }

    func checkParty(party []Hero) error {
    	return nil
    }

    func benched(err error) []string {
    	return nil
    }

    func main() {
    	party := []Hero{{"Ayla", 30, 5}, {"Bram", 0, 3}, {"Cora", 12, 0}}
    	err := checkParty(party)
    	fmt.Println(err)
    	// want:
    	// Bram has fallen
    	// Cora is out of stamina
    	fmt.Println(errors.Is(err, ErrFallen), benched(err)) // want true [Bram Cora]
    	fmt.Println(checkParty(party[:1]) == nil)            // want true
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var (
    	ErrFallen    = errors.New("has fallen")
    	ErrExhausted = errors.New("is out of stamina")
    )

    type Hero struct {
    	Name    string
    	HP      int
    	Stamina int
    }

    // HeroError says which hero can't set out, and why.
    type HeroError struct {
    	Hero string
    	Err  error
    }

    func (e *HeroError) Error() string {
    	return e.Hero + " " + e.Err.Error()
    }

    func (e *HeroError) Unwrap() error {
    	return e.Err
    }

    // checkHero returns a *HeroError if h can't set out, or nil.
    func checkHero(h Hero) error {
    	switch {
    	case h.HP <= 0:
    		return &HeroError{Hero: h.Name, Err: ErrFallen}
    	case h.Stamina <= 0:
    		return &HeroError{Hero: h.Name, Err: ErrExhausted}
    	}
    	return nil
    }

    // checkParty joins the problems of every hero, in party order, or returns nil.
    func checkParty(party []Hero) error {
    	var errs []error
    	for _, h := range party {
    		errs = append(errs, checkHero(h))
    	}
    	return errors.Join(errs...)
    }

    // benched returns the names of all heroes with a *HeroError in err's tree.
    func benched(err error) []string {
    	switch e := err.(type) {
    	case *HeroError:
    		return []string{e.Hero}
    	case interface{ Unwrap() []error }: // errors.Join, or fmt.Errorf with several %w
    		var names []string
    		for _, child := range e.Unwrap() {
    			names = append(names, benched(child)...)
    		}
    		return names
    	case interface{ Unwrap() error }: // fmt.Errorf with one %w
    		return benched(e.Unwrap())
    	}
    	return nil
    }

    func main() {
    	party := []Hero{{"Ayla", 30, 5}, {"Bram", 0, 3}, {"Cora", 12, 0}}
    	err := checkParty(party)
    	fmt.Println(err)
    	fmt.Println(errors.Is(err, ErrFallen), benched(err))
    	fmt.Println(checkParty(party[:1]) == nil)
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"slices"
    	"testing"
    )

    func TestCheckHero(t *testing.T) {
    	tests := []struct {
    		h    Hero
    		want error // nil, ErrFallen or ErrExhausted
    	}{
    		{Hero{"Ayla", 30, 5}, nil},
    		{Hero{"Bram", 0, 3}, ErrFallen},
    		{Hero{"Bram", -4, 3}, ErrFallen},
    		{Hero{"Cora", 12, 0}, ErrExhausted},
    		{Hero{"Dax", 0, 0}, ErrFallen},
    		{Hero{"Eli", 1, 1}, nil},
    	}
    	for _, tt := range tests {
    		err := checkHero(tt.h)
    		if tt.want == nil {
    			if err != nil {
    				t.Errorf("checkHero(%+v) = %#v, want nil. If that shows a nil *HeroError, you hit the nil-interface gotcha: return the literal nil", tt.h, err)
    			}
    			continue
    		}
    		he, ok := errors.AsType[*HeroError](err)
    		if !ok {
    			t.Errorf("checkHero(%+v) = %v, want a *HeroError", tt.h, err)
    			continue
    		}
    		if he.Hero != tt.h.Name || he.Err != tt.want {
    			t.Errorf("checkHero(%+v) = &HeroError{Hero: %q, Err: %v}, want Hero %q and Err %v", tt.h, he.Hero, he.Err, tt.h.Name, tt.want)
    		}
    		if !errors.Is(err, tt.want) {
    			t.Errorf("errors.Is(checkHero(%+v), %v) = false, want true: does HeroError have an Unwrap method?", tt.h, tt.want)
    		}
    		if want := tt.h.Name + " " + tt.want.Error(); err.Error() != want {
    			t.Errorf("checkHero(%+v).Error() = %q, want %q", tt.h, err.Error(), want)
    		}
    	}
    }

    func TestHeroErrorWrapsAnything(t *testing.T) {
    	cursed := errors.New("is cursed")
    	var err error = &HeroError{Hero: "Fen", Err: fmt.Errorf("wearing the ring: %w", cursed)}
    	if err.Error() != "Fen wearing the ring: is cursed" || !errors.Is(err, cursed) {
    		t.Errorf("HeroError{Hero: \"Fen\", Err: <wraps cursed>}: Error() = %q, errors.Is(err, cursed) = %v, want \"Fen wearing the ring: is cursed\", true", err.Error(), errors.Is(err, cursed))
    	}
    }

    func TestCheckParty(t *testing.T) {
    	party := []Hero{{"Ayla", 30, 5}, {"Bram", 0, 3}, {"Cora", 12, 0}, {"Dax", 8, 8}}
    	err := checkParty(party)
    	if err == nil {
    		t.Fatalf("checkParty(%v) = nil, want an error for Bram and Cora", party)
    	}
    	if want := "Bram has fallen\nCora is out of stamina"; err.Error() != want {
    		t.Errorf("checkParty(%v).Error() = %q, want %q", party, err.Error(), want)
    	}
    	if !errors.Is(err, ErrFallen) || !errors.Is(err, ErrExhausted) {
    		t.Errorf("checkParty error: errors.Is ErrFallen = %v, ErrExhausted = %v, want both true", errors.Is(err, ErrFallen), errors.Is(err, ErrExhausted))
    	}
    	if he, ok := errors.AsType[*HeroError](err); !ok || he.Hero != "Bram" {
    		t.Errorf("errors.AsType[*HeroError](checkParty(...)) should find Bram's error first, got %v, %v", he, ok)
    	}

    	for _, fine := range [][]Hero{nil, {}, {{"Ayla", 30, 5}}, {{"Ayla", 30, 5}, {"Eli", 1, 1}}} {
    		if err := checkParty(fine); err != nil {
    			t.Errorf("checkParty(%v) = %#v, want nil for a healthy party. A non-nil error holding nil values is the nil-interface gotcha", fine, err)
    		}
    	}
    }

    // oathError is the test's own error type, unrelated to heroes.
    type oathError struct{}

    func (oathError) Error() string { return "oath broken" }

    func TestBenched(t *testing.T) {
    	party := []Hero{{"Ayla", 0, 5}, {"Bram", 9, 3}, {"Cora", 12, 0}, {"Dax", -1, 0}}
    	tests := []struct {
    		name string
    		err  error
    		want []string
    	}{
    		{"nil", nil, nil},
    		{"a party", checkParty(party), []string{"Ayla", "Cora", "Dax"}},
    		{"a single hero", checkHero(Hero{"Eli", 0, 0}), []string{"Eli"}},
    		{"wrapped with %w", fmt.Errorf("quest cancelled: %w", checkHero(Hero{"Fen", 0, 1})), []string{"Fen"}},
    		{"unrelated error", oathError{}, nil},
    		{"joined with other errors", errors.Join(oathError{}, checkHero(Hero{"Gus", 5, 0}), errors.New("rain")), []string{"Gus"}},
    		{"joins inside joins", errors.Join(checkParty(party[:2]), fmt.Errorf("late: %w", checkParty(party[2:]))), []string{"Ayla", "Cora", "Dax"}},
    	}
    	for _, tt := range tests {
    		if got := benched(tt.err); !slices.Equal(got, tt.want) {
    			t.Errorf("%s: benched(%v) = %q, want %q", tt.name, tt.err, got, tt.want)
    		}
    	}
    }
---

Before a quest, the guild master checks every hero. A hero with 0 HP or less
**has fallen**; otherwise, a hero with 0 Stamina or less **is out of stamina**.
Report problems with a custom error type so callers can ask precise questions
about them with `errors.Is` and `errors.AsType`.

1. `*HeroError` implements `error`. `Error()` returns the hero's name, a space,
   and the wrapped error's message (`"Bram has fallen"`), and `Unwrap()` returns
   `Err`, so `errors.Is(err, ErrFallen)` works.
2. `checkHero(h)` returns a `*HeroError` wrapping `ErrFallen` or `ErrExhausted`
   (fallen wins if both apply), or **`nil`** for a healthy hero.
3. `checkParty(party)` returns all the heroes' errors joined in party order
   (`errors.Join` style: one per line), or `nil` if everyone is fine.
4. `benched(err)` returns the names of **every** hero with a `*HeroError`
   anywhere in `err`'s tree: through `errors.Join`, through `fmt.Errorf("...: %w")`
   wrapping, and through combinations of both, in the order they appear.

A healthy party must give an error that is `== nil`. Careful: an `error`
interface holding a nil `*HeroError` is **not** nil.

## Example

```go
party := []Hero{{"Ayla", 30, 5}, {"Bram", 0, 3}, {"Cora", 12, 0}}
err := checkParty(party)
fmt.Println(err)
// Bram has fallen
// Cora is out of stamina

errors.Is(err, ErrFallen)                       // true
benched(err)                                    // [Bram Cora]
benched(fmt.Errorf("quest cancelled: %w", err)) // [Bram Cora]
checkParty(party[:1]) == nil                    // true
```

## Constraints

- The tests also wrap other error types (ones that aren't `*HeroError`) into
  the trees they pass to `benched`. Skip those.
