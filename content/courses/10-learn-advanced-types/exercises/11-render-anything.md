---
title: Render Anything
difficulty: medium
after: interfaces-in-depth
hints:
  - 'You can''t type-assert on a type parameter directly, but you can on an interface value: `v := any(items[i])`. Check `v == nil` first, then `v.(fmt.Stringer)`, then `any(&items[i]).(fmt.Stringer)`, then `v.(error)`.'
  - 'Why `&items[i]`? A method with a pointer receiver is in the method set of `*T` but not of `T`. Slice elements are addressable, so taking `&items[i]` (not `&v` of a copy) gives the method access to the real element.'
  - 'For the rest, a type switch: `case string`, `case bool`, `case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr`, `case float32`, `case float64`. A case lists **exact** types, so a named `type Flag bool` falls through to `default`, which is exactly what the spec wants.'
exercise:
  starter: |
    package main

    import "fmt"

    func Render[T any](items []T) []string {
    	return nil
    }

    // Celsius has a String method on the POINTER receiver.
    type Celsius float64

    func (c *Celsius) String() string { return fmt.Sprintf("%.1f°C", float64(*c)) }

    type Flag bool

    func main() {
    	fmt.Printf("%q\n", Render([]any{"hi", 42, 2.50, true, nil, Flag(true), fmt.Errorf("boom")}))
    	// want ["\"hi\"" "42" "2.5" "yes" "nil" "true" "boom"]
    	fmt.Printf("%q\n", Render([]Celsius{21.5, -3}))
    	// want ["21.5°C" "-3.0°C"]
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strconv"
    )

    // Render turns every item into display text.
    func Render[T any](items []T) []string {
    	out := make([]string, len(items))
    	for i := range items {
    		out[i] = render(any(items[i]), any(&items[i]))
    	}
    	return out
    }

    func render(v, ptr any) string {
    	if v == nil {
    		return "nil"
    	}
    	if s, ok := v.(fmt.Stringer); ok {
    		return s.String()
    	}
    	if s, ok := ptr.(fmt.Stringer); ok {
    		return s.String()
    	}
    	if err, ok := v.(error); ok {
    		return err.Error()
    	}
    	switch x := v.(type) {
    	case string:
    		return strconv.Quote(x)
    	case bool:
    		if x {
    			return "yes"
    		}
    		return "no"
    	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
    		return fmt.Sprintf("%d", x)
    	case float32:
    		return strconv.FormatFloat(float64(x), 'g', -1, 32)
    	case float64:
    		return strconv.FormatFloat(x, 'g', -1, 64)
    	default:
    		return fmt.Sprintf("%v", x)
    	}
    }

    // Celsius has a String method on the POINTER receiver.
    type Celsius float64

    func (c *Celsius) String() string { return fmt.Sprintf("%.1f°C", float64(*c)) }

    type Flag bool

    func main() {
    	fmt.Printf("%q\n", Render([]any{"hi", 42, 2.50, true, nil, Flag(true), fmt.Errorf("boom")}))
    	fmt.Printf("%q\n", Render([]Celsius{21.5, -3}))
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"slices"
    	"testing"
    )

    // Key has a value-receiver String: both Key and *Key have it.
    type Key string

    func (k Key) String() string { return "key:" + string(k) }

    // Counter has a pointer-receiver String that reads the real element.
    type Counter struct{ N int }

    func (c *Counter) String() string { return fmt.Sprintf("counter(%d)", c.N) }

    // NotFound is an error type AND has a String method: String wins.
    type NotFound struct{ Key string }

    func (e NotFound) Error() string  { return "not found: " + e.Key }
    func (e NotFound) String() string { return "missing " + e.Key }

    type Level int

    type Point struct{ X, Y int }

    func check[T any](t *testing.T, name string, in []T, want []string) {
    	t.Helper()
    	got := Render(in)
    	if !slices.Equal(got, want) {
    		t.Errorf("%s: Render(%v) = %q, want %q", name, in, got, want)
    	}
    }

    func TestBasicKinds(t *testing.T) {
    	check(t, "strings", []string{"hi", "", `a"b`}, []string{`"hi"`, `""`, `"a\"b"`})
    	check(t, "ints", []int{0, -7, 1 << 40}, []string{"0", "-7", "1099511627776"})
    	check(t, "uint8", []uint8{255}, []string{"255"})
    	check(t, "bools", []bool{true, false}, []string{"yes", "no"})
    	check(t, "float64", []float64{2.5, 0.1, 1e21, -0}, []string{"2.5", "0.1", "1e+21", "0"})
    	check(t, "float32", []float32{0.1, 3}, []string{"0.1", "3"})
    	check(t, "empty", []int{}, []string{})
    }

    func TestAnyAndNil(t *testing.T) {
    	in := []any{"hi", 42, int8(-1), uint64(7), 2.50, true, nil, fmt.Errorf("boom"), Point{1, 2}, []int{1}}
    	want := []string{`"hi"`, "42", "-1", "7", "2.5", "yes", "nil", "boom", "{1 2}", "[1]"}
    	check(t, "mixed any", in, want)
    	check(t, "errors", []error{errors.New("disk full"), nil}, []string{"disk full", "nil"})
    }

    func TestNamedTypesUseDefault(t *testing.T) {
    	check(t, "named bool", []Flag{true, false}, []string{"true", "false"})
    	check(t, "named int", []Level{3}, []string{"3"})
    	check(t, "struct", []Point{{3, 4}}, []string{"{3 4}"})
    }

    func TestValueReceiverStringer(t *testing.T) {
    	check(t, "Key", []Key{"a", "b"}, []string{"key:a", "key:b"})
    	check(t, "Key in any", []any{Key("z")}, []string{"key:z"})
    	k := Key("p")
    	check(t, "*Key", []*Key{&k}, []string{"key:p"})
    }

    func TestPointerReceiverStringer(t *testing.T) {
    	check(t, "Celsius", []Celsius{21.5, -3}, []string{"21.5°C", "-3.0°C"})
    	check(t, "Counter", []Counter{{1}, {20}}, []string{"counter(1)", "counter(20)"})
    	check(t, "*Counter", []*Counter{{5}}, []string{"counter(5)"})
    	// A Counter stored in an any is a copy that isn't addressable: *Counter's
    	// String isn't reachable, so it falls through to the default.
    	check(t, "Counter in any", []any{Counter{9}}, []string{"{9}"})
    }

    func TestStringerBeatsError(t *testing.T) {
    	check(t, "NotFound", []NotFound{{"user:1"}}, []string{"missing user:1"})
    	check(t, "NotFound as error", []error{NotFound{"user:2"}}, []string{"missing user:2"})
    }

    func TestRenderDoesNotModify(t *testing.T) {
    	in := []Counter{{1}, {2}}
    	Render(in)
    	if in[0].N != 1 || in[1].N != 2 {
    		t.Errorf("Render changed its input to %v", in)
    	}
    }
---

Stash's admin console shows the contents of any collection as a column of
text. It gets handed a `[]T` for whatever `T` the collection holds, and each
value should be displayed in the most helpful way it supports.

Write `Render(items)`. It returns one string per item, in order, using the
**first** rule that applies:

1. A nil interface value (for example a `nil` in a `[]any` or `[]error`):
   `"nil"`.
2. The value has a `String() string` method (it's a `fmt.Stringer`): call it.
3. The value's **pointer** has a `String()` method (a pointer receiver on the
   element type): call it on the element in the slice.
4. The value is an `error`: its `Error()` text.
5. Otherwise, by exact type:
   - `string`: quoted with `strconv.Quote`, so `hi` becomes `"hi"`.
   - `bool`: `"yes"` or `"no"`.
   - any predeclared integer type (`int`, `int8`, ..., `uint64`, `uintptr`):
     its decimal value.
   - `float32` / `float64`: the shortest representation, as
     `strconv.FormatFloat(f, 'g', -1, bitSize)` gives it.
   - anything else, including named types like `type Flag bool`:
     `fmt.Sprintf("%v", v)`.

## Examples

```go
Render([]any{"hi", 42, 2.50, true, nil, Flag(true), fmt.Errorf("boom")})
// ["\"hi\"" "42" "2.5" "yes" "nil" "true" "boom"]

type Celsius float64
func (c *Celsius) String() string { ... }

Render([]Celsius{21.5, -3}) // ["21.5°C" "-3.0°C"]  (rule 3)
```

## Constraints

- Don't modify `items`.
- A value in a `[]any` is a copy that isn't addressable, so rule 3 can't reach
  pointer-receiver methods for it. The hidden tests check that too: a
  `Counter` with a pointer `String`, stored in a `[]any`, renders with `%v`.
- If a value is both a `Stringer` and an `error`, `String` wins.
