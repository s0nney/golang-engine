---
title: Debug Dump
difficulty: medium
after: https-and-security
hints:
  - 'For the URL, work on a copy (`u := *req.URL`, or Go 1.27''s `req.URL.Clone()`) so the request is untouched. `u.Redacted()` hides a userinfo password. For the query, `q := u.Query()`, replace secret values with `REDACTED`, then `u.RawQuery = q.Encode()`.'
  - '`slices.Sorted(maps.Keys(req.Header))` gives the header names in order. Keep a helper `redact(name, value string) string` that knows the rules, including keeping the scheme of `Authorization` (`strings.Cut(value, " ")`).'
  - 'Reading `req.Body` uses it up. Read it all with `io.ReadAll`, print it, then put a fresh reader back: `req.Body = io.NopCloser(bytes.NewReader(b))`. The server in the tests checks it gets the full body.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"io"
    	"net/http"
    	"os"
    	"strings"
    )

    func dumpRequest(w io.Writer, req *http.Request) error {
    	return nil
    }

    func main() {
    	req, err := http.NewRequest(http.MethodPost, "https://ana:hunter2@api.trackr.dev/v1/issues?api_key=k_9f2c&project=apollo",
    		strings.NewReader(`{"title": "Dark mode"}`))
    	if err != nil {
    		fmt.Println(err)
    		return
    	}
    	req.Header.Set("Authorization", "Bearer tk_live_8f3a")
    	req.Header.Set("Content-Type", "application/json")
    	req.Header.Set("X-Trackr-Session-Token", "s_77aa")
    	fmt.Println("--- trackr --debug ---")
    	if err := dumpRequest(os.Stdout, req); err != nil {
    		fmt.Println(err)
    	}
    	// want:
    	// --- trackr --debug ---
    	// POST https://ana:xxxxx@api.trackr.dev/v1/issues?api_key=REDACTED&project=apollo
    	// Authorization: Bearer REDACTED
    	// Content-Type: application/json
    	// X-Trackr-Session-Token: REDACTED
    	//
    	// {"title": "Dark mode"}
    }
  solution: |
    package main

    import (
    	"bytes"
    	"fmt"
    	"io"
    	"maps"
    	"net/http"
    	"os"
    	"slices"
    	"strings"
    )

    // secretParams are query parameters whose values are never logged.
    var secretParams = []string{"api_key", "access_token", "token", "sig"}

    func dumpRequest(w io.Writer, req *http.Request) error {
    	u := req.URL.Clone()
    	q := u.Query()
    	for _, name := range secretParams {
    		if q.Has(name) {
    			q.Set(name, "REDACTED")
    		}
    	}
    	u.RawQuery = q.Encode()
    	if _, err := fmt.Fprintf(w, "%s %s\n", req.Method, u.Redacted()); err != nil {
    		return err
    	}

    	for _, name := range slices.Sorted(maps.Keys(req.Header)) {
    		for _, v := range req.Header[name] {
    			if _, err := fmt.Fprintf(w, "%s: %s\n", name, redactHeader(name, v)); err != nil {
    				return err
    			}
    		}
    	}

    	if req.Body == nil || req.Body == http.NoBody {
    		return nil
    	}
    	b, err := io.ReadAll(req.Body)
    	req.Body.Close()
    	req.Body = io.NopCloser(bytes.NewReader(b)) // put it back for whoever sends req
    	if err != nil {
    		return err
    	}
    	_, err = fmt.Fprintf(w, "\n%s\n", b)
    	return err
    }

    func redactHeader(name, value string) string {
    	switch name {
    	case "Authorization", "Proxy-Authorization":
    		if scheme, _, ok := strings.Cut(value, " "); ok {
    			return scheme + " REDACTED"
    		}
    		return "REDACTED"
    	case "Cookie", "Set-Cookie":
    		return "REDACTED"
    	}
    	lower := strings.ToLower(name)
    	for _, word := range []string{"token", "key", "secret"} {
    		if strings.Contains(lower, word) {
    			return "REDACTED"
    		}
    	}
    	return value
    }

    func main() {
    	req, err := http.NewRequest(http.MethodPost, "https://ana:hunter2@api.trackr.dev/v1/issues?api_key=k_9f2c&project=apollo",
    		strings.NewReader(`{"title": "Dark mode"}`))
    	if err != nil {
    		fmt.Println(err)
    		return
    	}
    	req.Header.Set("Authorization", "Bearer tk_live_8f3a")
    	req.Header.Set("Content-Type", "application/json")
    	req.Header.Set("X-Trackr-Session-Token", "s_77aa")
    	fmt.Println("--- trackr --debug ---")
    	if err := dumpRequest(os.Stdout, req); err != nil {
    		fmt.Println(err)
    	}
    }
  tests: |
    package main

    import (
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync"
    	"testing"
    )

    type testCase struct {
    	name, method, url, body string
    	headers                 [][2]string // added in order with Header.Add
    	want                    string
    }

    var testCases = []testCase{
    	{
    		name: "POST with secrets everywhere", method: "POST",
    		url:  "https://ana:hunter2@api.trackr.dev/v1/issues?api_key=k_9f2c&project=apollo",
    		body: `{"title": "Dark mode"}`,
    		headers: [][2]string{{"Authorization", "Bearer tk_live_8f3a"}, {"Content-Type", "application/json"},
    			{"X-Trackr-Session-Token", "s_77aa"}},
    		want: "POST https://ana:xxxxx@api.trackr.dev/v1/issues?api_key=REDACTED&project=apollo\n" +
    			"Authorization: Bearer REDACTED\n" +
    			"Content-Type: application/json\n" +
    			"X-Trackr-Session-Token: REDACTED\n" +
    			"\n" +
    			"{\"title\": \"Dark mode\"}\n",
    	},
    	{
    		name: "GET without a body", method: "GET",
    		url:     "https://api.trackr.dev/v1/search?q=caf%C3%A9+%26+cr%C3%A8me&access_token=a1&access_token=a2&sig=zz",
    		headers: [][2]string{{"Accept", "application/json"}, {"User-Agent", "trackr/1.4"}, {"Cookie", "session=abc; theme=dark"}},
    		want: "GET https://api.trackr.dev/v1/search?access_token=REDACTED&q=caf%C3%A9+%26+cr%C3%A8me&sig=REDACTED\n" +
    			"Accept: application/json\n" +
    			"Cookie: REDACTED\n" +
    			"User-Agent: trackr/1.4\n",
    	},
    	{
    		name: "odd headers", method: "PATCH",
    		url:  "https://bot@api.trackr.dev/v1/issues/42?token=t0&page=2",
    		body: `{"state": "closed"}`,
    		headers: [][2]string{{"Authorization", "tk_live_nospace"}, {"X-Api-Key", "k1"}, {"X-Webhook-Secret", "shh"},
    			{"Proxy-Authorization", "Basic YWxhZGRpbjpvcGVuc2VzYW1l"}, {"Accept", "application/json"}, {"Accept", "text/plain"}},
    		want: "PATCH https://bot@api.trackr.dev/v1/issues/42?page=2&token=REDACTED\n" +
    			"Accept: application/json\n" +
    			"Accept: text/plain\n" +
    			"Authorization: REDACTED\n" +
    			"Proxy-Authorization: Basic REDACTED\n" +
    			"X-Api-Key: REDACTED\n" +
    			"X-Webhook-Secret: REDACTED\n" +
    			"\n" +
    			"{\"state\": \"closed\"}\n",
    	},
    	{
    		name: "nothing to hide", method: "DELETE",
    		url:  "http://127.0.0.1:8080/v1/issues/7/labels/good%20first%20issue",
    		want: "DELETE http://127.0.0.1:8080/v1/issues/7/labels/good%20first%20issue\n",
    	},
    }

    func testBuild(t *testing.T, tc testCase, base string) *http.Request {
    	t.Helper()
    	var body io.Reader
    	if tc.body != "" {
    		body = strings.NewReader(tc.body)
    	}
    	u := tc.url
    	if base != "" { // point the request at the test server, keeping path and query
    		u = base + tc.url[strings.Index(tc.url[len("https://"):], "/")+len("https://"):]
    	}
    	req, err := http.NewRequestWithContext(t.Context(), tc.method, u, body)
    	if err != nil {
    		t.Fatal(err)
    	}
    	for _, h := range tc.headers {
    		req.Header.Add(h[0], h[1])
    	}
    	return req
    }

    func TestDumpRequest(t *testing.T) {
    	for _, tc := range testCases {
    		var out strings.Builder
    		if err := dumpRequest(&out, testBuild(t, tc, "")); err != nil {
    			t.Errorf("%s: dumpRequest returned %v", tc.name, err)
    		}
    		if out.String() != tc.want {
    			t.Errorf("%s: dumpRequest wrote\n%s\nwant\n%s", tc.name, out.String(), tc.want)
    		}
    		for _, secret := range []string{"hunter2", "k_9f2c", "tk_live", "s_77aa", "a1", "a2", "abc", "t0", "k1", "shh", "YWxh"} {
    			if strings.Contains(out.String(), secret) && !strings.Contains(tc.want, secret) {
    				t.Errorf("%s: output leaks %q", tc.name, secret)
    			}
    		}
    	}
    }

    func TestDumpLeavesRequestUsable(t *testing.T) {
    	type seen struct {
    		method, uri, body string
    		header            http.Header
    	}
    	var mu sync.Mutex
    	var got seen
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		b, _ := io.ReadAll(r.Body)
    		mu.Lock()
    		got = seen{r.Method, r.RequestURI, string(b), r.Header.Clone()}
    		mu.Unlock()
    	}))
    	defer srv.Close()

    	for _, tc := range testCases {
    		req := testBuild(t, tc, srv.URL)
    		urlBefore := req.URL.String()
    		if err := dumpRequest(io.Discard, req); err != nil {
    			t.Errorf("%s: dumpRequest returned %v", tc.name, err)
    		}
    		if req.URL.String() != urlBefore {
    			t.Errorf("%s: dumpRequest changed req.URL from %s to %s; redact a copy", tc.name, urlBefore, req.URL)
    		}
    		resp, err := srv.Client().Do(req)
    		if err != nil {
    			t.Errorf("%s: sending the request after dumpRequest failed: %v", tc.name, err)
    			continue
    		}
    		resp.Body.Close()
    		mu.Lock()
    		g := got
    		mu.Unlock()
    		if g.body != tc.body {
    			t.Errorf("%s: after dumpRequest the server received body %q, want %q (put the body back after reading it)", tc.name, g.body, tc.body)
    		}
    		wantURI := urlBefore[len(srv.URL):]
    		if g.method != tc.method || g.uri != wantURI {
    			t.Errorf("%s: server received %s %s, want %s %s", tc.name, g.method, g.uri, tc.method, wantURI)
    		}
    		for _, h := range tc.headers {
    			if vals := g.header.Values(h[0]); !strings.Contains(strings.Join(vals, "\n"), h[1]) {
    				t.Errorf("%s: server received %s: %q, want the real value %q (don't redact the request itself)", tc.name, h[0], vals, h[1])
    			}
    		}
    	}
    }
---

`trackr --debug` prints every request before sending it, so users can paste it
into a bug report. Bug reports are public, so the dump must show everything
**except** secrets, and must not break the request it describes.

Complete `dumpRequest(w, req)`. It writes to `w`:

1. `<METHOD> <URL>` on the first line, where the URL has
   - any userinfo password replaced as `url.URL.Redacted` does (`ana:xxxxx@`),
   - the value of each of the query parameters `api_key`, `access_token`,
     `token` and `sig` replaced by a single `REDACTED`,
   - its query re-encoded with `url.Values.Encode` (so keys come out sorted).
2. One `Name: value` line per header value, with names sorted (a header with
   two values gets two lines, in their original order). Redact these values:
   - `Authorization` and `Proxy-Authorization`: keep the scheme,
     `Bearer tk_live_8f3a` becomes `Bearer REDACTED`. A value with no space
     becomes just `REDACTED`.
   - `Cookie`: `REDACTED`.
   - Any header whose name contains `token`, `key` or `secret`, in any case
     (`X-Api-Key`, `X-Trackr-Session-Token`): `REDACTED`.
3. If the request has a body: an empty line, then the body, then a newline.

**The request must still work afterwards**: same URL, same real header values,
and the full body still there to be sent. Return any error from writing or from
reading the body.

## Example

```
POST https://ana:xxxxx@api.trackr.dev/v1/issues?api_key=REDACTED&project=apollo
Authorization: Bearer REDACTED
Content-Type: application/json
X-Trackr-Session-Token: REDACTED

{"title": "Dark mode"}
```

## Constraints

- A `GET` without a body ends after the header lines, with no blank line.
- After dumping, the tests send each request to a fake API and check that it
  received the original method, path, query, headers and body.
