---
title: A Fake Trackr
difficulty: easy
after: testing-http-clients
hints:
  - '`mux := http.NewServeMux()` and `mux.HandleFunc("GET /issues/{id}", ...)` do the routing. The mux answers 404 for unknown paths and 405 for other methods on this one by itself.'
  - 'Inside the handler, check the header first, then `strconv.Atoi(r.PathValue("id"))` and look the issue up. For the error answers, set `Content-Type` **before** `w.WriteHeader(status)`: headers set after it are ignored.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    )

    // Issue is one Trackr issue.
    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    	State string `json:"state"`
    }

    // newFakeTrackr returns a fake Trackr API for tests. It serves
    //
    //	GET /issues/{id}
    //
    // Every answer has Content-Type: application/json.
    //
    //   - No "Authorization: Bearer test-token" header: 401 and {"error": "unauthorized"}.
    //   - An id that isn't in issues (or isn't a number): 404 and {"error": "issue <id> not found"}.
    //   - Otherwise: 200 and the issue as JSON, e.g. {"id":1,"title":"Login broken","state":"open"}.
    func newFakeTrackr(issues []Issue) http.Handler {
    	// Use an http.ServeMux with the pattern "GET /issues/{id}" and r.PathValue("id").
    	return http.NotFoundHandler()
    }

    func main() {
    	srv := httptest.NewServer(newFakeTrackr([]Issue{{1, "Login broken", "open"}}))
    	defer srv.Close()

    	for _, path := range []string{"/issues/1", "/issues/2"} {
    		req, _ := http.NewRequest(http.MethodGet, srv.URL+path, nil)
    		req.Header.Set("Authorization", "Bearer test-token")
    		resp, err := srv.Client().Do(req)
    		if err != nil {
    			fmt.Println(err)
    			return
    		}
    		body, _ := io.ReadAll(resp.Body)
    		resp.Body.Close()
    		fmt.Println(resp.StatusCode, resp.Header.Get("Content-Type"), string(body))
    	}
    	// want:
    	// 200 application/json {"id":1,"title":"Login broken","state":"open"}
    	// 404 application/json {"error":"issue 2 not found"}
    }
  solution: |
    package main

    import (
    	"encoding/json/v2"
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strconv"
    )

    // Issue is one Trackr issue.
    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    	State string `json:"state"`
    }

    func newFakeTrackr(issues []Issue) http.Handler {
    	byID := map[int]Issue{}
    	for _, iss := range issues {
    		byID[iss.ID] = iss
    	}
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /issues/{id}", func(w http.ResponseWriter, r *http.Request) {
    		if r.Header.Get("Authorization") != "Bearer test-token" {
    			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
    			return
    		}
    		id, err := strconv.Atoi(r.PathValue("id"))
    		iss, ok := byID[id]
    		if err != nil || !ok {
    			writeJSON(w, http.StatusNotFound, map[string]string{"error": "issue " + r.PathValue("id") + " not found"})
    			return
    		}
    		writeJSON(w, http.StatusOK, iss)
    	})
    	return mux
    }

    func writeJSON(w http.ResponseWriter, status int, v any) {
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(status)
    	json.MarshalWrite(w, v)
    }

    func main() {
    	srv := httptest.NewServer(newFakeTrackr([]Issue{{1, "Login broken", "open"}}))
    	defer srv.Close()

    	for _, path := range []string{"/issues/1", "/issues/2"} {
    		req, _ := http.NewRequest(http.MethodGet, srv.URL+path, nil)
    		req.Header.Set("Authorization", "Bearer test-token")
    		resp, err := srv.Client().Do(req)
    		if err != nil {
    			fmt.Println(err)
    			return
    		}
    		body, _ := io.ReadAll(resp.Body)
    		resp.Body.Close()
    		fmt.Println(resp.StatusCode, resp.Header.Get("Content-Type"), string(body))
    	}
    }
  tests: |
    package main

    import (
    	jsonv1 "encoding/json"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strconv"
    	"testing"
    )

    var testIssues = []Issue{{1, "Login broken", "open"}, {42, "Dark mode 🌙", "closed"}, {7, `Crash on "quotes"`, "open"}}

    func testDo(t *testing.T, srv *httptest.Server, method, path, auth string) (int, string, map[string]any) {
    	t.Helper()
    	req, err := http.NewRequestWithContext(t.Context(), method, srv.URL+path, nil)
    	if err != nil {
    		t.Fatal(err)
    	}
    	if auth != "" {
    		req.Header.Set("Authorization", auth)
    	}
    	resp, err := srv.Client().Do(req)
    	if err != nil {
    		t.Fatalf("%s %s: %v", method, path, err)
    	}
    	defer resp.Body.Close()
    	body, _ := io.ReadAll(resp.Body)
    	var v map[string]any
    	if len(body) > 0 && jsonv1.Unmarshal(body, &v) != nil {
    		v = map[string]any{"<not JSON>": string(body)}
    	}
    	return resp.StatusCode, resp.Header.Get("Content-Type"), v
    }

    func TestFakeTrackrServesIssues(t *testing.T) {
    	srv := httptest.NewServer(newFakeTrackr(testIssues))
    	defer srv.Close()
    	for _, iss := range testIssues {
    		path := "/issues/" + strconv.Itoa(iss.ID)
    		status, ct, v := testDo(t, srv, "GET", path, "Bearer test-token")
    		if status != 200 || ct != "application/json" {
    			t.Errorf("GET %s: status %d, Content-Type %q, want 200 and application/json", path, status, ct)
    		}
    		if v["id"] != float64(iss.ID) || v["title"] != iss.Title || v["state"] != iss.State || len(v) != 3 {
    			t.Errorf("GET %s: body %v, want {\"id\": %d, \"title\": %q, \"state\": %q}", path, v, iss.ID, iss.Title, iss.State)
    		}
    	}
    }

    func TestFakeTrackrNotFound(t *testing.T) {
    	srv := httptest.NewServer(newFakeTrackr(testIssues))
    	defer srv.Close()
    	for _, tt := range []struct{ path, id string }{{"/issues/2", "2"}, {"/issues/999", "999"}, {"/issues/abc", "abc"}, {"/issues/-1", "-1"}} {
    		status, ct, v := testDo(t, srv, "GET", tt.path, "Bearer test-token")
    		want := "issue " + tt.id + " not found"
    		if status != 404 || ct != "application/json" || v["error"] != want {
    			t.Errorf("GET %s: status %d, Content-Type %q, body %v, want 404, application/json, {\"error\": %q}", tt.path, status, ct, v, want)
    		}
    	}
    	// An empty fake has no issues at all.
    	empty := httptest.NewServer(newFakeTrackr(nil))
    	defer empty.Close()
    	if status, _, _ := testDo(t, empty, "GET", "/issues/1", "Bearer test-token"); status != 404 {
    		t.Errorf("fake with no issues: GET /issues/1 answered %d, want 404", status)
    	}
    }

    func TestFakeTrackrAuth(t *testing.T) {
    	srv := httptest.NewServer(newFakeTrackr(testIssues))
    	defer srv.Close()
    	for _, auth := range []string{"", "Bearer wrong", "test-token", "Basic dGVzdDp0b2tlbg=="} {
    		for _, path := range []string{"/issues/1", "/issues/999"} {
    			status, ct, v := testDo(t, srv, "GET", path, auth)
    			if status != 401 || ct != "application/json" || v["error"] != "unauthorized" {
    				t.Errorf("GET %s with Authorization %q: status %d, Content-Type %q, body %v, want 401, application/json, {\"error\": \"unauthorized\"}", path, auth, status, ct, v)
    			}
    		}
    	}
    }

    func TestFakeTrackrRouting(t *testing.T) {
    	srv := httptest.NewServer(newFakeTrackr(testIssues))
    	defer srv.Close()
    	if status, _, _ := testDo(t, srv, "DELETE", "/issues/1", "Bearer test-token"); status != 405 {
    		t.Errorf("DELETE /issues/1 answered %d, want 405 Method Not Allowed", status)
    	}
    	if status, _, _ := testDo(t, srv, "GET", "/projects/1", "Bearer test-token"); status != 404 {
    		t.Errorf("GET /projects/1 answered %d, want 404", status)
    	}
    	if status, _, _ := testDo(t, srv, "GET", "/issues/1/comments", "Bearer test-token"); status != 404 {
    		t.Errorf("GET /issues/1/comments answered %d, want 404", status)
    	}
    }
---

Every `trackr` test needs an API to talk to, and the real one is off-limits
(slow, flaky, needs credentials). So the test suite has a small **fake Trackr**
that serves canned issues. Time to write it.

Complete `newFakeTrackr(issues)`. It returns an `http.Handler` serving
`GET /issues/{id}`, and every answer it writes has
`Content-Type: application/json`:

- If the request doesn't carry exactly `Authorization: Bearer test-token`:
  `401` and `{"error": "unauthorized"}`. This is checked before anything else,
  so a bad token can't be used to probe which issues exist.
- If `{id}` isn't the id of one of `issues` (including ids that aren't numbers):
  `404` and `{"error": "issue <id> not found"}`, with the id as it appeared in
  the path.
- Otherwise: `200` and the issue as JSON, `{"id":1,"title":"Login broken","state":"open"}`.

Other paths should get a `404` and other methods on `/issues/{id}` a `405`. An
`http.ServeMux` with a method-and-path pattern does both for you.

## Example

```
GET /issues/1   (Bearer test-token)  -> 200 {"id":1,"title":"Login broken","state":"open"}
GET /issues/abc (Bearer test-token)  -> 404 {"error":"issue abc not found"}
GET /issues/1   (no token)           -> 401 {"error":"unauthorized"}
DELETE /issues/1                     -> 405
```

## Constraints

- The tests start your handler with `httptest.NewServer` and check the status,
  the `Content-Type` and the decoded JSON of every answer.
- Headers must be set **before** `WriteHeader` (or the first `Write`), or they
  are silently dropped.
