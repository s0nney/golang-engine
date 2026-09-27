---
title: Error Responses and Status Codes
quiz:
  - question: A mouse asks to delete a squeak that belongs to someone else. They *are* logged in. Which status fits best?
    options:
      - text: '`401 Unauthorized`'
      - text: '`403 Forbidden`'
        correct: true
      - text: '`404 Not Found`'
      - text: '`400 Bad Request`'
    explanation: |
      `401` means "I don't know who you are" (missing or bad credentials). `403` means "I
      know who you are, and you're not allowed". Some APIs deliberately answer `404` to hide
      that the resource exists, but the honest answer is `403`.
  - question: |
      The store fails with `pq: connection refused to 10.0.3.7:5432`. What should the
      client see?
    options:
      - text: '`500` with `{"error":"pq: connection refused to 10.0.3.7:5432"}` so they can report it'
      - text: '`500` with a generic `{"error":"something went wrong"}`; the details go to your logs'
        correct: true
      - text: '`400`, because the request failed'
      - text: '`200` with an empty list, so the app doesn''t crash'
    explanation: |
      Internal error messages leak your infrastructure (hosts, drivers, table names) to
      anyone who can trigger them. Log the details with the request ID, and send the client a
      generic message. They can quote the `X-Request-ID` when they report it.
exercise:
  starter: |
    package main

    import (
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"log"
    	"net/http"
    	"net/http/httptest"
    )

    var (
    	ErrNotFound   = errors.New("not found")
    	ErrForbidden  = errors.New("forbidden")
    	ErrEmailTaken = errors.New("email already registered")
    )

    // ValidationError reports a client mistake in one field of a request.
    type ValidationError struct {
    	Field   string
    	Problem string
    }

    func (e *ValidationError) Error() string { return e.Field + " " + e.Problem }

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

    type errorResponse struct {
    	Error string `json:"error"`
    }

    // respondWithError responds with {"error": msg} and the given status.
    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	// ?
    }

    // handleError picks the status and message for err. See the lesson.
    func handleError(w http.ResponseWriter, err error) {
    	// ?
    	respondWithError(w, http.StatusInternalServerError, err.Error())
    }

    func main() {
    	log.SetFlags(0)
    	for _, err := range []error{
    		fmt.Errorf("getting squeak 7: %w", ErrNotFound),
    		fmt.Errorf("deleting squeak 8: %w", ErrForbidden),
    		fmt.Errorf("creating user: %w", ErrEmailTaken),
    		fmt.Errorf("decoding: %w", &ValidationError{Field: "body", Problem: "is required"}),
    		fmt.Errorf("decoding: %w", &http.MaxBytesError{Limit: 1024}),
    		errors.New("dial tcp 10.0.3.7:5432: connection refused"),
    	} {
    		rec := httptest.NewRecorder()
    		handleError(rec, err)
    		fmt.Printf("%d %s\n", rec.Code, rec.Body.String())
    	}
    }
  solution: |
    package main

    import (
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"log"
    	"net/http"
    	"net/http/httptest"
    )

    var (
    	ErrNotFound   = errors.New("not found")
    	ErrForbidden  = errors.New("forbidden")
    	ErrEmailTaken = errors.New("email already registered")
    )

    // ValidationError reports a client mistake in one field of a request.
    type ValidationError struct {
    	Field   string
    	Problem string
    }

    func (e *ValidationError) Error() string { return e.Field + " " + e.Problem }

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

    type errorResponse struct {
    	Error string `json:"error"`
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	respondWithJSON(w, code, errorResponse{Error: msg})
    }

    func handleError(w http.ResponseWriter, err error) {
    	if ve, ok := errors.AsType[*ValidationError](err); ok {
    		respondWithError(w, http.StatusBadRequest, ve.Error())
    		return
    	}
    	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
    		respondWithError(w, http.StatusRequestEntityTooLarge, "request body too large")
    		return
    	}
    	switch {
    	case errors.Is(err, ErrNotFound):
    		respondWithError(w, http.StatusNotFound, "not found")
    	case errors.Is(err, ErrForbidden):
    		respondWithError(w, http.StatusForbidden, "forbidden")
    	case errors.Is(err, ErrEmailTaken):
    		respondWithError(w, http.StatusConflict, "email already registered")
    	default:
    		log.Printf("internal error: %v", err)
    		respondWithError(w, http.StatusInternalServerError, "something went wrong")
    	}
    }

    func main() {
    	log.SetFlags(0)
    	for _, err := range []error{
    		fmt.Errorf("getting squeak 7: %w", ErrNotFound),
    		fmt.Errorf("deleting squeak 8: %w", ErrForbidden),
    		fmt.Errorf("creating user: %w", ErrEmailTaken),
    		fmt.Errorf("decoding: %w", &ValidationError{Field: "body", Problem: "is required"}),
    		fmt.Errorf("decoding: %w", &http.MaxBytesError{Limit: 1024}),
    		errors.New("dial tcp 10.0.3.7:5432: connection refused"),
    	} {
    		rec := httptest.NewRecorder()
    		handleError(rec, err)
    		fmt.Printf("%d %s\n", rec.Code, rec.Body.String())
    	}
    }
  tests: |
    package main

    import (
    	"bytes"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"os"
    	"strings"
    	"testing"
    )

    func TestHandleError(t *testing.T) {
    	for _, tt := range []struct {
    		err      error
    		wantCode int
    		wantMsg  string
    	}{
    		{ErrNotFound, 404, "not found"},
    		{fmt.Errorf("getting squeak: %w", ErrNotFound), 404, "not found"},
    		{fmt.Errorf("deleting squeak: %w", ErrForbidden), 403, "forbidden"},
    		{fmt.Errorf("creating user: %w", ErrEmailTaken), 409, "email already registered"},
    		{&ValidationError{Field: "email", Problem: "is not an email address"}, 400, "email is not an email address"},
    		{fmt.Errorf("decoding: %w", &ValidationError{Field: "body", Problem: "is too long"}), 400, "body is too long"},
    		{fmt.Errorf("decoding: %w", &http.MaxBytesError{Limit: 1024}), 413, "request body too large"},
    		{errors.New("disk on fire"), 500, "something went wrong"},
    	} {
    		rec := httptest.NewRecorder()
    		handleError(rec, tt.err)
    		if rec.Code != tt.wantCode {
    			t.Errorf("handleError(%q): status = %d, want %d", tt.err, rec.Code, tt.wantCode)
    			continue
    		}
    		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
    			t.Errorf("handleError(%q): Content-Type = %q, want application/json", tt.err, ct)
    		}
    		var body struct {
    			Error string `json:"error"`
    		}
    		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
    			t.Errorf("handleError(%q): body %q isn't JSON like {\"error\":\"...\"}: %v", tt.err, rec.Body.String(), err)
    			continue
    		}
    		if body.Error != tt.wantMsg {
    			t.Errorf("handleError(%q): error message = %q, want %q", tt.err, body.Error, tt.wantMsg)
    		}
    	}
    }

    func TestInternalErrorsAreLoggedNotLeaked(t *testing.T) {
    	var logs bytes.Buffer
    	log.SetOutput(&logs)
    	t.Cleanup(func() { log.SetOutput(os.Stderr) })

    	rec := httptest.NewRecorder()
    	handleError(rec, errors.New("dial tcp 10.0.3.7:5432: connection refused"))
    	if strings.Contains(rec.Body.String(), "10.0.3.7") {
    		t.Errorf("the 500 response %q leaks the internal error; send a generic message", rec.Body.String())
    	}
    	if !strings.Contains(logs.String(), "10.0.3.7:5432: connection refused") {
    		t.Errorf("log output = %q, want it to contain the real error so you can debug it", logs.String())
    	}

    	logs.Reset()
    	handleError(httptest.NewRecorder(), ErrNotFound)
    	if logs.Len() != 0 {
    		t.Errorf("a 404 was logged as %q; only log the unexpected 500s", logs.String())
    	}
    }
---

Clients need to handle failures, and that's only pleasant when every error from your API
looks the same. Squeak's rule: **every error is JSON with an `error` field**, sent with
an accurate status code.

```json
{"error": "squeak is too long (max 140 characters)"}
```

## The error helper

It builds on `respondWithJSON`:

```go
type errorResponse struct {
	Error string `json:"error"`
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, errorResponse{Error: msg})
}
```

Now every handler's failure path is one line and one `return`.

### What about http.Error?

`http.Error(w, msg, code)` writes a **plain text** body. It's what the mux uses for its
own 404 and 405 responses. For a JSON API, clients then get plain text for routing
errors and JSON for everything else. You can smooth that over by registering your own
catch-all, such as `mux.HandleFunc("/", notFoundJSON)`, but remember from the routing
chapter that 405s would then need care too. Many APIs just live with plain-text routing
errors.

## Status codes you'll actually use

You met status codes as a client. Here's the server-side view: which one to *send*.

| Code | Name | Squeak uses it when |
|---|---|---|
| 200 | OK | a read or update succeeds |
| 201 | Created | a squeak or user was created (add a `Location` header) |
| 204 | No Content | a delete succeeds; no body |
| 400 | Bad Request | the body is malformed or breaks a rule |
| 401 | Unauthorized | the token is missing, expired or invalid |
| 403 | Forbidden | logged in, but not allowed (someone else's squeak) |
| 404 | Not Found | no squeak with that ID |
| 405 | Method Not Allowed | the mux sends this for you |
| 409 | Conflict | the email is already registered |
| 413 | Content Too Large (`StatusRequestEntityTooLarge` in Go) | the body is over the size limit |
| 429 | Too Many Requests | the client is rate limited |
| 500 | Internal Server Error | *your* bug or a failed dependency |

The big split: **4xx means the client should change something** and **5xx means the
server failed**. A client may retry a 5xx later. Retrying a 4xx unchanged is pointless.
Monitoring usually alerts on 5xx rates, so sending 500 for a validation error pages
someone at 3 a.m. for nothing.

## Don't leak internals

For 5xx errors, the message the client sees should be generic. The real error goes to
your logs:

```go
squeak, err := cfg.squeaks.CreateSqueak(r.Context(), authorID, body)
if err != nil {
	log.Printf("request %s: creating squeak: %v", requestIDFrom(r.Context()), err)
	respondWithError(w, http.StatusInternalServerError, "couldn't create squeak")
	return
}
```

Database hostnames, file paths and stack traces help attackers and confuse users. 4xx
messages are different: they're *about the client's input*, so be specific. "email is
required" helps, "bad request" doesn't.

## Mapping errors to statuses

As Squeak grows, lower layers return sentinel errors and the handler maps them to status
codes in one place:

```go
package main

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	errNotFound  = errors.New("not found")
	errForbidden = errors.New("forbidden")
)

func statusFor(err error) (int, string) {
	switch {
	case errors.Is(err, errNotFound):
		return http.StatusNotFound, "squeak not found"
	case errors.Is(err, errForbidden):
		return http.StatusForbidden, "you can only delete your own squeaks"
	default:
		return http.StatusInternalServerError, "something went wrong"
	}
}

func main() {
	for _, err := range []error{
		fmt.Errorf("deleting squeak 7: %w", errNotFound),
		fmt.Errorf("deleting squeak 8: %w", errForbidden),
		errors.New("disk on fire"),
	} {
		code, msg := statusFor(err)
		fmt.Printf("%d %s: %q\n", code, http.StatusText(code), msg)
	}
}
```

```
404 Not Found: "squeak not found"
403 Forbidden: "you can only delete your own squeaks"
500 Internal Server Error: "something went wrong"
```

`errors.Is` sees through the `%w` wrapping, so lower layers can add context freely.
The unknown error gets the generic 500.

## Location on 201

When you create something, `201 Created` plus a `Location` header pointing at the new
resource is the textbook response:

```go
w.Header().Set("Location", fmt.Sprintf("/api/squeaks/%v", squeak.ID))
respondWithJSON(w, http.StatusCreated, squeak)
```

Set the header *before* `respondWithJSON`, because it calls `WriteHeader`. You'll do
exactly this in the next lesson.

## Errors that carry details

Sentinels answer "which kind of error?". Sometimes the error also carries *data*, like
which field was invalid. That's a job for a custom error type, and for
`errors.AsType`, which digs through `%w` wrapping to find one:

```go
type ValidationError struct {
	Field   string
	Problem string
}

func (e *ValidationError) Error() string { return e.Field + " " + e.Problem }

if ve, ok := errors.AsType[*ValidationError](err); ok {
	respondWithError(w, http.StatusBadRequest, ve.Error()) // "body is required"
}
```

## Your task

Squeak's handlers are about to call one helper for every error they meet. Complete:

1. **`respondWithError`**: send `{"error": msg}` with the given status, built on
   `respondWithJSON`.
2. **`handleError`**: look at `err` (which may be wrapped) and respond with:

| Error | Status | `error` message |
|---|---|---|
| a `*ValidationError` | 400 | the validation error's own text |
| a `*http.MaxBytesError` | 413 | `request body too large` |
| `ErrNotFound` | 404 | `not found` |
| `ErrForbidden` | 403 | `forbidden` |
| `ErrEmailTaken` | 409 | `email already registered` |
| anything else | 500 | `something went wrong` |

For the 500 case only, also log the real error with `log.Printf`, so you can debug it
without showing it to the client.
