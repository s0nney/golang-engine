---
title: Human Sizes
difficulty: easy
after: the-type-system
hints:
  - '`Size` is a **defined type** whose underlying type is `int64`, so `s / KiB` and `s % KiB` work and give you a `Size` back. Walk the units from biggest to smallest and use the first one that is `<= s`.'
  - 'For the one decimal, use integer maths so nothing rounds up: `whole := s / unit` and `tenth := s % unit * 10 / unit`. Print `whole` alone when `tenth == 0`. Format `int64(whole)` with `%d`: printing a `Size` with `%v` inside `Size.String` would call `String` again, forever.'
  - '`type Quota Size` does **not** inherit `Size`''s methods: a type definition only copies the underlying type. Convert and reuse: `"quota " + Size(q).String()`.'
exercise:
  starter: |
    package main

    import "fmt"

    // Size is a number of bytes in Stash's storage reports.
    type Size int64

    const (
    	B   Size = 1
    	KiB      = 1024 * B
    	MiB      = 1024 * KiB
    	GiB      = 1024 * MiB
    	TiB      = 1024 * GiB
    )

    // String formats s with the biggest unit (B, KiB, MiB, GiB, TiB) that fits,
    // showing at most one decimal, truncated, never rounded:
    // 512 -> "512 B", 1536 -> "1.5 KiB", 2*MiB -> "2 MiB".
    func (s Size) String() string {
    	// 1. Pick the biggest unit that is <= s (bytes when s < KiB).
    	// 2. Split s into whole units and tenths using / and %.
    	return ""
    }

    // Quota is a storage limit. It is a new defined type based on Size.
    type Quota Size

    // String formats a quota as "quota " followed by the Size formatting,
    // e.g. Quota(3*GiB) -> "quota 3 GiB".
    func (q Quota) String() string {
    	// Quota doesn't have Size's methods. How do you reuse them?
    	return ""
    }

    func main() {
    	for _, s := range []Size{512, 1536, 2 * MiB, 1023*KiB + 1000} {
    		fmt.Printf("Size(%d) = %q\n", int64(s), s.String())
    	}
    	// want: "512 B", "1.5 KiB", "2 MiB", "1023.9 KiB"
    	fmt.Printf("Quota(3 * GiB) = %q\n", Quota(3*GiB).String()) // want "quota 3 GiB"
    }
  solution: |
    package main

    import "fmt"

    // Size is a number of bytes in Stash's storage reports.
    type Size int64

    const (
    	B   Size = 1
    	KiB      = 1024 * B
    	MiB      = 1024 * KiB
    	GiB      = 1024 * MiB
    	TiB      = 1024 * GiB
    )

    var units = []struct {
    	size Size
    	name string
    }{{TiB, "TiB"}, {GiB, "GiB"}, {MiB, "MiB"}, {KiB, "KiB"}, {B, "B"}}

    func (s Size) String() string {
    	for _, u := range units {
    		if s >= u.size || u.size == B {
    			whole := s / u.size
    			tenth := s % u.size * 10 / u.size
    			if tenth == 0 {
    				return fmt.Sprintf("%d %s", int64(whole), u.name)
    			}
    			return fmt.Sprintf("%d.%d %s", int64(whole), int64(tenth), u.name)
    		}
    	}
    	return ""
    }

    // Quota is a storage limit. It is a new defined type based on Size.
    type Quota Size

    func (q Quota) String() string {
    	return "quota " + Size(q).String()
    }

    func main() {
    	fmt.Println(Size(512), Size(1536), 2*MiB, 1023*KiB+1000)
    	fmt.Println(Quota(3 * GiB))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"testing"
    )

    func TestSizeString(t *testing.T) {
    	tests := []struct {
    		s    Size
    		want string
    	}{
    		{0, "0 B"},
    		{1, "1 B"},
    		{1023, "1023 B"},
    		{1024, "1 KiB"},
    		{1536, "1.5 KiB"},
    		{1024 + 102, "1 KiB"},
    		{1024 + 103, "1.1 KiB"},
    		{1023*KiB + 1000, "1023.9 KiB"},
    		{MiB - 1, "1023.9 KiB"},
    		{MiB, "1 MiB"},
    		{2*MiB + MiB/2, "2.5 MiB"},
    		{GiB, "1 GiB"},
    		{7*GiB + 9*GiB/10 + MiB, "7.9 GiB"},
    		{TiB, "1 TiB"},
    		{3000 * TiB, "3000 TiB"},
    	}
    	for _, tt := range tests {
    		if got := tt.s.String(); got != tt.want {
    			t.Errorf("Size(%d).String() = %q, want %q", int64(tt.s), got, tt.want)
    		}
    	}
    }

    func TestSizeWithFmt(t *testing.T) {
    	if got := fmt.Sprint(Size(1536)); got != "1.5 KiB" {
    		t.Errorf("fmt.Sprint(Size(1536)) = %q, want %q: fmt should find your String method", got, "1.5 KiB")
    	}
    }

    func TestQuotaString(t *testing.T) {
    	tests := []struct {
    		q    Quota
    		want string
    	}{
    		{Quota(3 * GiB), "quota 3 GiB"},
    		{Quota(1536), "quota 1.5 KiB"},
    		{0, "quota 0 B"},
    	}
    	for _, tt := range tests {
    		if got := fmt.Sprint(tt.q); got != tt.want {
    			t.Errorf("fmt.Sprint(Quota(%d)) = %q, want %q", int64(tt.q), got, tt.want)
    		}
    	}
    }
---

Stash's storage report prints sizes like `1.5 KiB` instead of `1536`. Sizes are
a defined type, `Size`, with typed constants for each unit.

Complete the two `String` methods:

- `Size.String` formats a size with the **biggest unit that fits**, one of `B`,
  `KiB`, `MiB`, `GiB` or `TiB` (each 1024 times the last). Sizes under 1 KiB
  use bytes. Show at most **one decimal**, truncated rather than rounded, and
  leave it off when it's `0`.
- `Quota.String` formats a `Quota` as `"quota "` followed by what `Size.String`
  would print for the same number of bytes.

## Examples

```go
fmt.Println(Size(512))        // 512 B
fmt.Println(Size(1536))       // 1.5 KiB
fmt.Println(MiB - 1)          // 1023.9 KiB (truncated, not "1024 KiB")
fmt.Println(2 * MiB)          // 2 MiB
fmt.Println(Quota(3 * GiB))   // quota 3 GiB
```

## Constraints

- Sizes are never negative. Anything of 1 TiB or more uses `TiB`, so
  `3000*TiB` is `3000 TiB`.
- Stick to integer arithmetic on `Size`; there's no need for `float64`.
- `Quota` is declared as `type Quota Size`. Think about which methods that
  declaration gives it.
