---
title: Stubs, Spies, Fakes and Mocks
quiz:
  - question: |
      A test double for `Store` keeps saved transactions in a map, so
      `Save` followed by `Load` really returns what was saved. What kind
      of double is it?
    options:
      - text: A stub
      - text: A spy
      - text: A fake
        correct: true
      - text: A dummy
    explanation: |
      A fake is a *working* implementation that takes a shortcut, like
      keeping data in memory instead of a database. A stub just returns
      canned answers and doesn't remember anything.
  - question: |
      A spy notifier records every message in a slice. Which assertion
      is the most likely to make the test brittle?
    options:
      - text: Checking that exactly one message was sent
      - text: Checking that the message mentions the account name
      - text: Checking the exact sequence of every internal call made on every dependency, in order
        correct: true
      - text: Checking that no message was sent when the balance is fine
    explanation: |
      Tests that pin down every interaction break whenever the
      implementation is refactored, even if the behaviour is unchanged.
      Assert on the outcomes the caller cares about, not the
      implementation's choreography.
exercise:
  starter: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"math"
    	"sync"
    )

    // stubRates is a RateSource that returns whatever rate and err it holds.
    type stubRates struct {
    	rate float64
    	err  error
    }

    func (s stubRates) Rate(ctx context.Context, from, to string) (float64, error) {
    	// ?
    	return 0, nil
    }

    // memStore is a fake Store that keeps transactions in memory, per account,
    // in the order they were saved.
    type memStore struct {
    	mu   sync.Mutex
    	txns map[string][]Transaction
    }

    func newMemStore() *memStore {
    	// ?
    	return &memStore{}
    }

    // Save records tx under tx.Account.
    func (m *memStore) Save(ctx context.Context, tx Transaction) error {
    	// ?
    	return nil
    }

    // List returns a copy of account's transactions, oldest first. Changing
    // the returned slice must not change what's stored.
    func (m *memStore) List(ctx context.Context, account string) ([]Transaction, error) {
    	// ?
    	return nil, nil
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    type Transaction struct {
    	Account string
    	Amount  Cents
    }

    var ErrNoRate = errors.New("ledgerly: no exchange rate")

    // RateSource looks up exchange rates, for example from a bank's API.
    type RateSource interface {
    	Rate(ctx context.Context, from, to string) (float64, error)
    }

    // Store keeps transactions, for example in a database.
    type Store interface {
    	Save(ctx context.Context, tx Transaction) error
    	List(ctx context.Context, account string) ([]Transaction, error)
    }

    // Convert converts amount between currencies, rounding to the nearest cent.
    func Convert(ctx context.Context, rates RateSource, amount Cents, from, to string) (Cents, error) {
    	if from == to {
    		return amount, nil
    	}
    	r, err := rates.Rate(ctx, from, to)
    	if err != nil {
    		return 0, fmt.Errorf("convert %s to %s: %w", from, to, err)
    	}
    	return Cents(math.Round(float64(amount) * r)), nil
    }

    // RecordForeign converts tx.Amount from currency to USD and saves it.
    // Nothing is saved if the conversion fails.
    func RecordForeign(ctx context.Context, s Store, rates RateSource, tx Transaction, currency string) error {
    	usd, err := Convert(ctx, rates, tx.Amount, currency, "USD")
    	if err != nil {
    		return err
    	}
    	tx.Amount = usd
    	return s.Save(ctx, tx)
    }

    func main() {
    	ctx := context.Background()
    	store := newMemStore()

    	err := RecordForeign(ctx, store, stubRates{rate: 1.5}, Transaction{Account: "travel", Amount: 1000}, "EUR")
    	fmt.Println("with rate 1.5:  ", err)
    	err = RecordForeign(ctx, store, stubRates{err: ErrNoRate}, Transaction{Account: "travel", Amount: 2000}, "JPY")
    	fmt.Println("with no rate:   ", err)

    	txns, err := store.List(ctx, "travel")
    	fmt.Println("travel:         ", txns, err)
    	if len(txns) > 0 {
    		txns[0].Amount = 999999 // must not change the store
    	}
    	txns, _ = store.List(ctx, "travel")
    	fmt.Println("travel again:   ", txns)
    }
  solution: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"math"
    	"slices"
    	"sync"
    )

    // stubRates is a RateSource that returns whatever rate and err it holds.
    type stubRates struct {
    	rate float64
    	err  error
    }

    func (s stubRates) Rate(ctx context.Context, from, to string) (float64, error) {
    	return s.rate, s.err
    }

    // memStore is a fake Store that keeps transactions in memory, per account,
    // in the order they were saved.
    type memStore struct {
    	mu   sync.Mutex
    	txns map[string][]Transaction
    }

    func newMemStore() *memStore {
    	return &memStore{txns: map[string][]Transaction{}}
    }

    // Save records tx under tx.Account.
    func (m *memStore) Save(ctx context.Context, tx Transaction) error {
    	m.mu.Lock()
    	defer m.mu.Unlock()
    	m.txns[tx.Account] = append(m.txns[tx.Account], tx)
    	return nil
    }

    // List returns a copy of account's transactions, oldest first. Changing
    // the returned slice must not change what's stored.
    func (m *memStore) List(ctx context.Context, account string) ([]Transaction, error) {
    	m.mu.Lock()
    	defer m.mu.Unlock()
    	return slices.Clone(m.txns[account]), nil
    }

    // ---- Ledgerly code (already done) ----

    type Cents int64

    type Transaction struct {
    	Account string
    	Amount  Cents
    }

    var ErrNoRate = errors.New("ledgerly: no exchange rate")

    // RateSource looks up exchange rates, for example from a bank's API.
    type RateSource interface {
    	Rate(ctx context.Context, from, to string) (float64, error)
    }

    // Store keeps transactions, for example in a database.
    type Store interface {
    	Save(ctx context.Context, tx Transaction) error
    	List(ctx context.Context, account string) ([]Transaction, error)
    }

    // Convert converts amount between currencies, rounding to the nearest cent.
    func Convert(ctx context.Context, rates RateSource, amount Cents, from, to string) (Cents, error) {
    	if from == to {
    		return amount, nil
    	}
    	r, err := rates.Rate(ctx, from, to)
    	if err != nil {
    		return 0, fmt.Errorf("convert %s to %s: %w", from, to, err)
    	}
    	return Cents(math.Round(float64(amount) * r)), nil
    }

    // RecordForeign converts tx.Amount from currency to USD and saves it.
    // Nothing is saved if the conversion fails.
    func RecordForeign(ctx context.Context, s Store, rates RateSource, tx Transaction, currency string) error {
    	usd, err := Convert(ctx, rates, tx.Amount, currency, "USD")
    	if err != nil {
    		return err
    	}
    	tx.Amount = usd
    	return s.Save(ctx, tx)
    }

    func main() {
    	ctx := context.Background()
    	store := newMemStore()

    	err := RecordForeign(ctx, store, stubRates{rate: 1.5}, Transaction{Account: "travel", Amount: 1000}, "EUR")
    	fmt.Println("with rate 1.5:  ", err)
    	err = RecordForeign(ctx, store, stubRates{err: ErrNoRate}, Transaction{Account: "travel", Amount: 2000}, "JPY")
    	fmt.Println("with no rate:   ", err)

    	txns, err := store.List(ctx, "travel")
    	fmt.Println("travel:         ", txns, err)
    	if len(txns) > 0 {
    		txns[0].Amount = 999999 // must not change the store
    	}
    	txns, _ = store.List(ctx, "travel")
    	fmt.Println("travel again:   ", txns)
    }
  tests: |
    package main

    import (
    	"errors"
    	"slices"
    	"testing"
    )

    func TestStubRates(t *testing.T) {
    	var src RateSource = stubRates{rate: 1.25}
    	if r, err := src.Rate(t.Context(), "EUR", "USD"); r != 1.25 || err != nil {
    		t.Errorf("stubRates{rate: 1.25}.Rate() = %v, %v; want 1.25, nil", r, err)
    	}
    	boom := errors.New("rates offline")
    	if r, err := (stubRates{err: boom}).Rate(t.Context(), "EUR", "USD"); err != boom || r != 0 {
    		t.Errorf("stubRates{err: boom}.Rate() = %v, %v; want 0, boom", r, err)
    	}
    }

    func TestConvertWithStub(t *testing.T) {
    	got, err := Convert(t.Context(), stubRates{rate: 1.5}, 1001, "EUR", "USD")
    	if err != nil || got != 1502 {
    		t.Errorf("Convert(1001 EUR) with rate 1.5 = %d, %v; want 1502, nil", got, err)
    	}
    	_, err = Convert(t.Context(), stubRates{err: ErrNoRate}, 1000, "EUR", "USD")
    	if !errors.Is(err, ErrNoRate) {
    		t.Errorf("Convert with a failing rate source: error = %v, want it to wrap ErrNoRate", err)
    	}
    }

    func TestMemStoreSaveAndList(t *testing.T) {
    	var s Store = newMemStore()
    	ctx := t.Context()
    	saves := []Transaction{{"travel", 100}, {"rent", -120000}, {"travel", 250}}
    	for _, tx := range saves {
    		if err := s.Save(ctx, tx); err != nil {
    			t.Fatalf("Save(%+v) error: %v", tx, err)
    		}
    	}
    	got, err := s.List(ctx, "travel")
    	if err != nil {
    		t.Fatalf("List(travel) error: %v", err)
    	}
    	if want := []Transaction{{"travel", 100}, {"travel", 250}}; !slices.Equal(got, want) {
    		t.Errorf("List(travel) = %+v, want %+v (only that account, oldest first)", got, want)
    	}
    	if got, err := s.List(ctx, "savings"); len(got) != 0 || err != nil {
    		t.Errorf("List(savings) with nothing saved = %+v, %v; want no transactions and no error", got, err)
    	}
    }

    func TestMemStoreListReturnsCopy(t *testing.T) {
    	s := newMemStore()
    	ctx := t.Context()
    	s.Save(ctx, Transaction{"travel", 100})
    	first, _ := s.List(ctx, "travel")
    	if len(first) != 1 {
    		t.Fatalf("List(travel) after one Save returned %d transactions, want 1", len(first))
    	}
    	first[0].Amount = 999999
    	again, _ := s.List(ctx, "travel")
    	if again[0].Amount != 100 {
    		t.Errorf("changing the slice List returned changed the store (Amount is now %d); return a copy, e.g. with slices.Clone", again[0].Amount)
    	}
    }

    func TestRecordForeignWithDoubles(t *testing.T) {
    	store := newMemStore()
    	ctx := t.Context()
    	if err := RecordForeign(ctx, store, stubRates{rate: 1.5}, Transaction{"travel", 1000}, "EUR"); err != nil {
    		t.Fatalf("RecordForeign with rate 1.5: %v", err)
    	}
    	err := RecordForeign(ctx, store, stubRates{err: ErrNoRate}, Transaction{"travel", 2000}, "JPY")
    	if !errors.Is(err, ErrNoRate) {
    		t.Errorf("RecordForeign with no rate: error = %v, want ErrNoRate", err)
    	}
    	got, _ := store.List(ctx, "travel")
    	if want := []Transaction{{"travel", 1500}}; !slices.Equal(got, want) {
    		t.Errorf("after one good and one failed RecordForeign, List(travel) = %+v, want %+v", got, want)
    	}
    }
---

"Mock" is often used for any test double, but the different kinds do different jobs. Knowing the vocabulary helps you pick the simplest one that works. Here they are, all hand-written, for Ledgerly's low-balance alerts:

```go
type RateSource interface {
	Rate(ctx context.Context, from, to string) (float64, error)
}

type Notifier interface {
	Notify(ctx context.Context, account, msg string) error
}

type Store interface {
	Save(ctx context.Context, tx Transaction) error
	List(ctx context.Context, account string) ([]Transaction, error)
}
```

## Stubs: canned answers

A **stub** returns whatever the test tells it to. It exists to steer the code under test down a particular path:

```go
type stubRates struct {
	rate float64
	err  error
}

func (s stubRates) Rate(context.Context, string, string) (float64, error) {
	return s.rate, s.err
}
```

```go
func TestConvertUsesRate(t *testing.T) {
	got, err := Convert(t.Context(), stubRates{rate: 1.5}, 1000, "EUR", "USD")
	// want 1500, nil
}

func TestConvertRateUnavailable(t *testing.T) {
	_, err := Convert(t.Context(), stubRates{err: ErrNoRate}, 1000, "EUR", "USD")
	// want errors.Is(err, ErrNoRate)
}
```

Stubs are the best way to test error paths. Getting a real exchange-rate service to fail on cue is hard. Getting a stub to fail is one field.

## Spies: recording what happened

A **spy** records how it was called, so the test can check afterwards:

```go
type spyNotifier struct {
	calls []string // "account: msg"
}

func (s *spyNotifier) Notify(_ context.Context, account, msg string) error {
	s.calls = append(s.calls, account+": "+msg)
	return nil
}
```

```go
func TestLowBalanceAlert(t *testing.T) {
	var spy spyNotifier
	w := NewWatcher(&spy, 5000) // alert below $50.00
	w.Check(t.Context(), "cash", 4200)
	w.Check(t.Context(), "savings", 90000)

	want := []string{"cash: balance $42.00 is below $50.00"}
	if !slices.Equal(spy.calls, want) {
		t.Errorf("notifications = %q, want %q", spy.calls, want)
	}
}
```

Note the pointer receiver: the spy must be shared between the test and the code, or the calls would be recorded on a copy. If the code under test calls the spy from several goroutines, guard `calls` with a `sync.Mutex`.

## Fakes: working shortcuts

A **fake** is a real, working implementation that cuts a corner. The classic one is an in-memory store:

```go
type memStore struct {
	mu   sync.Mutex
	txns map[string][]Transaction
}

func newMemStore() *memStore { return &memStore{txns: map[string][]Transaction{}} }

func (m *memStore) Save(_ context.Context, tx Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txns[tx.Account] = append(m.txns[tx.Account], tx)
	return nil
}

func (m *memStore) List(_ context.Context, account string) ([]Transaction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return slices.Clone(m.txns[account]), nil
}
```

Tests using a fake read naturally: save some things, run the code, look at the results. They don't care *how* the code used the store, only what ended up in it, which makes them robust to refactoring. Fakes take more work up front but get reused across many tests. Many projects ship them in a `xxxtest` package (like `net/http/httptest` or `testing/fstest`) so every caller can use them.

## Mocks: checking expectations

A **mock** is a spy that knows what it expects and fails the test itself when something's off:

```go
type mockNotifier struct {
	t        *testing.T
	wantAcct string
	called   bool
}

func (m *mockNotifier) Notify(_ context.Context, account, msg string) error {
	m.t.Helper()
	if account != m.wantAcct {
		m.t.Errorf("Notify(%q, ...), want account %q", account, m.wantAcct)
	}
	m.called = true
	return nil
}
```

Mocking libraries generate these from interfaces and let you script sequences of expected calls. They're popular in some teams, but in Go the hand-written doubles above cover almost everything, with no generated code to maintain. The danger with heavy mocking is **over-specification**: tests that assert every call, in order, with exact arguments. Those break on every refactor and mostly prove that the code does what the code does.

## Your turn: a stub and a fake

`RecordForeign` converts a foreign-currency transaction to US dollars and saves it. In production it talks to a bank's exchange-rate API and a database. Write the two doubles that let a test run it in microseconds:

1. **`stubRates`** is a stub. `Rate` ignores its arguments and returns the struct's `rate` and `err`, so a test can pick the path: `stubRates{rate: 1.5}` or `stubRates{err: ErrNoRate}`.
2. **`memStore`** is a fake. `newMemStore` returns a ready-to-use store (don't forget to make the map). `Save` appends the transaction under its account. `List` returns that account's transactions, oldest first, as a **copy**: a caller that changes the returned slice mustn't change what's stored. `slices.Clone` does it. Lock the mutex in both methods, as in the lesson, so the fake is safe if the code under test is concurrent.

The grader checks each double on its own, then uses both to test `Convert` and `RecordForeign`, including the rule that nothing is saved when the rate lookup fails.

## Choosing

| You need to... | Use |
| --- | --- |
| force a result or an error | stub |
| check that something was sent or called | spy |
| run realistic scenarios against a stateful dependency | fake |
| fail immediately on an unexpected interaction | mock (sparingly) |
| check pure logic | no double at all |

Prefer, in order: the real thing, a fake, a stub or spy. Reach for a mock last.

## Further reading

- [Learn Go with Tests: Mocking](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/mocking)
