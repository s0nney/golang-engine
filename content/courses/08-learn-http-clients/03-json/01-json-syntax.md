---
title: JSON Syntax
quiz:
  - question: Which of these is **valid** JSON?
    options:
      - text: "`{'id': 42}`"
      - text: '`{"id": 42,}`'
      - text: '`{"id": 42, "labels": ["bug", "ui"], "assignee": null}`'
        correct: true
      - text: '`{id: 42}`'
    explanation: |
      JSON is strict: keys must be strings in **double** quotes, and trailing commas
      aren't allowed. `null` is a real JSON value, so the third option is fine.
  - question: 'In JSON, what''s the difference between `"42"` and `42`?'
    options:
      - text: Nothing, JSON treats them the same
      - text: '`"42"` is a string and `42` is a number, and decoding one into a Go `int` field only works for `42`'
        correct: true
      - text: '`42` is invalid: numbers must be quoted'
    explanation: |
      Quotes make a string. Go's JSON decoder won't silently convert a JSON string
      into an `int`: it returns an error instead.
---

Almost every web API speaks **JSON** (JavaScript Object Notation). When `trackr` asks
for issue 42, the server answers with something like this:

```json
{
  "id": 42,
  "title": "Login button does nothing",
  "state": "open",
  "points": 3.5,
  "labels": ["bug", "ui"],
  "assignee": null,
  "project": {
    "slug": "apollo",
    "archived": false
  }
}
```

It's plain text, so humans can read it, and it's simple enough that every language
can parse it.

## Six kinds of value

| JSON | Example | Usual Go type |
| --- | --- | --- |
| string | `"open"` | `string` |
| number | `42`, `3.5`, `-1e3` | `int`, `float64`, ... |
| boolean | `true`, `false` | `bool` |
| null | `null` | a nil pointer, slice, map or interface |
| array | `["bug", "ui"]` | a slice, like `[]string` |
| object | `{"slug": "apollo"}` | a struct, or `map[string]T` |

Arrays and objects nest as deeply as you like. A list of issues is an array of
objects:

```json
[
  {"id": 1, "title": "Set up CI", "state": "closed"},
  {"id": 2, "title": "Write README", "state": "open"}
]
```

## The rules that bite

JSON looks like JavaScript but is much stricter:

- Object keys are **always strings in double quotes**: `{"id": 1}`, never `{id: 1}`
  or `{'id': 1}`.
- Strings use double quotes only. Special characters are escaped with a backslash:
  `"say \"hi\"\n"`.
- **No trailing commas.** `[1, 2, 3,]` is an error.
- **No comments.** Not `//`, not `/* */`.
- Numbers have no leading zeros (`07` is invalid), no `NaN`, no `Infinity` and no hex.
- `true`, `false` and `null` are lowercase.
- Whitespace between tokens doesn't matter.

## Types matter

`42` and `"42"` are different values: a number and a string. APIs are usually
consistent, but when one isn't, Go tells you rather than guessing:

```
json: cannot unmarshal string into Go struct field Issue.id of type int
```

That strictness is a feature. A silent conversion would hide a bug in the server or
in your struct.

## Numbers are just numbers

JSON has a single number type, with no separate integer and float. Go picks the
representation from your struct field: `int` for IDs, `float64` for story points. If
you decode into `any` instead, every number becomes a `float64`, which quietly loses
precision on very large IDs. Prefer concrete struct types.

## Checking JSON from Go

`json.Valid` reports whether some bytes are well-formed JSON:

```go
fmt.Println(json.Valid([]byte(`{"id": 42}`)))  // true
fmt.Println(json.Valid([]byte(`{"id": 42,}`))) // false
fmt.Println(json.Valid([]byte(`{id: 42}`)))    // false
```

You'll rarely call it directly, since decoding checks syntax as it goes. It's handy in
tests, though.

In the next lesson you'll map JSON objects onto Go structs with **struct tags**.

## Further reading

- JSON specification (one page, with railroad diagrams): https://www.json.org/
