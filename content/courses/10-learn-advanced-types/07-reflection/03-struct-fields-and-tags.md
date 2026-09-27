---
title: Struct Fields and Tags
quiz:
  - question: |
      For the field `` Size int64 `stash:"size,omitempty" json:"sz"` ``, what does `f.Tag.Get("stash")` return?
    options:
      - text: '`"size"`'
      - text: '`"size,omitempty"`'
        correct: true
      - text: '`stash:"size,omitempty"`'
      - text: '`"sz"`'
    explanation: |
      `Get` returns the whole quoted value for that key. Splitting off options like
      `omitempty` is up to your code; `strings.Cut(tag, ",")` is the usual way.
  - question: What's the difference between `f.Tag.Get("stash")` and `f.Tag.Lookup("stash")`?
    options:
      - text: '`Lookup` is faster'
      - text: '`Lookup` also reports whether the key was present, so you can tell `stash:""` from no `stash` tag at all'
        correct: true
      - text: '`Get` panics when the key is missing'
      - text: '`Lookup` searches embedded structs too'
    explanation: |
      `Get` returns `""` in both cases. When "present but empty" means something different
      from "absent", use `Lookup`.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"reflect"
    	"strings"
    )

    // Columns returns the column names Stash uses to store values of type T,
    // which must be a struct type (otherwise Columns returns nil).
    //
    // Consider every visible field (fields promoted from embedded structs
    // included, the embedded fields themselves excluded):
    //   - skip unexported fields
    //   - skip fields tagged stash:"-"
    //   - the name is the tag value before any comma: stash:"size,omitempty" -> "size"
    //   - with no tag, or an empty name, use the field name in lower case
    func Columns[T any]() []string {
    	t := reflect.TypeFor[T]()
    	// ?
    	_ = t
    	_ = strings.ToLower
    	return nil
    }

    type Audit struct {
    	CreatedBy string `stash:"created_by"`
    }

    type Entry struct {
    	Audit
    	Key   string `stash:"key"`
    	Size  int64  `stash:"size,omitempty"`
    	Notes string
    	Temp  []byte `stash:"-"`
    	owner string
    }

    func main() {
    	fmt.Println(Columns[Entry]())
    	fmt.Println(Columns[int]() == nil)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"reflect"
    	"strings"
    )

    // Columns returns the column names Stash uses to store values of type T,
    // which must be a struct type (otherwise Columns returns nil).
    //
    // Consider every visible field (fields promoted from embedded structs
    // included, the embedded fields themselves excluded):
    //   - skip unexported fields
    //   - skip fields tagged stash:"-"
    //   - the name is the tag value before any comma: stash:"size,omitempty" -> "size"
    //   - with no tag, or an empty name, use the field name in lower case
    func Columns[T any]() []string {
    	t := reflect.TypeFor[T]()
    	if t.Kind() != reflect.Struct {
    		return nil
    	}
    	var cols []string
    	for _, f := range reflect.VisibleFields(t) {
    		if f.Anonymous || !f.IsExported() {
    			continue
    		}
    		tag := f.Tag.Get("stash")
    		if tag == "-" {
    			continue
    		}
    		name, _, _ := strings.Cut(tag, ",")
    		if name == "" {
    			name = strings.ToLower(f.Name)
    		}
    		cols = append(cols, name)
    	}
    	return cols
    }

    type Audit struct {
    	CreatedBy string `stash:"created_by"`
    }

    type Entry struct {
    	Audit
    	Key   string `stash:"key"`
    	Size  int64  `stash:"size,omitempty"`
    	Notes string
    	Temp  []byte `stash:"-"`
    	owner string
    }

    func main() {
    	fmt.Println(Columns[Entry]())
    	fmt.Println(Columns[int]() == nil)
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    type User struct {
    	ID       int
    	Email    string `stash:"email_address"`
    	Password string `stash:"-"`
    	Nick     string `stash:",omitempty"`
    	internal bool
    }

    type Wrapped struct {
    	User
    	Role string `stash:"role"`
    }

    func TestColumnsEntry(t *testing.T) {
    	want := []string{"created_by", "key", "size", "notes"}
    	if got := Columns[Entry](); !slices.Equal(got, want) {
    		t.Errorf("Columns[Entry]() = %q, want %q", got, want)
    	}
    }

    func TestColumnsRules(t *testing.T) {
    	want := []string{"id", "email_address", "nick"}
    	if got := Columns[User](); !slices.Equal(got, want) {
    		t.Errorf("Columns[User]() = %q, want %q", got, want)
    	}
    	want = []string{"id", "email_address", "nick", "role"}
    	if got := Columns[Wrapped](); !slices.Equal(got, want) {
    		t.Errorf("Columns[Wrapped]() = %q, want %q (promoted fields first, embedded User itself skipped)", got, want)
    	}
    }

    func TestColumnsNonStruct(t *testing.T) {
    	if got := Columns[int](); got != nil {
    		t.Errorf("Columns[int]() = %q, want nil", got)
    	}
    	if got := Columns[*Entry](); got != nil {
    		t.Errorf("Columns[*Entry]() = %q, want nil (only struct types)", got)
    	}
    }
---

Struct tags are the little backquoted strings after a field's type. The compiler ignores them; they exist purely for reflection to read. They're how `encoding/json`, database libraries and validators learn what you meant.

## Reading fields

`reflect.Type` has `NumField()` and `Field(i)`, and since Go 1.26 a `Fields()` iterator that yields each `reflect.StructField` (there's a matching `Value.Fields()` yielding field-value pairs):

```go
package main

import (
	"fmt"
	"reflect"
)

type Audit struct {
	CreatedBy string `stash:"created_by"`
}

type Entry struct {
	Audit
	Key   string `stash:"key" json:"k"`
	Size  int64  `stash:"size,omitempty"`
	Notes string `stash:""`
	Temp  []byte `stash:"-"`
	owner string
}

func main() {
	t := reflect.TypeFor[Entry]()
	for f := range t.Fields() {
		tag, ok := f.Tag.Lookup("stash")
		fmt.Printf("%-6s %-8v exported=%-5v anon=%-5v stash=%q (%v) json=%q\n",
			f.Name, f.Type, f.IsExported(), f.Anonymous, tag, ok, f.Tag.Get("json"))
	}
	fmt.Println("---")
	for _, f := range reflect.VisibleFields(t) {
		fmt.Println(f.Name, f.Index)
	}
}
```

```
Audit  main.Audit exported=true  anon=true  stash="" (false) json=""
Key    string   exported=true  anon=false stash="key" (true) json="k"
Size   int64    exported=true  anon=false stash="size,omitempty" (true) json=""
Notes  string   exported=true  anon=false stash="" (true) json=""
Temp   []uint8  exported=true  anon=false stash="-" (true) json=""
owner  string   exported=false anon=false stash="" (false) json=""
---
Audit [0]
CreatedBy [0 0]
Key [1]
Size [2]
Notes [3]
Temp [4]
owner [5]
```

A `reflect.StructField` tells you:

- **`Name`** and **`Type`**, and **`IsExported()`**. Unexported fields are listed too; you just can't read them through `Interface()` or set them.
- **`Anonymous`**: true for an embedded field like `Audit`.
- **`Index`**: the path of field indexes to reach it, used by `Value.FieldByIndex`.
- **`Tag`**: a `reflect.StructTag`.

## Tag syntax

By convention, a tag is a space-separated list of `key:"value"` pairs:

```go
Key string `stash:"key" json:"k"`
```

- **`Tag.Get("stash")`** returns the value for a key, or `""` if it's missing.
- **`Tag.Lookup("stash")`** also returns whether the key was present, so `stash:""` and no tag can be told apart.

Values often carry options after a comma (`"size,omitempty"`). `Get` returns the whole string; split it yourself with `strings.Cut`. And `go vet` checks that tags are well-formed, so a typo like `` `stash:key` `` gets flagged.

## Embedded fields

`Type.Fields()` shows the struct exactly as declared: `Audit` is one field, and `CreatedBy` is inside it. What most encoders want is the **promoted** view, where `CreatedBy` behaves like a field of `Entry`. That's what **`reflect.VisibleFields(t)`** returns: every field reachable without ambiguity, including promoted ones (with longer `Index` paths, like `[0 0]`), in declaration order, and the embedded field itself too (with `Anonymous` set), so you can decide whether to skip it.

## Generics meet reflection

A nice combination: take the type as a **type parameter** and ask reflection about it with `reflect.TypeFor[T]()`. The caller gets a clean, value-free API (`Columns[Entry]()`), and there's no risk of passing a nil interface by accident.

## Your turn

Write `Columns[T]()`, which returns the column names Stash uses for a struct type:

- if `T` isn't a struct type, return `nil`;
- walk `reflect.VisibleFields`, skipping embedded fields themselves (their promoted fields still count) and unexported fields;
- skip fields tagged `stash:"-"`;
- the name is the part of the `stash` tag before any comma; if that's empty or there's no tag, use the field name in lower case.
