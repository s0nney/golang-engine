---
title: Type Definitions vs Aliases
quiz:
  - question: |
      What does this print?

      ```go
      type Tags []string     // definition
      type Labels = []string // alias

      func main() {
      	fmt.Printf("%T %T\n", Tags{}, Labels{})
      }
      ```
    options:
      - text: '`main.Tags main.Labels`'
      - text: '`main.Tags []string`'
        correct: true
      - text: '`[]string []string`'
      - text: '`[]string main.Labels`'
    explanation: |
      A definition creates a new type named `Tags`. An alias creates only a new *name*
      for an existing type, so `Labels{}` really is a `[]string`, and that's what `%T`
      reports.
  - question: Why can't you declare `func (s Set[T]) Has(v T) bool` when `type Set[T comparable] = map[T]struct{}`?
    options:
      - text: Generic types can't have methods
      - text: '`Set` is an alias for the type literal `map[T]struct{}`, and methods can only be declared on a defined type in the same package'
        correct: true
      - text: Method receivers can't mention type parameters
      - text: Maps can't have methods
    explanation: |
      An alias adds no new type to hang methods on; `Set[string]` *is*
      `map[string]struct{}`. If you want methods, make `Set` a defined type
      (`type Set[T comparable] map[T]struct{}`) or a struct.
exercise:
  starter: |
    package main

    import "fmt"

    // Index maps each key to the positions where it appears.
    // TODO: make this a generic alias for map[K][]int instead of a new type,
    // so an Index[string] can go anywhere a map[string][]int is expected.
    type Index[K comparable] map[K][]int

    // BuildIndex records, for every key, the positions (in increasing order)
    // at which it appears in keys.
    func BuildIndex[K comparable](keys []K) Index[K] {
    	// ?
    	return nil
    }

    func main() {
    	idx := BuildIndex([]string{"go", "rust", "go", "zig", "go"})
    	fmt.Printf("%T\n", idx)
    	fmt.Println(idx["go"], idx["rust"], idx["zig"])
    }
  solution: |
    package main

    import "fmt"

    // Index maps each key to the positions where it appears.
    type Index[K comparable] = map[K][]int

    // BuildIndex records, for every key, the positions (in increasing order)
    // at which it appears in keys.
    func BuildIndex[K comparable](keys []K) Index[K] {
    	idx := make(Index[K])
    	for i, k := range keys {
    		idx[k] = append(idx[k], i)
    	}
    	return idx
    }

    func main() {
    	idx := BuildIndex([]string{"go", "rust", "go", "zig", "go"})
    	fmt.Printf("%T\n", idx)
    	fmt.Println(idx["go"], idx["rust"], idx["zig"])
    }
  tests: |
    package main

    import (
    	"reflect"
    	"slices"
    	"testing"
    )

    func TestIndexIsAnAlias(t *testing.T) {
    	got := reflect.TypeFor[Index[string]]()
    	want := reflect.TypeFor[map[string][]int]()
    	if got != want {
    		t.Fatalf("Index[string] is the type %v, want it to be exactly %v (use a generic alias: type Index[K comparable] = ...)", got, want)
    	}
    }

    func TestBuildIndex(t *testing.T) {
    	idx := BuildIndex([]string{"go", "rust", "go", "zig", "go"})
    	for key, want := range map[string][]int{"go": {0, 2, 4}, "rust": {1}, "zig": {3}} {
    		if got := idx[key]; !slices.Equal(got, want) {
    			t.Errorf("BuildIndex(...)[%q] = %v, want %v", key, got, want)
    		}
    	}
    	if len(idx) != 3 {
    		t.Errorf("BuildIndex(...) has %d keys, want 3", len(idx))
    	}
    }

    func TestBuildIndexInts(t *testing.T) {
    	idx := BuildIndex([]int{7, 7, 7})
    	if got := idx[7]; !slices.Equal(got, []int{0, 1, 2}) {
    		t.Errorf("BuildIndex([7 7 7])[7] = %v, want [0 1 2]", got)
    	}
    	if empty := BuildIndex([]int{}); empty == nil || len(empty) != 0 {
    		t.Errorf("BuildIndex([]) = %#v, want an empty, non-nil map", empty)
    	}
    }
---

Go has two ways to introduce a type name, and they differ by a single `=`:

```go
type Tags []string     // type DEFINITION: a brand-new named type
type Labels = []string // type ALIAS: another name for []string
```

## Definitions create types

Everything so far in this chapter has been about definitions. `Tags` is a new named type with underlying type `[]string`: it's different from every other type, starts with an empty method set, and you can declare methods on it.

## Aliases create names

An alias doesn't create anything. `Labels` and `[]string` are **the same type**, spelled two ways. It's like a variable having two names. Anything true of one is true of the other, including its method set (and you can't add methods to an alias of a type from somewhere else). You already use two aliases all the time:

```go
type byte = uint8
type any = interface{}
```

That's why a `[]byte` and a `[]uint8` are interchangeable, and why `any` and `interface{}` are one type.

Aliases were added (in Go 1.9) for one main job: **moving a type between packages** without breaking every user at once. The old package keeps `type Client = newpkg.Client` while code migrates. They're also handy for shortening a long type literal you use everywhere.

## Generic aliases (Go 1.24)

Since Go 1.24, aliases can have type parameters:

```go
package main

import (
	"fmt"
	"reflect"
)

type Set[T comparable] = map[T]struct{}

type Pair[K comparable, V any] struct {
	Key K
	Val V
}

type StringPair[V any] = Pair[string, V]

func main() {
	s := Set[string]{"go": {}}
	var m map[string]struct{} = s // same type, no conversion
	fmt.Printf("%T %T\n", s, StringPair[int]{"a", 1})
	fmt.Println(reflect.TypeOf(s) == reflect.TypeFor[map[string]struct{}](), len(m))
}
```

```
map[string]struct {} main.Pair[string,int]
true 1
```

`Set[string]` is just `map[string]struct{}`, so `%T` prints the map type, and `reflect` sees exactly the same type. `StringPair[V]` fixes the first type argument of `Pair` and leaves the second open.

The big limitation: **no methods**. Trying to write `func (s Set[T]) Has(v T) bool` fails with `cannot define new methods on generic alias type Set[T comparable]`. If `Set` needs methods (and in chapter 5, Stash's will) it has to be a definition.

## Which one?

- Want **new behaviour, safety, or methods**? Use a **definition**. That's almost always what you want.
- Want a **shorter or transitional name** for a type that must stay *interchangeable* with the original? Use an **alias**.

A good sign you want an alias: you keep writing conversions like `map[string][]int(idx)` just to call someone else's function, and the new name doesn't need any methods.

## Your turn

Stash's `Index[K]` maps each key to the positions where it appears. Other code wants plain `map[K][]int` values, so:

1. Change `Index` into a **generic alias** for `map[K][]int`.
2. Complete `BuildIndex` so it returns a non-nil map where each key's slice lists the positions it occurred at, in increasing order.

## Further reading

- [Go 1.24 release notes: generic type aliases](https://go.dev/doc/go1.24#language)
