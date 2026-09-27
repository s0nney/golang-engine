---
title: 'Stash: A Tag-Driven Validator'
quiz:
  - question: |
      Stash's `Validate` joins its errors with `errors.Join`. How can a caller get at the first `*FieldError`?
    options:
      - text: '`err.(*FieldError)`'
      - text: '`errors.AsType[*FieldError](err)`'
        correct: true
      - text: '`strings.Split(err.Error(), "\n")[0]`'
      - text: It can't; joined errors hide their parts
    explanation: |
      A joined error has an `Unwrap() []error` method, and `errors.AsType` (like
      `errors.Is` and `errors.As`) walks into every joined error. A plain type assertion
      only looks at the outer error, which is the join, not a `*FieldError`.
  - question: Why does `Validate` stop checking a field after its first failing rule?
    options:
      - text: Reflection can only read a field once
      - text: 'So an empty required field reports just `required`, not also `min=3`; one clear message per field'
        correct: true
      - text: Because `errors.Join` accepts only one error per field
      - text: It's faster, and speed is the main concern
    explanation: |
      Rules are listed in order of importance (`required,min=3,max=64`), and later rules
      rarely add information once an earlier one fails. Stopping keeps error reports short
      and actionable.
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    	"reflect"
    	"strconv"
    	"strings"
    )

    // FieldError reports a field that failed a validation rule.
    type FieldError struct {
    	Field string
    	Rule  string
    }

    func (e *FieldError) Error() string { return e.Field + ": failed " + e.Rule }

    // Validate checks a struct (or pointer to one) against its `validate` tags.
    // Each field gets at most one error: the first rule it fails, in tag order.
    // Errors are joined with errors.Join, in field order.
    func Validate(v any) error {
    	rv := reflect.ValueOf(v)
    	for rv.Kind() == reflect.Pointer && !rv.IsNil() {
    		rv = rv.Elem()
    	}
    	if rv.Kind() != reflect.Struct {
    		return fmt.Errorf("stash: Validate needs a struct, got %T", v)
    	}
    	var errs []error
    	// ? For each exported field with a validate tag, check its rules in order
    	// (strings.SplitSeq splits the tag on commas). Stop at the field's first
    	// failure: a *FieldError if the rule failed, or "Field: <err>" (wrapping
    	// err) if checkRule returned an error.
    	return errors.Join(errs...)
    }

    // checkRule reports whether v passes one rule: "required", "min=N" or "max=N".
    // It returns an error for unknown rules, bad numbers, or kinds a rule
    // doesn't support.
    func checkRule(v reflect.Value, rule string) (bool, error) {
    	name, arg, _ := strings.Cut(rule, "=")
    	// ?
    	_, _ = name, arg
    	_ = strconv.ParseFloat
    	return true, nil
    }

    type Entry struct {
    	Key   string   `validate:"required,min=3,max=64"`
    	Size  int64    `validate:"min=1"`
    	Tags  []string `validate:"max=3"`
    	Ratio float64  `validate:"min=0,max=1"`
    	Notes string
    }

    func main() {
    	ok := Entry{Key: "logo.png", Size: 2048, Ratio: 0.5}
    	fmt.Println(Validate(ok))

    	bad := &Entry{Key: "a", Tags: []string{"x", "y", "z", "w"}, Ratio: 1.5}
    	err := Validate(bad)
    	fmt.Println(err)
    	if fe, ok := errors.AsType[*FieldError](err); ok {
    		fmt.Println("first failure:", fe.Field, fe.Rule)
    	}
    	fmt.Println(Validate(42))
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"reflect"
    	"strconv"
    	"strings"
    )

    // FieldError reports a field that failed a validation rule.
    type FieldError struct {
    	Field string
    	Rule  string
    }

    func (e *FieldError) Error() string { return e.Field + ": failed " + e.Rule }

    // Validate checks a struct (or pointer to one) against its `validate` tags.
    // Each field gets at most one error: the first rule it fails, in tag order.
    // Errors are joined with errors.Join, in field order.
    func Validate(v any) error {
    	rv := reflect.ValueOf(v)
    	for rv.Kind() == reflect.Pointer && !rv.IsNil() {
    		rv = rv.Elem()
    	}
    	if rv.Kind() != reflect.Struct {
    		return fmt.Errorf("stash: Validate needs a struct, got %T", v)
    	}
    	var errs []error
    	for f, fv := range rv.Fields() {
    		tag, ok := f.Tag.Lookup("validate")
    		if !ok || !f.IsExported() {
    			continue
    		}
    		for rule := range strings.SplitSeq(tag, ",") {
    			ok, err := checkRule(fv, rule)
    			if err != nil {
    				errs = append(errs, fmt.Errorf("%s: %w", f.Name, err))
    				break
    			}
    			if !ok {
    				errs = append(errs, &FieldError{Field: f.Name, Rule: rule})
    				break
    			}
    		}
    	}
    	return errors.Join(errs...)
    }

    // checkRule reports whether v passes one rule: "required", "min=N" or "max=N".
    // It returns an error for unknown rules, bad numbers, or kinds a rule
    // doesn't support.
    func checkRule(v reflect.Value, rule string) (bool, error) {
    	name, arg, _ := strings.Cut(rule, "=")
    	switch name {
    	case "required":
    		return !v.IsZero(), nil
    	case "min", "max":
    		limit, err := strconv.ParseFloat(arg, 64)
    		if err != nil {
    			return false, fmt.Errorf("bad number in rule %q", rule)
    		}
    		var x float64
    		switch v.Kind() {
    		case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
    			x = float64(v.Len())
    		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    			x = float64(v.Int())
    		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
    			x = float64(v.Uint())
    		case reflect.Float32, reflect.Float64:
    			x = v.Float()
    		default:
    			return false, fmt.Errorf("rule %q doesn't support kind %s", rule, v.Kind())
    		}
    		if name == "min" {
    			return x >= limit, nil
    		}
    		return x <= limit, nil
    	}
    	return false, fmt.Errorf("unknown rule %q", rule)
    }

    type Entry struct {
    	Key   string   `validate:"required,min=3,max=64"`
    	Size  int64    `validate:"min=1"`
    	Tags  []string `validate:"max=3"`
    	Ratio float64  `validate:"min=0,max=1"`
    	Notes string
    }

    func main() {
    	ok := Entry{Key: "logo.png", Size: 2048, Ratio: 0.5}
    	fmt.Println(Validate(ok))

    	bad := &Entry{Key: "a", Tags: []string{"x", "y", "z", "w"}, Ratio: 1.5}
    	err := Validate(bad)
    	fmt.Println(err)
    	if fe, ok := errors.AsType[*FieldError](err); ok {
    		fmt.Println("first failure:", fe.Field, fe.Rule)
    	}
    	fmt.Println(Validate(42))
    }
  tests: |
    package main

    import (
    	"errors"
    	"strings"
    	"testing"
    )

    func TestValidEntry(t *testing.T) {
    	ok := Entry{Key: "logo.png", Size: 2048, Tags: []string{"img"}, Ratio: 1}
    	if err := Validate(ok); err != nil {
    		t.Errorf("Validate(valid Entry) = %v, want nil", err)
    	}
    	if err := Validate(&ok); err != nil {
    		t.Errorf("Validate(&valid Entry) = %v, want nil", err)
    	}
    }

    func TestInvalidEntry(t *testing.T) {
    	bad := &Entry{Key: "a", Tags: []string{"x", "y", "z", "w"}, Ratio: 1.5}
    	err := Validate(bad)
    	want := "Key: failed min=3\nSize: failed min=1\nTags: failed max=3\nRatio: failed max=1"
    	if err == nil || err.Error() != want {
    		t.Fatalf("Validate(bad Entry) =\n%v\nwant\n%s", err, want)
    	}
    	fe, ok := errors.AsType[*FieldError](err)
    	if !ok || fe.Field != "Key" || fe.Rule != "min=3" {
    		t.Errorf("errors.AsType[*FieldError] = %+v, %v; want the Key/min=3 FieldError", fe, ok)
    	}
    }

    type Account struct {
    	Name  string          `validate:"required,min=2"`
    	Age   uint8           `validate:"min=18,max=130"`
    	Roles map[string]bool `validate:"min=1"`
    }

    func TestRequiredStopsAtFirstFailure(t *testing.T) {
    	err := Validate(Account{Age: 20, Roles: map[string]bool{"admin": true}})
    	if err == nil || err.Error() != "Name: failed required" {
    		t.Errorf("empty Name: got %v, want exactly \"Name: failed required\" (one error per field)", err)
    	}
    	err = Validate(Account{Name: "Al", Age: 12})
    	if err == nil || err.Error() != "Age: failed min=18\nRoles: failed min=1" {
    		t.Errorf("young, no roles: got %v, want Age and Roles errors", err)
    	}
    }

    type Broken struct {
    	Flag bool   `validate:"min=1"`
    	Code string `validate:"exactly=4"`
    	Max  int    `validate:"max=ten"`
    }

    func TestRuleErrors(t *testing.T) {
    	err := Validate(Broken{})
    	if err == nil {
    		t.Fatalf("Validate(Broken{}) = nil, want errors for all three fields")
    	}
    	msg := err.Error()
    	for _, want := range []string{"Flag:", "Code:", "unknown rule", "Max:"} {
    		if !strings.Contains(msg, want) {
    			t.Errorf("Validate(Broken{}) error %q should contain %q", msg, want)
    		}
    	}
    	if _, ok := errors.AsType[*FieldError](err); ok {
    		t.Errorf("rule errors (bad kinds, unknown rules, bad numbers) should not be FieldErrors")
    	}
    }

    func TestNotAStruct(t *testing.T) {
    	if err := Validate(42); err == nil {
    		t.Errorf("Validate(42) = nil, want an error")
    	}
    }
---

Time to cash in the whole chapter. Stash's `Validate` checks any struct against rules written in its tags:

```go
type Entry struct {
	Key   string   `validate:"required,min=3,max=64"`
	Size  int64    `validate:"min=1"`
	Tags  []string `validate:"max=3"`
	Ratio float64  `validate:"min=0,max=1"`
	Notes string
}
```

Calling `Validate(&Entry{Key: "a", Tags: []string{"x", "y", "z", "w"}, Ratio: 1.5})` returns an error that prints as:

```
Key: failed min=3
Size: failed min=1
Tags: failed max=3
Ratio: failed max=1
```

## The rules

- **`required`**: the field isn't its zero value (`Value.IsZero`).
- **`min=N` / `max=N`**: compare `N` against a number that depends on the field's **kind**:
  - strings, slices, maps and arrays: the **length** (`Len()`);
  - int kinds: `Int()`; uint kinds: `Uint()`; float kinds: `Float()`;
  - any other kind is a mistake in the tag, reported as an error.
- Anything else is an **unknown rule**, also an error.

Parsing `N` with `strconv.ParseFloat` lets one code path handle all the number kinds.

## Two kinds of failure

`Validate` distinguishes **invalid data** from **invalid rules**:

- Data that fails a rule produces a `*FieldError{Field, Rule}`, which prints as `Key: failed min=3`. Callers can find it with `errors.AsType[*FieldError](err)`, for example to highlight a form field.
- A broken rule (`validate:"min=ten"`, `min` on a `bool`, a typo like `exactly=4`) is the programmer's mistake. It's reported as a plain wrapped error, `Field: unknown rule "exactly=4"`, so it can't be mistaken for bad user input.

Each field reports **at most one** error: the first rule it fails. All the fields' errors are combined with `errors.Join`, in field order, so the caller sees everything that's wrong in one go.

## How it uses the chapter

- Accepts a struct or a pointer to one: loop `Elem()` while the kind is `Pointer` (lesson 4's "pointer, then `Elem`", though here we only read).
- `Value.Fields()` gives each `StructField` (for the name and tag) with its `Value` (lesson 3).
- `Tag.Lookup("validate")` skips untagged fields; unexported fields are skipped too.
- `strings.SplitSeq(tag, ",")` iterates the rules; `strings.Cut(rule, "=")` splits name from argument.
- A **Kind switch** picks the right getter (lesson 2).

## Your turn

The starter has the scaffolding: `FieldError`, the pointer-and-struct check at the top of `Validate`, and an `Entry` to play with. Finish:

1. **`checkRule(v, rule)`**, returning `(passed, err)`: handle `required`, `min=N` and `max=N` as described, returning an error for bad numbers, unsupported kinds and unknown rules.
2. **The loop in `Validate`**: for each exported field with a `validate` tag, check its rules in order. On the first failure, record a `*FieldError` (rule failed) or `fmt.Errorf("%s: %w", f.Name, err)` (rule error), and move on to the next field. Return `errors.Join(errs...)`.

In the capstone, `Validate` guards every write to Stash's repository.
