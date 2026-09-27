---
title: JSON Only, Please
difficulty: easy
after: handlers-and-middleware
hints:
  - 'A middleware is a function that takes the next `http.Handler` and returns a new one, usually an `http.HandlerFunc` closure that does its check and then calls `next.ServeHTTP(w, r)`. Returning **without** calling `next` is how you block a request.'
  - 'Don''t compare the header with `==`: clients send `application/json; charset=utf-8` and `Application/JSON` too. `mime.ParseMediaType(r.Header.Get("Content-Type"))` splits off the parameters and lowercases the type for you, and returns an error for an empty or garbled header.'
exercise:
  starter: |
    package main

    import (
    	"encoding/json/v2"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    )

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	data, _ := json.Marshal(map[string]string{"error": msg})
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    // requireJSON wraps next. POST, PUT and PATCH requests must have a
    // Content-Type whose media type is application/json (parameters such as
    // charset are fine). Others get a 415 Unsupported Media Type error from
    // respondWithError, and next is not called. Other methods always pass.
    func requireJSON(next http.Handler) http.Handler {
    	// Return an http.HandlerFunc that checks r.Method and the
    	// Content-Type header, then either rejects or calls next.
    	return next
    }

    func main() {
    	created := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.WriteHeader(http.StatusCreated)
    	})
    	h := requireJSON(created)
    	for _, ct := range []string{"application/json", "text/plain", ""} {
    		req := httptest.NewRequest("POST", "/api/squeaks", strings.NewReader(`{"body":"hi"}`))
    		req.Header.Set("Content-Type", ct)
    		rec := httptest.NewRecorder()
    		h.ServeHTTP(rec, req)
    		fmt.Printf("Content-Type %q -> %d %s\n", ct, rec.Code, rec.Body.String())
    	}
    }
  solution: |
    package main

    import (
    	"encoding/json/v2"
    	"fmt"
    	"mime"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    )

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	data, _ := json.Marshal(map[string]string{"error": msg})
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    func requireJSON(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		switch r.Method {
    		case http.MethodPost, http.MethodPut, http.MethodPatch:
    			mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
    			if err != nil || mt != "application/json" {
    				respondWithError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
    				return
    			}
    		}
    		next.ServeHTTP(w, r)
    	})
    }

    func main() {
    	created := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.WriteHeader(http.StatusCreated)
    	})
    	h := requireJSON(created)
    	for _, ct := range []string{"application/json", "text/plain", ""} {
    		req := httptest.NewRequest("POST", "/api/squeaks", strings.NewReader(`{"body":"hi"}`))
    		req.Header.Set("Content-Type", ct)
    		rec := httptest.NewRecorder()
    		h.ServeHTTP(rec, req)
    		fmt.Printf("Content-Type %q -> %d %s\n", ct, rec.Code, rec.Body.String())
    	}
    }
  tests: |
    package main

    import (
    	"encoding/json/v2"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"testing"
    )

    func try(method, contentType string) (rec *httptest.ResponseRecorder, called bool) {
    	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		called = true
    		w.WriteHeader(http.StatusTeapot)
    	})
    	req := httptest.NewRequest(method, "/api/squeaks", strings.NewReader(`{"body":"hi"}`))
    	if contentType != "-" {
    		req.Header.Set("Content-Type", contentType)
    	}
    	rec = httptest.NewRecorder()
    	requireJSON(next).ServeHTTP(rec, req)
    	return rec, called
    }

    func TestAccepted(t *testing.T) {
    	for _, tt := range []struct{ method, ct string }{
    		{"POST", "application/json"},
    		{"POST", "application/json; charset=utf-8"},
    		{"PUT", "Application/JSON"},
    		{"PATCH", "application/json;charset=UTF-8"},
    		{"GET", "-"},
    		{"DELETE", "text/plain"},
    		{"HEAD", "-"},
    	} {
    		rec, called := try(tt.method, tt.ct)
    		if !called || rec.Code != http.StatusTeapot {
    			t.Errorf("%s with Content-Type %q: next called = %v, status %d; want the request passed to next", tt.method, tt.ct, called, rec.Code)
    		}
    	}
    }

    func TestRejected(t *testing.T) {
    	for _, tt := range []struct{ method, ct string }{
    		{"POST", "-"},
    		{"POST", ""},
    		{"POST", "text/plain"},
    		{"PUT", "application/x-www-form-urlencoded"},
    		{"PATCH", "multipart/form-data; boundary=xyz"},
    		{"POST", "application/jsonp"},
    		{"POST", "application/json-seq"},
    		{"POST", ";;;"},
    	} {
    		rec, called := try(tt.method, tt.ct)
    		if called {
    			t.Errorf("%s with Content-Type %q: next was called, want the request blocked", tt.method, tt.ct)
    		}
    		if rec.Code != http.StatusUnsupportedMediaType {
    			t.Errorf("%s with Content-Type %q: status %d, want 415", tt.method, tt.ct, rec.Code)
    			continue
    		}
    		if got := rec.Header().Get("Content-Type"); got != "application/json" {
    			t.Errorf("%s with Content-Type %q: response Content-Type = %q, want application/json (use respondWithError)", tt.method, tt.ct, got)
    		}
    		var body map[string]string
    		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["error"] == "" {
    			t.Errorf("%s with Content-Type %q: body = %q, want {\"error\":\"...\"}", tt.method, tt.ct, rec.Body.String())
    		}
    	}
    }
---

Squeak's write endpoints only speak JSON. Instead of checking the
`Content-Type` header in every handler, put the check in one middleware.

Complete `requireJSON(next)`:

- For `POST`, `PUT` and `PATCH` requests, the `Content-Type` header's media type
  must be `application/json`. Parameters such as `; charset=utf-8` are allowed,
  and media types are case-insensitive.
- If it isn't (or the header is missing or garbled), respond with
  `respondWithError(w, http.StatusUnsupportedMediaType, ...)` and **don't** call
  `next`.
- Every other request, and every other method, goes straight to `next`.

## Examples

```
POST  Content-Type: application/json; charset=utf-8  -> next
PUT   Content-Type: Application/JSON                 -> next
GET   (no Content-Type)                              -> next
POST  Content-Type: text/plain                       -> 415 {"error":"Content-Type must be application/json"}
POST  Content-Type: application/jsonp                -> 415
POST  (no Content-Type)                              -> 415
```

## Constraints

- `application/jsonp` and `application/json-seq` are different media types, so
  a prefix check isn't enough.
- The exact error message is up to you. It must be a JSON `{"error": ...}` body.
