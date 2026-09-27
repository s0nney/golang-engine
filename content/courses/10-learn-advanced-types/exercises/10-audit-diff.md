---
title: Audit Diff
difficulty: medium
after: reflection
hints:
  - 'Start with `a, b := reflect.ValueOf(&before).Elem(), reflect.ValueOf(&after).Elem()` (this works even when `T` is an interface or pointer type). If `a.Kind()` is `reflect.Pointer`, swap each side for `reflect.Zero(elemType)` when it''s nil, or for `.Elem()` otherwise. Then check that the kind is `reflect.Struct`.'
  - 'Write a recursive helper `diff(prefix string, a, b reflect.Value, out *[]string)`. For each field (`a.Type().Field(i)` or the `Fields()` iterators): skip it if `!f.IsExported()` or its tag is `"-"`; work out its name from `f.Tag.Get("diff")` or `f.Name`; recurse when `f.Type.Kind() == reflect.Struct`, with `prefix + name + "."`.'
  - 'For every other field, compare `a.Field(i).Interface()` and `b.Field(i).Interface()` with `reflect.DeepEqual`, which handles slices, maps and pointers (pointers are equal when they point at deeply equal values).'
exercise:
  starter: |
    package main

    import "fmt"

    func Changes[T any](before, after T) ([]string, error) {
    	return nil, nil
    }

    type Limits struct {
    	MaxKeys int
    	Evict   bool
    }

    type Bucket struct {
    	Name    string
    	Tags    []string
    	Limits  Limits `diff:"limits"`
    	Secret  string `diff:"-"`
    	version int
    }

    func main() {
    	before := Bucket{Name: "sessions", Tags: []string{"hot"}, Limits: Limits{100, false}, Secret: "a", version: 1}
    	after := Bucket{Name: "sessions", Tags: []string{"hot", "eu"}, Limits: Limits{500, false}, Secret: "b", version: 2}
    	fmt.Println(Changes(before, after)) // want [Tags limits.MaxKeys] <nil>
    }
  solution: |
    package main

    import (
    	"fmt"
    	"reflect"
    )

    // Changes lists the exported fields that differ between before and after.
    func Changes[T any](before, after T) ([]string, error) {
    	a, b := reflect.ValueOf(&before).Elem(), reflect.ValueOf(&after).Elem()
    	if a.Kind() == reflect.Pointer {
    		a, b = derefOrZero(a), derefOrZero(b)
    	}
    	if a.Kind() != reflect.Struct {
    		return nil, fmt.Errorf("Changes: %v is not a struct or a pointer to one", reflect.TypeFor[T]())
    	}
    	var out []string
    	diff("", a, b, &out)
    	return out, nil
    }

    func derefOrZero(v reflect.Value) reflect.Value {
    	if v.IsNil() {
    		return reflect.Zero(v.Type().Elem())
    	}
    	return v.Elem()
    }

    func diff(prefix string, a, b reflect.Value, out *[]string) {
    	for f := range a.Type().Fields() {
    		if !f.IsExported() {
    			continue
    		}
    		tag := f.Tag.Get("diff")
    		if tag == "-" {
    			continue
    		}
    		name := f.Name
    		if tag != "" {
    			name = tag
    		}
    		fa, fb := a.FieldByIndex(f.Index), b.FieldByIndex(f.Index)
    		if f.Type.Kind() == reflect.Struct {
    			diff(prefix+name+".", fa, fb, out)
    			continue
    		}
    		if !reflect.DeepEqual(fa.Interface(), fb.Interface()) {
    			*out = append(*out, prefix+name)
    		}
    	}
    }

    type Limits struct {
    	MaxKeys int
    	Evict   bool
    }

    type Bucket struct {
    	Name    string
    	Tags    []string
    	Limits  Limits `diff:"limits"`
    	Secret  string `diff:"-"`
    	version int
    }

    func main() {
    	before := Bucket{Name: "sessions", Tags: []string{"hot"}, Limits: Limits{100, false}, Secret: "a", version: 1}
    	after := Bucket{Name: "sessions", Tags: []string{"hot", "eu"}, Limits: Limits{500, false}, Secret: "b", version: 2}
    	fmt.Println(Changes(before, after))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    type Address struct {
    	City string
    	Zip  string `diff:"postcode"`
    }

    type Meta struct {
    	Owner string
    	note  string
    }

    type Level int

    type Account struct {
    	Meta                        // embedded: its fields appear as Meta.Owner
    	ID      int                 `diff:"id"`
    	Level   Level               `diff:"level"`
    	Home    Address             `diff:"home"`
    	Work    *Address            `diff:"work"`
    	Roles   map[string]bool     `diff:"roles"`
    	Scores  []float64           `diff:"scores"`
    	Handler func()              `diff:"-"`
    	Extra   any                 `diff:"extra"`
    	cache   map[string][]string // unexported: ignored
    }

    func check[T any](t *testing.T, name string, before, after T, want []string) {
    	t.Helper()
    	got, err := Changes(before, after)
    	if err != nil {
    		t.Errorf("%s: Changes returned error %v, want %q", name, err, want)
    		return
    	}
    	if !slices.Equal(got, want) {
    		t.Errorf("%s: Changes = %q, want %q", name, got, want)
    	}
    }

    func TestBucketExample(t *testing.T) {
    	before := Bucket{Name: "sessions", Tags: []string{"hot"}, Limits: Limits{100, false}, Secret: "a", version: 1}
    	after := Bucket{Name: "sessions", Tags: []string{"hot", "eu"}, Limits: Limits{500, false}, Secret: "b", version: 2}
    	check(t, "Bucket example", before, after, []string{"Tags", "limits.MaxKeys"})
    	check(t, "identical Buckets", before, before, nil)
    	after2 := before
    	after2.Tags = []string{"hot"} // a different slice with equal contents is not a change
    	check(t, "equal slices", before, after2, nil)
    }

    func TestAccountFields(t *testing.T) {
    	base := Account{
    		Meta:   Meta{Owner: "ada", note: "x"},
    		ID:     1,
    		Level:  2,
    		Home:   Address{"Oslo", "0150"},
    		Work:   &Address{"Bergen", "5003"},
    		Roles:  map[string]bool{"admin": true},
    		Scores: []float64{1.5},
    		Extra:  "hi",
    		cache:  map[string][]string{"a": nil},
    	}
    	clone := func() Account {
    		a := base
    		w := *base.Work
    		a.Work = &w
    		a.Roles = map[string]bool{"admin": true}
    		a.Scores = []float64{1.5}
    		return a
    	}

    	check(t, "deep copy", base, clone(), nil)

    	c := clone()
    	c.note, c.cache, c.Handler = "changed", nil, func() {}
    	check(t, "only unexported and skipped fields", base, c, nil)

    	c = clone()
    	c.Owner = "grace"
    	c.Level = 3
    	check(t, "embedded and named-type fields", base, c, []string{"Meta.Owner", "level"})

    	c = clone()
    	c.Home.Zip = "0151"
    	c.Work.City = "Oslo"
    	check(t, "nested struct and pointer to struct", base, c, []string{"home.postcode", "work"})

    	c = clone()
    	c.Work = nil
    	c.Roles["reader"] = true
    	c.Scores = nil
    	c.Extra = 42
    	check(t, "nil pointer, map, slice and interface", base, c, []string{"work", "roles", "scores", "extra"})
    }

    func TestPointerType(t *testing.T) {
    	a := &Address{"Oslo", "0150"}
    	b := &Address{"Oslo", "9999"}
    	check(t, "*Address", a, b, []string{"postcode"})
    	check(t, "*Address, same pointer", a, a, nil)
    	check(t, "*Address, nil before", nil, b, []string{"City", "postcode"})
    	check[*Address](t, "*Address, both nil", nil, nil, nil)
    }

    func TestNotAStruct(t *testing.T) {
    	if _, err := Changes(1, 2); err == nil {
    		t.Errorf("Changes(1, 2) returned no error, want one: int is not a struct")
    	}
    	if _, err := Changes([]string{"a"}, nil); err == nil {
    		t.Errorf("Changes([]string...) returned no error, want one")
    	}
    	n := 3
    	if _, err := Changes(&n, &n); err == nil {
    		t.Errorf("Changes(*int...) returned no error, want one")
    	}
    }
---

Stash keeps an audit log. Whenever a record is updated, the log should say
**which fields** changed, like `limits.MaxKeys`, not dump both versions.

Write `Changes(before, after)`. It returns the names of the exported fields
whose values differ, in field order:

- `T` is a struct type or a **pointer** to one. A nil pointer counts as the
  zero struct. For any other `T`, return an error.
- **Unexported** fields are ignored, and so are fields tagged `diff:"-"`.
- A field's name is its `diff` tag if it has one, otherwise its Go name.
- A field whose type is a **struct** (including an embedded one) is compared
  field by field, and its changes are reported as `outer.inner` using the same
  naming rules at every level.
- Every other field (numbers, strings, pointers, slices, maps, interfaces...)
  is reported by its own name if `reflect.DeepEqual` says the two values differ.
  So two different slices with equal contents are *not* a change, and a pointer
  field changes when the values it points at differ.

## Example

```go
type Limits struct {
	MaxKeys int
	Evict   bool
}

type Bucket struct {
	Name    string
	Tags    []string
	Limits  Limits `diff:"limits"`
	Secret  string `diff:"-"`
	version int
}

before := Bucket{Name: "sessions", Tags: []string{"hot"}, Limits: Limits{100, false}}
after := Bucket{Name: "sessions", Tags: []string{"hot", "eu"}, Limits: Limits{500, false}}
Changes(before, after) // [Tags limits.MaxKeys], nil
```

## Constraints

- If nothing changed, return an empty (or nil) slice and a nil error.
- The hidden tests use their own structs with embedded structs, pointer, map,
  slice, `func` and `any` fields, named field types, and pointer-to-struct `T`.
- Reading an unexported field with `Interface()` panics, so skip unexported
  fields before touching their values.
