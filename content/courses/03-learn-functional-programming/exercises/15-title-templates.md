---
title: Title Templates
difficulty: hard
after: sum-types
hints:
  - 'Write `Eval` as one type switch with a `case` for each of the six variants and a `default` for everything else (including `nil`). `Concat`, `Upper`, `If` and `Default` evaluate their children by calling `Eval` recursively, and return any error from a child straight away.'
  - 'For a missing variable, look it up with `v, ok := env[name]` (reading a nil map is fine) and return `fmt.Errorf("%w %q", ErrUndefined, name)`. In `Default`, check `errors.Is(err, ErrUndefined)`: that kind of error means "use the fallback", but any other error must still be returned.'
  - 'For `Vars`, walk the tree with a helper `collect(e Expr, seen map[string]bool) error` that has the same six cases, recursing into **every** child (both branches of an `If`). Then turn the set into a sorted slice with `slices.Sorted(maps.Keys(seen))`.'
exercise:
  starter: |
    package main

    import (
    	"errors"
    	"fmt"
    )

    // Expr is one of: Lit, Var, Concat, Upper, If, Default.
    type Expr interface{ isExpr() }

    type Lit struct{ S string }             // the text S
    type Var struct{ Name string }          // the value of variable Name
    type Concat struct{ Parts []Expr }      // every part, joined
    type Upper struct{ E Expr }             // E in upper case
    type If struct{ Cond, Then, Else Expr } // Then if Cond is non-empty, else Else
    type Default struct{ E, Fallback Expr } // E, or Fallback if E is "" or undefined

    func (Lit) isExpr()     {}
    func (Var) isExpr()     {}
    func (Concat) isExpr()  {}
    func (Upper) isExpr()   {}
    func (If) isExpr()      {}
    func (Default) isExpr() {}

    var ErrUndefined = errors.New("undefined variable")

    func Eval(e Expr, env map[string]string) (string, error) {
    	return "", nil
    }

    func Vars(e Expr) ([]string, error) {
    	return nil, nil
    }

    func main() {
    	title := Concat{[]Expr{
    		Upper{Default{Var{"title"}, Lit{"untitled"}}},
    		If{Var{"draft"}, Lit{" (draft)"}, Lit{""}},
    		Lit{" by "},
    		Var{"author"},
    	}}
    	fmt.Println(Eval(title, map[string]string{"title": "Specs", "draft": "yes", "author": "Ada"})) // want: SPECS (draft) by Ada <nil>
    	fmt.Println(Eval(title, map[string]string{"draft": "", "author": "Linus"}))                    // want: UNTITLED by Linus <nil>
    	fmt.Println(Eval(title, map[string]string{"title": "Specs", "draft": ""}))                     // want:  undefined variable "author"
    	fmt.Println(Vars(title))                                                                       // want: [author draft title] <nil>
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"maps"
    	"slices"
    	"strings"
    )

    // Expr is one of: Lit, Var, Concat, Upper, If, Default.
    type Expr interface{ isExpr() }

    type Lit struct{ S string }             // the text S
    type Var struct{ Name string }          // the value of variable Name
    type Concat struct{ Parts []Expr }      // every part, joined
    type Upper struct{ E Expr }             // E in upper case
    type If struct{ Cond, Then, Else Expr } // Then if Cond is non-empty, else Else
    type Default struct{ E, Fallback Expr } // E, or Fallback if E is "" or undefined

    func (Lit) isExpr()     {}
    func (Var) isExpr()     {}
    func (Concat) isExpr()  {}
    func (Upper) isExpr()   {}
    func (If) isExpr()      {}
    func (Default) isExpr() {}

    var ErrUndefined = errors.New("undefined variable")

    func Eval(e Expr, env map[string]string) (string, error) {
    	switch e := e.(type) {
    	case Lit:
    		return e.S, nil
    	case Var:
    		v, ok := env[e.Name]
    		if !ok {
    			return "", fmt.Errorf("%w %q", ErrUndefined, e.Name)
    		}
    		return v, nil
    	case Concat:
    		var b strings.Builder
    		for _, p := range e.Parts {
    			s, err := Eval(p, env)
    			if err != nil {
    				return "", err
    			}
    			b.WriteString(s)
    		}
    		return b.String(), nil
    	case Upper:
    		s, err := Eval(e.E, env)
    		if err != nil {
    			return "", err
    		}
    		return strings.ToUpper(s), nil
    	case If:
    		cond, err := Eval(e.Cond, env)
    		if err != nil {
    			return "", err
    		}
    		if cond != "" {
    			return Eval(e.Then, env)
    		}
    		return Eval(e.Else, env)
    	case Default:
    		s, err := Eval(e.E, env)
    		if err != nil && !errors.Is(err, ErrUndefined) {
    			return "", err
    		}
    		if err != nil || s == "" {
    			return Eval(e.Fallback, env)
    		}
    		return s, nil
    	default:
    		return "", fmt.Errorf("unknown expression %T", e)
    	}
    }

    func Vars(e Expr) ([]string, error) {
    	seen := map[string]bool{}
    	if err := collectVars(e, seen); err != nil {
    		return nil, err
    	}
    	return slices.Sorted(maps.Keys(seen)), nil
    }

    func collectVars(e Expr, seen map[string]bool) error {
    	var children []Expr
    	switch e := e.(type) {
    	case Lit:
    	case Var:
    		seen[e.Name] = true
    	case Concat:
    		children = e.Parts
    	case Upper:
    		children = []Expr{e.E}
    	case If:
    		children = []Expr{e.Cond, e.Then, e.Else}
    	case Default:
    		children = []Expr{e.E, e.Fallback}
    	default:
    		return fmt.Errorf("unknown expression %T", e)
    	}
    	for _, c := range children {
    		if err := collectVars(c, seen); err != nil {
    			return err
    		}
    	}
    	return nil
    }

    func main() {
    	title := Concat{[]Expr{
    		Upper{Default{Var{"title"}, Lit{"untitled"}}},
    		If{Var{"draft"}, Lit{" (draft)"}, Lit{""}},
    		Lit{" by "},
    		Var{"author"},
    	}}
    	fmt.Println(Eval(title, map[string]string{"title": "Specs", "draft": "yes", "author": "Ada"}))
    	fmt.Println(Eval(title, map[string]string{"draft": "", "author": "Linus"}))
    	fmt.Println(Eval(title, map[string]string{"title": "Specs", "draft": ""}))
    	fmt.Println(Vars(title))
    }
  tests: |
    package main

    import (
    	"errors"
    	"slices"
    	"testing"
    )

    // Lower is an expression type that Eval and Vars don't know about.
    type Lower struct{ E Expr }

    func (Lower) isExpr() {}

    var env = map[string]string{"title": "Specs", "author": "Ada", "empty": "", "draft": "yes"}

    func TestEval(t *testing.T) {
    	tests := []struct {
    		name string
    		e    Expr
    		want string
    	}{
    		{"lit", Lit{"hello"}, "hello"},
    		{"var", Var{"title"}, "Specs"},
    		{"empty var", Var{"empty"}, ""},
    		{"concat", Concat{[]Expr{Var{"title"}, Lit{" by "}, Var{"author"}}}, "Specs by Ada"},
    		{"empty concat", Concat{}, ""},
    		{"upper", Upper{Var{"title"}}, "SPECS"},
    		{"nested", Upper{Concat{[]Expr{Lit{"a"}, Upper{Lit{"b"}}, Lit{"c"}}}}, "ABC"},
    		{"if true", If{Var{"draft"}, Lit{"DRAFT"}, Lit{"FINAL"}}, "DRAFT"},
    		{"if false", If{Var{"empty"}, Lit{"DRAFT"}, Lit{"FINAL"}}, "FINAL"},
    		{"if lit cond", If{Lit{""}, Lit{"x"}, Upper{Var{"author"}}}, "ADA"},
    		{"default set", Default{Var{"title"}, Lit{"untitled"}}, "Specs"},
    		{"default empty", Default{Var{"empty"}, Lit{"untitled"}}, "untitled"},
    		{"default undefined", Default{Var{"subtitle"}, Lit{"untitled"}}, "untitled"},
    		{"default chain", Default{Var{"a"}, Default{Var{"b"}, Var{"author"}}}, "Ada"},
    		{"default undefined deep inside", Default{Concat{[]Expr{Lit{"by "}, Var{"editor"}}}, Lit{"anon"}}, "anon"},
    	}
    	for _, tt := range tests {
    		got, err := Eval(tt.e, env)
    		if got != tt.want || err != nil {
    			t.Errorf("%s: Eval(%#v) = %q, %v, want %q, nil", tt.name, tt.e, got, err, tt.want)
    		}
    	}
    }

    func TestEvalErrors(t *testing.T) {
    	tests := []struct {
    		name      string
    		e         Expr
    		wantErr   string
    		undefined bool
    	}{
    		{"undefined var", Var{"subtitle"}, `undefined variable "subtitle"`, true},
    		{"undefined in concat", Concat{[]Expr{Lit{"x"}, Var{"nope"}}}, `undefined variable "nope"`, true},
    		{"undefined in upper", Upper{Var{"nope"}}, `undefined variable "nope"`, true},
    		{"undefined cond", If{Var{"nope"}, Lit{"a"}, Lit{"b"}}, `undefined variable "nope"`, true},
    		{"undefined fallback", Default{Var{"a"}, Var{"b"}}, `undefined variable "b"`, true},
    		{"nil", nil, "unknown expression <nil>", false},
    		{"unknown", Lower{Lit{"x"}}, "unknown expression main.Lower", false},
    		{"unknown deep inside", Concat{[]Expr{Upper{Lower{Lit{"x"}}}}}, "unknown expression main.Lower", false},
    		{"nil part", Concat{[]Expr{Lit{"a"}, nil}}, "unknown expression <nil>", false},
    		{"default doesn't hide unknown", Default{Lower{Lit{"x"}}, Lit{"fallback"}}, "unknown expression main.Lower", false},
    	}
    	for _, tt := range tests {
    		got, err := Eval(tt.e, env)
    		if err == nil || err.Error() != tt.wantErr || got != "" {
    			t.Errorf("%s: Eval(%#v) = %q, %v, want \"\" and the error %q", tt.name, tt.e, got, err, tt.wantErr)
    			continue
    		}
    		if errors.Is(err, ErrUndefined) != tt.undefined {
    			t.Errorf("%s: errors.Is(%q, ErrUndefined) = %v, want %v", tt.name, err, !tt.undefined, tt.undefined)
    		}
    	}
    }

    func TestEvalOnlyChosenBranch(t *testing.T) {
    	tests := []struct {
    		name string
    		e    Expr
    		want string
    	}{
    		{"if skips else", If{Lit{"y"}, Lit{"then"}, Var{"missing"}}, "then"},
    		{"if skips then", If{Lit{""}, Lower{nil}, Lit{"else"}}, "else"},
    		{"default skips fallback", Default{Var{"title"}, Var{"missing"}}, "Specs"},
    	}
    	for _, tt := range tests {
    		got, err := Eval(tt.e, env)
    		if got != tt.want || err != nil {
    			t.Errorf("%s: Eval(%#v) = %q, %v, want %q, nil: evaluate only the branch you need", tt.name, tt.e, got, err, tt.want)
    		}
    	}
    }

    func TestEvalDoesNotChangeEnv(t *testing.T) {
    	e := map[string]string{"a": "1"}
    	Eval(Concat{[]Expr{Var{"a"}, Default{Var{"b"}, Lit{"2"}}}}, e)
    	if len(e) != 1 || e["a"] != "1" {
    		t.Errorf("Eval changed its environment to %v, want map[a:1]", e)
    	}
    	if got, err := Eval(Var{"x"}, nil); err == nil || !errors.Is(err, ErrUndefined) {
    		t.Errorf("Eval(Var{x}, nil env) = %q, %v, want an ErrUndefined error", got, err)
    	}
    }

    func TestVars(t *testing.T) {
    	tests := []struct {
    		name string
    		e    Expr
    		want []string
    	}{
    		{"lit", Lit{"x"}, []string{}},
    		{"var", Var{"title"}, []string{"title"}},
    		{"sorted", Concat{[]Expr{Var{"b"}, Var{"c"}, Var{"a"}}}, []string{"a", "b", "c"}},
    		{"unique", Concat{[]Expr{Var{"a"}, Upper{Var{"a"}}, Var{"a"}}}, []string{"a"}},
    		{"every branch", If{Var{"cond"}, Var{"yes"}, Upper{Var{"no"}}}, []string{"cond", "no", "yes"}},
    		{"default", Default{Var{"x"}, Default{Lit{""}, Var{"y"}}}, []string{"x", "y"}},
    		{"empty concat", Concat{}, []string{}},
    	}
    	for _, tt := range tests {
    		got, err := Vars(tt.e)
    		if !slices.Equal(got, tt.want) || err != nil {
    			t.Errorf("%s: Vars(%#v) = %q, %v, want %q, nil", tt.name, tt.e, got, err, tt.want)
    		}
    	}
    }

    func TestVarsErrors(t *testing.T) {
    	for _, e := range []Expr{nil, Lower{Var{"a"}}, If{Var{"a"}, Lit{""}, Lower{Lit{""}}}, Default{Var{"a"}, nil}} {
    		got, err := Vars(e)
    		if err == nil || got != nil {
    			t.Errorf("Vars(%#v) = %q, %v, want nil and an \"unknown expression\" error", e, got, err)
    		}
    	}
    }

    func TestEvalDeep(t *testing.T) {
    	var e Expr = Var{"author"}
    	for range 300 {
    		e = Default{Upper{e}, Lit{"?"}}
    	}
    	if got, err := Eval(e, env); got != "ADA" || err != nil {
    		t.Errorf("Eval(300 nested Default{Upper{...}}) = %q, %v, want \"ADA\", nil", got, err)
    	}
    }
---

Doc2Doc lets users write **title templates** such as "the title in capitals,
or `UNTITLED`, plus ` (draft)` for drafts, then ` by ` and the author". The
parser has already turned each template into an expression tree. Expressions
form a **sum type**: every `Expr` is exactly one of these six variants.

| Variant | Value |
|---|---|
| `Lit{S}` | the text `S` |
| `Var{Name}` | `env[Name]`; if `Name` isn't in `env`, the error `undefined variable "Name"` wrapping `ErrUndefined` |
| `Concat{Parts}` | the values of all parts joined together (`""` for no parts) |
| `Upper{E}` | the value of `E` in upper case |
| `If{Cond, Then, Else}` | the value of `Then` if `Cond`'s value is non-empty, otherwise the value of `Else` |
| `Default{E, Fallback}` | the value of `E`, unless it's `""` **or** evaluating it failed with an `ErrUndefined` error; then the value of `Fallback` |

Write two functions that handle **every** variant:

- `Eval(e, env)` returns the value of `e`. Any error inside the tree makes it
  return `""` and that error. Any other expression type, including `nil`, is
  the error `fmt.Errorf("unknown expression %T", e)`, which `Default` does
  **not** swallow. `If` and `Default` evaluate only what they need: the branch
  not taken, and a fallback that isn't needed, are never evaluated.
- `Vars(e)` returns the sorted, de-duplicated names of every variable that
  appears anywhere in `e` (both branches of an `If`, both sides of a
  `Default`), or `nil` and an `unknown expression` error if the tree contains
  anything else.

## Example

```go
title := Concat{[]Expr{
	Upper{Default{Var{"title"}, Lit{"untitled"}}},
	If{Var{"draft"}, Lit{" (draft)"}, Lit{""}},
	Lit{" by "},
	Var{"author"},
}}
Eval(title, map[string]string{"title": "Specs", "draft": "yes", "author": "Ada"})
// "SPECS (draft) by Ada", nil
Eval(title, map[string]string{"draft": "", "author": "Linus"})
// "UNTITLED by Linus", nil
Eval(title, map[string]string{"title": "Specs", "draft": ""})
// "", undefined variable "author"
Vars(title)
// [author draft title], nil
```

## Constraints

- `env` may be `nil`, and must never be modified.
- Trees can be a few hundred levels deep.
