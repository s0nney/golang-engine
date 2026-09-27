---
title: Struct Tags
quiz:
  - question: |
      What does this print?

      ```go
      type Issue struct {
          ID     int      `json:"id"`
          Labels []string `json:"labels,omitempty"`
          secret string
          Note   string   `json:"-"`
      }

      data, _ := json.Marshal(Issue{ID: 1, secret: "s", Note: "n"})
      fmt.Println(string(data))
      ```
    options:
      - text: '`{"id":1,"labels":null}`'
      - text: '`{"id":1,"secret":"s","Note":"n"}`'
      - text: '`{"id":1}`'
        correct: true
      - text: '`{"id":1,"labels":[]}`'
    explanation: |
      `omitempty` drops the nil `Labels`. Unexported fields like `secret` are invisible
      to `encoding/json`, and the `"-"` tag hides `Note`. Only `id` is left.
  - question: 'A field is declared as `Title string` with **no tag**. What JSON key does `json.Marshal` use for it?'
    options:
      - text: '`"title"`'
      - text: '`"Title"`'
        correct: true
      - text: The field is skipped
    explanation: |
      Without a tag, the Go field name is used exactly as written. APIs usually want
      `snake_case` or `camelCase`, which is why you add tags.
---

`encoding/json` maps JSON objects onto Go structs. **Struct tags** tell it which JSON
key goes with which field.

## The Trackr Issue type

```go
type Issue struct {
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	State    string   `json:"state"`
	Labels   []string `json:"labels,omitempty"`
	Assignee *string  `json:"assignee"`
	Points   int      `json:"points,omitzero"`
	Secret   string   `json:"-"`
	Project  string
	internal string
}
```

A tag is a raw string after the field type. The `json:"..."` part holds the JSON key,
optionally followed by comma-separated options.

## Encoding: Go to JSON

`json.Marshal` turns a value into JSON bytes:

```go
iss := Issue{ID: 42, Title: "Login button does nothing", State: "open",
	Secret: "x", internal: "y", Project: "apollo"}
data, err := json.Marshal(iss)
if err != nil {
	return err
}
fmt.Println(string(data))
```

Output:

```
{"id":42,"title":"Login button does nothing","state":"open","assignee":null,"Project":"apollo"}
```

Go through it field by field:

- `id`, `title`, `state` use the names from their tags.
- `labels` is missing: it's nil, and `omitempty` leaves out empty slices, maps and strings.
- `assignee` is `null` because the pointer is nil. A pointer lets you tell "no
  assignee" (`null`) apart from an empty name.
- `points` is missing: `omitzero` leaves out a field holding its type's zero value.
- `Secret` is gone thanks to the `json:"-"` tag.
- `Project` has no tag, so its key is the Go name, capital P and all.
- `internal` is unexported (lowercase), so `encoding/json` can't see it at all.

`json.MarshalIndent(v, "", "  ")` does the same with line breaks and indentation, which
is nice for debugging output.

## Decoding: JSON to Go

`json.Unmarshal` goes the other way. Pass a **pointer** so it can fill in your value:

```go
var got Issue
err := json.Unmarshal([]byte(`{"id": 9, "title": "Crash on start",
	"state": "open", "assignee": "ana", "priority": "high"}`), &got)
fmt.Println(err, got.ID, got.Title, *got.Assignee)
// <nil> 9 Crash on start ana
```

Notice what's forgiving here:

- **Unknown keys are ignored.** `priority` has no field, and that's fine. APIs add
  fields all the time, and old clients keep working.
- **Missing keys leave the zero value.** Nothing said `points`, so `got.Points` is `0`.

And what isn't: a value of the wrong type is an error.

```
json: cannot unmarshal string into Go struct field Issue.id of type int
```

## Gotchas

- **Exported fields only.** A lowercase field is silently skipped in both directions.
  If a field "never gets filled", check its first letter.
- **Tag typos are silent** in `encoding/json`. `json:"title "` (with a space) or
  `json:title` (no quotes) just don't work. `go vet` catches some of these, so run it.
- **Pass a pointer to Unmarshal.** Passing `got` instead of `&got` returns an error,
  because Unmarshal can't change your copy.
- **Only decode what you need.** You don't have to model every field the API sends.
  A `trackr list` command might only need `ID`, `Title` and `State`.

## Further reading

- Go by Example, "JSON": https://gobyexample.com/json
