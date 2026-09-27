---
title: Responding with JSON
quiz:
  - question: |
      Using `encoding/json/v2`, what does this print?

      ```go
      type Squeak struct {
      	Body string   `json:"body"`
      	Tags []string `json:"tags"`
      }
      b, _ := json.Marshal(Squeak{Body: "hi"})
      fmt.Println(string(b))
      ```
    options:
      - text: '`{"body":"hi","tags":null}`'
      - text: '`{"body":"hi","tags":[]}`'
        correct: true
      - text: '`{"body":"hi"}`'
      - text: '`{"Body":"hi","Tags":[]}`'
    explanation: |
      json/v2 encodes a nil slice as an empty array `[]` (and a nil map as `{}`). The old
      `encoding/json` produced `null`, which forced clients to handle two "empty" cases.
  - question: Why does `respondWithJSON` marshal into a `[]byte` *before* calling `w.WriteHeader`?
    options:
      - text: '`WriteHeader` must be called after `Write`'
      - text: If marshaling fails, it can still send a clean 500, because nothing has been written yet
        correct: true
      - text: Marshaling to bytes is required by json/v2
      - text: It makes the response compress better
    explanation: |
      Once the status is written it can't be changed. Encoding first means an encoding
      failure (a channel or function in the value, say) becomes a proper 500 rather than
      a 200 with half a JSON document.
---

Squeak is a JSON API: every response body is JSON, from a single squeak to an error
message. You'll write the same four steps in every handler, so it's time for a helper.

## Which JSON package?

Squeak uses **`encoding/json/v2`**, which you met in
[Meet encoding/json/v2](/courses/learn-http-clients/json/json-v2) in the HTTP clients
course. The package name is still `json`:

```go
import "encoding/json/v2"
```

On the server side, three of its defaults matter most. Nil slices encode as `[]` (and nil
maps as `{}`) instead of `null`, `<` and `>` aren't HTML-escaped, and `omitzero` leaves out
zero values such as an unset `time.Time`.

## The response helper

```go
func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("encoding response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}
```

The **order** is the whole point:

1. **Marshal first.** If encoding fails, nothing has been sent, so you can still answer
   500. (A failed `json.Marshal` can return partial output, so never write `data`
   when `err != nil`.)
2. **Set headers** before the status. Headers set after `WriteHeader` are ignored.
3. **Write the status** with `WriteHeader(code)`.
4. **Write the body.**

Could you stream straight to `w` with `json.MarshalWrite(w, payload)`? Yes, and for
huge responses you'd want to. But then an error halfway through leaves the client with
a 200 and broken JSON. For API-sized payloads, marshal-then-write is the safer default.

## Shaping the output with struct tags

Response types are ordinary structs with `json` tags. Keep them separate from your
storage types when they differ, so you never leak a field by accident:

```go
package main

import (
	"encoding/json/v2"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

type squeakResponse struct {
	ID        int       `json:"id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	EditedAt  time.Time `json:"edited_at,omitzero"`
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("encoding response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func main() {
	rec := httptest.NewRecorder()
	respondWithJSON(rec, http.StatusOK, squeakResponse{
		ID:        1,
		Author:    "pip",
		Body:      "brie > cheddar & that's final",
		CreatedAt: time.Date(2026, 9, 1, 9, 30, 0, 0, time.UTC),
	})
	fmt.Println(rec.Code, rec.Header().Get("Content-Type"))
	fmt.Println(rec.Body.String())

	rec = httptest.NewRecorder()
	respondWithJSON(rec, http.StatusOK, map[string]any{"oops": make(chan int)})
	fmt.Println(rec.Code, rec.Body.Len())
}
```

```
200 application/json
{"id":1,"author":"pip","body":"brie > cheddar & that's final","tags":[],"created_at":"2026-09-01T09:30:00Z"}
2026/09/27 10:00:00 encoding response: json: unable to marshal from Go chan int within "/oops"
500 0
```

(The log line goes to stderr with the current date and time, so yours will differ. The
error text may also say `cannot marshal` instead of `unable to marshal`: json/v2 picks
one of the two phrasings at random each time a program starts, on purpose, so nobody
writes code that depends on the exact wording. Check errors with `errors.Is` and
`errors.AsType`, never by comparing strings.)

Things to notice:

- `time.Time` encodes as an RFC 3339 string, which every language can parse.
- `EditedAt` was zero, so `omitzero` left it out entirely.
- `Tags` was nil and still came out as `[]`.
- The channel couldn't be encoded, and the client got a clean 500 with no body.

## One more header: Content-Type

Always set `Content-Type: application/json`. Without it, Go *sniffs* the body and would
probably guess `text/plain; charset=utf-8`. Clients (and browsers, which may refuse to
render some types) rely on it.

## A list endpoint

With the helper, a handler is mostly about getting the data:

```go
func (cfg *apiConfig) handleListSqueaks(w http.ResponseWriter, r *http.Request) {
	// The store arrives in chapter 5. uuid.Nil() means "squeaks by every author".
	squeaks, err := cfg.squeaks.ListSqueaks(r.Context(), uuid.Nil())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, squeaks)
}
```

Next up: the other direction, reading JSON that clients send you.
