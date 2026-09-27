---
title: Walking Values by Kind
quiz:
  - question: |
      What is `reflect.ValueOf(nil).Kind()`?
    options:
      - text: '`reflect.Interface`'
      - text: '`reflect.Pointer`'
      - text: '`reflect.Invalid`'
        correct: true
      - text: It panics
    explanation: |
      A nil `any` has no dynamic type, so `ValueOf` returns the zero `reflect.Value`, whose
      kind is `Invalid`. Most methods panic on it, so recursive walkers should treat
      `Invalid` as "nothing here".
  - question: Why does the `Strings` walker check `v.IsNil()` before calling `v.Elem()` for pointers and interfaces?
    options:
      - text: '`Elem` is slow'
      - text: '`Elem` of a nil pointer returns the zero `Value` (kind `Invalid`), and without the check the walker would have to handle that everywhere; checking first keeps the recursion clean'
        correct: true
      - text: '`IsNil` is required before any method call'
      - text: To avoid an infinite loop
    explanation: |
      For a nil pointer or nil interface, `Elem()` returns an invalid `Value`. The walker
      would then call `Kind()` on it (fine, `Invalid`) but other code often calls methods
      that panic. Checking `IsNil` first is the clearest way to say "nothing to follow".
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"reflect"
    )

    // Weigh estimates how much data v carries:
    //
    //	bool                          1
    //	any int, uint or float kind   8
    //	string                        its length in bytes
    //	slice, array                  the sum of its elements
    //	map                           the sum of its keys and values
    //	struct                        the sum of its fields (exported or not)
    //	pointer, interface            0 if nil, else what it points to / holds
    //	anything else (nil, funcs...) 0
    func Weigh(v any) int {
    	return weigh(reflect.ValueOf(v))
    }

    func weigh(v reflect.Value) int {
    	switch v.Kind() {
    	case reflect.String:
    		return v.Len()
    		// ? the other kinds
    	}
    	return 0
    }

    type Meta struct {
    	Owner string
    	Size  int64
    }

    type Entry struct {
    	Key    string
    	Tags   []string
    	Meta   *Meta
    	Counts map[string]int
    	hidden bool
    }

    func main() {
    	e := Entry{
    		Key:    "logo.png",
    		Tags:   []string{"img", "brand"},
    		Meta:   &Meta{Owner: "ana", Size: 2048},
    		Counts: map[string]int{"views": 10},
    		hidden: true,
    	}
    	fmt.Println(Weigh(e))
    	fmt.Println(Weigh("hi"), Weigh(3.5), Weigh(nil), Weigh([]any{"ab", 1, nil}))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"reflect"
    )

    // Weigh estimates how much data v carries:
    //
    //	bool                          1
    //	any int, uint or float kind   8
    //	string                        its length in bytes
    //	slice, array                  the sum of its elements
    //	map                           the sum of its keys and values
    //	struct                        the sum of its fields (exported or not)
    //	pointer, interface            0 if nil, else what it points to / holds
    //	anything else (nil, funcs...) 0
    func Weigh(v any) int {
    	return weigh(reflect.ValueOf(v))
    }

    func weigh(v reflect.Value) int {
    	switch v.Kind() {
    	case reflect.String:
    		return v.Len()
    	case reflect.Bool:
    		return 1
    	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
    		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
    		reflect.Float32, reflect.Float64:
    		return 8
    	case reflect.Slice, reflect.Array:
    		total := 0
    		for i := range v.Len() {
    			total += weigh(v.Index(i))
    		}
    		return total
    	case reflect.Map:
    		total := 0
    		for k, val := range v.Seq2() {
    			total += weigh(k) + weigh(val)
    		}
    		return total
    	case reflect.Struct:
    		total := 0
    		for i := range v.NumField() {
    			total += weigh(v.Field(i))
    		}
    		return total
    	case reflect.Pointer, reflect.Interface:
    		if v.IsNil() {
    			return 0
    		}
    		return weigh(v.Elem())
    	}
    	return 0
    }

    type Meta struct {
    	Owner string
    	Size  int64
    }

    type Entry struct {
    	Key    string
    	Tags   []string
    	Meta   *Meta
    	Counts map[string]int
    	hidden bool
    }

    func main() {
    	e := Entry{
    		Key:    "logo.png",
    		Tags:   []string{"img", "brand"},
    		Meta:   &Meta{Owner: "ana", Size: 2048},
    		Counts: map[string]int{"views": 10},
    		hidden: true,
    	}
    	fmt.Println(Weigh(e))
    	fmt.Println(Weigh("hi"), Weigh(3.5), Weigh(nil), Weigh([]any{"ab", 1, nil}))
    }
  tests: |
    package main

    import "testing"

    type Bytes uint32

    func TestWeighBasics(t *testing.T) {
    	for _, tt := range []struct {
    		name string
    		v    any
    		want int
    	}{
    		{"string", "hello", 5},
    		{"bool", true, 1},
    		{"int", 42, 8},
    		{"float64", 2.5, 8},
    		{"named uint32", Bytes(7), 8},
    		{"int8", int8(1), 8},
    		{"nil", nil, 0},
    		{"func", func() {}, 0},
    	} {
    		if got := Weigh(tt.v); got != tt.want {
    			t.Errorf("Weigh(%s %#v) = %d, want %d", tt.name, tt.v, got, tt.want)
    		}
    	}
    }

    func TestWeighContainers(t *testing.T) {
    	if got := Weigh([]string{"ab", "cde"}); got != 5 {
    		t.Errorf(`Weigh([]string{"ab", "cde"}) = %d, want 5`, got)
    	}
    	if got := Weigh([3]int{1, 2, 3}); got != 24 {
    		t.Errorf("Weigh([3]int{1, 2, 3}) = %d, want 24", got)
    	}
    	if got := Weigh(map[string]int{"ab": 1, "c": 2}); got != 19 {
    		t.Errorf(`Weigh(map[string]int{"ab": 1, "c": 2}) = %d, want 19 (keys 3 + values 16)`, got)
    	}
    	if got := Weigh([]any{"ab", 1, nil, true}); got != 11 {
    		t.Errorf(`Weigh([]any{"ab", 1, nil, true}) = %d, want 11`, got)
    	}
    }

    func TestWeighStructsAndPointers(t *testing.T) {
    	var none *Meta
    	if got := Weigh(none); got != 0 {
    		t.Errorf("Weigh(nil *Meta) = %d, want 0", got)
    	}
    	if got := Weigh(&Meta{Owner: "ana", Size: 1}); got != 11 {
    		t.Errorf(`Weigh(&Meta{Owner: "ana", Size: 1}) = %d, want 11`, got)
    	}
    	e := Entry{
    		Key:    "logo.png",
    		Tags:   []string{"img", "brand"},
    		Meta:   &Meta{Owner: "ana", Size: 2048},
    		Counts: map[string]int{"views": 10},
    		hidden: true,
    	}
    	if got := Weigh(e); got != 41 {
    		t.Errorf("Weigh(the example Entry) = %d, want 41 (8 + 8 + 11 + 13 + 1, including the unexported bool)", got)
    	}
    }
---

Reflection code usually has one shape: a **recursive function that switches on `Kind`**, handles the leaves (strings, numbers, booleans) directly, and recurses into containers (slices, maps, structs, pointers). Once you've written one, `encoding/json`'s encoder won't look so mysterious.

## A walker

```go
package main

import (
	"fmt"
	"reflect"
)

type Entry struct {
	Key  string
	Tags []string
	Meta *Meta
}

type Meta struct {
	Owner string
	Size  int
}

// Strings collects every string reachable from v, depth first.
func Strings(v any) []string {
	var out []string
	collect(reflect.ValueOf(v), &out)
	return out
}

func collect(v reflect.Value, out *[]string) {
	switch v.Kind() {
	case reflect.String:
		*out = append(*out, v.String())
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			collect(v.Index(i), out)
		}
	case reflect.Struct:
		for i := range v.NumField() {
			collect(v.Field(i), out)
		}
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			collect(v.Elem(), out)
		}
	}
	// Other kinds (numbers, maps, funcs...) are skipped.
}

func main() {
	e := Entry{Key: "logo.png", Tags: []string{"img", "brand"}, Meta: &Meta{Owner: "ana", Size: 2048}}
	fmt.Println(Strings(e))
	fmt.Println(Strings([]any{"a", 1, []string{"b"}, nil, &e}))
	fmt.Println(len(Strings(42)), reflect.ValueOf(nil).Kind())
}
```

```
[logo.png img brand ana]
[a b logo.png img brand ana]
0 invalid
```

How each case navigates:

| Kind | How to get inside |
|---|---|
| `Slice`, `Array` | `v.Len()` and `v.Index(i)` |
| `Struct` | `v.NumField()` and `v.Field(i)`, or the `v.Fields()` iterator (next lesson) |
| `Map` | `v.MapKeys()` and `v.MapIndex(k)`, `v.MapRange()`, or the `v.Seq2()` iterator (Go 1.23) |
| `Pointer`, `Interface` | `v.IsNil()`, then `v.Elem()` |

Note what the walker does **not** do: it never calls `v.Interface()`. Reading through getters like `String()` and `Int()` works even on **unexported** fields, while `Interface()` on an unexported field panics. Reflection lets you look at private data, but not take it out as a normal value or change it.

## The nil cases

`reflect.ValueOf(nil)` has kind `Invalid`: there's no dynamic type at all. A walker's `switch` should quietly ignore it (here, by having no case for it). The other nil is a nil pointer, slice, map or interface *inside* a value: its kind is `Pointer` (or whichever), but `Elem()` on it gives an invalid `Value`, so check `IsNil()` first.

## Cycles

A pointer can point back to something that contains it, like a doubly linked list node, and a naive walker would recurse forever. Real encoders track the pointers they've visited (`v.Pointer()` gives the address as a `uintptr`) or limit the depth. Stash's data has no cycles, so the walkers in this chapter skip that.

## Map order

`v.MapKeys()`, `v.MapRange()` and `v.Seq2()` visit keys in the same random order as `range` does. If the output order matters, collect the keys and sort them, as `encoding/json` and `fmt` do for map keys. If you're only summing (as in this exercise), order doesn't matter.

## Your turn

Implement `weigh`, which estimates how much data a value carries by walking it:

- `bool` counts 1; every int, uint and float kind counts 8; a string counts its length;
- slices and arrays sum their elements; maps sum their keys and values (try `v.Seq2()`); structs sum all their fields, exported or not;
- pointers and interfaces count 0 if nil, otherwise whatever they point to or hold;
- everything else counts 0.

Multiple kinds can share a `case`, just like types in a type switch.
