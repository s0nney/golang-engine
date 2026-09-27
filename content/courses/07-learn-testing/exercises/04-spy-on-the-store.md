---
title: Spy on the Store
difficulty: easy
after: test-doubles
hints:
  - 'A spy is an ordinary type with a method that does two jobs: it **records** the call (`s.Calls = append(s.Calls, account)`) and then **answers** it, here from the `Balances` map.'
  - 'Use the comma-ok form, `c, ok := s.Balances[account]`, to tell a missing account from one with a zero balance, and wrap the sentinel: `fmt.Errorf("%w: %q", ErrNoAccount, account)`. Record the call before you return either way.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    // spyStore is a Store for tests. It answers from Balances and records
    // every account it was asked about, in order, in Calls.
    type spyStore struct {
    	Balances map[string]Cents
    	Calls    []string
    }

    // Balance records account in s.Calls, then returns its balance from
    // s.Balances. For an account that isn't in the map it returns 0 and an
    // error wrapping ErrNoAccount.
    func (s *spyStore) Balance(account string) (Cents, error) {
    	// ?
    	return 0, nil
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    var ErrNoAccount = errors.New("ledgerly: no such account")

    // Store looks up account balances. The real one queries a database.
    type Store interface {
    	Balance(account string) (Cents, error)
    }

    // CachedStore remembers successful lookups so each account hits the
    // underlying store only once. Failed lookups aren't cached.
    type CachedStore struct {
    	store Store
    	seen  map[string]Cents
    }

    func NewCachedStore(s Store) *CachedStore {
    	return &CachedStore{store: s, seen: map[string]Cents{}}
    }

    func (c *CachedStore) Balance(account string) (Cents, error) {
    	if b, ok := c.seen[account]; ok {
    		return b, nil
    	}
    	b, err := c.store.Balance(account)
    	if err != nil {
    		return 0, err
    	}
    	c.seen[account] = b
    	return b, nil
    }

    func main() {
    	spy := &spyStore{Balances: map[string]Cents{"rent": -120000, "cash": 5000}}
    	cache := NewCachedStore(spy)
    	for _, account := range []string{"rent", "rent", "cash", "boat", "rent", "boat"} {
    		b, err := cache.Balance(account)
    		fmt.Printf("Balance(%q) = %d, %v\n", account, b, err)
    	}
    	fmt.Printf("the store was asked for: %q\n", spy.Calls)
    	// want: ["rent" "cash" "boat" "boat"]
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    // spyStore is a Store for tests. It answers from Balances and records
    // every account it was asked about, in order, in Calls.
    type spyStore struct {
    	Balances map[string]Cents
    	Calls    []string
    }

    // Balance records account in s.Calls, then returns its balance from
    // s.Balances. For an account that isn't in the map it returns 0 and an
    // error wrapping ErrNoAccount.
    func (s *spyStore) Balance(account string) (Cents, error) {
    	s.Calls = append(s.Calls, account)
    	b, ok := s.Balances[account]
    	if !ok {
    		return 0, fmt.Errorf("%w: %q", ErrNoAccount, account)
    	}
    	return b, nil
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    var ErrNoAccount = errors.New("ledgerly: no such account")

    // Store looks up account balances. The real one queries a database.
    type Store interface {
    	Balance(account string) (Cents, error)
    }

    // CachedStore remembers successful lookups so each account hits the
    // underlying store only once. Failed lookups aren't cached.
    type CachedStore struct {
    	store Store
    	seen  map[string]Cents
    }

    func NewCachedStore(s Store) *CachedStore {
    	return &CachedStore{store: s, seen: map[string]Cents{}}
    }

    func (c *CachedStore) Balance(account string) (Cents, error) {
    	if b, ok := c.seen[account]; ok {
    		return b, nil
    	}
    	b, err := c.store.Balance(account)
    	if err != nil {
    		return 0, err
    	}
    	c.seen[account] = b
    	return b, nil
    }

    func main() {
    	spy := &spyStore{Balances: map[string]Cents{"rent": -120000, "cash": 5000}}
    	cache := NewCachedStore(spy)
    	for _, account := range []string{"rent", "rent", "cash", "boat", "rent", "boat"} {
    		b, err := cache.Balance(account)
    		fmt.Printf("Balance(%q) = %d, %v\n", account, b, err)
    	}
    	fmt.Printf("the store was asked for: %q\n", spy.Calls)
    	// want: ["rent" "cash" "boat" "boat"]
    }
  tests: |
    package main

    import (
    	"errors"
    	"slices"
    	"testing"
    )

    var _ Store = (*spyStore)(nil)

    func TestSpyAnswers(t *testing.T) {
    	spy := &spyStore{Balances: map[string]Cents{"rent": -120000, "empty": 0}}
    	if b, err := spy.Balance("rent"); b != -120000 || err != nil {
    		t.Errorf(`Balance("rent") = %d, %v; want -120000, nil`, b, err)
    	}
    	if b, err := spy.Balance("empty"); b != 0 || err != nil {
    		t.Errorf(`Balance("empty") = %d, %v; want 0, nil (the account exists, its balance is just zero)`, b, err)
    	}
    	b, err := spy.Balance("boat")
    	if !errors.Is(err, ErrNoAccount) || b != 0 {
    		t.Errorf(`Balance("boat") with no "boat" in Balances = %d, %v; want 0 and an error wrapping ErrNoAccount`, b, err)
    	}
    }

    func TestSpyRecords(t *testing.T) {
    	spy := &spyStore{Balances: map[string]Cents{"rent": 1, "cash": 2}}
    	for _, a := range []string{"cash", "rent", "boat", "cash"} {
    		spy.Balance(a)
    	}
    	want := []string{"cash", "rent", "boat", "cash"}
    	if !slices.Equal(spy.Calls, want) {
    		t.Errorf("after asking for %q, Calls = %q; want every call recorded in order, including failed ones", want, spy.Calls)
    	}
    }

    func TestSpyZeroValue(t *testing.T) {
    	var spy spyStore
    	if _, err := spy.Balance("rent"); !errors.Is(err, ErrNoAccount) {
    		t.Errorf("Balance on a spyStore with a nil Balances map returned %v, want ErrNoAccount", err)
    	}
    	if len(spy.Calls) != 1 {
    		t.Errorf("a zero spyStore recorded %q, want [\"rent\"]", spy.Calls)
    	}
    }

    // Using the spy for what it's for: proving the cache hits the store once.
    func TestSpyProvesCaching(t *testing.T) {
    	spy := &spyStore{Balances: map[string]Cents{"rent": -120000, "cash": 5000}}
    	cache := NewCachedStore(spy)
    	for _, a := range []string{"rent", "rent", "cash", "boat", "rent", "boat"} {
    		cache.Balance(a)
    	}
    	want := []string{"rent", "cash", "boat", "boat"}
    	if !slices.Equal(spy.Calls, want) {
    		t.Errorf("CachedStore over your spy: Calls = %q, want %q", spy.Calls, want)
    	}
    }
---

`CachedStore` promises that each account's balance is fetched from the
database only once. How would a test *prove* that? A fake store that returns
the right numbers isn't enough: the cache would also return the right numbers
if it never cached anything. You need a **spy**, a test double that remembers
how it was called.

Complete `spyStore.Balance`:

- Append `account` to `s.Calls`, every time, even when the lookup fails.
- Return the balance from `s.Balances` and a `nil` error.
- If the account isn't in the map, return `0` and an error that wraps
  `ErrNoAccount` (check it with `errors.Is`). An account with a zero balance
  *is* in the map, so it isn't an error.

## Example

```go
spy := &spyStore{Balances: map[string]Cents{"rent": -120000}}
spy.Balance("rent") // -120000, nil
spy.Balance("boat") // 0, ledgerly: no such account: "boat"
spy.Calls           // ["rent" "boat"]
```

**Run** puts your spy behind a `CachedStore` and prints what the store was
asked for. `"rent"` is requested three times but should reach the store once;
the missing `"boat"` isn't cached, so it reaches the store twice.

## Constraints

- A zero `spyStore` (nil map, nil slice) must work: reading a nil map is fine
  in Go, and so is appending to a nil slice.
