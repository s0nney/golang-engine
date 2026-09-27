---
title: Injecting Base URLs and Clients
quiz:
  - question: |
      Why is this hard to test?

      ```go
      func GetIssue(ctx context.Context, id int) (Issue, error) {
          url := fmt.Sprintf("https://api.trackr.dev/issues/%d", id)
          req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
          resp, err := http.DefaultClient.Do(req)
          // ...
      }
      ```
    options:
      - text: It uses a context
      - text: The base URL and the client are hard-coded, so a test can't point it at a fake server
        correct: true
      - text: '`fmt.Sprintf` is slow'
    explanation: |
      A test has no way in. Make the base URL and the `*http.Client` fields (or
      parameters) so tests can swap in `srv.URL` and `srv.Client()`.
  - question: |
      What does this test double do?

      ```go
      type roundTripFunc func(*http.Request) (*http.Response, error)

      func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
          return f(r)
      }
      ```
    options:
      - text: It starts an HTTP server
      - text: It lets any function act as an `http.RoundTripper`, so a client can be given canned responses without any server
        correct: true
      - text: It retries failed requests
    explanation: |
      It's the same trick as `http.HandlerFunc`: a function type with a method that
      calls itself. Set it as a client's `Transport` and every request goes to your
      function.
  - question: A test fails, and the message says it tried to reach `api.trackr.dev`. What's the most likely cause?
    options:
      - text: The fake server crashed
      - text: Somewhere the code ignored the injected base URL or client and fell back to the real one
        correct: true
      - text: DNS is broken on the test machine
    explanation: |
      That's exactly the kind of bug these tests catch: one code path that builds its
      own URL or uses `http.DefaultClient` instead of the injected client.
---

A fake server is useless if the code under test can't be pointed at it. The fix is an
old idea with a fancy name: **dependency injection**. Pass in the things that vary,
rather than reaching for globals.

## What to inject

For an HTTP client, two things:

1. **The base URL.** Production says `https://api.trackr.dev`, and a test says
   `srv.URL`.
2. **The `*http.Client`.** Production uses one with timeouts and a tuned transport.
   A test uses `srv.Client()`, a client with a fake transport, or a very short timeout.

`trackr`'s API client holds both:

```go
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// New returns a client for the real API.
func New(token string) *Client {
	return &Client{
		BaseURL:    "https://api.trackr.dev",
		Token:      token,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}
```

Tests build their own:

```go
c := &Client{BaseURL: srv.URL, Token: "test-token", HTTPClient: srv.Client()}
```

The standard library's own clients follow this pattern (`http.Client` has a
`Transport` field, `sql.DB` takes a driver), and so do most Go API client libraries.

## Every path must use them

Injection only works if **no** code path goes around it. Watch for these slips:

- `http.Get(...)` or `http.DefaultClient.Do(...)` in one forgotten function;
- a URL built from a hard-coded `"https://api.trackr.dev"` string;
- following a `Link` header or `Location` without keeping to the injected client.

A good fake server catches these: if a request escapes to the real host, the test
fails (or, with no network, errors out).

## Faking at the transport level

Sometimes a whole server is overkill. You just want "the next request returns this".
Because `http.Client.Transport` is an interface, a function can stand in for it:

```go
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestGetIssueNetworkError(t *testing.T) {
	boom := errors.New("connection reset by peer")
	c := &Client{
		BaseURL: "https://api.trackr.dev",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return nil, boom
		})},
	}
	_, err := c.GetIssue(t.Context(), 1)
	if !errors.Is(err, boom) {
		t.Errorf("GetIssue error = %v, want it to wrap %v", err, boom)
	}
}
```

To return a response instead, build one. `httptest.NewRecorder` helps: write to it like
a handler would, then call `Result()`:

```go
rec := httptest.NewRecorder()
rec.WriteHeader(http.StatusTeapot)
rec.WriteString(`{"error": "I'm a teapot"}`)
return rec.Result(), nil
```

Transport fakes are great for exact error values and odd cases. Server fakes are more
realistic: real headers, real connection handling, real body streaming. Most test
suites use both.

## Inject time, too

Anything nondeterministic is a dependency: the clock, randomness (jitter!) and
environment variables. You already saw `retryAfter(h, now, maxWait)` take `now` as a
parameter. For code that *sleeps*, you don't even need to inject a clock: lesson 5
uses `testing/synctest`, which fakes time for everything inside a bubble.

## Configuration as injection

`trackr` itself reads the base URL from config, falling back to the default:

```go
base := cmp.Or(os.Getenv("TRACKR_API_URL"), "https://api.trackr.dev")
```

`cmp.Or` returns its first non-zero argument. Now an end-to-end test can run the real
`trackr` binary against a fake server just by setting an environment variable.
