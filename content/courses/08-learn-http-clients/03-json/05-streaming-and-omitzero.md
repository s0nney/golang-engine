---
title: MarshalWrite, UnmarshalRead and omitzero
quiz:
  - question: |
      What does this print?

      ```go
      type Update struct {
          Points int       `json:"points,omitzero"`
          Due    time.Time `json:"due,omitzero"`
          Title  string    `json:"title,omitempty"`
      }
      json.MarshalWrite(os.Stdout, Update{Title: "Retitle"})
      ```
    options:
      - text: '`{"points":0,"due":"0001-01-01T00:00:00Z","title":"Retitle"}`'
      - text: '`{"due":"0001-01-01T00:00:00Z","title":"Retitle"}`'
      - text: '`{"title":"Retitle"}`'
        correct: true
    explanation: |
      `omitzero` drops `Points` because it's `0`, and drops `Due` because
      `time.Time` has an `IsZero` method that reports true. `Title` isn't empty, so it
      stays.
  - question: |
      The body is `{"id": 3} {"id": 4}`. What happens?

      ```go
      err := json.UnmarshalRead(resp.Body, &iss)
      ```
    options:
      - text: '`iss.ID` is 3 and `err` is nil'
      - text: '`iss.ID` is 4 and `err` is nil'
      - text: '`err` is non-nil, because the input must be a single JSON value'
        correct: true
    explanation: |
      `UnmarshalRead` reads the whole reader and insists on exactly one JSON value
      (plus whitespace). v1's `Decoder.Decode` would have returned the first object
      and ignored the rest.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"io"
    	"os"
    	"time"
    )

    // IssueUpdate is the body of PATCH /issues/{id}. Only the fields that are
    // set may appear in the JSON, so the server leaves everything else alone.
    // Add struct tags so the JSON keys are title, points, due and labels.
    type IssueUpdate struct {
    	Title  *string   // nil: leave alone
    	Points *int      // nil: leave alone. new(0): set the points to 0
    	Due    time.Time // zero time: leave alone
    	Labels []string  // nil: leave alone. []string{}: remove every label
    }

    // writeUpdate writes u to w as JSON.
    func writeUpdate(w io.Writer, u IssueUpdate) error {
    	// ?
    	return fmt.Errorf("writeUpdate: not implemented yet")
    }

    func main() {
    	updates := []IssueUpdate{
    		{Title: new("Login fails on Safari")},
    		{Points: new(0), Labels: []string{}},
    		{Due: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)},
    	}
    	for _, u := range updates {
    		if err := writeUpdate(os.Stdout, u); err != nil {
    			fmt.Println("error:", err)
    		}
    		fmt.Println()
    	}
    }
  solution: |
    package main

    import (
    	"encoding/json/v2"
    	"fmt"
    	"io"
    	"os"
    	"time"
    )

    type IssueUpdate struct {
    	Title  *string   `json:"title,omitzero"`
    	Points *int      `json:"points,omitzero"`
    	Due    time.Time `json:"due,omitzero"`
    	Labels []string  `json:"labels,omitzero"`
    }

    func writeUpdate(w io.Writer, u IssueUpdate) error {
    	return json.MarshalWrite(w, u)
    }

    func main() {
    	updates := []IssueUpdate{
    		{Title: new("Login fails on Safari")},
    		{Points: new(0), Labels: []string{}},
    		{Due: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)},
    	}
    	for _, u := range updates {
    		if err := writeUpdate(os.Stdout, u); err != nil {
    			fmt.Println("error:", err)
    		}
    		fmt.Println()
    	}
    }
  tests: |
    package main

    import (
    	"bytes"
    	"strings"
    	"testing"
    	"time"
    )

    func TestWriteUpdate(t *testing.T) {
    	due := time.Date(2026, 10, 1, 17, 30, 0, 0, time.UTC)
    	tests := []struct {
    		name string
    		u    IssueUpdate
    		want string
    	}{
    		{"nothing set", IssueUpdate{}, `{}`},
    		{"new title", IssueUpdate{Title: new("Retitled")}, `{"title":"Retitled"}`},
    		{"clear the title", IssueUpdate{Title: new("")}, `{"title":""}`},
    		{"points set to zero", IssueUpdate{Points: new(0)}, `{"points":0}`},
    		{"due date", IssueUpdate{Due: due}, `{"due":"2026-10-01T17:30:00Z"}`},
    		{"remove every label", IssueUpdate{Labels: []string{}}, `{"labels":[]}`},
    		{"points and labels", IssueUpdate{Points: new(3), Labels: []string{"bug", "ui"}}, `{"points":3,"labels":["bug","ui"]}`},
    		{"everything", IssueUpdate{Title: new("T"), Points: new(1), Due: due, Labels: []string{"x"}},
    			`{"title":"T","points":1,"due":"2026-10-01T17:30:00Z","labels":["x"]}`},
    	}
    	for _, tt := range tests {
    		var buf bytes.Buffer
    		if err := writeUpdate(&buf, tt.u); err != nil {
    			t.Errorf("%s: writeUpdate returned error %v", tt.name, err)
    			continue
    		}
    		if got := strings.TrimSpace(buf.String()); got != tt.want {
    			t.Errorf("%s: writeUpdate wrote %s, want %s", tt.name, got, tt.want)
    		}
    	}
    }
---

v2 also tidies up how you read and write JSON from **streams**, which is exactly what
HTTP bodies are.

## UnmarshalRead: decode a whole body

`json.UnmarshalRead(r, &v)` reads from an `io.Reader` until EOF and decodes it:

```go
var issues []Issue
if err := json.UnmarshalRead(resp.Body, &issues); err != nil {
	return nil, fmt.Errorf("decoding issues: %w", err)
}
```

Compared with v1's `json.NewDecoder(resp.Body).Decode(&issues)`:

- It's one call, with no decoder to create.
- It reads **to the end** and requires a **single** JSON value. Trailing junk is an
  error instead of being silently ignored.
- Reading to EOF also means the body is fully drained, which is exactly what connection
  reuse wants (chapter 6).

```go
err = json.UnmarshalRead(strings.NewReader(`{"id": 3} oops`), &got)
fmt.Println(err)
// jsontext: invalid character 'o' after top-level value after offset 10
```

For a stream of many values (like a log of events), v2 has `jsontext.Decoder`, but API
responses are almost always one value.

## MarshalWrite: encode straight to a writer

`json.MarshalWrite(w, v)` encodes `v` into any `io.Writer`: a file, `os.Stdout`, or a
`bytes.Buffer` that becomes a request body (chapter 4). Unlike v1's `Encoder.Encode`,
it **doesn't** add a trailing newline.

```go
var buf bytes.Buffer
if err := json.MarshalWrite(&buf, newIssue); err != nil {
	return err
}
// buf now holds the JSON, ready to send
```

## omitzero: leave out zero values

Say you're sending a partial update and want to leave out fields you didn't set. v1's
`omitempty` was confusing here: it omitted `0` and `false`, but never omitted a zero
`time.Time` struct, because a struct is never "empty".

`omitzero` has a simple rule: **omit the field if it holds its type's zero value**. If
the type has an `IsZero() bool` method, that decides instead. `time.Time` has one, so a
zero time is left out.

```go
package main

import (
	"encoding/json/v2"
	"fmt"
	"os"
	"strings"
	"time"
)

type Issue struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	Points   int       `json:"points,omitzero"`
	Due      time.Time `json:"due,omitzero"`
	Assignee *string   `json:"assignee,omitzero"`
	Labels   []string  `json:"labels,omitempty"`
}

func main() {
	a := Issue{ID: 1, Title: "Set up CI"}
	b := Issue{ID: 2, Title: "Ship v2", Points: 5,
		Due: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), Labels: []string{"release"}}

	for _, iss := range []Issue{a, b} {
		if err := json.MarshalWrite(os.Stdout, iss); err != nil {
			fmt.Println(err)
		}
		fmt.Println()
	}

	var got Issue
	err := json.UnmarshalRead(strings.NewReader(`{"id": 3, "title": "Fix login"}`), &got)
	fmt.Println(got.ID, got.Title, err)
}
```

Output:

```
{"id":1,"title":"Set up CI"}
{"id":2,"title":"Ship v2","points":5,"due":"2026-10-01T00:00:00Z","labels":["release"]}
3 Fix login <nil>
```

Issue 1 has no points, no due date, no assignee and no labels, so all four vanish.

### omitzero or omitempty?

| Field type | Use | Omitted when |
| --- | --- | --- |
| numbers, bools, pointers, structs like `time.Time` | `omitzero` | it's the zero value (or `IsZero()` says so) |
| strings, slices, maps | either | `omitempty`: empty. `omitzero`: `""`, or a **nil** slice or map |

The one subtle difference is `[]string{}`: an empty but non-nil slice is omitted by
`omitempty` and kept (as `[]`) by `omitzero`.

### The zero-value trap

`omitzero` can't tell "not set" from "set to zero". If `trackr` wants to set an issue's
points **to** 0, `Points int` with `omitzero` would drop the field and the server would
never hear about it. Use a pointer when zero is a meaningful value:

```go
type IssueUpdate struct {
	Points *int `json:"points,omitzero"` // nil: leave alone. new(0): set to zero.
}
```

`new(0)` (Go 1.26+) gives you a `*int` pointing at 0, which isn't the zero value of a
pointer, so it's sent.

`omitzero` works in v1's `encoding/json` too (since Go 1.24), so you can use it
either way.

## Your turn

`trackr edit 42 --points 0 --clear-labels` sends a PATCH whose body must contain
**only** the fields the user changed. Make `IssueUpdate` encode that way:

1. Add struct tags with the keys `title`, `points`, `due` and `labels`, choosing
   options so that:
   - a nil `Title` or `Points` is left out, but `new("")` and `new(0)` are sent;
   - a zero `Due` is left out;
   - a nil `Labels` is left out, but an empty, non-nil `[]string{}` is sent as `[]`
     (that's how `trackr` says "remove every label").
2. Complete `writeUpdate` so it writes `u` to `w` as JSON with `json.MarshalWrite`.
   Import `"encoding/json/v2"` yourself.

Careful: in v2, `omitempty` leaves out anything that encodes as `""` or `[]`, so it
would drop `new("")` and `[]string{}`. Check the table above to pick the right option.
