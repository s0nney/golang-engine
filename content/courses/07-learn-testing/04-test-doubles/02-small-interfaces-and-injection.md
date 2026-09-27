---
title: Small Interfaces and Dependency Injection
quiz:
  - question: |
      `Sync` only ever calls `client.Transactions(ctx)`, but it takes a
      `*BankClient` with 14 methods. What's the best change for testing?
    options:
      - text: Write a fake that implements all 14 methods
      - text: Declare a one-method interface next to `Sync` and accept that instead
        correct: true
      - text: Make `BankClient` a global variable that tests can replace
      - text: Use a mocking library to generate a fake `BankClient`
    explanation: |
      Go interfaces are satisfied implicitly, so the consumer can declare
      exactly what it needs. `*BankClient` satisfies the new one-method
      interface without any changes, and a test double needs only one
      method.
  - question: |
      Does this compile, given that `BankClient` was written before
      `TransactionSource` existed and doesn't mention it?

      ```go
      type TransactionSource interface {
          Transactions(ctx context.Context) ([]Transaction, error)
      }

      var _ TransactionSource = (*BankClient)(nil)
      ```
    options:
      - text: Yes, if `*BankClient` has a `Transactions` method with that exact signature
        correct: true
      - text: No, `BankClient` must declare `implements TransactionSource`
      - text: No, you can't assign `nil` to an interface
      - text: Only if they're in the same file
    explanation: |
      Satisfying an interface only requires the right method set. The
      `var _ Iface = (*T)(nil)` line is a compile-time check people use to
      prove a type still implements an interface.
---

In Go, you don't need a framework to inject dependencies. You pass them in. The trick is choosing *what type* to pass.

## Define interfaces where they're used

In many languages, the package that provides a type also publishes an interface for it (`IBankClient`), and everybody depends on that. Go turns this around. Because interfaces are satisfied implicitly, the **consumer** declares the small interface it needs:

```go
// package ledgerly/banksync

// TransactionSource is anything that can list transactions.
type TransactionSource interface {
	Transactions(ctx context.Context) ([]Transaction, error)
}
```

The bank package never hears about it. `*bank.Client` has a `Transactions` method with that signature, so it fits. So does any test double you write. Small interfaces mean small doubles: one method to fake instead of fourteen.

The Go proverb is *"The bigger the interface, the weaker the abstraction."* `io.Reader` has one method, and that's why everything implements it.

## Accept interfaces, return structs

The flip side: constructors should return concrete types.

```go
func NewClient(baseURL, token string) *Client { ... } // concrete
func Sync(ctx context.Context, l *Ledger, src TransactionSource, now time.Time) error // interface
```

Returning `*Client` lets callers use all its methods. Accepting `TransactionSource` lets `Sync` work with anything that fits. You don't have to predict every consumer's needs up front.

## Three ways to inject

**1. Function parameters**, for dependencies used once:

```go
func Sync(ctx context.Context, l *Ledger, src TransactionSource, now time.Time) error
```

**2. Struct fields set by a constructor**, for long-lived services:

```go
type Syncer struct {
	src    TransactionSource
	ledger *Ledger
	now    func() time.Time
}

func NewSyncer(src TransactionSource, l *Ledger) *Syncer {
	return &Syncer{src: src, ledger: l, now: time.Now}
}
```

Production code gets `time.Now`. A test in the same package can build a `Syncer` directly, or set `s.now` to return a fixed time, without widening the public API.

**3. Function types**, when the dependency is one operation. A named func type can even satisfy a one-method interface, the way `http.HandlerFunc` does:

```go
type SourceFunc func(ctx context.Context) ([]Transaction, error)

func (f SourceFunc) Transactions(ctx context.Context) ([]Transaction, error) {
	return f(ctx)
}
```

Now a test can write a double inline:

```go
src := SourceFunc(func(context.Context) ([]Transaction, error) {
	return nil, errors.New("bank is down")
})
err := Sync(t.Context(), &Ledger{}, src, testNow)
```

## Wiring it all up

Somewhere, real implementations have to be plugged in. That's `main`'s job (or a small `run` function called from `main`):

```go
func main() {
	client := bank.NewClient(os.Getenv("BANK_URL"), os.Getenv("BANK_TOKEN"))
	l := ledgerly.Load("ledger.csv")
	if err := banksync.Sync(context.Background(), l, client, time.Now()); err != nil {
		log.Fatal(err)
	}
}
```

This is all "dependency injection" means: construct things at the top, pass them down. No container, no reflection, no annotations.

## A compile-time check

If you want the compiler to confirm a type still satisfies an interface, add:

```go
var _ TransactionSource = (*bank.Client)(nil)
```

It costs nothing at runtime, and if someone changes `Transactions`' signature, the build fails right here instead of somewhere confusing.

## Further reading

- [Learn Go with Tests: Dependency Injection](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/dependency-injection)
