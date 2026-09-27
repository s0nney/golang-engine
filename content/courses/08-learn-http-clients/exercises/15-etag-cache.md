---
title: ETag Cache
difficulty: hard
after: testing-http-clients
hints:
  - 'Store `map[string]entry` where `entry` holds the `etag` and the `body` bytes, keyed by the full URL string, and guard it with a `sync.Mutex`. Look the entry up under the lock, **unlock**, do the request, then lock again to store the result. Holding the lock during a request would make every caller wait for every other caller''s network trip.'
  - 'With an entry, `req.Header.Set("If-None-Match", e.etag)`, exactly as the server sent it (weak `W/"..."` tags too). Then switch on the status: 304 means "use your copy", 200 means "here''s a new version" (store it if it has an ETag, **delete** the old entry if it doesn''t), anything else is an error.'
  - 'Slices share memory: if you return the cached `[]byte` itself, a caller that edits it corrupts the cache. Return `slices.Clone(e.body)`.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    // CachingClient GETs URLs, revalidating cached copies with ETags.
    type CachingClient struct {
    	client *http.Client
    	// your fields here
    }

    func NewCachingClient(client *http.Client) *CachingClient {
    	return &CachingClient{client: client}
    }

    func (c *CachingClient) Get(ctx context.Context, url string) ([]byte, error) {
    	return nil, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Printf("server got If-None-Match=%q\n", r.Header.Get("If-None-Match"))
    		if r.Header.Get("If-None-Match") == `"v1"` {
    			w.WriteHeader(http.StatusNotModified)
    			return
    		}
    		w.Header().Set("ETag", `"v1"`)
    		fmt.Fprint(w, `[{"id": 1, "name": "bug"}, {"id": 2, "name": "ui"}]`)
    	}))
    	defer srv.Close()

    	c := NewCachingClient(srv.Client())
    	for range 2 {
    		body, err := c.Get(context.Background(), srv.URL+"/labels")
    		fmt.Println(string(body), err)
    	}
    	// want:
    	// server got If-None-Match=""
    	// [{"id": 1, "name": "bug"}, {"id": 2, "name": "ui"}] <nil>
    	// server got If-None-Match="\"v1\""
    	// [{"id": 1, "name": "bug"}, {"id": 2, "name": "ui"}] <nil>
    }
  solution: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    	"sync"
    )

    type cacheEntry struct {
    	etag string
    	body []byte
    }

    // CachingClient GETs URLs, revalidating cached copies with ETags.
    type CachingClient struct {
    	client *http.Client

    	mu    sync.Mutex
    	cache map[string]cacheEntry
    }

    func NewCachingClient(client *http.Client) *CachingClient {
    	return &CachingClient{client: client, cache: map[string]cacheEntry{}}
    }

    func (c *CachingClient) Get(ctx context.Context, url string) ([]byte, error) {
    	c.mu.Lock()
    	e, cached := c.cache[url]
    	c.mu.Unlock()

    	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    	if err != nil {
    		return nil, err
    	}
    	if cached {
    		req.Header.Set("If-None-Match", e.etag)
    	}
    	resp, err := c.client.Do(req)
    	if err != nil {
    		return nil, err
    	}
    	defer resp.Body.Close()

    	switch resp.StatusCode {
    	case http.StatusNotModified:
    		if !cached {
    			return nil, errors.New("304 Not Modified for a URL that isn't cached")
    		}
    		return slices.Clone(e.body), nil
    	case http.StatusOK:
    		body, err := io.ReadAll(resp.Body)
    		if err != nil {
    			return nil, err
    		}
    		c.mu.Lock()
    		if etag := resp.Header.Get("ETag"); etag != "" {
    			c.cache[url] = cacheEntry{etag, slices.Clone(body)}
    		} else {
    			delete(c.cache, url)
    		}
    		c.mu.Unlock()
    		return body, nil
    	default:
    		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
    	}
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Printf("server got If-None-Match=%q\n", r.Header.Get("If-None-Match"))
    		if r.Header.Get("If-None-Match") == `"v1"` {
    			w.WriteHeader(http.StatusNotModified)
    			return
    		}
    		w.Header().Set("ETag", `"v1"`)
    		fmt.Fprint(w, `[{"id": 1, "name": "bug"}, {"id": 2, "name": "ui"}]`)
    	}))
    	defer srv.Close()

    	c := NewCachingClient(srv.Client())
    	for range 2 {
    		body, err := c.Get(context.Background(), srv.URL+"/labels")
    		fmt.Println(string(body), err)
    	}
    }
  tests: |
    package main

    import (
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync"
    	"sync/atomic"
    	"testing"
    	"time"
    )

    type testBodyTracker struct {
    	next           http.RoundTripper
    	opened, closed atomic.Int32
    }

    func (tb *testBodyTracker) RoundTrip(req *http.Request) (*http.Response, error) {
    	resp, err := tb.next.RoundTrip(req)
    	if err == nil {
    		tb.opened.Add(1)
    		resp.Body = &testCloseCounter{resp.Body, &tb.closed}
    	}
    	return resp, err
    }

    type testCloseCounter struct {
    	io.ReadCloser
    	n *atomic.Int32
    }

    func (c *testCloseCounter) Close() error {
    	c.n.Add(1)
    	return c.ReadCloser.Close()
    }

    // testResource is one URL on the fake API. It revalidates If-None-Match
    // against its current ETag, the way a real server does.
    type testResource struct {
    	etag   string // "" means: send no ETag
    	body   string
    	status int // non-zero: fail with this status
    }

    type testOrigin struct {
    	mu    sync.Mutex
    	res   map[string]*testResource // keyed by path+"?"+query
    	log   []string                 // "<key> If-None-Match=<value>"
    	delay time.Duration
    }

    func (o *testOrigin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    	if o.delay > 0 {
    		select {
    		case <-time.After(o.delay):
    		case <-r.Context().Done():
    			return
    		}
    	}
    	key := r.URL.Path + "?" + r.URL.RawQuery
    	o.mu.Lock()
    	o.log = append(o.log, fmt.Sprintf("%s If-None-Match=%s", key, r.Header.Get("If-None-Match")))
    	res, ok := o.res[key]
    	var cur testResource
    	if ok {
    		cur = *res
    	}
    	o.mu.Unlock()
    	switch {
    	case !ok:
    		http.NotFound(w, r)
    	case cur.status != 0:
    		w.WriteHeader(cur.status)
    		io.WriteString(w, `{"error": "nope"}`)
    	case cur.etag != "" && r.Header.Get("If-None-Match") == cur.etag:
    		w.Header().Set("ETag", cur.etag)
    		w.WriteHeader(http.StatusNotModified)
    	default:
    		if cur.etag != "" {
    			w.Header().Set("ETag", cur.etag)
    		}
    		io.WriteString(w, cur.body)
    	}
    }

    func (o *testOrigin) set(key string, r testResource) {
    	o.mu.Lock()
    	defer o.mu.Unlock()
    	o.res[key] = &r
    }

    func (o *testOrigin) takeLog() []string {
    	o.mu.Lock()
    	defer o.mu.Unlock()
    	l := o.log
    	o.log = nil
    	return l
    }

    func testSetup(t *testing.T) (*CachingClient, *testOrigin, string) {
    	o := &testOrigin{res: map[string]*testResource{}}
    	srv := httptest.NewServer(o)
    	tb := &testBodyTracker{next: srv.Client().Transport}
    	t.Cleanup(func() {
    		srv.Close()
    		if tb.opened.Load() != tb.closed.Load() {
    			t.Errorf("%d responses but %d body closes: close every body, 304s included", tb.opened.Load(), tb.closed.Load())
    		}
    	})
    	return NewCachingClient(&http.Client{Transport: tb}), o, srv.URL
    }

    // testStep calls Get and checks the body and what the server received.
    func testStep(t *testing.T, c *CachingClient, o *testOrigin, url, wantBody, wantLog string) {
    	t.Helper()
    	got, err := c.Get(t.Context(), url)
    	if err != nil || string(got) != wantBody {
    		t.Errorf("Get(%s) = %q, %v, want %q, nil", url[strings.LastIndex(url, "/"):], got, err, wantBody)
    	}
    	if l := o.takeLog(); len(l) != 1 || l[0] != wantLog {
    		t.Errorf("server received %q, want [%q]", l, wantLog)
    	}
    }

    func TestCacheRevalidates(t *testing.T) {
    	c, o, base := testSetup(t)
    	o.set("/labels?", testResource{etag: `"v1"`, body: `["bug","ui"]`})
    	testStep(t, c, o, base+"/labels", `["bug","ui"]`, `/labels? If-None-Match=`)
    	testStep(t, c, o, base+"/labels", `["bug","ui"]`, `/labels? If-None-Match="v1"`)
    	testStep(t, c, o, base+"/labels", `["bug","ui"]`, `/labels? If-None-Match="v1"`)

    	o.set("/labels?", testResource{etag: `"v2"`, body: `["bug","ui","docs"]`})
    	testStep(t, c, o, base+"/labels", `["bug","ui","docs"]`, `/labels? If-None-Match="v1"`)
    	testStep(t, c, o, base+"/labels", `["bug","ui","docs"]`, `/labels? If-None-Match="v2"`)
    }

    func TestCacheWeakETagsAndKeys(t *testing.T) {
    	c, o, base := testSetup(t)
    	o.set("/issues?state=open", testResource{etag: `W/"open-3"`, body: `open`})
    	o.set("/issues?state=closed", testResource{etag: `W/"closed-9"`, body: `closed`})
    	o.set("/issues?q=caf%C3%A9", testResource{etag: `"q"`, body: `café`})
    	testStep(t, c, o, base+"/issues?state=open", `open`, `/issues?state=open If-None-Match=`)
    	testStep(t, c, o, base+"/issues?state=closed", `closed`, `/issues?state=closed If-None-Match=`)
    	testStep(t, c, o, base+"/issues?q=caf%C3%A9", `café`, `/issues?q=caf%C3%A9 If-None-Match=`)
    	testStep(t, c, o, base+"/issues?state=open", `open`, `/issues?state=open If-None-Match=W/"open-3"`)
    	testStep(t, c, o, base+"/issues?state=closed", `closed`, `/issues?state=closed If-None-Match=W/"closed-9"`)
    	testStep(t, c, o, base+"/issues?q=caf%C3%A9", `café`, `/issues?q=caf%C3%A9 If-None-Match="q"`)
    }

    func TestCacheNoETag(t *testing.T) {
    	c, o, base := testSetup(t)
    	o.set("/me?", testResource{body: `{"login": "ana"}`})
    	testStep(t, c, o, base+"/me", `{"login": "ana"}`, `/me? If-None-Match=`)
    	testStep(t, c, o, base+"/me", `{"login": "ana"}`, `/me? If-None-Match=`)

    	// A resource that stops sending ETags must not be revalidated with the old one.
    	o.set("/project?", testResource{etag: `"p1"`, body: `v1`})
    	testStep(t, c, o, base+"/project", `v1`, `/project? If-None-Match=`)
    	o.set("/project?", testResource{body: `v2`})
    	testStep(t, c, o, base+"/project", `v2`, `/project? If-None-Match="p1"`)
    	testStep(t, c, o, base+"/project", `v2`, `/project? If-None-Match=`)
    }

    func TestCacheErrors(t *testing.T) {
    	c, o, base := testSetup(t)
    	o.set("/labels?", testResource{etag: `"v1"`, body: `["bug"]`})
    	testStep(t, c, o, base+"/labels", `["bug"]`, `/labels? If-None-Match=`)

    	for _, status := range []int{500, 404, 403} {
    		o.set("/labels?", testResource{etag: `"v1"`, status: status})
    		got, err := c.Get(t.Context(), base+"/labels")
    		if err == nil || got != nil {
    			t.Errorf("server answered %d: Get = %q, %v, want nil and an error (don't serve stale data on an error)", status, got, err)
    		}
    		o.takeLog()
    	}
    	// The cached copy survives the errors.
    	o.set("/labels?", testResource{etag: `"v1"`, body: `["bug"]`})
    	testStep(t, c, o, base+"/labels", `["bug"]`, `/labels? If-None-Match="v1"`)

    	got, err := c.Get(t.Context(), base+"/nowhere")
    	if err == nil || got != nil {
    		t.Errorf("Get(/nowhere) = %q, %v, want nil and an error for a 404", got, err)
    	}
    }

    func TestCacheUnexpected304(t *testing.T) {
    	c, _, _ := testSetup(t)
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.WriteHeader(http.StatusNotModified)
    	}))
    	defer srv.Close()
    	got, err := c.Get(t.Context(), srv.URL+"/labels")
    	if err == nil || got != nil {
    		t.Errorf("304 for a URL that was never cached: Get = %q, %v, want nil and an error", got, err)
    	}
    }

    func TestCacheReturnsCopies(t *testing.T) {
    	c, o, base := testSetup(t)
    	o.set("/labels?", testResource{etag: `"v1"`, body: `bug`})
    	first, _ := c.Get(t.Context(), base+"/labels")
    	if len(first) > 0 {
    		first[0] = 'X'
    	}
    	second, _ := c.Get(t.Context(), base+"/labels")
    	if len(second) > 0 {
    		second[0] = 'Y'
    	}
    	third, _ := c.Get(t.Context(), base+"/labels")
    	if string(third) != "bug" {
    		t.Errorf("callers edited the bodies they got, and now Get returns %q, want %q: return a copy of the cached bytes", third, "bug")
    	}
    }

    func TestCacheConcurrent(t *testing.T) {
    	c, o, base := testSetup(t)
    	for i := range 5 {
    		o.set(fmt.Sprintf("/issues/%d?", i), testResource{etag: fmt.Sprintf(`"e%d"`, i), body: fmt.Sprintf("issue %d", i)})
    	}
    	var wg sync.WaitGroup
    	for g := range 20 {
    		wg.Go(func() {
    			for n := range 10 {
    				i := (g + n) % 5
    				got, err := c.Get(t.Context(), fmt.Sprintf("%s/issues/%d", base, i))
    				if err != nil || string(got) != fmt.Sprintf("issue %d", i) {
    					t.Errorf("concurrent Get(/issues/%d) = %q, %v, want %q, nil", i, got, err, fmt.Sprintf("issue %d", i))
    					return
    				}
    			}
    		})
    	}
    	wg.Wait()
    	full := 0
    	for _, l := range o.takeLog() {
    		if strings.HasSuffix(l, "If-None-Match=") {
    			full++
    		}
    	}
    	if full > 20*5 {
    		t.Errorf("%d unconditional requests out of 200: the cache isn't being used", full)
    	}
    }

    func TestCacheDoesNotSerialiseRequests(t *testing.T) {
    	c, o, base := testSetup(t)
    	o.delay = 200 * time.Millisecond
    	for i := range 10 {
    		o.set(fmt.Sprintf("/issues/%d?", i), testResource{etag: `"x"`, body: "ok"})
    	}
    	start := time.Now()
    	var wg sync.WaitGroup
    	for i := range 10 {
    		wg.Go(func() { c.Get(t.Context(), fmt.Sprintf("%s/issues/%d", base, i)) })
    	}
    	wg.Wait()
    	if elapsed := time.Since(start); elapsed > time.Second {
    		t.Errorf("10 concurrent Gets to a server taking 200ms each took %v, want about 200ms: don't hold the mutex while a request is in flight", elapsed)
    	}
    }
---

`trackr` fetches the same slow-changing data over and over: the label list, the
project settings, your own profile. The Trackr API supports **conditional
requests**, so a client can ask "has this changed?" instead of downloading it
again:

```
GET /labels                           -> 200, ETag: "v1", [...the labels...]
GET /labels   If-None-Match: "v1"     -> 304 Not Modified (no body: use your copy)
GET /labels   If-None-Match: "v1"     -> 200, ETag: "v2", [...new labels...]
```

Build `CachingClient`:

- `NewCachingClient(client)` returns a cache that sends every request through
  `client`.
- `Get(ctx, url)` GETs `url` and returns the response body.
  - Cache entries are keyed by the **full URL**, query included:
    `/issues?state=open` and `/issues?state=closed` are different entries.
  - If a copy of `url` is cached, send `If-None-Match` with its ETag, exactly
    as the server sent it (weak tags like `W/"open-3"` included).
  - `304 Not Modified`: return the cached body. (A 304 for a URL that isn't
    cached is an error.)
  - `200 OK`: return the body. If the response has an `ETag`, cache the body
    under it, replacing any older copy. If it has none, **forget** any older
    copy, so the next request is unconditional.
  - Anything else (or a failed request): return `nil` and an error, and keep
    any cached copy for next time. Never serve cached data for an error.
- Callers own the slices they get back: changing one must not change what the
  cache returns later.
- `Get` must be safe to call from many goroutines at once, **without**
  making them wait for each other's requests.

## Example

```go
c := NewCachingClient(http.DefaultClient)
c.Get(ctx, base+"/labels") // GET, no If-None-Match     -> 200 "v1": cached, returned
c.Get(ctx, base+"/labels") // GET, If-None-Match: "v1"  -> 304: cached copy returned
```

## Constraints

- The fake API really revalidates: it answers 304 only if `If-None-Match`
  matches its current ETag. The tests log the `If-None-Match` of every request.
- One test makes 200 concurrent `Get`s (run the tests with `-race` locally if
  you can). Another sends 10 concurrent `Get`s to a server that takes 200ms to
  answer and expects them to finish in about 200ms, not 2s.
- Close every response body, 304s included.
