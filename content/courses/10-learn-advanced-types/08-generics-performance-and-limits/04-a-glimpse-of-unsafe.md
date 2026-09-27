---
title: A Glimpse of unsafe
quiz:
  - question: |
      On a 64-bit platform, what's `unsafe.Sizeof` of this struct?

      ```go
      type Loose struct {
      	Active bool
      	Size   int64
      	Pinned bool
      }
      ```
    options:
      - text: '`10`'
      - text: '`16`'
      - text: '`24`'
        correct: true
      - text: '`17`'
    explanation: |
      `Size` must start at a multiple of 8, so 7 bytes of padding follow `Active`, and
      the struct's total size is rounded up to a multiple of 8 after `Pinned`:
      8 + 8 + 8 = 24. Putting the two `bool`s together after `Size` gets it down to 16.
  - question: Which statement about `unsafe.Sizeof` is true?
    options:
      - text: It measures everything a value references, like a slice's backing array
      - text: It reports only the value's own fixed size (a slice header is 24 bytes on 64-bit, however long the slice), and it's evaluated at compile time
        correct: true
      - text: It allocates memory to measure the value
      - text: It's unsafe to call and can crash the program
    explanation: |
      `Sizeof`, `Alignof` and `Offsetof` are compile-time constants describing layout.
      They're harmless to call. The *unsafe* parts of the package are `unsafe.Pointer`
      and friends, which let you bypass the type system.
---

The `unsafe` package is Go's escape hatch from the type system. This course doesn't teach you to use it, and most Go programmers rarely should. But three of its functions are perfectly safe, and they let you *see* the memory layouts this course has been describing.

## Sizes and layout

```go
package main

import (
	"fmt"
	"unsafe"
)

type Loose struct {
	Active bool
	Size   int64
	Pinned bool
}

type Tight struct {
	Size   int64
	Active bool
	Pinned bool
}

func main() {
	var (
		n int
		s string
		b []byte
		a any
		m map[string]int
		p *Loose
	)
	fmt.Println(unsafe.Sizeof(n), unsafe.Sizeof(s), unsafe.Sizeof(b), unsafe.Sizeof(a), unsafe.Sizeof(m), unsafe.Sizeof(p))

	fmt.Println(unsafe.Sizeof(Loose{}), unsafe.Sizeof(Tight{}))
	var l Loose
	fmt.Println(unsafe.Offsetof(l.Size), unsafe.Offsetof(l.Pinned), unsafe.Alignof(l.Size))
}
```

On a 64-bit platform this prints:

```
8 16 24 16 8 8
24 16
8 16 8
```

The first line confirms the pictures from earlier chapters:

| Type | Size | Why |
|---|---|---|
| `int` | 8 | one machine word |
| `string` | 16 | pointer + length |
| `[]byte` | 24 | pointer + length + capacity |
| `any` | 16 | type word + data word (chapter 6) |
| `map[string]int` | 8 | a map value is a pointer to the runtime's map structure |
| `*Loose` | 8 | one pointer |

These are the sizes of the values **themselves**. `unsafe.Sizeof` never follows pointers: a million-element slice is still 24 bytes of header.

## Padding

The second and third lines show **alignment** at work. An `int64` must start at an address that's a multiple of 8 (`Alignof` reports 8), so in `Loose` the compiler inserts 7 bytes of padding after `Active` (`Size` is at offset 8), and more after `Pinned` to round the struct up to 24 bytes. `Tight` has the same fields ordered largest first and needs only 16.

Go doesn't reorder fields for you. For a struct you keep millions of in a slice, like Stash's cache entries, ordering fields from largest to smallest alignment can save real memory. For everything else, order fields for readability. The `fieldalignment` analyzer (in `golang.org/x/tools`) can point out structs worth reordering.

`Sizeof`, `Alignof` and `Offsetof` are evaluated at **compile time** and are constants, so calling them costs nothing and can't go wrong.

## The actually unsafe part

The rest of the package lets you step outside Go's type safety:

- **`unsafe.Pointer`** can be converted to and from any pointer type, and to `uintptr`, so you can reinterpret memory as a different type.
- **`unsafe.Slice`, `unsafe.String`, `unsafe.SliceData` and `unsafe.StringData`** build slices and strings from raw pointers, or get at their backing memory without copying.
- **`unsafe.Add`** does pointer arithmetic.

These are how parts of the runtime, `reflect`, `sync/atomic` and a few high-performance libraries are built. They come with a strict list of valid patterns in the `unsafe.Pointer` documentation; anything else may break silently with a future compiler or garbage collector change, and `go vet` catches only some misuses. Programs that import `unsafe` also give up the guarantees of the Go 1 compatibility promise for that code.

The rule of thumb: if you're not writing a runtime, a serializer squeezing out the last nanosecond, or an interface to C, you don't need `unsafe.Pointer`. Everything in this course, including reflection, gets by without it. If you ever do need it, read `go doc unsafe.Pointer` in full first.
