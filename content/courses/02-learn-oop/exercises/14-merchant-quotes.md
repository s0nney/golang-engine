---
title: Merchant Quotes
difficulty: medium
after: design-in-practice
hints:
  - '`Shop` stores the `PriceList` **interface** it was given, never a copy of the prices, and calls `s.prices.Price(item)` inside `Quote`. That''s what lets the tests hand in their own price list and change prices between quotes.'
  - 'Functional options: `type Option func(*Shop)`. `NewShop` builds a `Shop` with the defaults first (0% discount, 0% tax, a do-nothing ledger), then runs each option on it in order, so a later option overwrites an earlier one. Clamp percentages inside the option with `min` and `max`.'
  - 'A default `Ledger` that does nothing (`type noLedger struct{}` with an empty `Record`) saves a `nil` check in `Quote`; `WithLedger(nil)` can install it too. Only call `Record` once the whole order has been validated. For the maths: `total := subtotal * (100 - discount) / 100`, then `total += total * tax / 100`.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var (
    	ErrUnknownItem = errors.New("unknown item")
    	ErrBadQty      = errors.New("quantity must be at least 1")
    )

    // PriceList looks up the price of one unit of an item.
    type PriceList interface {
    	Price(item string) (gold int, ok bool)
    }

    // Ledger records completed sales.
    type Ledger interface {
    	Record(total int)
    }

    // Prices is a simple map-backed PriceList.
    type Prices map[string]int

    func (p Prices) Price(item string) (int, bool) {
    	return 0, false
    }

    // Line is one line of an order.
    type Line struct {
    	Item string
    	Qty  int
    }

    type Shop struct {
    }

    type Option func(*Shop)

    func WithDiscount(percent int) Option {
    	return func(s *Shop) {}
    }

    func WithTax(percent int) Option {
    	return func(s *Shop) {}
    }

    func WithLedger(l Ledger) Option {
    	return func(s *Shop) {}
    }

    func NewShop(prices PriceList, opts ...Option) *Shop {
    	return &Shop{}
    }

    func (s *Shop) Quote(order []Line) (int, error) {
    	return 0, nil
    }

    func main() {
    	prices := Prices{"potion": 25, "rope": 4, "lantern": 30}
    	shop := NewShop(prices, WithDiscount(10), WithTax(5))
    	fmt.Println(shop.Quote([]Line{{"potion", 2}, {"rope", 5}})) // want 66 <nil>
    	fmt.Println(shop.Quote([]Line{{"dragon egg", 1}}))          // want 0 and an unknown item error
    	fmt.Println(NewShop(prices).Quote([]Line{{"lantern", 1}}))  // want 30 <nil>
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    var (
    	ErrUnknownItem = errors.New("unknown item")
    	ErrBadQty      = errors.New("quantity must be at least 1")
    )

    // PriceList looks up the price of one unit of an item.
    type PriceList interface {
    	Price(item string) (gold int, ok bool)
    }

    // Ledger records completed sales.
    type Ledger interface {
    	Record(total int)
    }

    // Prices is a simple map-backed PriceList.
    type Prices map[string]int

    func (p Prices) Price(item string) (int, bool) {
    	gold, ok := p[item]
    	return gold, ok
    }

    var _ PriceList = Prices(nil)

    // Line is one line of an order.
    type Line struct {
    	Item string
    	Qty  int
    }

    // Shop quotes orders using an injected price list.
    type Shop struct {
    	prices   PriceList
    	ledger   Ledger
    	discount int // percent
    	tax      int // percent
    }

    // Option configures a Shop.
    type Option func(*Shop)

    // WithDiscount takes percent off every order (clamped to 0..100).
    func WithDiscount(percent int) Option {
    	return func(s *Shop) { s.discount = min(100, max(0, percent)) }
    }

    // WithTax adds percent tax after any discount (negative counts as 0).
    func WithTax(percent int) Option {
    	return func(s *Shop) { s.tax = max(0, percent) }
    }

    // WithLedger records every successful quote in l. A nil l turns recording off.
    func WithLedger(l Ledger) Option {
    	return func(s *Shop) {
    		if l == nil {
    			l = noLedger{}
    		}
    		s.ledger = l
    	}
    }

    // noLedger is the default Ledger: it forgets everything.
    type noLedger struct{}

    func (noLedger) Record(int) {}

    // NewShop returns a shop that prices orders with prices.
    func NewShop(prices PriceList, opts ...Option) *Shop {
    	s := &Shop{prices: prices, ledger: noLedger{}}
    	for _, opt := range opts {
    		opt(s)
    	}
    	return s
    }

    // Quote returns the total for order, or an error if any line is invalid.
    func (s *Shop) Quote(order []Line) (int, error) {
    	subtotal := 0
    	for _, line := range order {
    		if line.Qty < 1 {
    			return 0, fmt.Errorf("%s x%d: %w", line.Item, line.Qty, ErrBadQty)
    		}
    		price, ok := s.prices.Price(line.Item)
    		if !ok {
    			return 0, fmt.Errorf("%q: %w", line.Item, ErrUnknownItem)
    		}
    		subtotal += price * line.Qty
    	}
    	total := subtotal * (100 - s.discount) / 100
    	total += total * s.tax / 100
    	s.ledger.Record(total)
    	return total, nil
    }

    func main() {
    	prices := Prices{"potion": 25, "rope": 4, "lantern": 30}
    	shop := NewShop(prices, WithDiscount(10), WithTax(5))
    	fmt.Println(shop.Quote([]Line{{"potion", 2}, {"rope", 5}})) // 70 * 90% = 63, + 3 tax = 66
    	fmt.Println(shop.Quote([]Line{{"dragon egg", 1}}))
    	fmt.Println(NewShop(prices).Quote([]Line{{"lantern", 1}}))
    }
  tests: |
    package main

    import (
    	"errors"
    	"slices"
    	"strings"
    	"testing"
    )

    var stock = Prices{"potion": 25, "rope": 4, "lantern": 30, "map": 7}

    func TestPrices(t *testing.T) {
    	if g, ok := stock.Price("rope"); g != 4 || !ok {
    		t.Errorf("Prices.Price(\"rope\") = %d, %v, want 4, true", g, ok)
    	}
    	if g, ok := stock.Price("crown"); g != 0 || ok {
    		t.Errorf("Prices.Price(\"crown\") = %d, %v, want 0, false", g, ok)
    	}
    }

    func TestQuote(t *testing.T) {
    	order := []Line{{"potion", 2}, {"rope", 5}} // 70 gold
    	tests := []struct {
    		name string
    		opts []Option
    		want int
    	}{
    		{"no options", nil, 70},
    		{"10% off", []Option{WithDiscount(10)}, 63},
    		{"5% tax", []Option{WithTax(5)}, 73},                               // 70 + 3.5, rounded down
    		{"discount, then tax", []Option{WithDiscount(10), WithTax(5)}, 66}, // 63 + 3
    		{"option order doesn't change the maths", []Option{WithTax(5), WithDiscount(10)}, 66},
    		{"the last discount wins", []Option{WithDiscount(10), WithDiscount(50)}, 35},
    		{"discount above 100% is 100%", []Option{WithDiscount(150), WithTax(20)}, 0},
    		{"negative discount is 0%", []Option{WithDiscount(-20)}, 70},
    		{"negative tax is 0%", []Option{WithTax(-5)}, 70},
    		{"33% off rounds down", []Option{WithDiscount(33)}, 46}, // 46.9
    	}
    	for _, tt := range tests {
    		got, err := NewShop(stock, tt.opts...).Quote(order)
    		if got != tt.want || err != nil {
    			t.Errorf("%s: Quote(2 potions, 5 rope) = %d, %v, want %d, <nil>", tt.name, got, err, tt.want)
    		}
    	}
    	if got, err := NewShop(stock, WithTax(50)).Quote(nil); got != 0 || err != nil {
    		t.Errorf("Quote(empty order) = %d, %v, want 0, <nil>", got, err)
    	}
    }

    func TestQuoteErrors(t *testing.T) {
    	shop := NewShop(stock)
    	tests := []struct {
    		order   []Line
    		want    error
    		mention string
    	}{
    		{[]Line{{"potion", 1}, {"dragon egg", 1}}, ErrUnknownItem, "dragon egg"},
    		{[]Line{{"rope", 0}}, ErrBadQty, "rope"},
    		{[]Line{{"map", -2}, {"crown", 1}}, ErrBadQty, "map"},
    		{[]Line{{"crown", 1}, {"map", -2}}, ErrUnknownItem, "crown"},
    		{[]Line{{"crown", 0}}, ErrBadQty, "crown"},
    	}
    	for _, tt := range tests {
    		got, err := shop.Quote(tt.order)
    		if got != 0 || !errors.Is(err, tt.want) {
    			t.Errorf("Quote(%v) = %d, %v, want 0 and an error wrapping %q (report the first bad line)", tt.order, got, err, tt.want)
    		} else if !strings.Contains(err.Error(), tt.mention) {
    			t.Errorf("Quote(%v) error %q should name the item %q", tt.order, err, tt.mention)
    		}
    	}
    }

    // spyPrices is the test's own PriceList. It counts lookups, and its prices
    // can change between quotes, as a merchant's do.
    type spyPrices struct {
    	gold    map[string]int
    	lookups int
    }

    func (s *spyPrices) Price(item string) (int, bool) {
    	s.lookups++
    	g, ok := s.gold[item]
    	return g, ok
    }

    func TestUsesInjectedPriceList(t *testing.T) {
    	spy := &spyPrices{gold: map[string]int{"scale": 100}}
    	shop := NewShop(spy, WithDiscount(20))
    	if got, _ := shop.Quote([]Line{{"scale", 3}}); got != 240 {
    		t.Errorf("Quote(3 scales at 100, 20%% off) = %d, want 240", got)
    	}
    	if spy.lookups != 1 {
    		t.Errorf("the shop called Price %d times for a one-line order, want 1", spy.lookups)
    	}
    	spy.gold["scale"] = 50 // prices changed after NewShop
    	spy.gold["claw"] = 10
    	if got, _ := shop.Quote([]Line{{"scale", 1}, {"claw", 2}}); got != 56 {
    		t.Errorf("after the price list changed, Quote = %d, want 56: ask the PriceList on every quote", got)
    	}
    }

    // book is the test's own Ledger.
    type book struct{ totals []int }

    func (b *book) Record(total int) { b.totals = append(b.totals, total) }

    func TestLedger(t *testing.T) {
    	b := &book{}
    	shop := NewShop(stock, WithLedger(b), WithTax(10))
    	shop.Quote([]Line{{"lantern", 1}})
    	shop.Quote([]Line{{"lantern", 1}, {"crown", 1}}) // fails: not recorded
    	shop.Quote([]Line{{"map", 3}})
    	shop.Quote(nil)
    	if !slices.Equal(b.totals, []int{33, 23, 0}) {
    		t.Errorf("ledger recorded %v, want [33 23 0]: one entry per successful quote, none for failures", b.totals)
    	}

    	for _, s := range []*Shop{NewShop(stock), NewShop(stock, WithLedger(b), WithLedger(nil))} {
    		if got, err := s.Quote([]Line{{"rope", 1}}); got != 4 || err != nil {
    			t.Errorf("shop without a ledger: Quote = %d, %v, want 4, <nil>", got, err)
    		}
    	}
    	if len(b.totals) > 3 {
    		t.Errorf("WithLedger(nil) should turn recording off, but the ledger got %v", b.totals)
    	}
    }

    func TestShopsAreIndependent(t *testing.T) {
    	cheap := NewShop(stock, WithDiscount(50))
    	full := NewShop(stock)
    	cheap.Quote([]Line{{"potion", 1}})
    	if got, _ := full.Quote([]Line{{"potion", 1}}); got != 25 {
    		t.Errorf("a shop with no options quoted %d for a potion, want 25: are options stored in a package variable?", got)
    	}
    }
---

The travelling merchant's shop has to work in every town: each town has its own
price list, some towns tax sales, and the guild wants a ledger of every sale.
Build `Shop` so all of that is **injected** rather than hard-coded:

- `NewShop(prices, opts...)` takes a `PriceList` interface (the shop never cares
  where prices come from) and any number of **functional options**:
  - `WithDiscount(percent)`: take `percent`% off the order. Values outside
    0 to 100 are clamped into that range.
  - `WithTax(percent)`: add `percent`% tax, computed **after** the discount. A
    negative tax counts as 0.
  - `WithLedger(l)`: call `l.Record(total)` after every **successful** quote.
    `WithLedger(nil)` turns recording off.

  Options apply in order, so when the same option is given twice, the last one
  wins. With no options there's no discount, no tax and no ledger.
- `Quote(order)` prices a slice of `Line{Item, Qty}` using the price list **at
  the time of the quote** (prices change). It returns the total in whole gold:
  `subtotal * (100 - discount) / 100`, then plus `that * tax / 100`, both with
  integer division. An empty order costs 0.
- A line with `Qty < 1` is an error wrapping `ErrBadQty`, and an item the price
  list doesn't know is an error wrapping `ErrUnknownItem`. Report the **first**
  bad line (checking its quantity before its price), mention the item's name in
  the message, and return 0.
- `Prices` is a ready-made map-backed `PriceList`: give it its `Price` method.

## Example

```go
prices := Prices{"potion": 25, "rope": 4, "lantern": 30}
shop := NewShop(prices, WithDiscount(10), WithTax(5))

shop.Quote([]Line{{"potion", 2}, {"rope", 5}}) // 66, nil: 70 → 63 after 10% off → 66 with 5% tax
shop.Quote([]Line{{"dragon egg", 1}})          // 0, error wrapping ErrUnknownItem
NewShop(prices).Quote([]Line{{"lantern", 1}})  // 30, nil
```

## Constraints

- The tests pass in their own `PriceList` and `Ledger` implementations and
  check that the shop really uses them.
