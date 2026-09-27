---
title: Auth Middleware
quiz:
  - question: Why do APIs send tokens in the `Authorization` header rather than the URL, like `/api/squeaks?token=eyJ...`?
    options:
      - text: URLs can't contain dots
      - text: URLs end up in server logs, browser history and `Referer` headers, spreading the token around
        correct: true
      - text: Query strings are limited to 16 characters
      - text: The mux strips query strings
    explanation: |
      Access logs, proxies, analytics and browser history all record full URLs. Headers
      usually aren't logged. Anything bearing a secret belongs in a header (or a secure
      cookie).
  - question: |
      What does this return for the header `Authorization: Bearer`?

      ```go
      scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
      ```
    options:
      - text: '`scheme="Bearer"`, `token=""`, `ok=false`'
        correct: true
      - text: '`scheme="Bearer"`, `token=""`, `ok=true`'
      - text: '`scheme=""`, `token="Bearer"`, `ok=false`'
      - text: It panics
    explanation: |
      There's no space to cut on, so `Cut` returns the whole string as the first result, an
      empty second result and `ok=false`. Your code must treat that (and an empty token after
      the space) as "no token".
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
    	"strings"
    	"uuid"
    )

    var ErrNoToken = errors.New("missing bearer token")

    type apiConfig struct {
    	validateToken func(token string) (uuid.UUID, error)
    }

    type userIDKey struct{}

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, err := json.Marshal(payload)
    	if err != nil {
    		w.WriteHeader(http.StatusInternalServerError)
    		return
    	}
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	respondWithJSON(w, code, map[string]string{"error": msg})
    }

    // getBearerToken extracts the token from "Authorization: Bearer <token>".
    func getBearerToken(h http.Header) (string, error) {
    	// ?
    	return h.Get("Authorization"), nil
    }

    // requireAuth rejects requests without a valid token (401) and otherwise
    // stores the user ID in the request context under userIDKey{}.
    func (cfg *apiConfig) requireAuth(next http.Handler) http.Handler {
    	// ?
    	return next
    }

    // userIDFrom returns the user ID that requireAuth stored in ctx.
    func userIDFrom(ctx context.Context) (uuid.UUID, bool) {
    	// ?
    	return uuid.Nil(), true
    }

    func (cfg *apiConfig) handleMe(w http.ResponseWriter, r *http.Request) {
    	userID, ok := userIDFrom(r.Context())
    	if !ok {
    		respondWithError(w, http.StatusInternalServerError, "auth middleware missing")
    		return
    	}
    	respondWithJSON(w, http.StatusOK, map[string]string{"id": userID.String()})
    }

    func newRouter(cfg *apiConfig) *http.ServeMux {
    	mux := http.NewServeMux()
    	mux.Handle("GET /api/users/me", cfg.requireAuth(http.HandlerFunc(cfg.handleMe)))
    	return mux
    }

    func main() {
    	pip := uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	cfg := &apiConfig{
    		validateToken: func(token string) (uuid.UUID, error) {
    			if token == "pips-token" {
    				return pip, nil
    			}
    			return uuid.Nil(), errors.New("invalid token")
    		},
    	}
    	mux := newRouter(cfg)

    	for _, auth := range []string{"", "Bearer pips-token", "Bearer forged", "Basic cGlwOmNoZWVzZQ=="} {
    		req := httptest.NewRequest("GET", "/api/users/me", nil)
    		if auth != "" {
    			req.Header.Set("Authorization", auth)
    		}
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, req)
    		fmt.Printf("%-30q -> %d %s\n", auth, rec.Code, strings.TrimSpace(rec.Body.String()))
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
    	"strings"
    	"uuid"
    )

    var ErrNoToken = errors.New("missing bearer token")

    type apiConfig struct {
    	validateToken func(token string) (uuid.UUID, error)
    }

    type userIDKey struct{}

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, err := json.Marshal(payload)
    	if err != nil {
    		w.WriteHeader(http.StatusInternalServerError)
    		return
    	}
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	respondWithJSON(w, code, map[string]string{"error": msg})
    }

    func getBearerToken(h http.Header) (string, error) {
    	scheme, token, ok := strings.Cut(h.Get("Authorization"), " ")
    	token = strings.TrimSpace(token)
    	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
    		return "", ErrNoToken
    	}
    	return token, nil
    }

    func unauthorized(w http.ResponseWriter, msg string) {
    	w.Header().Set("WWW-Authenticate", `Bearer realm="squeak"`)
    	respondWithError(w, http.StatusUnauthorized, msg)
    }

    func (cfg *apiConfig) requireAuth(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		token, err := getBearerToken(r.Header)
    		if err != nil {
    			unauthorized(w, "missing bearer token")
    			return
    		}
    		userID, err := cfg.validateToken(token)
    		if err != nil {
    			unauthorized(w, "invalid or expired token")
    			return
    		}
    		ctx := context.WithValue(r.Context(), userIDKey{}, userID)
    		next.ServeHTTP(w, r.WithContext(ctx))
    	})
    }

    func userIDFrom(ctx context.Context) (uuid.UUID, bool) {
    	id, ok := ctx.Value(userIDKey{}).(uuid.UUID)
    	return id, ok
    }

    func (cfg *apiConfig) handleMe(w http.ResponseWriter, r *http.Request) {
    	userID, ok := userIDFrom(r.Context())
    	if !ok {
    		respondWithError(w, http.StatusInternalServerError, "auth middleware missing")
    		return
    	}
    	respondWithJSON(w, http.StatusOK, map[string]string{"id": userID.String()})
    }

    func newRouter(cfg *apiConfig) *http.ServeMux {
    	mux := http.NewServeMux()
    	mux.Handle("GET /api/users/me", cfg.requireAuth(http.HandlerFunc(cfg.handleMe)))
    	return mux
    }

    func main() {
    	pip := uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	cfg := &apiConfig{
    		validateToken: func(token string) (uuid.UUID, error) {
    			if token == "pips-token" {
    				return pip, nil
    			}
    			return uuid.Nil(), errors.New("invalid token")
    		},
    	}
    	mux := newRouter(cfg)

    	for _, auth := range []string{"", "Bearer pips-token", "Bearer forged", "Basic cGlwOmNoZWVzZQ=="} {
    		req := httptest.NewRequest("GET", "/api/users/me", nil)
    		if auth != "" {
    			req.Header.Set("Authorization", auth)
    		}
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, req)
    		fmt.Printf("%-30q -> %d %s\n", auth, rec.Code, strings.TrimSpace(rec.Body.String()))
    	}
    }
  tests: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"testing"
    	"uuid"
    )

    var pip = uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")

    func fakeConfig(calls *[]string) *apiConfig {
    	return &apiConfig{validateToken: func(token string) (uuid.UUID, error) {
    		*calls = append(*calls, token)
    		if token == "pips-token" {
    			return pip, nil
    		}
    		return uuid.Nil(), errors.New("invalid token")
    	}}
    }

    func TestGetBearerToken(t *testing.T) {
    	for _, tt := range []struct {
    		header, want string
    		wantErr      bool
    	}{
    		{"Bearer abc.def.ghi", "abc.def.ghi", false},
    		{"bearer abc", "abc", false},
    		{"BEARER abc", "abc", false},
    		{"Bearer   abc  ", "abc", false},
    		{"", "", true},
    		{"Bearer", "", true},
    		{"Bearer ", "", true},
    		{"Bearer    ", "", true},
    		{"Basic cGlwOmNoZWVzZQ==", "", true},
    		{"Token abc", "", true},
    		{"abc", "", true},
    	} {
    		h := http.Header{}
    		if tt.header != "" {
    			h.Set("Authorization", tt.header)
    		}
    		got, err := getBearerToken(h)
    		if tt.wantErr {
    			if !errors.Is(err, ErrNoToken) {
    				t.Errorf("getBearerToken(%q) error = %v, want ErrNoToken", tt.header, err)
    			}
    			continue
    		}
    		if err != nil || got != tt.want {
    			t.Errorf("getBearerToken(%q) = %q, %v; want %q, nil", tt.header, got, err, tt.want)
    		}
    	}
    }

    func TestRequireAuthRejects(t *testing.T) {
    	for _, auth := range []string{"", "Bearer forged", "Bearer ", "Basic cGlwOmNoZWVzZQ=="} {
    		var calls []string
    		cfg := fakeConfig(&calls)
    		reached := false
    		h := cfg.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    			reached = true
    		}))
    		req := httptest.NewRequest("GET", "/api/users/me", nil)
    		if auth != "" {
    			req.Header.Set("Authorization", auth)
    		}
    		rec := httptest.NewRecorder()
    		h.ServeHTTP(rec, req)
    		if reached {
    			t.Errorf("Authorization %q: the wrapped handler ran, but the request should have been rejected", auth)
    		}
    		if rec.Code != http.StatusUnauthorized {
    			t.Errorf("Authorization %q: status = %d, want 401", auth, rec.Code)
    		}
    		if !strings.HasPrefix(rec.Header().Get("WWW-Authenticate"), "Bearer") {
    			t.Errorf("Authorization %q: WWW-Authenticate = %q, want it to start with Bearer", auth, rec.Header().Get("WWW-Authenticate"))
    		}
    		var body map[string]string
    		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["error"] == "" {
    			t.Errorf("Authorization %q: body = %q, want JSON like {\"error\":\"...\"}", auth, rec.Body.String())
    		}
    	}
    }

    func TestRequireAuthAccepts(t *testing.T) {
    	var calls []string
    	cfg := fakeConfig(&calls)
    	var seen uuid.UUID
    	var seenOK bool
    	h := cfg.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		seen, seenOK = userIDFrom(r.Context())
    		w.WriteHeader(http.StatusTeapot)
    	}))
    	req := httptest.NewRequest("GET", "/api/users/me", nil)
    	req.Header.Set("Authorization", "Bearer pips-token")
    	rec := httptest.NewRecorder()
    	h.ServeHTTP(rec, req)
    	if rec.Code != http.StatusTeapot {
    		t.Fatalf("valid token: status = %d, want the wrapped handler's 418", rec.Code)
    	}
    	if !seenOK || seen != pip {
    		t.Errorf("valid token: userIDFrom in handler = %s, %v; want %s, true", seen, seenOK, pip)
    	}
    	if len(calls) != 1 || calls[0] != "pips-token" {
    		t.Errorf("validateToken was called with %q, want exactly [\"pips-token\"]", calls)
    	}
    }

    func TestUserIDFromEmpty(t *testing.T) {
    	if id, ok := userIDFrom(context.Background()); ok || id != uuid.Nil() {
    		t.Errorf("userIDFrom(empty context) = %s, %v; want uuid.Nil(), false", id, ok)
    	}
    }

    func TestRouterEndToEnd(t *testing.T) {
    	var calls []string
    	mux := newRouter(fakeConfig(&calls))
    	req := httptest.NewRequest("GET", "/api/users/me", nil)
    	req.Header.Set("Authorization", "Bearer pips-token")
    	rec := httptest.NewRecorder()
    	mux.ServeHTTP(rec, req)
    	var body map[string]string
    	json.Unmarshal(rec.Body.Bytes(), &body)
    	if rec.Code != 200 || body["id"] != pip.String() {
    		t.Errorf("GET /api/users/me with pip's token = %d %q, want 200 with id %s", rec.Code, rec.Body.String(), pip)
    	}
    }
---

Squeak can issue and verify tokens. Now it needs to *demand* them. Posting a squeak,
deleting one, and `GET /api/users/me` all need to know who's calling. That's a job
for middleware.

## The Authorization header

Clients send an access token like this:

```
GET /api/users/me
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3Mi...
```

`Bearer` is the *scheme*: "whoever bears this token is authorised". Other schemes exist,
like `Basic` (base64 `user:password`) and `ApiKey` (you'll see it with webhooks). The
scheme name is case-insensitive, so accept `bearer` too.

```go
var ErrNoToken = errors.New("missing bearer token")

func getBearerToken(h http.Header) (string, error) {
	scheme, token, ok := strings.Cut(h.Get("Authorization"), " ")
	token = strings.TrimSpace(token)
	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return "", ErrNoToken
	}
	return token, nil
}
```

## The middleware

```go
func (cfg *apiConfig) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := getBearerToken(r.Header)
		if err != nil {
			unauthorized(w, "missing bearer token")
			return
		}
		userID, err := cfg.validateToken(token)
		if err != nil {
			unauthorized(w, "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey{}, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func unauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="squeak"`)
	respondWithError(w, http.StatusUnauthorized, msg)
}
```

It's the request ID pattern from chapter 3 again: middleware does the work once and
stores the result in the context under a private key. Handlers read it with a helper:

```go
type userIDKey struct{}

func userIDFrom(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey{}).(uuid.UUID)
	return id, ok
}
```

The `WWW-Authenticate` header is what the HTTP spec expects with a 401: it tells the
client *how* to authenticate. Well-behaved clients use it, and it costs one line.

## Applying it to routes

Only some routes need auth. Wrap those, and leave login, registration and public reads
alone:

```go
mux.HandleFunc("POST /api/users", cfg.handleCreateUser)
mux.HandleFunc("POST /api/login", cfg.handleLogin)
mux.HandleFunc("GET /api/squeaks", cfg.handleListSqueaks)

auth := cfg.requireAuth
mux.Handle("GET /api/users/me", auth(http.HandlerFunc(cfg.handleMe)))
mux.Handle("POST /api/squeaks", auth(http.HandlerFunc(cfg.handleCreateSqueak)))
mux.Handle("DELETE /api/squeaks/{id}", auth(http.HandlerFunc(cfg.handleDeleteSqueak)))
```

Handlers behind `requireAuth` can rely on the user ID being there. The `ok` from
`userIDFrom` still deserves a check: if someone registers a handler and forgets the
middleware, you want a loud 500, not a squeak posted by `uuid.Nil()`.

```go
func (cfg *apiConfig) handleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "auth middleware missing")
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"id": userID.String()})
}
```

## Where do validateToken and the secret come from?

The `validateToken` function from the last exercise needs the signing secret and the
current time. Squeak's `apiConfig` stores the secret (loaded from an environment variable
in the final chapter) and exposes a method that closes over it. In this exercise it's a
plain function field on `apiConfig`, so tests can plug in a fake validator:

```go
type apiConfig struct {
	validateToken func(token string) (uuid.UUID, error)
}
```

## Your task

1. **`getBearerToken`** returns the token from an `Authorization: Bearer <token>` header.
   The scheme is case-insensitive, and surrounding spaces around the token are trimmed.
   A missing header, a different scheme or an empty token returns `ErrNoToken`.
2. **`requireAuth`** responds **401** with a JSON error body and a `WWW-Authenticate`
   header starting with `Bearer` when the token is missing or `cfg.validateToken` fails.
   Otherwise it stores the user ID in the context under `userIDKey{}` and calls `next`.
3. **`userIDFrom`** returns the user ID from the context, and `false` if there isn't one.

`handleMe` and the router are already written. **Run** shows four requests: with the
starter, all of them wrongly get through.
