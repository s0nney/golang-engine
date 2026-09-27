---
title: A Contract for Every Store
difficulty: medium
after: test-doubles
hints:
  - 'Write the contract as a sequence of small steps on a **fresh** store from `newStore()`, returning a descriptive error at the first step that goes wrong: `return fmt.Errorf("History(%q) after Append 5, -3 = %v, want [5 -3]", ...)`. `slices.Equal` compares two `[]Cents`.'
  - 'Some bugs only show if you poke at them: change an element of the slice `History` returned, then call `History` again; append to two accounts and check neither sees the other''s amounts; call `newStore()` twice and check the second store starts empty.'
  - 'Remember both sides of every error rule: `History` and `Delete` of an unknown account must return an error wrapping `ErrNoAccount` (check with `errors.Is`), and after a successful `Delete` the account is unknown again.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"slices"
    )

    // checkStore runs the Store contract against fresh stores made by
    // newStore. It returns nil if every rule holds, or an error describing
    // the first rule that's broken.
    func checkStore(newStore func() Store) error {
    	return nil
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    var ErrNoAccount = errors.New("ledgerly: no such account")

    // Store keeps the amounts posted to each account. The real one is a
    // database; MemStore is the fake the rest of Ledgerly's tests use.
    type Store interface {
    	// Append posts amount to account, creating the account if needed.
    	Append(account string, amount Cents)
    	// History returns a copy of account's amounts in the order they were
    	// appended, or an error wrapping ErrNoAccount if it doesn't exist.
    	History(account string) ([]Cents, error)
    	// Delete removes account, or returns an error wrapping ErrNoAccount
    	// if it doesn't exist.
    	Delete(account string) error
    }

    type MemStore struct {
    	accounts map[string][]Cents
    }

    func NewMemStore() Store { return &MemStore{accounts: map[string][]Cents{}} }

    func (m *MemStore) Append(account string, amount Cents) {
    	m.accounts[account] = append(m.accounts[account], amount)
    }

    func (m *MemStore) History(account string) ([]Cents, error) {
    	h, ok := m.accounts[account]
    	if !ok {
    		return nil, fmt.Errorf("%w: %q", ErrNoAccount, account)
    	}
    	return slices.Clone(h), nil
    }

    func (m *MemStore) Delete(account string) error {
    	if _, ok := m.accounts[account]; !ok {
    		return fmt.Errorf("%w: %q", ErrNoAccount, account)
    	}
    	delete(m.accounts, account)
    	return nil
    }

    // leakyStore is a MemStore whose History hands out its internal slice.
    type leakyStore struct{ MemStore }

    func (l *leakyStore) History(account string) ([]Cents, error) {
    	h, ok := l.accounts[account]
    	if !ok {
    		return nil, fmt.Errorf("%w: %q", ErrNoAccount, account)
    	}
    	return h, nil
    }

    func main() {
    	fmt.Println("MemStore:  ", checkStore(NewMemStore))
    	fmt.Println("leakyStore:", checkStore(func() Store {
    		return &leakyStore{MemStore{accounts: map[string][]Cents{}}}
    	}))
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"slices"
    )

    // checkStore runs the Store contract against fresh stores made by
    // newStore. It returns nil if every rule holds, or an error describing
    // the first rule that's broken.
    func checkStore(newStore func() Store) error {
    	s := newStore()
    	if h, err := s.History("rent"); !errors.Is(err, ErrNoAccount) {
    		return fmt.Errorf("History(%q) on a new store = %v, %v; want an error wrapping ErrNoAccount", "rent", h, err)
    	}
    	if err := s.Delete("rent"); !errors.Is(err, ErrNoAccount) {
    		return fmt.Errorf("Delete(%q) on a new store = %v; want an error wrapping ErrNoAccount", "rent", err)
    	}

    	s.Append("rent", -120000)
    	s.Append("cash", 5000)
    	s.Append("rent", -1500)
    	s.Append("rent", 0)
    	for account, want := range map[string][]Cents{
    		"rent": {-120000, -1500, 0},
    		"cash": {5000},
    	} {
    		if h, err := s.History(account); err != nil || !slices.Equal(h, want) {
    			return fmt.Errorf("History(%q) = %v, %v; want %v, nil", account, h, err, want)
    		}
    	}

    	h, _ := s.History("rent")
    	h[0] = 999
    	if h, _ := s.History("rent"); len(h) == 0 || h[0] != -120000 {
    		return fmt.Errorf("changing the slice History returned changed the store: History(%q) = %v", "rent", h)
    	}

    	if err := s.Delete("rent"); err != nil {
    		return fmt.Errorf("Delete(%q) = %v, want nil", "rent", err)
    	}
    	if h, err := s.History("rent"); !errors.Is(err, ErrNoAccount) {
    		return fmt.Errorf("History(%q) after Delete = %v, %v; want an error wrapping ErrNoAccount", "rent", h, err)
    	}
    	if h, err := s.History("cash"); err != nil || !slices.Equal(h, []Cents{5000}) {
    		return fmt.Errorf("History(%q) after deleting rent = %v, %v; want [5000], nil", "cash", h, err)
    	}

    	if h, err := newStore().History("cash"); !errors.Is(err, ErrNoAccount) {
    		return fmt.Errorf("History(%q) on a second new store = %v, %v; want ErrNoAccount (stores must not share data)", "cash", h, err)
    	}
    	return nil
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    var ErrNoAccount = errors.New("ledgerly: no such account")

    // Store keeps the amounts posted to each account. The real one is a
    // database; MemStore is the fake the rest of Ledgerly's tests use.
    type Store interface {
    	// Append posts amount to account, creating the account if needed.
    	Append(account string, amount Cents)
    	// History returns a copy of account's amounts in the order they were
    	// appended, or an error wrapping ErrNoAccount if it doesn't exist.
    	History(account string) ([]Cents, error)
    	// Delete removes account, or returns an error wrapping ErrNoAccount
    	// if it doesn't exist.
    	Delete(account string) error
    }

    type MemStore struct {
    	accounts map[string][]Cents
    }

    func NewMemStore() Store { return &MemStore{accounts: map[string][]Cents{}} }

    func (m *MemStore) Append(account string, amount Cents) {
    	m.accounts[account] = append(m.accounts[account], amount)
    }

    func (m *MemStore) History(account string) ([]Cents, error) {
    	h, ok := m.accounts[account]
    	if !ok {
    		return nil, fmt.Errorf("%w: %q", ErrNoAccount, account)
    	}
    	return slices.Clone(h), nil
    }

    func (m *MemStore) Delete(account string) error {
    	if _, ok := m.accounts[account]; !ok {
    		return fmt.Errorf("%w: %q", ErrNoAccount, account)
    	}
    	delete(m.accounts, account)
    	return nil
    }

    // leakyStore is a MemStore whose History hands out its internal slice.
    type leakyStore struct{ MemStore }

    func (l *leakyStore) History(account string) ([]Cents, error) {
    	h, ok := l.accounts[account]
    	if !ok {
    		return nil, fmt.Errorf("%w: %q", ErrNoAccount, account)
    	}
    	return h, nil
    }

    func main() {
    	fmt.Println("MemStore:  ", checkStore(NewMemStore))
    	fmt.Println("leakyStore:", checkStore(func() Store {
    		return &leakyStore{MemStore{accounts: map[string][]Cents{}}}
    	}))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"slices"
    	"testing"
    )

    // run calls checkStore and turns a panic into an error message.
    func run(newStore func() Store) (err error, panicked string) {
    	defer func() {
    		if p := recover(); p != nil {
    			panicked = fmt.Sprint(p)
    		}
    	}()
    	return checkStore(newStore), ""
    }

    func TestContractAcceptsMemStore(t *testing.T) {
    	err, p := run(NewMemStore)
    	if p != "" {
    		t.Fatalf("checkStore(NewMemStore) panicked: %s", p)
    	}
    	if err != nil {
    		t.Fatalf("checkStore(NewMemStore) = %v, want nil (MemStore is correct)", err)
    	}
    }

    // A second correct store, built differently, must pass too.
    type listStore struct {
    	names []string
    	rows  [][]Cents
    }

    func (l *listStore) find(a string) int { return slices.Index(l.names, a) }
    func (l *listStore) Append(a string, c Cents) {
    	if i := l.find(a); i >= 0 {
    		l.rows[i] = append(l.rows[i], c)
    		return
    	}
    	l.names = append(l.names, a)
    	l.rows = append(l.rows, []Cents{c})
    }
    func (l *listStore) History(a string) ([]Cents, error) {
    	if i := l.find(a); i >= 0 {
    		return append([]Cents{}, l.rows[i]...), nil
    	}
    	return nil, fmt.Errorf("listStore: %q: %w", a, ErrNoAccount)
    }
    func (l *listStore) Delete(a string) error {
    	i := l.find(a)
    	if i < 0 {
    		return fmt.Errorf("listStore: %q: %w", a, ErrNoAccount)
    	}
    	l.names = slices.Delete(l.names, i, i+1)
    	l.rows = slices.Delete(l.rows, i, i+1)
    	return nil
    }

    func TestContractAcceptsAnotherCorrectStore(t *testing.T) {
    	err, p := run(func() Store { return &listStore{} })
    	if p != "" || err != nil {
    		t.Fatalf("checkStore rejected a second, correct Store implementation: %v %s", err, p)
    	}
    }

    // ---- broken stores, each a MemStore with one bug ----

    func mem() MemStore { return MemStore{accounts: map[string][]Cents{}} }

    type aliasing struct{ MemStore }

    func (s *aliasing) History(a string) ([]Cents, error) {
    	if h, ok := s.accounts[a]; ok {
    		return h, nil
    	}
    	return s.MemStore.History(a)
    }

    type nilForUnknown struct{ MemStore }

    func (s *nilForUnknown) History(a string) ([]Cents, error) {
    	if _, ok := s.accounts[a]; !ok {
    		return nil, nil
    	}
    	return s.MemStore.History(a)
    }

    type deleteNoop struct{ MemStore }

    func (s *deleteNoop) Delete(a string) error {
    	if _, ok := s.accounts[a]; !ok {
    		return s.MemStore.Delete(a)
    	}
    	return nil
    }

    type deleteUnknownOK struct{ MemStore }

    func (s *deleteUnknownOK) Delete(a string) error {
    	s.MemStore.Delete(a)
    	return nil
    }

    type newestFirst struct{ MemStore }

    func (s *newestFirst) History(a string) ([]Cents, error) {
    	h, err := s.MemStore.History(a)
    	slices.Reverse(h)
    	return h, err
    }

    type oneBucket struct{ MemStore }

    func (s *oneBucket) Append(a string, c Cents) {
    	for name := range s.accounts {
    		s.accounts[name] = append(s.accounts[name], c)
    	}
    	if _, ok := s.accounts[a]; !ok {
    		s.accounts[a] = []Cents{c}
    	}
    }

    type dropsZero struct{ MemStore }

    func (s *dropsZero) Append(a string, c Cents) {
    	if c == 0 {
    		if _, ok := s.accounts[a]; !ok {
    			s.accounts[a] = nil
    		}
    		return
    	}
    	s.MemStore.Append(a, c)
    }

    var sharedAccounts map[string][]Cents

    func TestContractCatchesBugs(t *testing.T) {
    	bugs := []struct {
    		bug      string
    		newStore func() Store
    	}{
    		{"returns its internal slice from History, so callers can change the store", func() Store { return &aliasing{mem()} }},
    		{"returns nil, nil from History for an unknown account", func() Store { return &nilForUnknown{mem()} }},
    		{"does nothing on Delete (but returns nil)", func() Store { return &deleteNoop{mem()} }},
    		{"returns nil from Delete of an unknown account", func() Store { return &deleteUnknownOK{mem()} }},
    		{"returns History newest first", func() Store { return &newestFirst{mem()} }},
    		{"appends every amount to every existing account", func() Store { return &oneBucket{mem()} }},
    		{"silently drops zero amounts", func() Store { return &dropsZero{mem()} }},
    		{"shares one map between all stores made by newStore", func() Store {
    			if sharedAccounts == nil {
    				sharedAccounts = map[string][]Cents{}
    			}
    			return &MemStore{accounts: sharedAccounts}
    		}},
    	}
    	for _, b := range bugs {
    		sharedAccounts = nil
    		err, p := run(b.newStore)
    		if p != "" {
    			t.Errorf("checkStore panicked on a store that %s: %s (return an error instead)", b.bug, p)
    		} else if err == nil {
    			t.Errorf("checkStore returned nil for a store that %s", b.bug)
    		}
    	}
    }
---

The rest of Ledgerly's tests use `MemStore`, an in-memory **fake** of the
database-backed store. A fake is only useful if it behaves like the real
thing. Otherwise your tests pass against a store that doesn't exist. The fix
is a **contract test**: one function that any `Store` implementation must
pass, run against the real store in integration tests and against every fake
in unit tests.

Write `checkStore(newStore)`. It builds fresh stores with `newStore()`,
exercises them, and returns `nil` if the contract holds or an error describing
the first rule that's broken.

## The contract

- `History` of an account that was never appended to, and `Delete` of one,
  return an error wrapping `ErrNoAccount`.
- `Append` creates the account if needed. `History` returns every amount,
  **including zeros**, in the order it was appended.
- Accounts are independent: appending to `"cash"` doesn't change `"rent"`.
- `History` returns a **copy**: changing the returned slice doesn't change the
  store.
- After `Delete`, the account is unknown again (`ErrNoAccount`), and other
  accounts are untouched.
- Each call to `newStore()` returns a new, empty store.

## How you're graded

The grader runs your `checkStore` against `MemStore` and a second correct
store (both must return `nil`), then against **eight** broken stores, each a
`MemStore` with one bug. Each must get a non-nil error, and `checkStore`
must never panic.

**Run** checks `MemStore` and `leakyStore`, whose `History` hands out its
internal slice. Your function should accept the first and reject the second.

## Constraints

- Only call the store through the `Store` interface, since that's all the
  real database store has in common with the fakes.
- Don't compare error messages, just `errors.Is(err, ErrNoAccount)`.
