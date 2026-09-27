---
title: Setting Values
quiz:
  - question: |
      What does this print?

      ```go
      cfg := Config{Name: "stash"}
      v := reflect.ValueOf(cfg)
      fmt.Println(v.Field(0).CanSet())
      ```
    options:
      - text: '`true`'
      - text: '`false`, because `ValueOf(cfg)` holds a copy of `cfg`, and changing the copy would be pointless'
        correct: true
      - text: It panics
      - text: '`false`, because `Name` is a string'
    explanation: |
      `ValueOf` receives `cfg` through an `any`, which holds a copy (chapter 6). Reflection
      refuses to let you set a copy that nobody can see. Pass `&cfg` and call `.Elem()` to
      get a settable view of the original.
  - question: Which call **panics**, given `p := reflect.ValueOf(&cfg).Elem()` and an int8 field `Workers`?
    options:
      - text: '`p.FieldByName("Workers").SetInt(100)`'
      - text: '`p.FieldByName("Workers").OverflowInt(300)`'
      - text: '`p.FieldByName("Workers").SetString("4")`'
        correct: true
      - text: '`p.FieldByName("Workers").SetZero()`'
    explanation: |
      Every `Set` method checks the kind: `SetString` on an `int8` panics. `SetInt(100)`
      fits, `OverflowInt(300)` just reports `true` (300 doesn't fit in an `int8`), and
      `SetZero` works for any kind. `SetInt(300)` wouldn't panic, it would silently wrap,
      which is why you check `OverflowInt` first.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"reflect"
    	"strconv"
    )

    // ApplyDefaults fills zero-valued exported fields of the struct that ptr
    // points to from their `default:"..."` tags. Supported kinds: string, bool,
    // every int kind (reject values that overflow the field) and float64.
    // Non-zero fields are left alone. It returns an error if ptr isn't a
    // non-nil pointer to a struct, or a default can't be parsed; the error
    // must mention the field name.
    func ApplyDefaults(ptr any) error {
    	v := reflect.ValueOf(ptr)
    	if v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct {
    		return errors.New("ApplyDefaults: need a non-nil pointer to a struct")
    	}
    	// ?
    	_ = strconv.ParseInt
    	return nil
    }

    type Config struct {
    	Name    string  `default:"stash"`
    	Workers int8    `default:"4"`
    	Ratio   float64 `default:"0.75"`
    	Debug   bool    `default:"true"`
    	Port    int     `default:"8080"`
    	Notes   string
    }

    func main() {
    	cfg := Config{Port: 9090}
    	err := ApplyDefaults(&cfg)
    	fmt.Printf("%+v %v\n", cfg, err)
    	fmt.Println(ApplyDefaults(cfg))
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"reflect"
    	"strconv"
    )

    // ApplyDefaults fills zero-valued exported fields of the struct that ptr
    // points to from their `default:"..."` tags. Supported kinds: string, bool,
    // every int kind (reject values that overflow the field) and float64.
    // Non-zero fields are left alone. It returns an error if ptr isn't a
    // non-nil pointer to a struct, or a default can't be parsed; the error
    // must mention the field name.
    func ApplyDefaults(ptr any) error {
    	v := reflect.ValueOf(ptr)
    	if v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct {
    		return errors.New("ApplyDefaults: need a non-nil pointer to a struct")
    	}
    	s := v.Elem()
    	for f, fv := range s.Fields() {
    		def, ok := f.Tag.Lookup("default")
    		if !ok || !f.IsExported() || !fv.IsZero() {
    			continue
    		}
    		if err := set(fv, def); err != nil {
    			return fmt.Errorf("field %s: %w", f.Name, err)
    		}
    	}
    	return nil
    }

    func set(fv reflect.Value, s string) error {
    	switch fv.Kind() {
    	case reflect.String:
    		fv.SetString(s)
    	case reflect.Bool:
    		b, err := strconv.ParseBool(s)
    		if err != nil {
    			return err
    		}
    		fv.SetBool(b)
    	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    		n, err := strconv.ParseInt(s, 10, 64)
    		if err != nil {
    			return err
    		}
    		if fv.OverflowInt(n) {
    			return fmt.Errorf("%d overflows %s", n, fv.Type())
    		}
    		fv.SetInt(n)
    	case reflect.Float64:
    		x, err := strconv.ParseFloat(s, 64)
    		if err != nil {
    			return err
    		}
    		fv.SetFloat(x)
    	default:
    		return fmt.Errorf("unsupported kind %s", fv.Kind())
    	}
    	return nil
    }

    type Config struct {
    	Name    string  `default:"stash"`
    	Workers int8    `default:"4"`
    	Ratio   float64 `default:"0.75"`
    	Debug   bool    `default:"true"`
    	Port    int     `default:"8080"`
    	Notes   string
    }

    func main() {
    	cfg := Config{Port: 9090}
    	err := ApplyDefaults(&cfg)
    	fmt.Printf("%+v %v\n", cfg, err)
    	fmt.Println(ApplyDefaults(cfg))
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    )

    func TestAppliesDefaults(t *testing.T) {
    	cfg := Config{Port: 9090}
    	if err := ApplyDefaults(&cfg); err != nil {
    		t.Fatalf("ApplyDefaults returned %v", err)
    	}
    	want := Config{Name: "stash", Workers: 4, Ratio: 0.75, Debug: true, Port: 9090}
    	if cfg != want {
    		t.Errorf("after ApplyDefaults: %+v\nwant %+v (Port was set, so it keeps 9090)", cfg, want)
    	}
    }

    type Limits struct {
    	Max    int16 `default:"-7"`
    	hidden int   `default:"5"`
    	Label  string
    }

    func TestSkipsUnexportedAndUntagged(t *testing.T) {
    	var l Limits
    	if err := ApplyDefaults(&l); err != nil {
    		t.Fatalf("ApplyDefaults(&Limits{}) returned %v", err)
    	}
    	if l.Max != -7 || l.hidden != 0 || l.Label != "" {
    		t.Errorf("got %+v, want Max -7, hidden untouched (unexported), Label empty (no tag)", l)
    	}
    }

    type Tiny struct {
    	Level int8 `default:"300"`
    }

    type BadBool struct {
    	On bool `default:"maybe"`
    }

    func TestErrors(t *testing.T) {
    	if err := ApplyDefaults(&Tiny{}); err == nil || !strings.Contains(err.Error(), "Level") {
    		t.Errorf("an int8 default of 300 should fail with an error mentioning Level, got %v", err)
    	}
    	if err := ApplyDefaults(&BadBool{}); err == nil || !strings.Contains(err.Error(), "On") {
    		t.Errorf(`a bool default of "maybe" should fail with an error mentioning On, got %v`, err)
    	}
    	if err := ApplyDefaults(Config{}); err == nil {
    		t.Errorf("ApplyDefaults(non-pointer) should fail")
    	}
    	var nilCfg *Config
    	if err := ApplyDefaults(nilCfg); err == nil {
    		t.Errorf("ApplyDefaults(nil *Config) should fail")
    	}
    }
---

Reading through reflection is safe. **Writing** through it is where the rules get strict, and the rules are all about one question: *would this change be visible to anyone?*

## Settability

A `reflect.Value` is **settable** only if it refers to a real, addressable location that your code is allowed to modify. Check with `CanSet()`:

```go
package main

import (
	"fmt"
	"reflect"
)

type Config struct {
	Name    string
	Workers int8
	secret  string
}

func main() {
	cfg := Config{Name: "stash", Workers: 4}

	v := reflect.ValueOf(cfg)
	fmt.Println(v.Field(0).CanSet()) // a copy: not settable

	p := reflect.ValueOf(&cfg).Elem() // the struct cfg itself
	fmt.Println(p.Field(0).CanSet(), p.Field(2).CanSet())

	p.Field(0).SetString("stash-prod")
	w := p.FieldByName("Workers")
	fmt.Println(w.OverflowInt(300), w.OverflowInt(100))
	w.SetInt(100)
	fmt.Printf("%+v\n", cfg)

	p.Field(0).SetZero()
	fmt.Printf("%q\n", cfg.Name)

	defer func() { fmt.Println("panic:", recover()) }()
	p.Field(2).SetString("nope")
}
```

```
false
true false
true false
{Name:stash-prod Workers:100 secret:}
""
panic: reflect: reflect.Value.SetString using value obtained using unexported field
```

Step by step:

1. **`reflect.ValueOf(cfg)`** gets a *copy* of `cfg` (it went through an `any`). Setting a field of that copy would change nothing anyone can see, so it's not settable.
2. **`reflect.ValueOf(&cfg).Elem()`** starts from a pointer, and `Elem()` follows it to the real `cfg`. Now exported fields are settable. This "pointer, then `Elem`" move is the standard way in.
3. **Unexported fields** are never settable through reflection, even when their struct is. Trying panics with `using value obtained using unexported field`.
4. **Setters check kinds**: `SetString`, `SetInt`, `SetFloat`, `SetBool`, and the general `Set(x Value)` (which needs an assignable type). `SetInt` takes an `int64` for *every* int kind, so check `OverflowInt` before narrowing, otherwise 300 silently wraps in an `int8`.
5. **`SetZero()`** (Go 1.20) resets any settable value to its zero value.

## Creating values

To build a new value of a type known only at runtime, use `reflect.New(t)`: it returns a `Value` holding a `*T` pointing at a fresh zero `T`, so `.Elem()` is settable. `reflect.MakeSlice`, `reflect.MakeMap` and `reflect.Zero` cover the other common cases. That's how `encoding/json` fills in slices and maps it finds inside your structs.

## Zero checks

`v.IsZero()` reports whether a value equals its type's zero value, for any kind. It's the reflective answer to chapter 4's "you need `comparable` to compare with zero": reflection doesn't need a constraint, because it checks at runtime.

## Your turn

Stash configs carry defaults in tags: `` Workers int8 `default:"4"` ``. Write `ApplyDefaults(ptr any) error`:

- `ptr` must be a non-nil pointer to a struct (that check is done for you);
- for each **exported** field that is **zero** and has a `default` tag, parse the tag into the field: `string`, `bool` (`strconv.ParseBool`), every **int** kind (`strconv.ParseInt`, then reject values where `OverflowInt` is true), and `float64`;
- non-zero fields are left alone;
- on a parse or overflow error, return an error that includes the field's name (for example `field Level: ...`).

`s.Fields()` on the struct's `Value` yields each `StructField` together with its (settable) field `Value`.
