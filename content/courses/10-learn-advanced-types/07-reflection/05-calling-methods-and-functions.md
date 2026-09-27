---
title: Calling Methods and Functions
quiz:
  - question: |
      With `func (a *Admin) Count() int` and `func (a Admin) Version() string`, what is `reflect.TypeFor[Admin]().NumMethod()`?
    options:
      - text: '`2`'
      - text: '`1`'
        correct: true
      - text: '`0`'
      - text: It panics, because `Admin` isn't a pointer
    explanation: |
      Reflection follows the language's method set rules: the method set of `Admin` has
      only the value-receiver method `Version`. `reflect.TypeFor[*Admin]()` would report
      both.
  - question: What happens if you call a method through reflection with the wrong number of arguments?
    options:
      - text: The missing arguments are filled with zero values
      - text: The compiler reports an error
      - text: '`Call` panics at runtime'
        correct: true
      - text: '`Call` returns an error value'
    explanation: |
      Reflection has no compile-time checking: `Value.Call` panics if the argument count
      or types don't match. That's why reflective dispatchers check `Type().NumIn()` and
      `In(i)` before calling, and why ordinary interfaces are preferred when possible.
---

Reflection can also *call* things: methods looked up by name at runtime, and any function value. It's the machinery behind `text/template` calling your methods from `{{.Count}}`, and behind RPC frameworks that turn a method name in a request into a call.

## Methods by name

```go
package main

import (
	"fmt"
	"reflect"
	"strings"
)

// Admin holds Stash's admin commands. Every exported method is a command.
type Admin struct {
	entries map[string]int
}

func (a *Admin) Count() int { return len(a.entries) }

func (a *Admin) Put(key string, size int) string {
	a.entries[key] = size
	return "stored " + key
}

func (a Admin) Version() string { return "stash 0.7" }

func main() {
	a := &Admin{entries: map[string]int{}}

	// The method sets differ: *Admin has all three, Admin only Version.
	for m := range reflect.TypeOf(a).Methods() {
		fmt.Println(m.Name, m.Type)
	}
	fmt.Println(reflect.TypeFor[Admin]().NumMethod())

	// Call a method by name, with arguments.
	put := reflect.ValueOf(a).MethodByName("Put")
	out := put.Call([]reflect.Value{reflect.ValueOf("logo.png"), reflect.ValueOf(2048)})
	fmt.Println(out[0].String())

	count, _ := reflect.TypeAssert[int](reflect.ValueOf(a).MethodByName("Count").Call(nil)[0])
	fmt.Println(count)

	// Any function value can be called too.
	upper := reflect.ValueOf(strings.ToUpper)
	fmt.Println(upper.Call([]reflect.Value{reflect.ValueOf("done")})[0])

	fmt.Println(reflect.ValueOf(a).MethodByName("Delete").IsValid())
}
```

```
Count func(*main.Admin) int
Put func(*main.Admin, string, int) string
Version func(*main.Admin) string
1
stored logo.png
1
DONE
false
```

What's going on:

- **`Type.Methods()`** (Go 1.26) iterates the type's exported methods in lexicographic order, yielding `reflect.Method` values. Before 1.26, you'd loop with `NumMethod()` and `Method(i)`. `Value.Methods()` does the same for a value, yielding each method together with a callable **method value**. The `Type` of a method obtained from a *type* includes the receiver as its first argument: `func(*main.Admin) int`.
- **Method sets are the language's.** `*Admin` has three methods; `Admin` has only `Version`. Reflection can't call a pointer method on a non-addressable value any more than normal code can.
- **`Value.MethodByName("Put")`** returns a method value with the receiver already bound, so `Call` takes only the real arguments. For a name that doesn't exist, you get an invalid `Value` (check `IsValid()`), not a panic.
- **`Value.Call(args)`** takes and returns `[]reflect.Value`. Arguments must be assignable to the parameter types, or it panics. Variadic functions get their extra arguments as individual values (or use `CallSlice`).
- **Any function value** works: `reflect.ValueOf(strings.ToUpper).Call(...)`.

## Checking before calling

Everything the compiler normally checks becomes your job. A robust dispatcher looks at the method's type first:

```go
m := v.MethodByName(name)
if !m.IsValid() {
	return fmt.Errorf("unknown command %q", name)
}
t := m.Type()
if t.NumIn() != len(args) {
	return fmt.Errorf("%s takes %d arguments, got %d", name, t.NumIn(), len(args))
}
for i, a := range args {
	if !a.Type().AssignableTo(t.In(i)) {
		return fmt.Errorf("argument %d: have %s, want %s", i, a.Type(), t.In(i))
	}
}
```

Since Go 1.26, `t.Ins()` and `t.Outs()` iterate the parameter and result types too.

## A cost you don't see

Reflective method lookups by name have a hidden cost at **link** time. Normally the linker drops methods nothing calls. But once a program calls `MethodByName` with a non-constant name, or uses `Value.Method`/`Methods` in general, the linker can't know which methods might be needed, so it keeps all exported methods of every reachable type. Binaries get bigger. It's one more reason to reach for an interface, a `map[string]func(...)`, or a `switch` when the set of commands is known in advance, and keep reflective dispatch for plugins, templates and similar truly dynamic cases.
