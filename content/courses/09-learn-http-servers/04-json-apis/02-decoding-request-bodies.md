---
title: Decoding Request Bodies
quiz:
  - question: |
      With `encoding/json/v2` and the call below, which request body decodes **without**
      an error? `params` has a single field, `Body string`, tagged `json:"body"`.

      ```go
      json.UnmarshalRead(r.Body, &params, json.RejectUnknownMembers(true))
      ```
    options:
      - text: '`{"body":"hi","mood":"happy"}`'
      - text: '`{"body":"hi"} {"body":"again"}`'
      - text: '`{"body":"hi","body":"bye"}`'
      - text: '`{"body":"hi"}` followed by a newline'
        correct: true
    explanation: |
      Trailing whitespace is fine. The others fail: `mood` is an unknown member (rejected
      because of the option), a second JSON value after the first is an error in v2, and
      v2 rejects duplicate member names by default.
  - question: What does `http.MaxBytesReader` do when a client sends more than the limit?
    options:
      - text: It silently truncates the body to the limit
      - text: Reads past the limit fail with an `*http.MaxBytesError`, and the server is told to close the connection afterwards
        correct: true
      - text: It immediately sends a `413` response for you
      - text: It panics
    explanation: |
      It's a reader wrapper. Your decode call gets an error it can recognise with
      `errors.AsType[*http.MaxBytesError](err)`, and *you* decide the response, usually
      `413 Request Entity Too Large`.
  - question: |
      In json/v2, a client sends `{"Body":"hi"}` to a struct field tagged `json:"body"`.
      What happens by default?
    options:
      - text: The field gets `"hi"`, because matching ignores case
      - text: The field stays empty; `Body` doesn't match `body` because v2 matches names case-sensitively
        correct: true
      - text: Unmarshal returns a syntax error
      - text: It panics
    explanation: |
      v1 matched names case-insensitively, and v2 is case-sensitive by default. With
      `RejectUnknownMembers(true)` you'd get an "unknown object member" error, which tells
      the client exactly what's wrong.
---

Responding was the easy half. Now a mouse wants to post a squeak:

```
POST /api/squeaks
Content-Type: application/json

{"body": "the cheese is a lie"}
```

The body is a stream of bytes from a stranger on the internet. It might be huge, it
might be broken JSON, it might have fields you've never heard of. A good decode step
defends against all of it.

## Decoding with json/v2

`r.Body` is an `io.ReadCloser`, and `json.UnmarshalRead` decodes straight from a reader:

```go
type createSqueakParams struct {
	Body string `json:"body"`
}

var params createSqueakParams
err := json.UnmarshalRead(r.Body, &params)
```

The server closes `r.Body` for you after the handler returns, so no `defer` is needed.

Compared with the old `encoding/json`, v2 is strict out of the box:

- **Names match case-sensitively.** `{"BODY": ...}` does *not* fill `Body`.
- **Duplicate names are errors.** `{"body":"a","body":"b"}` is rejected, rather than
  "last one wins". Attackers have abused that ambiguity between parsers.
- **Trailing data is an error.** `UnmarshalRead` reads to EOF and fails if another value
  follows the first.
- **Invalid UTF-8 is an error.**

## Rejecting unknown fields

By default, members that don't match any struct field are silently ignored. For an API
that's often a trap: a client sends `{"bdy": "hi"}`, gets a confusing "body is required",
and can't tell why. Turn on `RejectUnknownMembers`:

```go
err := json.UnmarshalRead(r.Body, &params, json.RejectUnknownMembers(true))
// json: cannot unmarshal JSON string into Go main.createSqueakParams:
//   unknown object member name "bdy"
```

In v1 the equivalent is a `Decoder` with `DisallowUnknownFields()`, and you'd also have
to check for trailing data yourself:

```go
dec := json.NewDecoder(r.Body) // encoding/json (v1)
dec.DisallowUnknownFields()
err := dec.Decode(&params)
```

Rejecting unknown fields makes an API stricter to *evolve*: an older server will refuse
requests from newer clients that send extra fields. Many public APIs stay lenient for
that reason. For Squeak, where you control the clients, strict is the better default.

## Limiting the size

Without a limit, a client can send a 10 GB body and your decoder will happily try to
read it. `http.MaxBytesReader` caps it:

```go
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB
```

Once more than the limit is read, reads fail with an `*http.MaxBytesError`. (Passing
`w` lets it tell the server to close the connection instead of reading the rest.)

## A decode helper

Put it together and turn each failure into the right status code:

```go
package main

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

type createSqueakParams struct {
	Body string `json:"body"`
}

// decodeJSON reads r's body into dst. It returns the HTTP status to use on failure.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) (int, error) {
	r.Body = http.MaxBytesReader(w, r.Body, 64) // tiny, to show the limit
	err := json.UnmarshalRead(r.Body, dst, json.RejectUnknownMembers(true))
	if err == nil {
		return 0, nil
	}
	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		return http.StatusRequestEntityTooLarge, errors.New("request body too large")
	}
	return http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err)
}

func main() {
	for _, body := range []string{
		`{"body":"the cheese is a lie"}`,
		`{"body":"hi","mood":"sneaky"}`,
		`{"body":`,
		`{"body":42}`,
		`{"body":"` + strings.Repeat("e", 100) + `"}`,
	} {
		req := httptest.NewRequest("POST", "/api/squeaks", strings.NewReader(body))
		var params createSqueakParams
		code, err := decodeJSON(httptest.NewRecorder(), req, &params)
		if err != nil {
			fmt.Println(code, err)
			continue
		}
		fmt.Printf("ok: %q\n", params.Body)
	}
}
```

```
ok: "the cheese is a lie"
400 invalid JSON: json: cannot unmarshal JSON string into Go main.createSqueakParams: unknown object member name "mood"
400 invalid JSON: jsontext: unexpected EOF within "/body" after offset 8
400 invalid JSON: json: cannot unmarshal JSON number into Go string within "/body"
413 request body too large
```

Notice how v2's errors point at the problem with a JSON Pointer such as `"/body"`.
Your run may say `unable to unmarshal` where this says `cannot unmarshal`. That's
deliberate: json/v2 randomly picks one of the two phrasings per process, so that no one
can build code that depends on the exact error text.

## Should you check Content-Type?

A strict API can reject requests whose `Content-Type` isn't `application/json` with
`415 Unsupported Media Type`. It's a good habit (it also blocks some cross-site form
tricks), but many clients forget the header, so decide deliberately. Squeak keeps it
simple and doesn't check.

## Don't echo raw errors forever

Sending the decoder's error text back is handy while developing. In a public API you
may prefer a friendlier message, and you should never include anything *internal* in
an error response. More on that in the error-responses lesson.

The body is decoded, but that doesn't mean it's *valid*. `{"body": ""}` decoded
perfectly. That's the next lesson.
