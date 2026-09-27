---
title: Underlying Types
quiz:
  - question: |
      What is the underlying type of `Quota`?

      ```go
      type Bytes int64
      type Quota Bytes
      ```
    options:
      - text: '`Bytes`'
      - text: '`int64`'
        correct: true
      - text: '`Quota` itself'
      - text: It has no underlying type
    explanation: |
      A type definition takes the underlying type of whatever it's defined from. `Bytes`
      has underlying type `int64`, so `Quota` does too. The chain always bottoms out at a
      predeclared type or a type literal, never at another defined type.
  - question: |
      What does this print?

      ```go
      type Bytes int64

      func (b Bytes) String() string { return fmt.Sprintf("%dB", int64(b)) }

      type Quota Bytes

      func main() {
      	var q Quota = 512
      	fmt.Println(q, Bytes(q))
      }
      ```
    options:
      - text: '`512B 512B`'
      - text: '`512 512`'
      - text: '`512 512B`'
        correct: true
      - text: It doesn't compile, because `Quota` can't be converted to `Bytes`
    explanation: |
      `Quota` gets `Bytes`'s *underlying type* but none of its methods, so it prints as a
      plain number. Converting to `Bytes` (allowed, since both share underlying type
      `int64`) brings the `String` method back.
---

Every type in Go has an **underlying type**. It's the answer to "what is this, really, once you strip off the names?"

## The rules

- For a **predeclared** type (`int`, `string`...) or a **type literal** (`[]string`, `map[Key]int`, `struct{...}`), the underlying type is the type itself.
- For a **defined type** `type T X`, the underlying type is *X's* underlying type.

So the chain always ends at a predeclared type or a type literal:

```go
type Bytes int64      // underlying: int64
type Quota Bytes      // underlying: int64 (not Bytes!)
type Tags []string    // underlying: []string
type Labels Tags      // underlying: []string
type Row map[Key]Tags // underlying: map[Key]Tags
```

Look at that last one: the underlying type of `Row` is the literal `map[Key]Tags`. Only the *outermost* name is stripped. The `Key` and `Tags` inside stay exactly as they are.

## Operations come from the underlying type

What you can *do* with a value (index it, `len` it, add it, range over it, compare it) is decided by its underlying type. That's why this works:

```go
package main

import "fmt"

type Tags []string

type Bytes int64

func main() {
	t := Tags{"go", "types"}
	t = append(t, "generics") // Tags is a slice underneath
	fmt.Println(len(t), t[2])

	var total Bytes
	for _, n := range []Bytes{512, 1024} {
		total += n // + comes from int64
	}
	fmt.Println(total * 2)
}
```

```
3 generics
3072
```

`append`, `len`, indexing and `+` all see through the name. Hold on to this idea: in the next chapter, a constraint like `~[]E` means exactly "any type whose underlying type is `[]E`".

## Methods don't come along

A new type definition copies the underlying type but **not the methods**:

```go
type Bytes int64

func (b Bytes) String() string { return fmt.Sprintf("%dB", int64(b)) }

type Quota Bytes // Quota has NO String method
```

That's deliberate. It's how you take an existing type and give it a fresh, empty method set, perhaps to replace behaviour you don't want. The standard library does this too: `sort.StringSlice` and friends wrap `[]string` precisely to attach new methods.

There are two exceptions, and both follow from the rule rather than breaking it:

1. **Interface types.** An interface's methods *are* its type, so `type Stringer2 fmt.Stringer` has the same method set.
2. **Embedded fields.** If the underlying type is a struct with an embedded field, the struct literal (and so its promoted methods) carries over:

```go
package main

import (
	"fmt"
	"time"
)

type Meta struct{ time.Time }

type Entry Meta // underlying: struct{ time.Time }

func main() {
	e := Entry{time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)}
	fmt.Println(e.Year()) // promoted from the embedded time.Time
}
```

```
2026
```

`Entry` has no methods of its own, but its underlying struct still embeds `time.Time`, so `Year` is promoted exactly as it is for `Meta`.

## Underlying types and identity

Two different named types can share an underlying type (`Bytes` and `Quota` are both `int64` underneath) and still be different types. The underlying type isn't about identity. It's about what the compiler lets you **convert** between and what operations are legal, which is the next lesson.
