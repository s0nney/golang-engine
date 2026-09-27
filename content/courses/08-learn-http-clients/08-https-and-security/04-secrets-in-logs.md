---
title: Keeping Secrets Out of Logs
quiz:
  - question: |
      What does this print when nothing is listening on port 1?

      ```go
      _, err := http.Get("http://127.0.0.1:1/issues?api_key=tk_live_123")
      fmt.Println(err)
      ```
    options:
      - text: '`connection refused`'
      - text: 'An error that includes the full URL, `api_key=tk_live_123` and all'
        correct: true
      - text: An error with the query string removed
    explanation: |
      Client errors are `*url.Error` values whose message quotes the URL. Go hides a
      password in the userinfo (`ana:***@`), but it can't know that `api_key` is secret.
      Anything that logs `err` logs the key.
  - question: 'Why does `safeURL` work on `u.Clone()` instead of changing `req.URL` directly?'
    options:
      - text: '`req.URL` is read-only'
      - text: The request still has to go out with the real key. Only the log line should be redacted.
        correct: true
      - text: '`Clone` is faster'
    explanation: |
      A logging layer must never change what it logs. Redacting `req.URL` in place would
      send `REDACTED` to the server and break authentication. Clone first, and redact
      the copy.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"net/url"
    	"os"
    )

    // secretParams are query parameters whose values must never be logged.
    var secretParams = []string{"api_key", "token", "access_token"}

    // safeURL returns u as a string that's safe to log: any password in the
    // userinfo is hidden, and the value of every secret query parameter is
    // replaced with "REDACTED". It must not modify u.
    func safeURL(u *url.URL) string {
    	// ?
    	return u.String()
    }

    // loggingTransport logs one line per request, then passes it on to next.
    type loggingTransport struct {
    	next http.RoundTripper
    	out  io.Writer
    }

    // RoundTrip is finished for you: it sends the request, then logs
    // "GET <safe URL> -> 200 OK" or "GET <safe URL> -> error: <err>".
    func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    	resp, err := t.next.RoundTrip(req)
    	if err != nil {
    		fmt.Fprintf(t.out, "%s %s -> error: %v\n", req.Method, safeURL(req.URL), err)
    		return nil, err
    	}
    	fmt.Fprintf(t.out, "%s %s -> %s\n", req.Method, safeURL(req.URL), resp.Status)
    	return resp, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprintf(w, "server saw api_key=%s", r.URL.Query().Get("api_key"))
    	}))
    	defer srv.Close()

    	client := &http.Client{Transport: &loggingTransport{next: http.DefaultTransport, out: os.Stdout}}
    	resp, err := client.Get(srv.URL + "/issues?state=open&api_key=tk_live_123")
    	if err != nil {
    		fmt.Println(err)
    		return
    	}
    	defer resp.Body.Close()
    	body, _ := io.ReadAll(resp.Body)
    	fmt.Println(string(body))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"net/url"
    	"os"
    )

    // secretParams are query parameters whose values must never be logged.
    var secretParams = []string{"api_key", "token", "access_token"}

    func safeURL(u *url.URL) string {
    	c := u.Clone()
    	q := c.Query()
    	for _, name := range secretParams {
    		if q.Has(name) {
    			q.Set(name, "REDACTED")
    		}
    	}
    	c.RawQuery = q.Encode()
    	return c.Redacted()
    }

    // loggingTransport logs one line per request, then passes it on to next.
    type loggingTransport struct {
    	next http.RoundTripper
    	out  io.Writer
    }

    // RoundTrip is finished for you: it sends the request, then logs
    // "GET <safe URL> -> 200 OK" or "GET <safe URL> -> error: <err>".
    func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    	resp, err := t.next.RoundTrip(req)
    	if err != nil {
    		fmt.Fprintf(t.out, "%s %s -> error: %v\n", req.Method, safeURL(req.URL), err)
    		return nil, err
    	}
    	fmt.Fprintf(t.out, "%s %s -> %s\n", req.Method, safeURL(req.URL), resp.Status)
    	return resp, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprintf(w, "server saw api_key=%s", r.URL.Query().Get("api_key"))
    	}))
    	defer srv.Close()

    	client := &http.Client{Transport: &loggingTransport{next: http.DefaultTransport, out: os.Stdout}}
    	resp, err := client.Get(srv.URL + "/issues?state=open&api_key=tk_live_123")
    	if err != nil {
    		fmt.Println(err)
    		return
    	}
    	defer resp.Body.Close()
    	body, _ := io.ReadAll(resp.Body)
    	fmt.Println(string(body))
    }
  tests: |
    package main

    import (
    	"bytes"
    	"net/http"
    	"net/http/httptest"
    	"net/url"
    	"strings"
    	"testing"
    )

    func TestSafeURL(t *testing.T) {
    	tests := []struct {
    		in       string
    		want     []string
    		mustSkip []string
    	}{
    		{"https://api.trackr.dev/issues?state=open", []string{"/issues", "state=open"}, nil},
    		{"https://api.trackr.dev/issues?api_key=tk_live_123&state=open",
    			[]string{"api_key=REDACTED", "state=open"}, []string{"tk_live_123"}},
    		{"https://api.trackr.dev/me?token=abc&access_token=def",
    			[]string{"token=REDACTED", "access_token=REDACTED"}, []string{"abc", "def"}},
    		{"https://ana:hunter2@api.trackr.dev/issues",
    			[]string{"ana", "api.trackr.dev/issues"}, []string{"hunter2"}},
    		{"https://api.trackr.dev/search?q=token", []string{"q=token"}, nil},
    	}
    	for _, tt := range tests {
    		u, err := url.Parse(tt.in)
    		if err != nil {
    			t.Fatal(err)
    		}
    		got := safeURL(u)
    		for _, w := range tt.want {
    			if !strings.Contains(got, w) {
    				t.Errorf("safeURL(%s) = %s, want it to contain %q", tt.in, got, w)
    			}
    		}
    		for _, s := range tt.mustSkip {
    			if strings.Contains(got, s) {
    				t.Errorf("safeURL(%s) = %s, LEAKS the secret %q", tt.in, got, s)
    			}
    		}
    		if u.String() != tt.in {
    			t.Errorf("safeURL changed its argument: URL is now %s, was %s (work on a Clone)", u, tt.in)
    		}
    	}
    }

    func TestTransportLogsSafelyAndSendsRealRequest(t *testing.T) {
    	var sawKey, sawAuth string
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		sawKey = r.URL.Query().Get("api_key")
    		sawAuth = r.Header.Get("Authorization")
    		w.WriteHeader(http.StatusTeapot)
    	}))
    	defer srv.Close()

    	var log bytes.Buffer
    	client := &http.Client{Transport: &loggingTransport{next: http.DefaultTransport, out: &log}}
    	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/issues?api_key=SECRET-KEY&state=open", nil)
    	req.Header.Set("Authorization", "Bearer SECRET-TOKEN")
    	resp, err := client.Do(req)
    	if err != nil {
    		t.Fatalf("request failed: %v", err)
    	}
    	resp.Body.Close()

    	if sawKey != "SECRET-KEY" {
    		t.Errorf("server saw api_key=%q, want the real key: redact the log line, not the request", sawKey)
    	}
    	if sawAuth != "Bearer SECRET-TOKEN" {
    		t.Errorf("server saw Authorization %q, want the real header", sawAuth)
    	}
    	line := log.String()
    	if strings.Contains(line, "SECRET") {
    		t.Errorf("log output leaks a secret:\n%s", line)
    	}
    	for _, w := range []string{"GET", "/issues", "state=open", "api_key=REDACTED", "418"} {
    		if !strings.Contains(line, w) {
    			t.Errorf("log output %q should contain %q", line, w)
    		}
    	}
    }
---

Logs are copied, shipped, indexed, pasted into bug reports and kept for years. A token
that lands in a log is effectively published. HTTP clients are especially leaky,
because the things you most want to log (the URL, the headers) are where credentials
live.

## What must never be logged

- `Authorization`, `Cookie`, `Set-Cookie`, `Proxy-Authorization`, and API-key headers
  like `X-Api-Key`;
- secret query parameters: `api_key`, `token`, `access_token` and friends;
- passwords in URLs (`https://user:pass@...`);
- request bodies of login, token and password-change calls.

## Leak #1: error messages

Every error from `client.Do` is a `*url.Error`, and its message includes the URL:

```
Get "http://ana:***@127.0.0.1:1/issues?api_key=tk_live_123": dial tcp 127.0.0.1:1: connect: connection refused
```

Go masks the userinfo password as `***`, but the query is printed as is. So
`log.Println(err)` just leaked the key. This is the strongest argument for putting
credentials in **headers** (chapter 5): the error text never includes headers.

## Leak #2: printing the whole request

`fmt.Printf("%+v", req)` or `httputil.DumpRequestOut(req, true)` print every header,
`Authorization` included. They're great for a quick local debug session and must never
be left in code that runs for real.

## Make secrets hard to print

Give tokens their own type that refuses to print itself:

```go
// Token is a secret API token. It refuses to be printed or logged.
type Token string

func (Token) String() string       { return "[REDACTED]" }
func (Token) LogValue() slog.Value { return slog.StringValue("[REDACTED]") }
func (t Token) Header() string     { return "Bearer " + string(t) }
```

```go
tok := Token("tk_live_8f3a")
logger.Info("calling Trackr", "token", tok, "project", "apollo")
fmt.Println("token is", tok)
// time=... level=INFO msg="calling Trackr" token=[REDACTED] project=apollo
// token is [REDACTED]
```

`fmt` uses the `String` method, and `log/slog` uses `LogValue` (the `slog.LogValuer`
interface). The only way to get the real value is to ask for it explicitly with
`tok.Header()` or `string(tok)`, which is easy to spot in code review. (`%#v` still
shows the raw value, so a truly paranoid type also implements `GoString`.)

## Log through a RoundTripper

The cleanest place for request logging is the **transport**. An `http.RoundTripper`
is anything with one method:

```go
type RoundTripper interface {
	RoundTrip(*http.Request) (*http.Response, error)
}
```

`http.Transport` is one. You can wrap it in your own to add behaviour (logging,
metrics, auth headers, even retries) to every request a client makes:

```go
client := &http.Client{
	Transport: &loggingTransport{next: http.DefaultTransport, out: os.Stderr},
}
```

Two rules for RoundTrippers, straight from the docs: **don't modify the request** (the
caller still owns it), and **don't interpret the response status** (a 404 is not an
error at this level). A logging transport just observes.

## Your turn

`loggingTransport.RoundTrip` is written. It logs lines like this:

```
GET https://api.trackr.dev/issues?api_key=REDACTED&state=open -> 200 OK
```

using `safeURL` to turn the URL into something safe to print. Right now `safeURL` just
returns `u.String()`, so run the program and watch the key leak into the log.

Complete `safeURL(u)` so that it:

1. works on `u.Clone()` and never modifies `u` (the request must still carry the
   real key);
2. replaces the value of every parameter named in `secretParams` that's present in the
   query with `REDACTED`, leaving other parameters alone;
3. hides any userinfo password (`(*url.URL).Redacted` does exactly that);
4. returns the result as a string.

Hint: read the clone's query with `Query()`, `Set` the secret ones, and write it back
with `RawQuery = q.Encode()`, just like chapter 2.
