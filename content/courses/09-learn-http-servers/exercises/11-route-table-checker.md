---
title: Route Table Checker
difficulty: medium
after: testing-servers
hints:
  - 'For each case: `req := httptest.NewRequest(c.Method, c.Target, strings.NewReader(c.Body))`, set every `c.Header` entry with `req.Header.Set`, then `rec := httptest.NewRecorder()` and `h.ServeHTTP(rec, req)`. A fresh recorder per case keeps cases independent.'
  - 'Don''t stop at the first mismatch: append a message for the status, then one per wrong header, then one for the body. Loop over `slices.Sorted(maps.Keys(c.WantHeader))` so header messages come out in a predictable order.'
  - 'Comparing JSON as strings fails on `{"a":1,"b":2}` vs `{"b":2, "a":1}`. Unmarshal both sides into an `any` and compare with `reflect.DeepEqual`. If the response doesn''t even parse, that''s a body mismatch too.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"net/http"
    )

    type RouteCase struct {
    	Name       string
    	Method     string
    	Target     string
    	Body       string
    	Header     map[string]string
    	WantStatus int
    	WantHeader map[string]string
    	WantJSON   string
    }

    func checkRoutes(h http.Handler, cases []RouteCase) []string {
    	return nil
    }

    func main() {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
    		w.Header().Set("Content-Type", "application/json")
    		fmt.Fprint(w, `{"status": "ok", "version": 3}`)
    	})
    	failures := checkRoutes(mux, []RouteCase{
    		{Name: "health", Method: "GET", Target: "/api/healthz", WantStatus: 200,
    			WantHeader: map[string]string{"Content-Type": "application/json"},
    			WantJSON:   `{"version":3,"status":"ok"}`},
    		{Name: "wrong method", Method: "POST", Target: "/api/healthz", WantStatus: 404},
    	})
    	fmt.Println(len(failures), "failures") // want: 1 failures
    	for _, f := range failures {
    		fmt.Println(f)
    	}
    }
  solution: |
    package main

    import (
    	"encoding/json/v2"
    	"fmt"
    	"maps"
    	"net/http"
    	"net/http/httptest"
    	"reflect"
    	"slices"
    	"strings"
    )

    type RouteCase struct {
    	Name       string
    	Method     string
    	Target     string
    	Body       string
    	Header     map[string]string
    	WantStatus int
    	WantHeader map[string]string
    	WantJSON   string
    }

    func jsonEqual(a, b []byte) bool {
    	var x, y any
    	if json.Unmarshal(a, &x) != nil || json.Unmarshal(b, &y) != nil {
    		return false
    	}
    	return reflect.DeepEqual(x, y)
    }

    func checkRoutes(h http.Handler, cases []RouteCase) []string {
    	var failures []string
    	for _, c := range cases {
    		req := httptest.NewRequest(c.Method, c.Target, strings.NewReader(c.Body))
    		for k, v := range c.Header {
    			req.Header.Set(k, v)
    		}
    		rec := httptest.NewRecorder()
    		h.ServeHTTP(rec, req)

    		if rec.Code != c.WantStatus {
    			failures = append(failures, fmt.Sprintf("%s: status %d, want %d", c.Name, rec.Code, c.WantStatus))
    		}
    		for _, k := range slices.Sorted(maps.Keys(c.WantHeader)) {
    			if got := rec.Header().Get(k); got != c.WantHeader[k] {
    				failures = append(failures, fmt.Sprintf("%s: header %s = %q, want %q", c.Name, k, got, c.WantHeader[k]))
    			}
    		}
    		if c.WantJSON != "" && !jsonEqual(rec.Body.Bytes(), []byte(c.WantJSON)) {
    			failures = append(failures, fmt.Sprintf("%s: body %s, want JSON %s", c.Name, strings.TrimSpace(rec.Body.String()), c.WantJSON))
    		}
    	}
    	return failures
    }

    func main() {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
    		w.Header().Set("Content-Type", "application/json")
    		fmt.Fprint(w, `{"status": "ok", "version": 3}`)
    	})
    	failures := checkRoutes(mux, []RouteCase{
    		{Name: "health", Method: "GET", Target: "/api/healthz", WantStatus: 200,
    			WantHeader: map[string]string{"Content-Type": "application/json"},
    			WantJSON:   `{"version":3,"status":"ok"}`},
    		{Name: "wrong method", Method: "POST", Target: "/api/healthz", WantStatus: 404},
    	})
    	fmt.Println(len(failures), "failures")
    	for _, f := range failures {
    		fmt.Println(f)
    	}
    }
  tests: |
    package main

    import (
    	"fmt"
    	"io"
    	"net/http"
    	"strings"
    	"sync/atomic"
    	"testing"
    )

    // squeakAPI is a tiny API to point the checker at.
    func squeakAPI() http.Handler {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/squeaks/{id}", func(w http.ResponseWriter, r *http.Request) {
    		if r.PathValue("id") != "1" {
    			w.Header().Set("Content-Type", "application/json")
    			w.WriteHeader(404)
    			io.WriteString(w, `{"error":"squeak not found"}`)
    			return
    		}
    		w.Header().Set("Content-Type", "application/json")
    		w.Header().Set("Cache-Control", "max-age=60")
    		io.WriteString(w, `{"id": 1, "body": "cheese", "tags": ["food", "life"], "likes": 12.0}`+"\n")
    	})
    	mux.HandleFunc("POST /api/squeaks", func(w http.ResponseWriter, r *http.Request) {
    		if r.Header.Get("Authorization") != "Bearer pip" {
    			w.WriteHeader(401)
    			return
    		}
    		b, _ := io.ReadAll(r.Body)
    		w.Header().Set("Content-Type", "application/json")
    		w.Header().Set("Location", "/api/squeaks/2")
    		w.WriteHeader(201)
    		fmt.Fprintf(w, `{"id":2,"echo":%q}`, b)
    	})
    	mux.HandleFunc("GET /api/text", func(w http.ResponseWriter, r *http.Request) {
    		io.WriteString(w, "not json")
    	})
    	return mux
    }

    func TestAllPass(t *testing.T) {
    	cases := []RouteCase{
    		{Name: "get squeak", Method: "GET", Target: "/api/squeaks/1", WantStatus: 200,
    			WantHeader: map[string]string{"Content-Type": "application/json", "Cache-Control": "max-age=60"},
    			WantJSON:   `{"likes":12,"tags":["food","life"],"body":"cheese","id":1}`},
    		{Name: "missing squeak", Method: "GET", Target: "/api/squeaks/9", WantStatus: 404, WantJSON: `{"error": "squeak not found"}`},
    		{Name: "create", Method: "POST", Target: "/api/squeaks", Body: "hello",
    			Header:     map[string]string{"Authorization": "Bearer pip"},
    			WantStatus: 201, WantHeader: map[string]string{"Location": "/api/squeaks/2"},
    			WantJSON: `{"echo":"hello","id":2}`},
    		{Name: "no auth", Method: "POST", Target: "/api/squeaks", WantStatus: 401},
    		{Name: "wrong method", Method: "DELETE", Target: "/api/squeaks/1", WantStatus: 405, WantHeader: map[string]string{"Allow": "GET, HEAD"}},
    		{Name: "absent header", Method: "GET", Target: "/api/text", WantStatus: 200, WantHeader: map[string]string{"Location": ""}},
    	}
    	if got := checkRoutes(squeakAPI(), cases); len(got) != 0 {
    		t.Errorf("checkRoutes on a correct API returned %d failures, want none:\n%s", len(got), strings.Join(got, "\n"))
    	}
    	if got := checkRoutes(squeakAPI(), nil); len(got) != 0 {
    		t.Errorf("checkRoutes with no cases = %q, want none", got)
    	}
    }

    func TestReportsEveryMismatch(t *testing.T) {
    	cases := []RouteCase{
    		{Name: "status", Method: "GET", Target: "/api/squeaks/1", WantStatus: 201},
    		{Name: "json value", Method: "GET", Target: "/api/squeaks/1", WantStatus: 200, WantJSON: `{"id":1,"body":"brie","tags":["food","life"],"likes":12}`},
    		{Name: "json order in array", Method: "GET", Target: "/api/squeaks/1", WantStatus: 200, WantJSON: `{"id":1,"body":"cheese","tags":["life","food"],"likes":12}`},
    		{Name: "json extra field", Method: "GET", Target: "/api/squeaks/1", WantStatus: 200, WantJSON: `{"id":1,"body":"cheese","tags":["food","life"]}`},
    		{Name: "not json", Method: "GET", Target: "/api/text", WantStatus: 200, WantJSON: `"not json"`},
    		{Name: "three problems", Method: "POST", Target: "/api/squeaks", Body: "x",
    			Header:     map[string]string{"Authorization": "Bearer pip"},
    			WantStatus: 200, WantHeader: map[string]string{"Location": "/api/squeaks/3", "Content-Type": "text/plain"},
    			WantJSON: `{"id":2,"echo":"y"}`},
    		{Name: "header needs request header", Method: "POST", Target: "/api/squeaks", WantStatus: 401, WantHeader: map[string]string{"Location": "/api/squeaks/2"}},
    	}
    	got := checkRoutes(squeakAPI(), cases)
    	want := []struct {
    		name     string
    		contains []string
    	}{
    		{"status", []string{"status", "200", "201"}},
    		{"json value", []string{"body", "brie"}},
    		{"json order in array", []string{"body"}},
    		{"json extra field", []string{"body"}},
    		{"not json", []string{"body"}},
    		{"three problems", []string{"status", "201", "200"}},
    		{"three problems", []string{"header Content-Type", `"application/json"`, `"text/plain"`}},
    		{"three problems", []string{"header Location", `"/api/squeaks/2"`, `"/api/squeaks/3"`}},
    		{"three problems", []string{"body", `"y"`}},
    		{"header needs request header", []string{"header Location", `""`, `"/api/squeaks/2"`}},
    	}
    	if len(got) != len(want) {
    		t.Fatalf("checkRoutes returned %d failures, want %d:\n%s", len(got), len(want), strings.Join(got, "\n"))
    	}
    	for i, w := range want {
    		if !strings.HasPrefix(got[i], w.name+": ") {
    			t.Errorf("failure %d = %q, want it to start with %q (in case order: status, headers by name, body)", i, got[i], w.name+": ")
    			continue
    		}
    		for _, s := range w.contains {
    			if !strings.Contains(got[i], s) {
    				t.Errorf("failure %d = %q, want it to mention %s", i, got[i], s)
    			}
    		}
    	}
    }

    func TestFreshRequestEachCase(t *testing.T) {
    	var calls atomic.Int64
    	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		n := calls.Add(1)
    		b, _ := io.ReadAll(r.Body)
    		w.Header().Set("X-Call", fmt.Sprint(n))
    		fmt.Fprintf(w, `{"method":%q,"path":%q,"body":%q,"key":%q}`, r.Method, r.URL.RequestURI(), b, r.Header.Get("X-Key"))
    	})
    	cases := []RouteCase{
    		{Name: "a", Method: "PUT", Target: "/x?y=1", Body: "one", Header: map[string]string{"X-Key": "k1"}, WantStatus: 200,
    			WantHeader: map[string]string{"X-Call": "1"}, WantJSON: `{"method":"PUT","path":"/x?y=1","body":"one","key":"k1"}`},
    		{Name: "b", Method: "PATCH", Target: "/z", Body: "two", WantStatus: 200,
    			WantHeader: map[string]string{"X-Call": "2"}, WantJSON: `{"method":"PATCH","path":"/z","body":"two","key":""}`},
    	}
    	if got := checkRoutes(h, cases); len(got) != 0 {
    		t.Errorf("each case must send its own method, target, body and headers to a fresh recorder; failures:\n%s", strings.Join(got, "\n"))
    	}
    	if n := calls.Load(); n != 2 {
    		t.Errorf("handler called %d times for 2 cases, want 2", n)
    	}
    }
---

Squeak has grown to dozens of routes, and its tests had turned into pages of
copy-pasted `httptest` code. Time to build a tiny **table-test runner**: you
describe each request and what should come back, and `checkRoutes` does the
rest.

Complete `checkRoutes(h, cases)`. For each `RouteCase`, in order:

1. Build a request from `Method`, `Target`, `Body` and the `Header` entries, and
   serve it with `h` into a fresh `httptest.ResponseRecorder`.
2. Collect a failure message for **every** mismatch in this order:
   - the status, if it isn't `WantStatus`:
     `"<Name>: status <got>, want <want>"`
   - each `WantHeader` entry whose response header differs, sorted by header
     name: `"<Name>: header <Key> = "<got>", want "<want>""`. A want of `""`
     means the header must be absent.
   - the body, if `WantJSON` isn't empty and the response isn't **the same
     JSON value**: `"<Name>: body <got>, want JSON <want>"`.

Return all messages (none means everything passed).

"The same JSON value" ignores whitespace and the order of object keys, but
not the order of array elements. `{"a":1,"b":[1,2]}` equals
`{ "b": [1, 2], "a": 1.0 }`, and doesn't equal `{"a":1,"b":[2,1]}`.

## Example

```go
checkRoutes(mux, []RouteCase{
	{Name: "health", Method: "GET", Target: "/api/healthz", WantStatus: 200,
		WantJSON: `{"version":3,"status":"ok"}`},
	{Name: "wrong method", Method: "POST", Target: "/api/healthz", WantStatus: 404},
})
// ["wrong method: status 405, want 404"]
```

## Constraints

- The tests check each message's `"<Name>: "` prefix and that it mentions the
  values involved. The exact spacing of the rest is up to you.
- A response that isn't JSON at all is a body mismatch, not a crash.
