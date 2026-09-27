---
title: Authentication Headers
quiz:
  - question: Which is the correct way to send a bearer token?
    options:
      - text: '`req.Header.Set("Authorization", token)`'
      - text: '`req.Header.Set("Authorization", "Bearer "+token)`'
        correct: true
      - text: '`req.Header.Set("Bearer", token)`'
      - text: '`req.URL.RawQuery = "token=" + token`'
    explanation: |
      The `Authorization` header holds a *scheme* and then the credentials, separated
      by a space: `Bearer tk_live_...`. Forgetting the `Bearer ` prefix is a very common
      cause of 401s.
  - question: |
      `trackr` sends `X-Api-Key: tk_live_123` to `api.trackr.dev`, which answers with a
      redirect to `cdn.example.net`. What does Go's client do with the header?
    options:
      - text: Drops it, like it drops `Authorization`
      - text: Forwards it to `cdn.example.net`
        correct: true
      - text: Refuses to follow the redirect
    explanation: |
      Go strips `Authorization`, `WWW-Authenticate` and `Cookie` when a redirect leaves
      the original domain, but it has no idea that `X-Api-Key` is secret, so custom
      headers are copied along. Keep that in mind when you use API-key headers.
exercise:
  starter: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    )

    // Client talks to the Trackr API as one user.
    type Client struct {
    	baseURL string
    	token   string
    }

    // newRequest builds a request for c.baseURL + path with the headers every
    // Trackr call needs:
    //
    //	Authorization: Bearer <token>
    //	Accept: application/json
    //	User-Agent: trackr/1.4
    //
    // It returns an error, without building anything, if c.token is empty.
    func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
    	// ?
    	return http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
    }

    // whoami returns the username of the token's owner, from GET /me.
    // It's finished: it relies on newRequest.
    func (c *Client) whoami(ctx context.Context) (string, error) {
    	req, err := c.newRequest(ctx, http.MethodGet, "/me", nil)
    	if err != nil {
    		return "", err
    	}
    	resp, err := http.DefaultClient.Do(req)
    	if err != nil {
    		return "", err
    	}
    	defer resp.Body.Close()

    	if resp.StatusCode == http.StatusUnauthorized {
    		return "", errors.New("your token was rejected: run `trackr login`")
    	}
    	if resp.StatusCode != http.StatusOK {
    		return "", fmt.Errorf("GET /me: unexpected status %s", resp.Status)
    	}
    	var me struct {
    		Username string `json:"username"`
    	}
    	if err := json.UnmarshalRead(resp.Body, &me); err != nil {
    		return "", err
    	}
    	return me.Username, nil
    }

    func main() {
    	// A fake Trackr API that only answers callers with the right token.
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Printf("  (server: Authorization=%q User-Agent=%q)\n", r.Header.Get("Authorization"), r.Header.Get("User-Agent"))
    		if r.Header.Get("Authorization") != "Bearer tk_test_42" {
    			w.WriteHeader(http.StatusUnauthorized)
    			return
    		}
    		fmt.Fprint(w, `{"username": "ana"}`)
    	}))
    	defer srv.Close()

    	for _, token := range []string{"tk_test_42", "tk_wrong", ""} {
    		c := &Client{baseURL: srv.URL, token: token}
    		name, err := c.whoami(context.Background())
    		fmt.Printf("token %q -> %q %v\n", token, name, err)
    	}
    }
  solution: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    )

    // Client talks to the Trackr API as one user.
    type Client struct {
    	baseURL string
    	token   string
    }

    func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
    	if c.token == "" {
    		return nil, errors.New("not logged in: set TRACKR_TOKEN")
    	}
    	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
    	if err != nil {
    		return nil, err
    	}
    	req.Header.Set("Authorization", "Bearer "+c.token)
    	req.Header.Set("Accept", "application/json")
    	req.Header.Set("User-Agent", "trackr/1.4")
    	return req, nil
    }

    // whoami returns the username of the token's owner, from GET /me.
    // It's finished: it relies on newRequest.
    func (c *Client) whoami(ctx context.Context) (string, error) {
    	req, err := c.newRequest(ctx, http.MethodGet, "/me", nil)
    	if err != nil {
    		return "", err
    	}
    	resp, err := http.DefaultClient.Do(req)
    	if err != nil {
    		return "", err
    	}
    	defer resp.Body.Close()

    	if resp.StatusCode == http.StatusUnauthorized {
    		return "", errors.New("your token was rejected: run `trackr login`")
    	}
    	if resp.StatusCode != http.StatusOK {
    		return "", fmt.Errorf("GET /me: unexpected status %s", resp.Status)
    	}
    	var me struct {
    		Username string `json:"username"`
    	}
    	if err := json.UnmarshalRead(resp.Body, &me); err != nil {
    		return "", err
    	}
    	return me.Username, nil
    }

    func main() {
    	// A fake Trackr API that only answers callers with the right token.
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Printf("  (server: Authorization=%q User-Agent=%q)\n", r.Header.Get("Authorization"), r.Header.Get("User-Agent"))
    		if r.Header.Get("Authorization") != "Bearer tk_test_42" {
    			w.WriteHeader(http.StatusUnauthorized)
    			return
    		}
    		fmt.Fprint(w, `{"username": "ana"}`)
    	}))
    	defer srv.Close()

    	for _, token := range []string{"tk_test_42", "tk_wrong", ""} {
    		c := &Client{baseURL: srv.URL, token: token}
    		name, err := c.whoami(context.Background())
    		fmt.Printf("token %q -> %q %v\n", token, name, err)
    	}
    }
  tests: |
    package main

    import (
    	"context"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"testing"
    )

    func TestNewRequestHeaders(t *testing.T) {
    	c := &Client{baseURL: "https://api.trackr.dev/v1", token: "tk_live_8f3a"}
    	req, err := c.newRequest(context.Background(), http.MethodPost, "/projects/apollo/issues", strings.NewReader(`{"title":"x"}`))
    	if err != nil {
    		t.Fatalf("newRequest returned error %v", err)
    	}
    	if req.Method != http.MethodPost {
    		t.Errorf("method = %q, want POST", req.Method)
    	}
    	if got, want := req.URL.String(), "https://api.trackr.dev/v1/projects/apollo/issues"; got != want {
    		t.Errorf("URL = %q, want %q", got, want)
    	}
    	for name, want := range map[string]string{
    		"Authorization": "Bearer tk_live_8f3a",
    		"Accept":        "application/json",
    		"User-Agent":    "trackr/1.4",
    	} {
    		if got := req.Header.Get(name); got != want {
    			t.Errorf("header %s = %q, want %q", name, got, want)
    		}
    	}
    	if body, _ := io.ReadAll(req.Body); string(body) != `{"title":"x"}` {
    		t.Errorf("request body = %q, want the body that was passed in", body)
    	}
    }

    func TestNewRequestNoToken(t *testing.T) {
    	c := &Client{baseURL: "https://api.trackr.dev", token: ""}
    	req, err := c.newRequest(context.Background(), http.MethodGet, "/me", nil)
    	if err == nil {
    		t.Errorf("newRequest with an empty token = (%v, nil), want an error (don't send requests that can only fail with 401)", req.URL)
    	}
    }

    func TestWhoami(t *testing.T) {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		if r.URL.Path != "/me" {
    			http.NotFound(w, r)
    			return
    		}
    		if r.Header.Get("Authorization") != "Bearer tk_test_bo" {
    			w.WriteHeader(http.StatusUnauthorized)
    			return
    		}
    		io.WriteString(w, `{"username": "bo"}`)
    	}))
    	defer srv.Close()

    	good := &Client{baseURL: srv.URL, token: "tk_test_bo"}
    	if name, err := good.whoami(context.Background()); err != nil || name != "bo" {
    		t.Errorf("whoami with a valid token = (%q, %v), want (\"bo\", nil)", name, err)
    	}
    	bad := &Client{baseURL: srv.URL, token: "tk_test_eve"}
    	if name, err := bad.whoami(context.Background()); err == nil {
    		t.Errorf("whoami with a rejected token = (%q, nil), want an error", name)
    	}
    }
---

Most of the Trackr API needs to know **who** is calling. Without credentials you get
`401 Unauthorized`. With valid ones that don't allow the action, you get
`403 Forbidden`. Credentials travel in headers.

## Bearer tokens

The most common scheme today. You get a token from the API's settings page (or from an
OAuth flow) and send it in the `Authorization` header with the word `Bearer` in front:

```go
req.Header.Set("Authorization", "Bearer "+token)
```

```
Authorization: Bearer tk_live_8f3a...
```

"Bearer" means "whoever holds this token is allowed in", so treat a token like a
password. Anyone who copies it *is* you as far as the API is concerned.

## API keys

Some APIs use their own header instead:

```go
req.Header.Set("X-Api-Key", key)
```

The name varies (`X-Api-Key`, `Api-Key`, `X-Trackr-Token` and so on). Read the API's
docs. Functionally it's the same idea as a bearer token.

## Basic auth

The oldest scheme sends a username and password, base64-encoded:

```go
req.SetBasicAuth("ana", "s3cret")
fmt.Println(req.Header.Get("Authorization"))
// Basic YW5hOnMzY3JldA==
```

Base64 is **encoding, not encryption**. Anyone who sees that header can decode it in
a second. Basic auth is only acceptable over HTTPS (chapter 8), and so are the other
schemes.

## Where tokens should come from

Never hard-code a token in source code, because it ends up in git history forever.
`trackr` reads it from the environment or a config file:

```go
token := os.Getenv("TRACKR_TOKEN")
if token == "" {
	return errors.New("TRACKR_TOKEN is not set; create a token at https://trackr.dev/settings/tokens")
}
```

## Where tokens should NOT go

- **Not in the URL.** `?api_key=...` ends up in server access logs, proxy logs, shell
  history and error messages. Remember that `url.Error` includes the full URL in its
  text: `Get "https://api.trackr.dev/issues?api_key=tk_live_123": ...`.
- **Not in logs.** Chapter 8 shows how to log requests without leaking headers.

## Redirects and credentials

When a server redirects you to a **different domain**, Go's client drops the
`Authorization`, `WWW-Authenticate` and `Cookie` headers you set, so your Trackr token
isn't handed to someone else's server. Redirects within the same domain (or to a
subdomain of it) keep them.

Custom headers are a different story. Go doesn't know `X-Api-Key` is a secret, so it
**forwards** it on every redirect. Here's a fake API that redirects `/other-host` to a
different hostname:

```
B 127.0.0.1:33351 got Authorization="Bearer secret" X-Api-Key="k"   (same host)
B localhost:33351 got Authorization="" X-Api-Key="k"                (different host)
```

If you must use a custom key header and don't trust where redirects might lead, set
`Client.CheckRedirect` to stop following them (chapter 6 shows how to build your own
client).

## Adding auth to every request

Setting the header in every function gets repetitive, and sooner or later one function
forgets. `trackr` centralizes it in one place that builds requests:

```go
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "trackr/1.4")
	return req, nil
}
```

Every API call goes through `newRequest`, so none of them can forget the token.

## Your turn

`trackr whoami` prints who the configured token belongs to. `whoami` is written, and
it builds its request with `newRequest`, which doesn't add any headers yet. Run the
program and watch the fake server reject every call.

Complete `newRequest` so that it:

1. returns an error straight away if `c.token` is empty (a request without a token can
   only earn a 401, so say "not logged in" instead);
2. builds the request for `c.baseURL + path` with `http.NewRequestWithContext`,
   passing on `ctx`, `method` and `body`;
3. sets `Authorization: Bearer <token>`, `Accept: application/json` and
   `User-Agent: trackr/1.4`.
