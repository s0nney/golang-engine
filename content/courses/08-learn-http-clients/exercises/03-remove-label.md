---
title: Remove a Label
difficulty: easy
after: http-methods
hints:
  - '`http.Get` only does GETs. Build the request with `http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)` and send it with `client.Do(req)`.'
  - 'A label is one path **segment**, so escape it with `url.PathEscape` before joining: otherwise `ui/ux` turns into two segments. `url.JoinPath(baseURL, "issues", strconv.Itoa(id), "labels", url.PathEscape(label))` also copes with a trailing slash on `baseURL`.'
  - 'Once `Do` succeeds you own `resp.Body`: `defer resp.Body.Close()` before you look at the status.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    // removeLabel sends
    //
    //	DELETE <baseURL>/issues/<id>/labels/<label>
    //
    // with the label escaped as a single path segment. baseURL may end in "/".
    // 204 No Content means success. Any other status is an error that mentions
    // the status, e.g. "removing label: 404 Not Found".
    func removeLabel(ctx context.Context, client *http.Client, baseURL string, id int, label string) error {
    	// 1. Build the URL (net/url can join and escape for you).
    	// 2. Build a DELETE request with http.NewRequestWithContext.
    	// 3. Send it with client.Do and close the body.
    	// 4. Check the status code.
    	return nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Println("server got:", r.Method, r.URL.EscapedPath())
    		w.WriteHeader(http.StatusNoContent)
    	}))
    	defer srv.Close()

    	err := removeLabel(context.Background(), srv.Client(), srv.URL+"/v1/", 42, "good first issue")
    	fmt.Println("error:", err)
    	// want:
    	// server got: DELETE /v1/issues/42/labels/good%20first%20issue
    	// error: <nil>
    }
  solution: |
    package main

    import (
    	"context"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"net/url"
    	"strconv"
    )

    func removeLabel(ctx context.Context, client *http.Client, baseURL string, id int, label string) error {
    	u, err := url.JoinPath(baseURL, "issues", strconv.Itoa(id), "labels", url.PathEscape(label))
    	if err != nil {
    		return err
    	}
    	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
    	if err != nil {
    		return err
    	}
    	resp, err := client.Do(req)
    	if err != nil {
    		return err
    	}
    	defer resp.Body.Close()
    	if resp.StatusCode != http.StatusNoContent {
    		return fmt.Errorf("removing label: %s", resp.Status)
    	}
    	return nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Println("server got:", r.Method, r.URL.EscapedPath())
    		w.WriteHeader(http.StatusNoContent)
    	}))
    	defer srv.Close()

    	err := removeLabel(context.Background(), srv.Client(), srv.URL+"/v1/", 42, "good first issue")
    	fmt.Println("error:", err)
    }
  tests: |
    package main

    import (
    	"context"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    	"strings"
    	"sync"
    	"sync/atomic"
    	"testing"
    )

    // testBodyTracker wraps a transport and counts response bodies that were closed.
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

    type testReceived struct {
    	method, path string
    	body         string
    }

    type testLog struct {
    	mu   sync.Mutex
    	reqs []testReceived
    }

    func (l *testLog) all() []testReceived {
    	l.mu.Lock()
    	defer l.mu.Unlock()
    	return slices.Clone(l.reqs)
    }

    // testAPI answers every request with status and records what it testReceived.
    func testAPI(t *testing.T, status int) (*httptest.Server, *testLog) {
    	got := &testLog{}
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		b, _ := io.ReadAll(r.Body)
    		got.mu.Lock()
    		got.reqs = append(got.reqs, testReceived{r.Method, r.URL.EscapedPath(), string(b)})
    		got.mu.Unlock()
    		if status != http.StatusNoContent {
    			w.Header().Set("Content-Type", "application/json")
    			w.WriteHeader(status)
    			io.WriteString(w, `{"error": "nope"}`)
    			return
    		}
    		w.WriteHeader(status)
    	}))
    	t.Cleanup(srv.Close)
    	return srv, got
    }

    func TestRemoveLabelRequest(t *testing.T) {
    	tests := []struct {
    		base, label string
    		id          int
    		wantPath    string
    	}{
    		{"", "bug", 42, "/issues/42/labels/bug"},
    		{"/v1", "bug", 7, "/v1/issues/7/labels/bug"},
    		{"/v1/", "good first issue", 42, "/v1/issues/42/labels/good%20first%20issue"},
    		{"/v1", "ui/ux", 3, "/v1/issues/3/labels/ui%2Fux"},
    		{"/v1", "100%", 3, "/v1/issues/3/labels/100%25"},
    		{"/v1", "día?", 9, "/v1/issues/9/labels/d%C3%ADa%3F"},
    	}
    	for _, tt := range tests {
    		srv, got := testAPI(t, http.StatusNoContent)
    		err := removeLabel(t.Context(), srv.Client(), srv.URL+tt.base, tt.id, tt.label)
    		if err != nil {
    			t.Errorf("removeLabel(base %q, %d, %q) = %v, want nil for a 204", tt.base, tt.id, tt.label, err)
    		}
    		reqs := got.all()
    		if len(reqs) != 1 {
    			t.Errorf("removeLabel(base %q, %d, %q) sent %d requests, want 1", tt.base, tt.id, tt.label, len(reqs))
    			continue
    		}
    		r := reqs[0]
    		if r.method != http.MethodDelete || r.path != tt.wantPath {
    			t.Errorf("removeLabel(base %q, %d, %q) sent %s %s, want DELETE %s", tt.base, tt.id, tt.label, r.method, r.path, tt.wantPath)
    		}
    		if r.body != "" {
    			t.Errorf("removeLabel sent a request body %q, want none", r.body)
    		}
    	}
    }

    func TestRemoveLabelStatus(t *testing.T) {
    	for _, status := range []int{http.StatusOK, http.StatusNotFound, http.StatusForbidden, http.StatusInternalServerError} {
    		srv, _ := testAPI(t, status)
    		err := removeLabel(t.Context(), srv.Client(), srv.URL, 1, "bug")
    		if err == nil {
    			t.Errorf("server answered %d: removeLabel returned nil, want an error", status)
    			continue
    		}
    		if !strings.Contains(err.Error(), http.StatusText(status)) {
    			t.Errorf("server answered %d: error %q should mention the status %q", status, err, http.StatusText(status))
    		}
    	}
    }

    func TestRemoveLabelClosesBody(t *testing.T) {
    	for _, status := range []int{http.StatusNoContent, http.StatusNotFound} {
    		srv, _ := testAPI(t, status)
    		tb := &testBodyTracker{next: srv.Client().Transport}
    		removeLabel(t.Context(), &http.Client{Transport: tb}, srv.URL, 1, "bug")
    		if tb.opened.Load() != 1 || tb.closed.Load() != 1 {
    			t.Errorf("server answered %d: %d response(s), %d body close(s), want 1 and 1 (use the client you're given, and always close the body)", status, tb.opened.Load(), tb.closed.Load())
    		}
    	}
    }

    func TestRemoveLabelUsesContext(t *testing.T) {
    	srv, got := testAPI(t, http.StatusNoContent)
    	ctx, cancel := context.WithCancel(t.Context())
    	cancel()
    	err := removeLabel(ctx, srv.Client(), srv.URL, 1, "bug")
    	if n := len(got.all()); err == nil || n != 0 {
    		t.Errorf("with a cancelled context removeLabel = %v and sent %d requests, want an error and 0 requests (build the request with http.NewRequestWithContext)", err, n)
    	}
    }
---

`trackr label rm 42 "good first issue"` takes a label off an issue. The Trackr
API models each label as its own resource, so removing one is a `DELETE`:

```
DELETE /issues/42/labels/good%20first%20issue
```

Complete `removeLabel(ctx, client, baseURL, id, label)`:

- Send `DELETE <baseURL>/issues/<id>/labels/<label>` with **no body**, using
  `client` and a request built with `ctx`.
- The label is **one** path segment, whatever it contains. `ui/ux` must arrive
  as `ui%2Fux`, not as two segments.
- `baseURL` may have a path (`.../v1`) and may end in `/`.
- `204 No Content` means success. Any other status is an error whose message
  includes the status text, e.g. `removing label: 404 Not Found`.
- Always close the response body.

## Examples

```
removeLabel(ctx, c, "https://api.trackr.dev/v1/", 42, "good first issue")
// DELETE /v1/issues/42/labels/good%20first%20issue -> 204 -> nil

removeLabel(ctx, c, "https://api.trackr.dev/v1", 3, "ui/ux")
// DELETE /v1/issues/3/labels/ui%2Fux -> 404 -> error "removing label: 404 Not Found"
```

## Constraints

- The tests check the method and the **escaped** path the server received,
  including labels with `/`, `%`, `?` and non-ASCII letters.
- They also count response bodies, so make sure every one is closed, including
  on error statuses.
