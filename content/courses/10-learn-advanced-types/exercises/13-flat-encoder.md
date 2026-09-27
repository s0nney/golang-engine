---
title: Flat Encoder
difficulty: hard
after: reflection
hints:
  - 'Split the work in two recursive helpers: `encodeStruct(prefix string, v reflect.Value)` walks the fields, and `encodeValue(path string, v reflect.Value)` writes one value by `v.Kind()`. Collect lines in a `strings.Builder` (or a `[]string`) and return the first error you hit. Join names with a helper: `join("", "id")` is `"id"`, `join("owner", "id")` is `"owner.id"`.'
  - 'For each field: skip `!f.IsExported()`; `name, opts, _ := strings.Cut(f.Tag.Get("stash"), ",")`; skip `name == "-"`; default the name to `f.Name`; skip `opts == "omitempty" && fv.IsZero()`. For an embedded struct without a tag name (`f.Anonymous`), recurse with the **same** prefix instead of adding a name.'
  - 'In `encodeValue`, switch on the kind. `Pointer` and `Interface`: write `nil` or recurse into `v.Elem()` with the same path. `Slice`/`Array`: `[]` when empty, else recurse with `path.0`, `path.1`... `Map`: check `v.Type().Key().Kind() == reflect.String`, write `{}` when empty, else sort the keys (`slices.SortFunc` with `cmp.Compare(a.String(), b.String())`) and recurse with `path.key`. `Struct`: `encodeStruct(path, v)`. Numbers via `strconv`, strings via `strconv.Quote`. Anything else is an error that names the path.'
exercise:
  starter: |
    package main

    import "fmt"

    func Encode(v any) (string, error) {
    	return "", nil
    }

    type Limits struct {
    	MaxKeys int  `stash:"max_keys"`
    	Evict   bool `stash:"evict,omitempty"`
    }

    type Owner struct {
    	Name string `stash:"name"`
    }

    type Bucket struct {
    	ID     int               `stash:"id"`
    	Tags   []string          `stash:"tags"`
    	Limits Limits            `stash:"limits"`
    	Owner  *Owner            `stash:"owner"`
    	Labels map[string]string `stash:"labels,omitempty"`
    	Secret string            `stash:"-"`
    	cache  []byte
    }

    func main() {
    	b := Bucket{
    		ID:     7,
    		Tags:   []string{"hot", "eu"},
    		Limits: Limits{MaxKeys: 500},
    		Owner:  &Owner{"ada"},
    		Secret: "hunter2",
    	}
    	out, err := Encode(b)
    	fmt.Printf("%s(err: %v)\n", out, err)
    	// want:
    	// id=7
    	// tags.0="hot"
    	// tags.1="eu"
    	// limits.max_keys=500
    	// owner.name="ada"
    	// (err: <nil>)
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"reflect"
    	"slices"
    	"strconv"
    	"strings"
    )

    // Encode flattens the struct v (or a non-nil pointer to one) into
    // "path=value" lines.
    func Encode(v any) (string, error) {
    	rv := reflect.ValueOf(v)
    	if rv.Kind() == reflect.Pointer && !rv.IsNil() {
    		rv = rv.Elem()
    	}
    	if rv.Kind() != reflect.Struct {
    		return "", fmt.Errorf("stash: Encode needs a struct or a non-nil pointer to one, got %T", v)
    	}
    	e := &encoder{}
    	if err := e.encodeStruct("", rv); err != nil {
    		return "", err
    	}
    	return e.b.String(), nil
    }

    type encoder struct {
    	b strings.Builder
    }

    func join(prefix, name string) string {
    	if prefix == "" {
    		return name
    	}
    	return prefix + "." + name
    }

    func (e *encoder) line(path, value string) {
    	e.b.WriteString(path)
    	e.b.WriteByte('=')
    	e.b.WriteString(value)
    	e.b.WriteByte('\n')
    }

    func (e *encoder) encodeStruct(prefix string, v reflect.Value) error {
    	for f, fv := range v.Fields() {
    		if !f.IsExported() {
    			continue
    		}
    		name, opts, _ := strings.Cut(f.Tag.Get("stash"), ",")
    		if name == "-" {
    			continue
    		}
    		if opts == "omitempty" && fv.IsZero() {
    			continue
    		}
    		if f.Anonymous && name == "" && f.Type.Kind() == reflect.Struct {
    			if err := e.encodeStruct(prefix, fv); err != nil {
    				return err
    			}
    			continue
    		}
    		if name == "" {
    			name = f.Name
    		}
    		if err := e.encodeValue(join(prefix, name), fv); err != nil {
    			return err
    		}
    	}
    	return nil
    }

    func (e *encoder) encodeValue(path string, v reflect.Value) error {
    	switch v.Kind() {
    	case reflect.String:
    		e.line(path, strconv.Quote(v.String()))
    	case reflect.Bool:
    		e.line(path, strconv.FormatBool(v.Bool()))
    	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    		e.line(path, strconv.FormatInt(v.Int(), 10))
    	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
    		e.line(path, strconv.FormatUint(v.Uint(), 10))
    	case reflect.Float32, reflect.Float64:
    		e.line(path, strconv.FormatFloat(v.Float(), 'g', -1, v.Type().Bits()))
    	case reflect.Pointer, reflect.Interface:
    		if v.IsNil() {
    			e.line(path, "nil")
    			return nil
    		}
    		return e.encodeValue(path, v.Elem())
    	case reflect.Struct:
    		return e.encodeStruct(path, v)
    	case reflect.Slice, reflect.Array:
    		if v.Len() == 0 {
    			e.line(path, "[]")
    			return nil
    		}
    		for i := range v.Len() {
    			if err := e.encodeValue(join(path, strconv.Itoa(i)), v.Index(i)); err != nil {
    				return err
    			}
    		}
    	case reflect.Map:
    		if v.Type().Key().Kind() != reflect.String {
    			return fmt.Errorf("stash: cannot encode %s at %s: map keys must be strings", v.Type(), path)
    		}
    		if v.Len() == 0 {
    			e.line(path, "{}")
    			return nil
    		}
    		keys := v.MapKeys()
    		slices.SortFunc(keys, func(a, b reflect.Value) int { return cmp.Compare(a.String(), b.String()) })
    		for _, k := range keys {
    			if err := e.encodeValue(join(path, k.String()), v.MapIndex(k)); err != nil {
    				return err
    			}
    		}
    	default:
    		return fmt.Errorf("stash: cannot encode %s at %s", v.Type(), path)
    	}
    	return nil
    }

    type Limits struct {
    	MaxKeys int  `stash:"max_keys"`
    	Evict   bool `stash:"evict,omitempty"`
    }

    type Owner struct {
    	Name string `stash:"name"`
    }

    type Bucket struct {
    	ID     int               `stash:"id"`
    	Tags   []string          `stash:"tags"`
    	Limits Limits            `stash:"limits"`
    	Owner  *Owner            `stash:"owner"`
    	Labels map[string]string `stash:"labels,omitempty"`
    	Secret string            `stash:"-"`
    	cache  []byte
    }

    func main() {
    	b := Bucket{
    		ID:     7,
    		Tags:   []string{"hot", "eu"},
    		Limits: Limits{MaxKeys: 500},
    		Owner:  &Owner{"ada"},
    		Secret: "hunter2",
    	}
    	out, err := Encode(b)
    	fmt.Printf("%s(err: %v)\n", out, err)
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    	"time"
    )

    type Region string

    type Level int8

    type Audit struct {
    	CreatedBy string `stash:"created_by"`
    	Rev       uint16
    }

    type Point struct {
    	X, Y float64
    }

    type hidden struct {
    	Visible string
    }

    type Record struct {
    	Audit                   // embedded without a tag: fields are promoted
    	Region   Region         `stash:"region"`
    	Level    Level          `stash:"level"`
    	Score    float32        `stash:"score"`
    	TTL      time.Duration  `stash:"ttl"`
    	Path     []Point        `stash:"path"`
    	Grid     [2][2]int      `stash:"grid"`
    	Parent   *Record        `stash:"parent"`
    	Meta     map[string]any `stash:"meta"`
    	Note     *string        `stash:"note,omitempty"`
    	Extra    any            `stash:"extra"`
    	Empty    []int          `stash:"empty"`
    	NoTag    bool
    	Opt      int    `stash:",omitempty"`
    	internal hidden // unexported: skipped entirely
    	counts   map[string]int
    	Skip     func()                `stash:"-"`
    	Wrapped  struct{ A, B string } `stash:"wrapped"`
    }

    func encode(t *testing.T, v any) string {
    	t.Helper()
    	out, err := Encode(v)
    	if err != nil {
    		t.Fatalf("Encode(%+v) returned error %v, want none", v, err)
    	}
    	return out
    }

    func diffLines(t *testing.T, name, got, want string) {
    	t.Helper()
    	if got == want {
    		return
    	}
    	g, w := strings.Split(got, "\n"), strings.Split(want, "\n")
    	for i := range max(len(g), len(w)) {
    		var gl, wl string
    		if i < len(g) {
    			gl = g[i]
    		}
    		if i < len(w) {
    			wl = w[i]
    		}
    		if gl != wl {
    			t.Errorf("%s: line %d is %q, want %q\nfull output:\n%s\nwant:\n%s", name, i+1, gl, wl, got, want)
    			return
    		}
    	}
    }

    func TestBucketExample(t *testing.T) {
    	b := Bucket{ID: 7, Tags: []string{"hot", "eu"}, Limits: Limits{MaxKeys: 500}, Owner: &Owner{"ada"}, Secret: "hunter2"}
    	want := "id=7\ntags.0=\"hot\"\ntags.1=\"eu\"\nlimits.max_keys=500\nowner.name=\"ada\"\n"
    	diffLines(t, "Bucket example", encode(t, b), want)
    	diffLines(t, "pointer to Bucket", encode(t, &b), want)

    	b.Limits.Evict = true
    	b.Owner = nil
    	b.Tags = nil
    	b.Labels = map[string]string{"tier": "gold", "env": "prod"}
    	want = "id=7\ntags=[]\nlimits.max_keys=500\nlimits.evict=true\nowner=nil\nlabels.env=\"prod\"\nlabels.tier=\"gold\"\n"
    	diffLines(t, "Bucket with evict, nil owner, labels", encode(t, b), want)
    }

    func TestRecord(t *testing.T) {
    	note := "hi \"there\""
    	r := Record{
    		Audit:    Audit{CreatedBy: "ada", Rev: 3},
    		Region:   "eu-west",
    		Level:    -2,
    		Score:    0.1,
    		TTL:      90 * time.Second,
    		Path:     []Point{{1.5, 2}, {0, -0.25}},
    		Grid:     [2][2]int{{1, 2}, {3, 4}},
    		Parent:   &Record{Region: "root", Grid: [2][2]int{}, Meta: map[string]any{}},
    		Meta:     map[string]any{"b": []string{"x"}, "a": 1, "c": nil},
    		Note:     &note,
    		Extra:    Point{9, 8},
    		NoTag:    true,
    		internal: hidden{"no"},
    		counts:   map[string]int{"no": 1},
    		Skip:     func() {},
    	}
    	r.Wrapped.A = "a"
    	want := strings.Join([]string{
    		`created_by="ada"`,
    		`Rev=3`,
    		`region="eu-west"`,
    		`level=-2`,
    		`score=0.1`,
    		`ttl=90000000000`,
    		`path.0.X=1.5`,
    		`path.0.Y=2`,
    		`path.1.X=0`,
    		`path.1.Y=-0.25`,
    		`grid.0.0=1`,
    		`grid.0.1=2`,
    		`grid.1.0=3`,
    		`grid.1.1=4`,
    		`parent.created_by=""`,
    		`parent.Rev=0`,
    		`parent.region="root"`,
    		`parent.level=0`,
    		`parent.score=0`,
    		`parent.ttl=0`,
    		`parent.path=[]`,
    		`parent.grid.0.0=0`,
    		`parent.grid.0.1=0`,
    		`parent.grid.1.0=0`,
    		`parent.grid.1.1=0`,
    		`parent.parent=nil`,
    		`parent.meta={}`,
    		`parent.extra=nil`,
    		`parent.empty=[]`,
    		`parent.NoTag=false`,
    		`parent.wrapped.A=""`,
    		`parent.wrapped.B=""`,
    		`meta.a=1`,
    		`meta.b.0="x"`,
    		`meta.c=nil`,
    		`note="hi \"there\""`,
    		`extra.X=9`,
    		`extra.Y=8`,
    		`empty=[]`,
    		`NoTag=true`,
    		`wrapped.A="a"`,
    		`wrapped.B=""`,
    	}, "\n") + "\n"
    	diffLines(t, "Record", encode(t, r), want)

    	r2 := Record{Opt: 5}
    	out := encode(t, r2)
    	if !strings.Contains(out, "\nOpt=5\n") {
    		t.Errorf("a field tagged `stash:\",omitempty\"` should use its Go name: Encode(Record{Opt: 5}) =\n%s", out)
    	}
    	if strings.Contains(out, "note=") {
    		t.Errorf("a nil Note tagged omitempty should be left out: Encode(Record{Opt: 5}) =\n%s", out)
    	}
    }

    func TestEmptyStruct(t *testing.T) {
    	if got := encode(t, struct{}{}); got != "" {
    		t.Errorf("Encode(struct{}{}) = %q, want \"\"", got)
    	}
    	if got := encode(t, struct{ x int }{1}); got != "" {
    		t.Errorf("Encode(struct with only unexported fields) = %q, want \"\"", got)
    	}
    }

    func TestErrors(t *testing.T) {
    	for _, v := range []any{nil, 42, "str", []Point{{1, 2}}, (*Record)(nil)} {
    		if out, err := Encode(v); err == nil {
    			t.Errorf("Encode(%#v) = %q, nil, want an error: only structs and non-nil pointers to structs can be encoded", v, out)
    		}
    	}
    	type withChan struct {
    		OK   int
    		Pipe chan int `stash:"pipe"`
    	}
    	if _, err := Encode(withChan{Pipe: make(chan int)}); err == nil || !strings.Contains(err.Error(), "pipe") {
    		t.Errorf("Encode(struct with a chan field) error = %v, want an error mentioning the path \"pipe\"", err)
    	}
    	type deep struct {
    		Items []map[string]any `stash:"items"`
    	}
    	bad := deep{Items: []map[string]any{{"ok": 1}, {"fn": func() {}}}}
    	if _, err := Encode(bad); err == nil || !strings.Contains(err.Error(), "items.1.fn") {
    		t.Errorf("Encode(func nested in a slice of maps) error = %v, want an error mentioning the path \"items.1.fn\"", err)
    	}
    	type intKeys struct {
    		M map[int]string `stash:"m"`
    	}
    	if _, err := Encode(intKeys{M: map[int]string{1: "a"}}); err == nil || !strings.Contains(err.Error(), "m") {
    		t.Errorf("Encode(map with int keys) error = %v, want an error mentioning \"m\"", err)
    	}
    }
---

Stash exports records to a plain-text format that's easy to `grep`: one
`path=value` line per leaf value, with nested fields joined by dots. Write the
encoder with reflection so it works for **any** struct.

Implement `Encode(v)`. `v` must be a struct or a non-nil pointer to one;
anything else is an error. Output one line per value, each ending in `\n`, in
field order:

**Field names**

- Unexported fields are skipped, and so are fields tagged `stash:"-"`.
- The name is the tag's name part (`stash:"max_keys"`), or the Go field name
  when there's no tag or the name part is empty (`stash:",omitempty"`).
- The `omitempty` option skips a field whose value is its zero value
  (`reflect.Value.IsZero`).
- An **embedded** struct without a tag name adds no prefix: its fields are
  written as if they belonged to the outer struct.

**Values**, by kind

| Kind | Written as |
| --- | --- |
| string | quoted with `strconv.Quote`: `name="ada"` |
| bool, integers | `true`/`false`, decimal: `id=7`, `level=-2` |
| floats | `strconv.FormatFloat(f, 'g', -1, bits)`: `score=0.1` |
| struct | each field, with `path.` in front: `limits.max_keys=500` |
| pointer, interface | `nil` if nil, otherwise the value it holds, same path |
| slice, array | `[]` if empty, otherwise each element at `path.0`, `path.1`, ... |
| map with string keys | `{}` if empty, otherwise each value at `path.key`, keys sorted |

Anything else (channels, funcs, complex numbers, maps with non-string keys)
is an error whose message includes the path, like `items.1.fn`.

## Example

```go
b := Bucket{ID: 7, Tags: []string{"hot", "eu"}, Limits: Limits{MaxKeys: 500},
	Owner: &Owner{"ada"}, Secret: "hunter2"}
out, _ := Encode(b)
fmt.Print(out)
```

```
id=7
tags.0="hot"
tags.1="eu"
limits.max_keys=500
owner.name="ada"
```

(`evict` and `labels` are `omitempty` and zero, `Secret` is tagged `-`, and
`cache` is unexported.)

## Constraints

- The hidden tests use their own types: named strings and integers,
  `time.Duration`, arrays of arrays, slices of structs, a pointer to the same
  struct type, `map[string]any` holding slices and `nil`, anonymous struct
  fields and unexported struct fields.
- Values never contain pointer cycles.
- Calling `Interface()` on an unexported field panics. You don't need
  `Interface()` at all: use `v.String()`, `v.Int()`, `v.Elem()` and friends.
