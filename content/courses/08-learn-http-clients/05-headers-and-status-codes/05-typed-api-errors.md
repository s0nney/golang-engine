---
title: A Typed APIError
quiz:
  - question: |
      `getIssue` returns an error that wraps an `*APIError`:
      `fmt.Errorf("fetching issue: %w", apiErr)`. Which check finds it?
    options:
      - text: '`err.(*APIError)`'
      - text: '`apiErr, ok := errors.AsType[*APIError](err)`'
        correct: true
      - text: '`err == &APIError{}`'
    explanation: |
      A type assertion only looks at the outermost error. `errors.AsType` (Go 1.26)
      walks the whole `%w` chain and returns the first `*APIError` it finds.
  - question: Why does `checkResponse` fall back to `http.StatusText` instead of returning the decode error?
    options:
      - text: Because decode errors can't be wrapped
      - text: Error bodies aren't always JSON (a proxy might send HTML), and the status code is the important fact either way
        correct: true
      - text: '`json.UnmarshalRead` never fails on error responses'
    explanation: |
      When the server is failing, its body is the least reliable part of the response.
      The caller should still get an `*APIError` with the right status code, rather
      than a confusing JSON syntax error.
exercise:
  starter: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    	State string `json:"state"`
    }

    // APIError is returned for any response outside the 2xx range.
    type APIError struct {
    	StatusCode int    // e.g. 404
    	Message    string // from the body's "error" field, or http.StatusText(StatusCode)
    }

    // Error formats the error as "trackr API: <code> <message>",
    // e.g. "trackr API: 404 issue 999 not found".
    func (e *APIError) Error() string {
    	// ?
    	return ""
    }

    // checkResponse returns nil for a 2xx response. Otherwise it returns an
    // *APIError with the status code and the message from the JSON body
    // ({"error": "..."}), falling back to http.StatusText if there isn't one.
    func checkResponse(resp *http.Response) error {
    	// ?
    	return nil
    }

    // getIssue fetches one issue. It's finished: it relies on checkResponse.
    func getIssue(ctx context.Context, baseURL string, id int) (Issue, error) {
    	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/issues/%d", baseURL, id), nil)
    	if err != nil {
    		return Issue{}, err
    	}
    	resp, err := http.DefaultClient.Do(req)
    	if err != nil {
    		return Issue{}, err
    	}
    	defer resp.Body.Close()

    	if err := checkResponse(resp); err != nil {
    		return Issue{}, err
    	}
    	var iss Issue
    	if err := json.UnmarshalRead(resp.Body, &iss); err != nil {
    		return Issue{}, fmt.Errorf("decoding issue %d: %w", id, err)
    	}
    	return iss, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.Header().Set("Content-Type", "application/json")
    		if r.URL.Path != "/issues/42" {
    			w.WriteHeader(http.StatusNotFound)
    			fmt.Fprintf(w, `{"error": "no issue at %s"}`, r.URL.Path)
    			return
    		}
    		fmt.Fprint(w, `{"id": 42, "title": "Login broken", "state": "open"}`)
    	}))
    	defer srv.Close()

    	for _, id := range []int{42, 999} {
    		iss, err := getIssue(context.Background(), srv.URL, id)
    		if apiErr, ok := errors.AsType[*APIError](err); ok {
    			fmt.Printf("API error! status=%d message=%q\n  %v\n", apiErr.StatusCode, apiErr.Message, err)
    			continue
    		}
    		fmt.Printf("%+v %v\n", iss, err)
    	}
    }
  solution: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    	State string `json:"state"`
    }

    // APIError is returned for any response outside the 2xx range.
    type APIError struct {
    	StatusCode int    // e.g. 404
    	Message    string // from the body's "error" field, or http.StatusText(StatusCode)
    }

    func (e *APIError) Error() string {
    	return fmt.Sprintf("trackr API: %d %s", e.StatusCode, e.Message)
    }

    func checkResponse(resp *http.Response) error {
    	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
    		return nil
    	}
    	apiErr := &APIError{StatusCode: resp.StatusCode}

    	var body struct {
    		Error string `json:"error"`
    	}
    	if err := json.UnmarshalRead(resp.Body, &body); err == nil && body.Error != "" {
    		apiErr.Message = body.Error
    	} else {
    		apiErr.Message = http.StatusText(resp.StatusCode)
    	}
    	return apiErr
    }

    // getIssue fetches one issue. It's finished: it relies on checkResponse.
    func getIssue(ctx context.Context, baseURL string, id int) (Issue, error) {
    	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/issues/%d", baseURL, id), nil)
    	if err != nil {
    		return Issue{}, err
    	}
    	resp, err := http.DefaultClient.Do(req)
    	if err != nil {
    		return Issue{}, err
    	}
    	defer resp.Body.Close()

    	if err := checkResponse(resp); err != nil {
    		return Issue{}, err
    	}
    	var iss Issue
    	if err := json.UnmarshalRead(resp.Body, &iss); err != nil {
    		return Issue{}, fmt.Errorf("decoding issue %d: %w", id, err)
    	}
    	return iss, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.Header().Set("Content-Type", "application/json")
    		if r.URL.Path != "/issues/42" {
    			w.WriteHeader(http.StatusNotFound)
    			fmt.Fprintf(w, `{"error": "no issue at %s"}`, r.URL.Path)
    			return
    		}
    		fmt.Fprint(w, `{"id": 42, "title": "Login broken", "state": "open"}`)
    	}))
    	defer srv.Close()

    	for _, id := range []int{42, 999} {
    		iss, err := getIssue(context.Background(), srv.URL, id)
    		if apiErr, ok := errors.AsType[*APIError](err); ok {
    			fmt.Printf("API error! status=%d message=%q\n  %v\n", apiErr.StatusCode, apiErr.Message, err)
    			continue
    		}
    		fmt.Printf("%+v %v\n", iss, err)
    	}
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"testing"
    )

    func fake(status int, contentType, body string) *httptest.Server {
    	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.Header().Set("Content-Type", contentType)
    		w.WriteHeader(status)
    		fmt.Fprint(w, body)
    	}))
    }

    func TestSuccess(t *testing.T) {
    	for _, status := range []int{http.StatusOK, 203} {
    		srv := fake(status, "application/json", `{"id": 7, "title": "Dark mode", "state": "open"}`)
    		iss, err := getIssue(context.Background(), srv.URL, 7)
    		srv.Close()
    		if err != nil {
    			t.Errorf("status %d: getIssue returned error %v, want nil (every 2xx is a success)", status, err)
    			continue
    		}
    		if iss.ID != 7 || iss.Title != "Dark mode" {
    			t.Errorf("status %d: getIssue = %+v, want issue 7 \"Dark mode\"", status, iss)
    		}
    	}
    }

    func TestAPIErrorFromJSON(t *testing.T) {
    	tests := []struct {
    		status int
    		body   string
    		msg    string
    	}{
    		{http.StatusNotFound, `{"error": "issue 999 not found", "code": "not_found"}`, "issue 999 not found"},
    		{http.StatusUnauthorized, `{"error": "token expired"}`, "token expired"},
    		{http.StatusUnprocessableEntity, `{"error": "title is required"}`, "title is required"},
    	}
    	for _, tt := range tests {
    		srv := fake(tt.status, "application/json", tt.body)
    		_, err := getIssue(context.Background(), srv.URL, 999)
    		srv.Close()
    		if err == nil {
    			t.Errorf("status %d: getIssue returned nil error, want an *APIError", tt.status)
    			continue
    		}
    		apiErr, ok := errors.AsType[*APIError](err)
    		if !ok {
    			t.Errorf("status %d: error %v (type %T) isn't an *APIError", tt.status, err, err)
    			continue
    		}
    		if apiErr.StatusCode != tt.status || apiErr.Message != tt.msg {
    			t.Errorf("status %d: got APIError{StatusCode: %d, Message: %q}, want {%d, %q}",
    				tt.status, apiErr.StatusCode, apiErr.Message, tt.status, tt.msg)
    		}
    		want := fmt.Sprintf("trackr API: %d %s", tt.status, tt.msg)
    		if got := err.Error(); got != want {
    			t.Errorf("status %d: err.Error() = %q, want %q", tt.status, got, want)
    		}
    	}
    }

    func TestAPIErrorFallback(t *testing.T) {
    	tests := []struct {
    		status      int
    		contentType string
    		body        string
    	}{
    		{http.StatusInternalServerError, "text/html", `<html><h1>Oops</h1></html>`},
    		{http.StatusBadGateway, "text/plain", ``},
    		{http.StatusServiceUnavailable, "application/json", `{"retry": true}`},
    	}
    	for _, tt := range tests {
    		srv := fake(tt.status, tt.contentType, tt.body)
    		_, err := getIssue(context.Background(), srv.URL, 1)
    		srv.Close()
    		apiErr, ok := errors.AsType[*APIError](err)
    		if !ok {
    			t.Errorf("status %d with body %q: error %v (type %T) isn't an *APIError", tt.status, tt.body, err, err)
    			continue
    		}
    		if want := http.StatusText(tt.status); apiErr.StatusCode != tt.status || apiErr.Message != want {
    			t.Errorf("status %d with body %q: got APIError{%d, %q}, want {%d, %q} (fall back to http.StatusText)",
    				tt.status, tt.body, apiErr.StatusCode, apiErr.Message, tt.status, want)
    		}
    		if !strings.Contains(err.Error(), fmt.Sprint(tt.status)) {
    			t.Errorf("err.Error() = %q, want it to include the status code", err.Error())
    		}
    	}
    }
---

`fmt.Errorf("unexpected status %s", resp.Status)` gets the job done, but callers can't
do much with a string. `trackr` wants to react differently to different failures:

- **404**: print "issue #999 doesn't exist".
- **401**: print "your token expired, run `trackr login`".
- **429 or 503**: wait and retry (chapter 6).

For that, the caller needs the status code as **data**, not buried in a message.
Enter a custom error type.

## APIError

```go
// APIError is returned for any response outside the 2xx range.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("trackr API: %d %s", e.StatusCode, e.Message)
}
```

Anything with an `Error() string` method is an `error`. The pointer receiver means
`*APIError` is the error type, which is the usual choice for struct errors.

## One place that checks every response

Rather than checking status codes in every function, `trackr` has one helper:

```go
func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return nil
	}
	// ... build and return an *APIError
}
```

Every API call does the same thing after `Do`:

```go
resp, err := c.http.Do(req)
if err != nil {
	return Issue{}, err // network trouble: not an APIError
}
defer resp.Body.Close()
if err := checkResponse(resp); err != nil {
	return Issue{}, err // the server said no: an *APIError
}
```

That gives callers two clearly different kinds of failure: **transport errors** (no
response at all) and **API errors** (a response with a bad status).

## Inspecting it: errors.AsType

Errors get wrapped as they travel up (`fmt.Errorf("closing issue: %w", err)`), so use
`errors.AsType` (new in Go 1.26) to find the `*APIError` anywhere in the chain:

```go
iss, err := getIssue(ctx, baseURL, 999)
if apiErr, ok := errors.AsType[*APIError](err); ok {
	switch apiErr.StatusCode {
	case http.StatusNotFound:
		fmt.Println("no such issue")
	case http.StatusUnauthorized:
		fmt.Println("token expired: run `trackr login`")
	default:
		fmt.Println(apiErr.Message)
	}
	return
}
if err != nil {
	fmt.Println("couldn't reach Trackr:", err)
	return
}
```

`errors.AsType[T](err)` returns `(T, bool)`. It's the generic replacement for the older
`errors.As(err, &target)`, and it doesn't need a pre-declared target variable.

## Sentinel errors, too

Some callers just want "was it a 404?". Give `APIError` an `Is` method, and it will
match a sentinel with `errors.Is`:

```go
var ErrNotFound = errors.New("not found")

func (e *APIError) Is(target error) bool {
	return target == ErrNotFound && e.StatusCode == http.StatusNotFound
}

// elsewhere
if errors.Is(err, ErrNotFound) { ... }
```

`errors.Is` calls your `Is` method at each step of the chain, so this works through
any amount of wrapping.

## Your turn

`getIssue` is finished for you, and it relies on `checkResponse`. Make it work:

1. Implement `(*APIError).Error` to return `trackr API: <code> <message>`, for example
   `trackr API: 404 issue 999 not found`.
2. Implement `checkResponse`:
   - return `nil` for any status from 200 to 299;
   - otherwise return an `*APIError` with the `StatusCode`, and a `Message` taken
     from the JSON body's `"error"` field, like `{"error": "issue 999 not found"}`;
   - if the body isn't JSON, or has no (or an empty) `"error"` field, use
     `http.StatusText(resp.StatusCode)` as the message.

Don't close the body in `checkResponse`, because `getIssue` already defers that.
