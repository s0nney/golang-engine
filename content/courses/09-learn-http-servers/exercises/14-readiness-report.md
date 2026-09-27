---
title: Readiness Report
difficulty: easy
after: production-readiness
hints:
  - 'Loop over the checks, call each with `r.Context()`, and collect the **names** whose check returned an error. Sort them with `slices.Sort` at the end, since map order is random.'
  - 'Give the response struct a `Failing []string` field with the tag `json:"failing,omitempty"`, so the healthy answer has no `failing` field at all. Set both headers **before** `w.WriteHeader`.'
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

    // readyz returns the handler for GET /readyz. It runs every check with
    // the request's context and answers, with Cache-Control: no-store and
    // Content-Type: application/json:
    //
    //	200 {"status":"ok"}                                  if every check passes
    //	503 {"status":"unavailable","failing":["cache","db"]} otherwise
    //
    // "failing" lists the names of failed checks in sorted order. The checks'
    // error messages never appear in the response.
    func readyz(checks map[string]func(context.Context) error) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		// 1. Run each check and collect the names that return an error.
    		// 2. Sort the names.
    		// 3. Set the headers, then write 200 or 503 with the JSON body.
    	})
    }

    func main() {
    	h := readyz(map[string]func(context.Context) error{
    		"db":    func(ctx context.Context) error { return nil },
    		"cache": func(ctx context.Context) error { return errors.New("dial tcp 10.0.0.9:6379: refused") },
    	})
    	rec := httptest.NewRecorder()
    	h.ServeHTTP(rec, httptest.NewRequest("GET", "/readyz", nil))
    	fmt.Println(rec.Code, rec.Body.String()) // want: 503 {"status":"unavailable","failing":["cache"]}
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
    	"slices"
    )

    type readiness struct {
    	Status  string   `json:"status"`
    	Failing []string `json:"failing,omitempty"`
    }

    func readyz(checks map[string]func(context.Context) error) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		var failing []string
    		for name, check := range checks {
    			if err := check(r.Context()); err != nil {
    				failing = append(failing, name)
    			}
    		}
    		slices.Sort(failing)

    		resp, code := readiness{Status: "ok"}, http.StatusOK
    		if len(failing) > 0 {
    			resp, code = readiness{Status: "unavailable", Failing: failing}, http.StatusServiceUnavailable
    		}
    		data, _ := json.Marshal(resp)
    		w.Header().Set("Cache-Control", "no-store")
    		w.Header().Set("Content-Type", "application/json")
    		w.WriteHeader(code)
    		w.Write(data)
    	})
    }

    func main() {
    	h := readyz(map[string]func(context.Context) error{
    		"db":    func(ctx context.Context) error { return nil },
    		"cache": func(ctx context.Context) error { return errors.New("dial tcp 10.0.0.9:6379: refused") },
    	})
    	rec := httptest.NewRecorder()
    	h.ServeHTTP(rec, httptest.NewRequest("GET", "/readyz", nil))
    	fmt.Println(rec.Code, rec.Body.String())
    }
  tests: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"net/http/httptest"
    	"slices"
    	"strings"
    	"testing"
    )

    type ctxKey struct{}

    func probe(t *testing.T, checks map[string]func(context.Context) error) (int, map[string]any, *httptest.ResponseRecorder) {
    	t.Helper()
    	rec := httptest.NewRecorder()
    	req := httptest.NewRequest("GET", "/readyz", nil)
    	req = req.WithContext(context.WithValue(req.Context(), ctxKey{}, "probe"))
    	readyz(checks).ServeHTTP(rec, req)
    	var body map[string]any
    	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
    		t.Fatalf("body %q isn't JSON: %v", rec.Body.String(), err)
    	}
    	if got := rec.Header().Get("Content-Type"); got != "application/json" {
    		t.Errorf("Content-Type = %q, want application/json", got)
    	}
    	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
    		t.Errorf("Cache-Control = %q, want no-store (a cached answer would hide an outage)", got)
    	}
    	return rec.Code, body, rec
    }

    var (
    	ok   = func(ctx context.Context) error { return nil }
    	fail = func(ctx context.Context) error {
    		return errors.New("pq: password authentication failed for user \"squeak\" at 10.0.3.7")
    	}
    )

    func TestAllHealthy(t *testing.T) {
    	for _, checks := range []map[string]func(context.Context) error{
    		{"db": ok, "cache": ok, "queue": ok},
    		{},
    	} {
    		code, body, rec := probe(t, checks)
    		if code != 200 || body["status"] != "ok" {
    			t.Errorf("%d healthy checks: %d %s, want 200 {\"status\":\"ok\"}", len(checks), code, rec.Body.String())
    		}
    		if _, has := body["failing"]; has {
    			t.Errorf("%d healthy checks: body %s has a \"failing\" field, want none", len(checks), rec.Body.String())
    		}
    	}
    }

    func TestSomeFailing(t *testing.T) {
    	code, body, rec := probe(t, map[string]func(context.Context) error{"queue": fail, "db": fail, "cache": ok, "auth": fail, "search": ok})
    	if code != 503 || body["status"] != "unavailable" {
    		t.Errorf("3 failing checks: %d %s, want 503 with status \"unavailable\"", code, rec.Body.String())
    	}
    	var got []string
    	list, _ := body["failing"].([]any)
    	for _, v := range list {
    		s, _ := v.(string)
    		got = append(got, s)
    	}
    	if want := []string{"auth", "db", "queue"}; !slices.Equal(got, want) {
    		t.Errorf("failing = %q, want %q (sorted names of the failed checks)", got, want)
    	}
    	if s := rec.Body.String(); strings.Contains(s, "10.0.3.7") || strings.Contains(s, "password") {
    		t.Errorf("body %s includes a check's error text; list only the names", s)
    	}
    }

    func TestChecksGetRequestContext(t *testing.T) {
    	var seen []string
    	check := func(name string) func(context.Context) error {
    		return func(ctx context.Context) error {
    			if ctx.Value(ctxKey{}) != "probe" {
    				return errors.New("wrong context")
    			}
    			seen = append(seen, name)
    			return nil
    		}
    	}
    	code, _, _ := probe(t, map[string]func(context.Context) error{"db": check("db"), "cache": check("cache")})
    	slices.Sort(seen)
    	if code != 200 || !slices.Equal(seen, []string{"cache", "db"}) {
    		t.Errorf("each check must run once with r.Context(); ran %q, status %d", seen, code)
    	}
    }
---

Squeak's load balancer asks every instance `GET /readyz` a few times a second:
"can you take traffic right now?" An instance that answers `503` gets no new
requests until it recovers. Squeak depends on a few things (the database, the
cache, the job queue), and each has a small check function.

Complete `readyz(checks)`. It returns a handler that runs **every** check with
the request's context, and then answers:

- `200` with `{"status":"ok"}` if every check returned `nil` (or there are no
  checks).
- `503` with `{"status":"unavailable","failing":[...]}` otherwise, where
  `failing` holds the **names** of the failed checks, sorted.

Both answers carry `Cache-Control: no-store` and `Content-Type: application/json`.

## Example

```
checks: db ok, cache fails with "dial tcp 10.0.0.9:6379: refused"

GET /readyz
503 {"status":"unavailable","failing":["cache"]}
```

## Constraints

- Never include a check's error message: probes are often reachable by more
  people than your logs are, and error text can contain addresses or usernames.
- The healthy answer has no `failing` field at all.
