---
title: Keep the Token Home
difficulty: easy
after: https-and-security
hints:
  - 'The check is `req.URL.Scheme == "https" && req.URL.Host == t.Host`. If it fails, return `t.Next.RoundTrip(req)` unchanged.'
  - 'A RoundTripper must not modify the request it''s given. Make a copy with `r := req.Clone(req.Context())`, set the header on `r.Header`, and send `r`.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    )

    // authTransport adds "Authorization: Bearer <Token>" to requests that go to
    // Host (exactly, including any port) over https, and sends every request on
    // through Next. Requests anywhere else pass through untouched.
    type authTransport struct {
    	Token string
    	Host  string // e.g. "api.trackr.dev" or "127.0.0.1:8443"
    	Next  http.RoundTripper
    }

    func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    	// 1. If the request isn't https to t.Host, just send it with t.Next.
    	// 2. Otherwise clone it, set the Authorization header on the clone,
    	//    and send the clone. Never modify req itself.
    	return t.Next.RoundTrip(req)
    }

    func main() {
    	api := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprintf(w, "api saw Authorization=%q\n", r.Header.Get("Authorization"))
    	}))
    	defer api.Close()
    	other := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprintf(w, "other saw Authorization=%q\n", r.Header.Get("Authorization"))
    	}))
    	defer other.Close()

    	client := &http.Client{Transport: &authTransport{
    		Token: "tk_live_8f3a",
    		Host:  strings.TrimPrefix(api.URL, "https://"),
    		Next:  api.Client().Transport,
    	}}
    	for _, u := range []string{api.URL, other.URL} {
    		resp, err := client.Get(u)
    		if err != nil {
    			fmt.Println(err)
    			continue
    		}
    		body, _ := io.ReadAll(resp.Body)
    		resp.Body.Close()
    		fmt.Print(string(body))
    	}
    	// want:
    	// api saw Authorization="Bearer tk_live_8f3a"
    	// other saw Authorization=""
    }
  solution: |
    package main

    import (
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    )

    // authTransport adds "Authorization: Bearer <Token>" to requests that go to
    // Host (exactly, including any port) over https, and sends every request on
    // through Next. Requests anywhere else pass through untouched.
    type authTransport struct {
    	Token string
    	Host  string // e.g. "api.trackr.dev" or "127.0.0.1:8443"
    	Next  http.RoundTripper
    }

    func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    	if req.URL.Scheme != "https" || req.URL.Host != t.Host {
    		return t.Next.RoundTrip(req)
    	}
    	r := req.Clone(req.Context())
    	r.Header.Set("Authorization", "Bearer "+t.Token)
    	return t.Next.RoundTrip(r)
    }

    func main() {
    	api := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprintf(w, "api saw Authorization=%q\n", r.Header.Get("Authorization"))
    	}))
    	defer api.Close()
    	other := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Fprintf(w, "other saw Authorization=%q\n", r.Header.Get("Authorization"))
    	}))
    	defer other.Close()

    	client := &http.Client{Transport: &authTransport{
    		Token: "tk_live_8f3a",
    		Host:  strings.TrimPrefix(api.URL, "https://"),
    		Next:  api.Client().Transport,
    	}}
    	for _, u := range []string{api.URL, other.URL} {
    		resp, err := client.Get(u)
    		if err != nil {
    			fmt.Println(err)
    			continue
    		}
    		body, _ := io.ReadAll(resp.Body)
    		resp.Body.Close()
    		fmt.Print(string(body))
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

    // testRecorder is a server that remembers the Authorization header of each request.
    type testRecorder struct {
    	mu   sync.Mutex
    	auth []string
    }

    func (r *testRecorder) handler(redirectTo *string) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    		r.mu.Lock()
    		r.auth = append(r.auth, req.Header.Get("Authorization"))
    		r.mu.Unlock()
    		if redirectTo != nil && req.URL.Path == "/attachments/7" {
    			http.Redirect(w, req, *redirectTo+"/cdn/screenshot.png", http.StatusFound)
    			return
    		}
    		io.WriteString(w, "ok")
    	})
    }

    func (r *testRecorder) seen() []string {
    	r.mu.Lock()
    	defer r.mu.Unlock()
    	return append([]string(nil), r.auth...)
    }

    func testGet(t *testing.T, client *http.Client, u string) {
    	t.Helper()
    	resp, err := client.Get(u)
    	if err != nil {
    		t.Fatalf("GET %s: %v", u, err)
    	}
    	io.Copy(io.Discard, resp.Body)
    	resp.Body.Close()
    }

    func TestTokenForAPIOnly(t *testing.T) {
    	var cdnURL string
    	apiRec, otherRec, plainRec := &testRecorder{}, &testRecorder{}, &testRecorder{}
    	api := httptest.NewTLSServer(apiRec.handler(&cdnURL))
    	defer api.Close()
    	other := httptest.NewTLSServer(otherRec.handler(nil))
    	defer other.Close()
    	plain := httptest.NewServer(plainRec.handler(nil))
    	defer plain.Close()
    	cdnURL = other.URL

    	apiHost := strings.TrimPrefix(api.URL, "https://")
    	client := &http.Client{Transport: &authTransport{Token: "tk_live_8f3a", Host: apiHost, Next: api.Client().Transport}}

    	testGet(t, client, api.URL+"/v1/issues")
    	if got := apiRec.seen(); len(got) != 1 || got[0] != "Bearer tk_live_8f3a" {
    		t.Errorf("request to the API (https://%s) arrived with Authorization %q, want [\"Bearer tk_live_8f3a\"]", apiHost, got)
    	}

    	testGet(t, client, other.URL+"/v1/issues")
    	if got := otherRec.seen(); len(got) != 1 || got[0] != "" {
    		t.Errorf("request to another host (%s) arrived with Authorization %q, want none: the token only goes to %s", other.URL, got, apiHost)
    	}

    	// Same IP, different port: still a different host.
    	plainHostAsHTTPS := &authTransport{Token: "tk_live_8f3a", Host: strings.TrimPrefix(plain.URL, "http://"), Next: api.Client().Transport}
    	testGet(t, &http.Client{Transport: plainHostAsHTTPS}, plain.URL+"/v1/issues")
    	if got := plainRec.seen(); len(got) != 1 || got[0] != "" {
    		t.Errorf("plain-http request to Host %s arrived with Authorization %q, want none: never send the token without TLS", strings.TrimPrefix(plain.URL, "http://"), got)
    	}
    }

    func TestTokenNotFollowingRedirect(t *testing.T) {
    	var cdnURL string
    	apiRec, cdnRec := &testRecorder{}, &testRecorder{}
    	api := httptest.NewTLSServer(apiRec.handler(&cdnURL))
    	defer api.Close()
    	cdn := httptest.NewTLSServer(cdnRec.handler(nil))
    	defer cdn.Close()
    	cdnURL = cdn.URL

    	client := &http.Client{Transport: &authTransport{Token: "tk_live_8f3a", Host: strings.TrimPrefix(api.URL, "https://"), Next: api.Client().Transport}}
    	testGet(t, client, api.URL+"/attachments/7")
    	if got := apiRec.seen(); len(got) != 1 || got[0] != "Bearer tk_live_8f3a" {
    		t.Errorf("API saw Authorization %q, want [\"Bearer tk_live_8f3a\"]", got)
    	}
    	if got := cdnRec.seen(); len(got) != 1 || got[0] != "" {
    		t.Errorf("after the API redirected to %s, that host saw Authorization %q, want none", cdn.URL, got)
    	}
    }

    func TestTokenLeavesRequestAlone(t *testing.T) {
    	rec := &testRecorder{}
    	api := httptest.NewTLSServer(rec.handler(nil))
    	defer api.Close()
    	tr := &authTransport{Token: "tk_live_8f3a", Host: strings.TrimPrefix(api.URL, "https://"), Next: api.Client().Transport}

    	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, api.URL+"/v1/me", nil)
    	if err != nil {
    		t.Fatal(err)
    	}
    	req.Header.Set("Accept", "application/json")
    	resp, err := tr.RoundTrip(req)
    	if err != nil {
    		t.Fatalf("RoundTrip: %v", err)
    	}
    	resp.Body.Close()
    	if got := req.Header.Get("Authorization"); got != "" {
    		t.Errorf("after RoundTrip, the caller's request has Authorization %q: a RoundTripper must not modify the request it's given (clone it)", got)
    	}
    	if got := rec.seen(); len(got) != 1 || got[0] != "Bearer tk_live_8f3a" {
    		t.Errorf("server saw Authorization %q, want [\"Bearer tk_live_8f3a\"]", got)
    	}

    	// Another host: the caller's own header, if any, goes through unchanged.
    	other := httptest.NewTLSServer(rec.handler(nil))
    	defer other.Close()
    	req2, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, other.URL+"/", nil)
    	req2.Header.Set("Authorization", "Basic b3RoZXI6c2VydmljZQ==")
    	resp, err = tr.RoundTrip(req2)
    	if err != nil {
    		t.Fatalf("RoundTrip: %v", err)
    	}
    	resp.Body.Close()
    	if got := rec.seen(); len(got) != 2 || got[1] != "Basic b3RoZXI6c2VydmljZQ==" {
    		t.Errorf("request to another host with its own Authorization arrived with %q, want it unchanged", got)
    	}
    }
---

`trackr` adds your API token to requests in one place: a custom
`http.RoundTripper`. The first version added it to **every** request, which
turned out to be a leak. Issue attachments redirect to a CDN on another host,
and the CDN's access logs were full of `Bearer tk_live_...`.

Complete `(*authTransport).RoundTrip(req)`:

- If the request goes to exactly `t.Host` (host **and** port, compared with
  `req.URL.Host`) over `https`, send a copy of it with
  `Authorization: Bearer <t.Token>` through `t.Next`.
- Send every other request through `t.Next` untouched: other hosts, other
  ports, and plain `http` even to the right host. A token sent without TLS is a
  token anyone on the network can read.
- Never modify `req` itself. The `http.RoundTripper` contract forbids it, and
  the caller may reuse the request.

## Example

```
Host: "api.trackr.dev"
GET https://api.trackr.dev/v1/issues           -> Authorization: Bearer tk_live_8f3a
GET https://cdn.trackr-files.net/shot.png      -> no Authorization
GET http://api.trackr.dev/v1/issues            -> no Authorization (not https)
GET https://api.trackr.dev:8443/v1/issues      -> no Authorization (different port)
```

## Constraints

- The tests use `httptest.NewTLSServer` for the API and a second TLS server
  for the "CDN", including an API endpoint that redirects there. The CDN must
  never see your token.
- A request to another host that already has its own `Authorization` header
  must arrive with that header unchanged.
