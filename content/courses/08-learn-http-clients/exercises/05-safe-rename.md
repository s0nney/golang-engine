---
title: Safe Rename
difficulty: medium
after: headers-and-status-codes
hints:
  - 'Split it into two helpers: `get` (GET, check for 200, read `resp.Header.Get("ETag")`, decode) and `patch` (PATCH with `If-Match`, return the status and decoded issue). Then `renameIssue` is a loop of at most 3 rounds.'
  - 'Send the ETag back **exactly** as you got it, quotes and all: `req.Header.Set("If-Match", etag)`. For the body, encode a struct with only a `Title` field, so you can''t overwrite anybody else''s changes to other fields.'
  - 'A `412 Precondition Failed` means "someone changed it since you read it": `continue` the loop to GET the fresh version. Only after the third 412 return `fmt.Errorf("renaming issue %d: %w", id, ErrConflict)`.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    // Issue is one Trackr issue.
    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    	State string `json:"state"`
    }

    // ErrConflict means the issue kept changing under us.
    var ErrConflict = errors.New("issue was modified by someone else")

    func renameIssue(ctx context.Context, client *http.Client, baseURL string, id int, title string) (Issue, error) {
    	return Issue{}, nil
    }

    func main() {
    	version, title := 1, "Dark mode"
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		etag := fmt.Sprintf(`"v%d"`, version)
    		switch r.Method {
    		case http.MethodGet:
    			w.Header().Set("ETag", etag)
    		case http.MethodPatch:
    			if r.Header.Get("If-Match") != etag {
    				w.WriteHeader(http.StatusPreconditionFailed)
    				return
    			}
    			version, title = version+1, "Dark theme" // pretend we decoded the body
    		}
    		fmt.Fprintf(w, `{"id": 42, "title": %q, "state": "open"}`, title)
    	}))
    	defer srv.Close()

    	fmt.Println(renameIssue(context.Background(), srv.Client(), srv.URL, 42, "Dark theme"))
    	// want: {42 Dark theme open} <nil>
    }
  solution: |
    package main

    import (
    	"bytes"
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"net/url"
    	"strconv"
    )

    // Issue is one Trackr issue.
    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    	State string `json:"state"`
    }

    // ErrConflict means the issue kept changing under us.
    var ErrConflict = errors.New("issue was modified by someone else")

    func renameIssue(ctx context.Context, client *http.Client, baseURL string, id int, title string) (Issue, error) {
    	u, err := url.JoinPath(baseURL, "issues", strconv.Itoa(id))
    	if err != nil {
    		return Issue{}, err
    	}
    	for range 3 {
    		iss, etag, err := getIssue(ctx, client, u)
    		if err != nil {
    			return Issue{}, err
    		}
    		if iss.Title == title {
    			return iss, nil
    		}
    		iss, status, err := patchTitle(ctx, client, u, etag, title)
    		if err != nil {
    			return Issue{}, err
    		}
    		if status == http.StatusPreconditionFailed {
    			continue
    		}
    		return iss, nil
    	}
    	return Issue{}, fmt.Errorf("renaming issue %d: %w", id, ErrConflict)
    }

    func getIssue(ctx context.Context, client *http.Client, u string) (Issue, string, error) {
    	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
    	if err != nil {
    		return Issue{}, "", err
    	}
    	resp, err := client.Do(req)
    	if err != nil {
    		return Issue{}, "", err
    	}
    	defer resp.Body.Close()
    	if resp.StatusCode != http.StatusOK {
    		return Issue{}, "", fmt.Errorf("fetching issue: %s", resp.Status)
    	}
    	etag := resp.Header.Get("ETag")
    	if etag == "" {
    		return Issue{}, "", errors.New("fetching issue: no ETag, can't update safely")
    	}
    	var iss Issue
    	if err := json.UnmarshalRead(resp.Body, &iss); err != nil {
    		return Issue{}, "", fmt.Errorf("fetching issue: %w", err)
    	}
    	return iss, etag, nil
    }

    // patchTitle returns the status code alongside the issue so the caller can spot a 412.
    func patchTitle(ctx context.Context, client *http.Client, u, etag, title string) (Issue, int, error) {
    	var body bytes.Buffer
    	if err := json.MarshalWrite(&body, struct {
    		Title string `json:"title"`
    	}{title}); err != nil {
    		return Issue{}, 0, err
    	}
    	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, u, &body)
    	if err != nil {
    		return Issue{}, 0, err
    	}
    	req.Header.Set("Content-Type", "application/json")
    	req.Header.Set("If-Match", etag)
    	resp, err := client.Do(req)
    	if err != nil {
    		return Issue{}, 0, err
    	}
    	defer resp.Body.Close()
    	switch resp.StatusCode {
    	case http.StatusOK:
    		var iss Issue
    		if err := json.UnmarshalRead(resp.Body, &iss); err != nil {
    			return Issue{}, 0, fmt.Errorf("updating issue: %w", err)
    		}
    		return iss, resp.StatusCode, nil
    	case http.StatusPreconditionFailed:
    		return Issue{}, resp.StatusCode, nil
    	default:
    		return Issue{}, 0, fmt.Errorf("updating issue: %s", resp.Status)
    	}
    }

    func main() {
    	version, title := 1, "Dark mode"
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		etag := fmt.Sprintf(`"v%d"`, version)
    		switch r.Method {
    		case http.MethodGet:
    			w.Header().Set("ETag", etag)
    		case http.MethodPatch:
    			if r.Header.Get("If-Match") != etag {
    				w.WriteHeader(http.StatusPreconditionFailed)
    				return
    			}
    			version, title = version+1, "Dark theme" // pretend we decoded the body
    		}
    		fmt.Fprintf(w, `{"id": 42, "title": %q, "state": "open"}`, title)
    	}))
    	defer srv.Close()

    	fmt.Println(renameIssue(context.Background(), srv.Client(), srv.URL, 42, "Dark theme"))
    }
  tests: |
    package main

    import (
    	jsonv1 "encoding/json"
    	"errors"
    	"fmt"
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

    // testTracker is a fake Trackr API holding issue 42. meddle > 0 makes someone
    // else edit the issue (changing its state) right before each of the next
    // meddle PATCHes arrives.
    type testTracker struct {
    	mu            sync.Mutex
    	t             *testing.T
    	version       int
    	iss           Issue
    	meddle        int
    	getStatus     int    // non-zero: GET answers with this status
    	patchStatus   int    // non-zero: PATCH answers with this status
    	noETag        bool   // GET sends no ETag
    	getBody       string // non-empty: GET sends this body
    	gets, patches int
    	log           []string
    }

    func (f *testTracker) etag() string { return fmt.Sprintf(`"r%d-%s"`, f.version, "b7f") }

    func (f *testTracker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    	f.mu.Lock()
    	defer f.mu.Unlock()
    	f.log = append(f.log, r.Method+" "+r.URL.Path)
    	if r.URL.Path != "/v2/issues/42" {
    		http.NotFound(w, r)
    		return
    	}
    	w.Header().Set("Content-Type", "application/json")
    	switch r.Method {
    	case http.MethodGet:
    		f.gets++
    		if f.getStatus != 0 {
    			w.WriteHeader(f.getStatus)
    			return
    		}
    		if !f.noETag {
    			w.Header().Set("ETag", f.etag())
    		}
    		if f.getBody != "" {
    			io.WriteString(w, f.getBody)
    			return
    		}
    		jsonv1.NewEncoder(w).Encode(f.iss)
    	case http.MethodPatch:
    		f.patches++
    		if f.meddle > 0 {
    			f.meddle--
    			f.version++
    			f.iss.State = fmt.Sprintf("edited-%d", f.version)
    		}
    		if got := r.Header.Get("If-Match"); got != f.etag() {
    			w.WriteHeader(http.StatusPreconditionFailed)
    			fmt.Fprintf(w, `{"error": "If-Match %s doesn't match %s"}`, got, f.etag())
    			return
    		}
    		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
    			f.t.Errorf("PATCH Content-Type = %q, want application/json", ct)
    		}
    		var patch map[string]any
    		if err := jsonv1.NewDecoder(r.Body).Decode(&patch); err != nil {
    			f.t.Errorf("PATCH body isn't JSON: %v", err)
    		}
    		if len(patch) != 1 || patch["title"] == nil {
    			f.t.Errorf("PATCH body = %v, want only a \"title\" key (don't send fields you aren't changing)", patch)
    		}
    		if f.patchStatus != 0 {
    			w.WriteHeader(f.patchStatus)
    			return
    		}
    		f.iss.Title, _ = patch["title"].(string)
    		f.version++
    		w.Header().Set("ETag", f.etag())
    		jsonv1.NewEncoder(w).Encode(f.iss)
    	default:
    		w.WriteHeader(http.StatusMethodNotAllowed)
    	}
    }

    func testRun(t *testing.T, f *testTracker, title string) (Issue, error) {
    	t.Helper()
    	f.t = t
    	if f.iss.ID == 0 {
    		f.iss = Issue{42, "Dark mode", "open"}
    		f.version = 7
    	}
    	srv := httptest.NewServer(f)
    	defer srv.Close()
    	tb := &testBodyTracker{next: srv.Client().Transport}
    	iss, err := renameIssue(t.Context(), &http.Client{Transport: tb}, srv.URL+"/v2/", 42, title)
    	if tb.opened.Load() != tb.closed.Load() {
    		t.Errorf("%d response(s) but %d body close(s): close every body", tb.opened.Load(), tb.closed.Load())
    	}
    	return iss, err
    }

    func TestRenameSimple(t *testing.T) {
    	f := &testTracker{}
    	got, err := testRun(t, f, "Dark theme")
    	want := Issue{42, "Dark theme", "open"}
    	if err != nil || got != want {
    		t.Errorf("renameIssue = %+v, %v, want %+v, nil", got, err, want)
    	}
    	if strings.Join(f.log, ", ") != "GET /v2/issues/42, PATCH /v2/issues/42" {
    		t.Errorf("server received %v, want [GET /v2/issues/42 PATCH /v2/issues/42]", f.log)
    	}
    }

    func TestRenameUnicodeTitle(t *testing.T) {
    	f := &testTracker{}
    	title := `Crash when title has "quotes" & émojis 🐛`
    	got, err := testRun(t, f, title)
    	if err != nil || got.Title != title {
    		t.Errorf("renameIssue(%q) = %+v, %v, want the new title back", title, got, err)
    	}
    }

    func TestRenameAlreadyNamed(t *testing.T) {
    	f := &testTracker{}
    	got, err := testRun(t, f, "Dark mode")
    	if err != nil || got != (Issue{42, "Dark mode", "open"}) {
    		t.Errorf("renameIssue to the current title = %+v, %v, want the issue unchanged", got, err)
    	}
    	if f.patches != 0 {
    		t.Errorf("title was already right, but renameIssue sent %d PATCH(es), want 0", f.patches)
    	}
    }

    func TestRenameRetriesAfterConflict(t *testing.T) {
    	for meddle := 1; meddle <= 2; meddle++ {
    		f := &testTracker{meddle: meddle}
    		got, err := testRun(t, f, "Dark theme")
    		want := Issue{42, "Dark theme", fmt.Sprintf("edited-%d", 7+meddle)}
    		if err != nil || got != want {
    			t.Errorf("after %d conflicting edit(s): renameIssue = %+v, %v, want %+v, nil (GET again after a 412 and keep their change)", meddle, got, err, want)
    		}
    		if f.gets != meddle+1 || f.patches != meddle+1 {
    			t.Errorf("after %d conflicting edit(s): server saw %d GETs and %d PATCHes, want %d of each", meddle, f.gets, f.patches, meddle+1)
    		}
    	}
    }

    func TestRenameGivesUp(t *testing.T) {
    	f := &testTracker{meddle: 100}
    	got, err := testRun(t, f, "Dark theme")
    	if !errors.Is(err, ErrConflict) {
    		t.Errorf("issue changes before every PATCH: renameIssue = %+v, %v, want an error wrapping ErrConflict", got, err)
    	}
    	if f.gets != 3 || f.patches != 3 {
    		t.Errorf("issue changes before every PATCH: server saw %d GETs and %d PATCHes, want exactly 3 of each", f.gets, f.patches)
    	}
    }

    func TestRenameFailures(t *testing.T) {
    	tests := []struct {
    		why string
    		f   *testTracker
    	}{
    		{"GET 404", &testTracker{getStatus: http.StatusNotFound}},
    		{"GET 500", &testTracker{getStatus: http.StatusInternalServerError}},
    		{"GET without an ETag", &testTracker{noETag: true}},
    		{"GET with malformed JSON", &testTracker{getBody: `{"id": 42, "title": `}},
    	}
    	for _, tt := range tests {
    		got, err := testRun(t, tt.f, "Dark theme")
    		if err == nil {
    			t.Errorf("%s: renameIssue = %+v, nil, want an error", tt.why, got)
    		}
    		if errors.Is(err, ErrConflict) {
    			t.Errorf("%s: error %v wraps ErrConflict, but nothing conflicted", tt.why, err)
    		}
    		if tt.f.patches != 0 {
    			t.Errorf("%s: renameIssue still sent %d PATCH(es), want 0", tt.why, tt.f.patches)
    		}
    	}
    	for _, status := range []int{http.StatusForbidden, http.StatusUnprocessableEntity, http.StatusInternalServerError} {
    		f := &testTracker{patchStatus: status}
    		got, err := testRun(t, f, "Dark theme")
    		if err == nil || errors.Is(err, ErrConflict) {
    			t.Errorf("PATCH answered %d: renameIssue = %+v, %v, want a non-conflict error", status, got, err)
    		}
    		if f.patches != 1 {
    			t.Errorf("PATCH answered %d: server saw %d PATCHes, want 1 (only a 412 is worth another try)", status, f.patches)
    		}
    	}
    }
---

Two people rename the same issue at once. Without care, the slower one silently
overwrites the faster one. The Trackr API prevents this with **optimistic
concurrency**: every `GET` returns an `ETag` header (a version tag such as
`"r7-b7f"`), and a `PATCH` sent with `If-Match: <that tag>` only succeeds if the
issue hasn't changed since. If it has, the API answers
`412 Precondition Failed` and changes nothing.

Complete `renameIssue(ctx, client, baseURL, id, title)`:

1. `GET <baseURL>/issues/<id>`. Anything but `200 OK`, a missing `ETag` or
   malformed JSON is an error (no PATCH is sent).
2. If the issue already has this title, return it without sending a PATCH.
3. `PATCH <baseURL>/issues/<id>` with `Content-Type: application/json`,
   `If-Match` set to the ETag **exactly** as received, and a body containing
   **only** the title: `{"title": "..."}`.
4. `200 OK`: return the updated issue from the response.
   `412 Precondition Failed`: someone else got there first. Go back to step 1
   and try again with the fresh issue. Any other status is an error, with no
   retry.
5. Make at most **3** rounds. If the third PATCH also gets a 412, return
   `Issue{}` and an error wrapping `ErrConflict`.

## Example

```
GET   /issues/42                    -> 200, ETag "r7-b7f", {"title": "Dark mode", ...}
PATCH /issues/42  If-Match "r7-b7f" -> 412 (someone closed the issue meanwhile)
GET   /issues/42                    -> 200, ETag "r8-b7f", {"state": "closed", ...}
PATCH /issues/42  If-Match "r8-b7f" -> 200, {"title": "Dark theme", "state": "closed"}
```

## Constraints

- `baseURL` may end in `/`.
- Sending the whole issue back in the PATCH would undo the other person's change
  (the tests check that the body holds only `title`).
- Close every response body, including 412s.
