---
title: Meet encoding/json/v2
quiz:
  - question: |
      With `encoding/json/v2`, what happens here?

      ```go
      type Issue struct {
          ID int `json:"id"`
      }
      var iss Issue
      err := json.Unmarshal([]byte(`{"ID": 7}`), &iss)
      ```
    options:
      - text: '`iss.ID` is 7 and `err` is nil'
      - text: '`iss.ID` is 0 and `err` is nil'
        correct: true
      - text: '`err` reports an unknown member'
    explanation: |
      v2 matches names **case-sensitively**, so `"ID"` doesn't match the `id` tag. The
      key is treated as unknown, and unknown keys are ignored by default, so no error.
      (v1 would have matched it case-insensitively and set 7.)
  - question: 'How does `encoding/json/v2` marshal a nil `[]string` field with no options?'
    options:
      - text: '`null`'
      - text: '`[]`'
        correct: true
      - text: The field is left out
    explanation: |
      v1 wrote `null` for nil slices and maps, which surprised clients in other
      languages. v2 writes `[]` for a nil slice and `{}` for a nil map.
  - question: 'In Go 1.27, what do you need to do to use `encoding/json/v2`?'
    options:
      - text: Set `GOEXPERIMENT=jsonv2` when building
      - text: Run `go get` to download it
      - text: 'Just import `"encoding/json/v2"`'
        correct: true
    explanation: |
      It was an experiment in Go 1.25 and 1.26, behind `GOEXPERIMENT=jsonv2`. In Go 1.27
      it's on by default and part of the standard library. The package name is still
      `json`.
---

`encoding/json` has served Go since 1.0, but it has some rough edges that can't be
fixed without breaking old programs. So Go 1.27 ships a second version alongside it:
**`encoding/json/v2`**.

## Using it

It's in the standard library, and in Go 1.27 it's **on by default**. (In Go 1.25 and
1.26 it existed only behind `GOEXPERIMENT=jsonv2`. You can still switch it off with
`GOEXPERIMENT=nojsonv2`, but you won't want to.) The package is still called `json`:

```go
import "encoding/json/v2"

data, err := json.Marshal(issue)
err = json.Unmarshal(data, &issue)
```

`Marshal` and `Unmarshal` have the same shape as before, plus optional options at the
end. Your struct tags carry over, too.

There's also a companion package, `encoding/json/jsontext`, that handles the raw
syntax (tokens, quoting, formatting). You'll only need it for a few options.

## What changed, side by side

The same input through both versions:

```go
package main

import (
	jsonv1 "encoding/json"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
)

type Issue struct {
	ID     int      `json:"id"`
	Title  string   `json:"title"`
	Labels []string `json:"labels"`
}

func main() {
	input := []byte(`{"ID": 7, "title": "Dark mode", "title": "Light mode"}`)

	var a Issue
	errA := jsonv1.Unmarshal(input, &a)
	fmt.Printf("v1: %+v err=%v\n", a, errA)

	var b Issue
	errB := json.Unmarshal(input, &b)
	fmt.Printf("v2: %+v err=%v\n", b, errB)

	var c Issue
	errC := json.Unmarshal(input, &c, jsontext.AllowDuplicateNames(true), json.MatchCaseInsensitiveNames(true))
	fmt.Printf("v2 relaxed: %+v err=%v\n", c, errC)

	out1, _ := jsonv1.Marshal(Issue{ID: 1})
	out2, _ := json.Marshal(Issue{ID: 1})
	fmt.Println("v1:", string(out1))
	fmt.Println("v2:", string(out2))
}
```

Output:

```
v1: {ID:7 Title:Light mode Labels:[]} err=<nil>
v2: {ID:0 Title:Dark mode Labels:[]} err=jsontext: duplicate object member name "title"
v2 relaxed: {ID:7 Title:Light mode Labels:[]} err=<nil>
v1: {"id":1,"title":"","labels":null}
v2: {"id":1,"title":"","labels":[]}
```

(The two packages are both named `json`, so the program renames v1 to `jsonv1`.)

## The big differences

**Stricter by default**, to close security holes:

- **Duplicate keys are an error.** With v1, `{"title": "a", "title": "b"}` quietly
  keeps the last one. If two services disagree about which one wins, an attacker can
  show each of them something different. v2 refuses.
- **Invalid UTF-8 is an error.** v1 silently replaced bad bytes with `�`. v2 reports it.
- **Names match case-sensitively.** v1 would fill `id` from `"ID"` or `"Id"`. In v2,
  only `"id"` matches.

**Friendlier output:**

- A nil slice becomes `[]` and a nil map becomes `{}`, not `null`.
- HTML characters like `<` and `&` aren't escaped any more (v1 wrote `<` as `\u003c`).

**Other changes worth knowing:**

- Map keys are **not sorted** by default. Pass `json.Deterministic(true)` if you need
  stable output, for example in a golden-file test.
- `time.Duration` has no default format and returns an error. Store durations as
  strings or integers with a clear unit.
- `omitempty` now means "omit if the JSON would be empty (`null`, `""`, `{}` or `[]`)".
  For numbers and bools, use `omitzero` (next lesson).

Notice one more thing in the output: when v2 hit the duplicate, it had already set
`Title`. After a decode error, don't trust the half-filled value.

## Loosening it up when you must

Each strict default has an option to turn it off for one call:

| Option | Effect |
| --- | --- |
| `jsontext.AllowDuplicateNames(true)` | last duplicate wins, like v1 |
| `jsontext.AllowInvalidUTF8(true)` | accept bad bytes |
| `json.MatchCaseInsensitiveNames(true)` | v1-style name matching |
| `json.RejectUnknownMembers(true)` | the opposite: fail on keys you didn't expect |

You can also put `case:ignore` in a single field's tag: `` `json:"id,case:ignore"` ``.

## What about the old package?

`encoding/json` isn't going anywhere, and your existing code keeps working. In Go 1.27
it's actually implemented **on top of v2**, with options that recreate the old
behaviour. Its docs now recommend v2 for new code, so the rest of this course imports
`"encoding/json/v2"`.
