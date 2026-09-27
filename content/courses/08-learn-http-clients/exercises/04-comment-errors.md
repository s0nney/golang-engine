---
title: Comments and API Errors
difficulty: medium
after: headers-and-status-codes
hints:
  - 'Don''t compare `Content-Type` with `==`: servers send `application/json; charset=utf-8` and even `Application/JSON`. `mime.ParseMediaType` returns the lower-cased media type without the parameters.'
  - 'For the error path, decode into a small struct shaped like the body, `struct{ Error struct{ Code, Message string } }` with tags. If decoding fails or the message is empty, fall back to `http.StatusText(resp.StatusCode)`.'
  - 'Return `&APIError{...}` (a pointer) so `errors.AsType[*APIError](err)` finds it. And make every path after `client.Do` close the body: one `defer resp.Body.Close()` right after the error check does it.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    // Comment is a comment on a Trackr issue.
    type Comment struct {
    	ID      int    `json:"id"`
    	IssueID int    `json:"issue_id"`
    	Body    string `json:"body"`
    	Author  string `json:"author"`
    }

    // APIError is a non-2xx answer from the Trackr API.
    type APIError struct {
    	StatusCode int
    	Code       string // machine-readable, e.g. "invalid_body"; may be ""
    	Message    string
    }

    func (e *APIError) Error() string {
    	if e.Code == "" {
    		return fmt.Sprintf("trackr API: %d %s", e.StatusCode, e.Message)
    	}
    	return fmt.Sprintf("trackr API: %d %s: %s", e.StatusCode, e.Code, e.Message)
    }

    func addComment(ctx context.Context, client *http.Client, baseURL string, issueID int, text string) (Comment, error) {
    	return Comment{}, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.Header().Set("Content-Type", "application/json; charset=utf-8")
    		if r.URL.Path == "/issues/404/comments" {
    			w.WriteHeader(http.StatusNotFound)
    			fmt.Fprint(w, `{"error": {"code": "not_found", "message": "issue 404 not found"}}`)
    			return
    		}
    		w.WriteHeader(http.StatusCreated)
    		fmt.Fprint(w, `{"id": 1, "issue_id": 42, "body": "Can reproduce on 1.4", "author": "ana"}`)
    	}))
    	defer srv.Close()

    	fmt.Println(addComment(context.Background(), srv.Client(), srv.URL, 42, "Can reproduce on 1.4"))
    	// want: {1 42 Can reproduce on 1.4 ana} <nil>
    	fmt.Println(addComment(context.Background(), srv.Client(), srv.URL, 404, "hello?"))
    	// want: {0 0  } trackr API: 404 not_found: issue 404 not found
    }
  solution: |
    package main

    import (
    	"bytes"
    	"context"
    	"encoding/json/v2"
    	"fmt"
    	"mime"
    	"net/http"
    	"net/http/httptest"
    	"net/url"
    	"strconv"
    )

    // Comment is a comment on a Trackr issue.
    type Comment struct {
    	ID      int    `json:"id"`
    	IssueID int    `json:"issue_id"`
    	Body    string `json:"body"`
    	Author  string `json:"author"`
    }

    // APIError is a non-2xx answer from the Trackr API.
    type APIError struct {
    	StatusCode int
    	Code       string // machine-readable, e.g. "invalid_body"; may be ""
    	Message    string
    }

    func (e *APIError) Error() string {
    	if e.Code == "" {
    		return fmt.Sprintf("trackr API: %d %s", e.StatusCode, e.Message)
    	}
    	return fmt.Sprintf("trackr API: %d %s: %s", e.StatusCode, e.Code, e.Message)
    }

    func addComment(ctx context.Context, client *http.Client, baseURL string, issueID int, text string) (Comment, error) {
    	u, err := url.JoinPath(baseURL, "issues", strconv.Itoa(issueID), "comments")
    	if err != nil {
    		return Comment{}, err
    	}
    	var body bytes.Buffer
    	if err := json.MarshalWrite(&body, struct {
    		Body string `json:"body"`
    	}{text}); err != nil {
    		return Comment{}, err
    	}
    	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, &body)
    	if err != nil {
    		return Comment{}, err
    	}
    	req.Header.Set("Content-Type", "application/json")
    	req.Header.Set("Accept", "application/json")

    	resp, err := client.Do(req)
    	if err != nil {
    		return Comment{}, err
    	}
    	defer resp.Body.Close()

    	if resp.StatusCode < 200 || resp.StatusCode > 299 {
    		apiErr := &APIError{StatusCode: resp.StatusCode}
    		var e struct {
    			Error struct {
    				Code    string `json:"code"`
    				Message string `json:"message"`
    			} `json:"error"`
    		}
    		if json.UnmarshalRead(resp.Body, &e) == nil && e.Error.Message != "" {
    			apiErr.Code, apiErr.Message = e.Error.Code, e.Error.Message
    		} else {
    			apiErr.Message = http.StatusText(resp.StatusCode)
    		}
    		return Comment{}, apiErr
    	}

    	mt, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
    	if err != nil || mt != "application/json" {
    		return Comment{}, fmt.Errorf("adding comment: unexpected Content-Type %q", resp.Header.Get("Content-Type"))
    	}
    	var c Comment
    	if err := json.UnmarshalRead(resp.Body, &c); err != nil {
    		return Comment{}, fmt.Errorf("adding comment: decoding response: %w", err)
    	}
    	return c, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.Header().Set("Content-Type", "application/json; charset=utf-8")
    		if r.URL.Path == "/issues/404/comments" {
    			w.WriteHeader(http.StatusNotFound)
    			fmt.Fprint(w, `{"error": {"code": "not_found", "message": "issue 404 not found"}}`)
    			return
    		}
    		w.WriteHeader(http.StatusCreated)
    		fmt.Fprint(w, `{"id": 1, "issue_id": 42, "body": "Can reproduce on 1.4", "author": "ana"}`)
    	}))
    	defer srv.Close()

    	fmt.Println(addComment(context.Background(), srv.Client(), srv.URL, 42, "Can reproduce on 1.4"))
    	fmt.Println(addComment(context.Background(), srv.Client(), srv.URL, 404, "hello?"))
    }
  tests: |
    package main

    import (
    	jsonv1 "encoding/json"
    	"errors"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync"
    	"sync/atomic"
    	"testing"
    )

    type testBodyTracker struct {
    	next           http.RoundTripper
    	opened, closed atomic.Int32
    }

    func (tb *testBodyTracker) RoundTrip(req *http.Request) (*http.Response, error) {
    	resp, err := tb.next.RoundTrip(req)
    	if err == nil {
    		tb.opened.Add(1)
    		resp.Body = &testCloseCounter{resp.Body, &tb.closed}
    	}
    	return resp, err
    }

    type testCloseCounter struct {
    	io.ReadCloser
    	n *atomic.Int32
    }

    func (c *testCloseCounter) Close() error {
    	c.n.Add(1)
    	return c.ReadCloser.Close()
    }

    // testServer answers every request with the given status, Content-Type and body.
    func testServer(t *testing.T, status int, contentType, body string) (*http.Client, string) {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		if contentType != "" {
    			w.Header().Set("Content-Type", contentType)
    		}
    		w.WriteHeader(status)
    		io.WriteString(w, body)
    	}))
    	t.Cleanup(srv.Close)
    	tb := &testBodyTracker{next: srv.Client().Transport}
    	t.Cleanup(func() {
    		if tb.opened.Load() != tb.closed.Load() {
    			t.Errorf("server answered %d: %d response(s) but %d body close(s); close every body, even for errors", status, tb.opened.Load(), tb.closed.Load())
    		}
    	})
    	return &http.Client{Transport: tb}, srv.URL
    }

    func TestAddCommentRequest(t *testing.T) {
    	var mu sync.Mutex
    	var method, path, ctype, accept string
    	var sent map[string]any
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		mu.Lock()
    		defer mu.Unlock()
    		method, path = r.Method, r.URL.Path
    		ctype, accept = r.Header.Get("Content-Type"), r.Header.Get("Accept")
    		sent = nil
    		if err := jsonv1.NewDecoder(r.Body).Decode(&sent); err != nil {
    			t.Errorf("request body isn't valid JSON: %v", err)
    		}
    		w.Header().Set("Content-Type", "application/json")
    		w.WriteHeader(http.StatusCreated)
    		io.WriteString(w, `{"id": 5, "issue_id": 42, "body": "x", "author": "ana"}`)
    	}))
    	defer srv.Close()

    	for _, text := range []string{"Can reproduce on 1.4", `He said "it's fine" \o/`, "Ça marche 👍\nsecond line", ""} {
    		if _, err := addComment(t.Context(), srv.Client(), srv.URL+"/v1/", 42, text); err != nil {
    			t.Errorf("addComment(%q) = %v, want no error", text, err)
    		}
    		mu.Lock()
    		if method != http.MethodPost || path != "/v1/issues/42/comments" {
    			t.Errorf("addComment sent %s %s, want POST /v1/issues/42/comments (base URL %s/v1/)", method, path, srv.URL)
    		}
    		if ctype != "application/json" || accept != "application/json" {
    			t.Errorf("addComment sent Content-Type %q and Accept %q, want both \"application/json\"", ctype, accept)
    		}
    		if len(sent) != 1 || sent["body"] != text {
    			t.Errorf("addComment(%q) sent JSON %v, want exactly {\"body\": %q}", text, sent, text)
    		}
    		mu.Unlock()
    	}
    }

    func TestAddCommentSuccess(t *testing.T) {
    	want := Comment{ID: 5, IssueID: 42, Body: "Café ☕ fixed it", Author: "björn"}
    	body := `{"id": 5, "issue_id": 42, "body": "Café ☕ fixed it", "author": "björn"}`
    	for _, ct := range []string{"application/json", "application/json; charset=utf-8", "Application/JSON;charset=UTF-8"} {
    		for _, status := range []int{http.StatusCreated, http.StatusOK} {
    			client, base := testServer(t, status, ct, body)
    			got, err := addComment(t.Context(), client, base, 42, "Café ☕ fixed it")
    			if err != nil || got != want {
    				t.Errorf("status %d, Content-Type %q: addComment = %+v, %v, want %+v, nil", status, ct, got, err, want)
    			}
    		}
    	}
    }

    func TestAddCommentBadSuccessResponse(t *testing.T) {
    	tests := []struct {
    		why, ct, body string
    	}{
    		{"HTML instead of JSON", "text/html; charset=utf-8", `<html>Login</html>`},
    		{"missing Content-Type", "", `{"id": 5}`},
    		{"malformed JSON", "application/json", `{"id": 5, "body": `},
    		{"wrong JSON type", "application/json", `{"id": "five"}`},
    		{"empty body", "application/json", ``},
    	}
    	for _, tt := range tests {
    		client, base := testServer(t, http.StatusCreated, tt.ct, tt.body)
    		got, err := addComment(t.Context(), client, base, 42, "hi")
    		if err == nil {
    			t.Errorf("%s: addComment = %+v, nil, want an error", tt.why, got)
    			continue
    		}
    		if _, ok := errors.AsType[*APIError](err); ok {
    			t.Errorf("%s: a 201 response gave an *APIError (%v); APIError is only for non-2xx statuses", tt.why, err)
    		}
    	}
    }

    func TestAddCommentAPIErrors(t *testing.T) {
    	tests := []struct {
    		status   int
    		ct, body string
    		want     APIError
    	}{
    		{422, "application/json", `{"error": {"code": "invalid_body", "message": "body must not be empty"}}`,
    			APIError{422, "invalid_body", "body must not be empty"}},
    		{404, "application/json; charset=utf-8", `{"error": {"code": "not_found", "message": "issue 42 not found"}}`,
    			APIError{404, "not_found", "issue 42 not found"}},
    		{403, "application/json", `{"error": {"message": "read-only token"}}`, APIError{403, "", "read-only token"}},
    		{500, "text/html", `<h1>Internal Server Error</h1>`, APIError{500, "", "Internal Server Error"}},
    		{502, "application/json", `{"error": {"code": "bad_gateway", `, APIError{502, "", "Bad Gateway"}},
    		{503, "", ``, APIError{503, "", "Service Unavailable"}},
    		{429, "application/json", `{"error": {"code": "rate_limited", "message": ""}}`, APIError{429, "", "Too Many Requests"}},
    	}
    	for _, tt := range tests {
    		client, base := testServer(t, tt.status, tt.ct, tt.body)
    		got, err := addComment(t.Context(), client, base, 42, "hi")
    		apiErr, ok := errors.AsType[*APIError](err)
    		if !ok {
    			t.Errorf("status %d with body %q: addComment error = %v, want an *APIError", tt.status, tt.body, err)
    			continue
    		}
    		if *apiErr != tt.want {
    			t.Errorf("status %d with body %q: got %#v, want %#v", tt.status, tt.body, *apiErr, tt.want)
    		}
    		if got != (Comment{}) {
    			t.Errorf("status %d: addComment returned %+v with its error, want Comment{}", tt.status, got)
    		}
    		if !strings.Contains(err.Error(), tt.want.Message) {
    			t.Errorf("status %d: error text %q should include %q", tt.status, err, tt.want.Message)
    		}
    	}
    }
---

`trackr comment 42 "Can reproduce on 1.4"` posts a comment. When it fails,
`trackr` wants to tell "that issue doesn't exist" apart from "your comment is
empty" and "the server is down", so errors come back as a typed `*APIError`.

Complete `addComment(ctx, client, baseURL, issueID, text)`.

**The request**

- `POST <baseURL>/issues/<issueID>/comments` (`baseURL` may end in `/`), sent
  with `client` and built with `ctx`.
- The body is the JSON object `{"body": <text>}`. Encode it properly: comments
  contain quotes, backslashes, newlines and emoji.
- Headers `Content-Type: application/json` and `Accept: application/json`.

**A 2xx response**

- Its media type must be `application/json`. Parameters such as
  `; charset=utf-8` are fine, and media types are case-insensitive. Anything else
  (including no `Content-Type` at all) is an error.
- Decode the body into a `Comment`. Malformed JSON is an error.
- These errors are **not** `*APIError`s: the API said yes, but the answer was bad.

**Any other status** returns `Comment{}` and an `*APIError` with:

- `StatusCode`: the status code.
- `Code` and `Message`: from a JSON body shaped like
  `{"error": {"code": "invalid_body", "message": "body must not be empty"}}`.
- If the body doesn't decode into that shape, or its `message` is empty, `Code`
  is `""` and `Message` is `http.StatusText(StatusCode)`.

## Examples

```
201, application/json; charset=utf-8, {"id": 1, "issue_id": 42, ...}
    -> Comment{ID: 1, IssueID: 42, ...}, nil
404, {"error": {"code": "not_found", "message": "issue 404 not found"}}
    -> &APIError{404, "not_found", "issue 404 not found"}
500, <h1>Internal Server Error</h1>
    -> &APIError{500, "", "Internal Server Error"}
201, text/html, <html>Login</html>
    -> an error, but not an *APIError
```

## Constraints

- The tests use `errors.AsType[*APIError](err)`, so return a pointer (wrapping
  it with `%w` is fine).
- The fake API checks what it received: method, path, both headers and the
  decoded JSON body.
- Every response body must be closed, on every path.
