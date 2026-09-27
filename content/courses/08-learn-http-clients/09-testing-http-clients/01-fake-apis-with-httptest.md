---
title: Fake APIs with httptest
quiz:
  - question: Inside an `httptest` handler, which is safe to call when a request looks wrong?
    options:
      - text: '`t.Fatalf`'
      - text: '`t.Errorf`'
        correct: true
      - text: '`panic`'
    explanation: |
      The handler runs on the server's goroutine, not the test's. `t.Errorf` is safe
      from any goroutine, but `t.Fatal` (and `FailNow`) must only be called from the
      goroutine running the test function.
  - question: 'What''s special about a server from `httptest.NewTestServer(t, h)` (Go 1.27)?'
    options:
      - text: It listens on a real port chosen by the OS
      - text: It uses an in-memory network, `srv.Client()` sends every request to it whatever the hostname, and it shuts down automatically when the test ends
        correct: true
      - text: It can only serve HTTPS
    explanation: |
      No port, no loopback networking, and cleanup is registered with `t.Cleanup`. It
      also works inside `testing/synctest` bubbles, which you'll use for retries.
---

You've been using `httptest` since chapter 1 so the exercises could run without the
internet. Now it's time to use it on purpose: to **test** `trackr`.

## Why not test against the real API?

- It needs network access and valid credentials.
- It's slow, and it's flaky whenever the network is.
- You can't make it return a 503, a malformed body or a 30-second stall on demand.
- Tests that create issues leave garbage behind.

A fake server fixes all four. It runs in your test process, answers in microseconds,
and does exactly what you tell it.

## httptest.NewServer

`httptest.NewServer(handler)` starts a real HTTP server on `127.0.0.1` with a random
port. `srv.URL` is its base URL. The handler is an ordinary `http.Handler`, where you
both **check the request** and **script the response**:

```go
func TestGetIssue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check what the client sent...
		if r.Method != http.MethodGet || r.URL.Path != "/issues/42" {
			t.Errorf("got %s %s, want GET /issues/42", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
		}
		// ...and send back a canned response.
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id": 42, "title": "Login broken"}`)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "test-token", HTTPClient: srv.Client()}
	iss, err := c.GetIssue(t.Context(), 42)
	if err != nil {
		t.Fatalf("GetIssue: %v", err)
	}
	if iss.ID != 42 || iss.Title != "Login broken" {
		t.Errorf("GetIssue = %+v, want issue 42 \"Login broken\"", iss)
	}
}
```

A few habits worth copying:

- Use **`t.Errorf`** inside the handler, never `t.Fatalf`. The handler runs on a
  server goroutine, and `Fatal` may only be called from the test's own goroutine.
- **`defer srv.Close()`** shuts the server down and waits for in-flight requests.
- **`srv.Client()`** returns a client configured for this server. For a plain HTTP
  server any client works, but for `NewTLSServer` it's the one that trusts the test
  certificate.
- **`t.Context()`** (Go 1.24) is a context that's cancelled when the test ends, which
  is ideal for requests made by the test.

## Routing with ServeMux

A fake that serves several endpoints can use a `ServeMux` with method-and-path
patterns:

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /issues/{id}", func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `{"id": %s, "title": "Issue %s"}`, r.PathValue("id"), r.PathValue("id"))
})
mux.HandleFunc("POST /projects/{slug}/issues", func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, `{"id": 101, "title": "new"}`)
})
srv := httptest.NewServer(mux)
```

Requests that match nothing get a 404, and the wrong method on a known path gets a
405. Both are handy for catching a client that builds the wrong URL.

## New in Go 1.27: NewTestServer

`httptest.NewTestServer(t, handler)` is the new recommended way to get a test server:

```go
srv := httptest.NewTestServer(t, handler)
c := &Client{BaseURL: "https://api.trackr.dev", Token: "t", HTTPClient: srv.Client()}
```

It differs from `NewServer` in three useful ways:

1. It uses an **in-memory network** instead of a real port, so there's no port
   exhaustion and no flakiness from the OS network stack.
2. **`srv.Client()` sends every request to the fake**, whatever the host. You can
   keep the production base URL, `https://api.trackr.dev`, in the test (HTTPS works
   too, over TLS in memory). Only that client is wired up, so always use `srv.Client()`
   or its `Transport`. (`srv.URL` is `http://example.com`, and it's only set once
   `Client` has been called.)
3. It's tied to `t`: it **closes itself** when the test ends, and it fails the test
   if the handler panics.

It also runs inside `testing/synctest` bubbles, which lesson 5 relies on.

`NewServer` is still fine, especially when something outside your Go test (a
subprocess, a browser) has to reach the server over a real port.

## Further reading

- Learn Go with Tests, "Select" (uses httptest for racing URLs): https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/select
