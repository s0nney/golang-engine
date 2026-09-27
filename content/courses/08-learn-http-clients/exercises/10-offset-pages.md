---
title: Offset Pages
difficulty: medium
after: paging-and-errors
hints:
  - 'Parse `baseURL` once, then for each page build the URL with `u.JoinPath("projects", url.PathEscape(project), "issues")` and a fresh `url.Values{"limit": ..., "offset": ...}` encoded into `RawQuery`.'
  - 'The loop condition is the whole trick: keep going while `len(all) < total`. That stops after the last page without asking for an empty one, and handles `total == 0` after a single request.'
  - 'A server that claims more than it has would make that loop spin forever. Also stop when a page comes back empty. Start with `all := []Issue{}` so an empty project gives a non-nil slice.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strconv"
    )

    // Issue is one Trackr issue.
    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    }

    func listProjectIssues(ctx context.Context, client *http.Client, baseURL, project string, limit int) ([]Issue, error) {
    	return nil, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Println("server got:", r.URL.EscapedPath(), r.URL.RawQuery)
    		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
    		switch offset {
    		case 0:
    			fmt.Fprint(w, `{"issues": [{"id": 1, "title": "Login broken"}, {"id": 2, "title": "Dark mode"}], "total": 3}`)
    		default:
    			fmt.Fprint(w, `{"issues": [{"id": 3, "title": "Typo on home page"}], "total": 3}`)
    		}
    	}))
    	defer srv.Close()

    	fmt.Println(listProjectIssues(context.Background(), srv.Client(), srv.URL+"/v1/", "web app", 2))
    	// want:
    	// server got: /v1/projects/web%20app/issues limit=2&offset=0
    	// server got: /v1/projects/web%20app/issues limit=2&offset=2
    	// [{1 Login broken} {2 Dark mode} {3 Typo on home page}] <nil>
    }
  solution: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"net/url"
    	"strconv"
    )

    // Issue is one Trackr issue.
    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    }

    type issuePage struct {
    	Issues []Issue `json:"issues"`
    	Total  int     `json:"total"`
    }

    func listProjectIssues(ctx context.Context, client *http.Client, baseURL, project string, limit int) ([]Issue, error) {
    	base, err := url.Parse(baseURL)
    	if err != nil {
    		return nil, err
    	}
    	u := base.JoinPath("projects", url.PathEscape(project), "issues")
    	all := []Issue{}
    	for {
    		u.RawQuery = url.Values{
    			"limit":  {strconv.Itoa(limit)},
    			"offset": {strconv.Itoa(len(all))},
    		}.Encode()
    		page, err := fetchOffsetPage(ctx, client, u.String())
    		if err != nil {
    			return nil, fmt.Errorf("listing %s at offset %d: %w", project, len(all), err)
    		}
    		all = append(all, page.Issues...)
    		if len(page.Issues) == 0 || len(all) >= page.Total {
    			return all, nil
    		}
    	}
    }

    func fetchOffsetPage(ctx context.Context, client *http.Client, u string) (issuePage, error) {
    	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
    	if err != nil {
    		return issuePage{}, err
    	}
    	resp, err := client.Do(req)
    	if err != nil {
    		return issuePage{}, err
    	}
    	defer resp.Body.Close()
    	if resp.StatusCode != http.StatusOK {
    		return issuePage{}, fmt.Errorf("unexpected status %s", resp.Status)
    	}
    	var p issuePage
    	if err := json.UnmarshalRead(resp.Body, &p); err != nil {
    		return issuePage{}, err
    	}
    	return p, nil
    }

    func main() {
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		fmt.Println("server got:", r.URL.EscapedPath(), r.URL.RawQuery)
    		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
    		switch offset {
    		case 0:
    			fmt.Fprint(w, `{"issues": [{"id": 1, "title": "Login broken"}, {"id": 2, "title": "Dark mode"}], "total": 3}`)
    		default:
    			fmt.Fprint(w, `{"issues": [{"id": 3, "title": "Typo on home page"}], "total": 3}`)
    		}
    	}))
    	defer srv.Close()

    	fmt.Println(listProjectIssues(context.Background(), srv.Client(), srv.URL+"/v1/", "web app", 2))
    }
  tests: |
    package main

    import (
    	jsonv1 "encoding/json"
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    	"strconv"
    	"strings"
    	"sync"
    	"sync/atomic"
    	"testing"
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

    // testAPI serves n issues of one project with offset/limit paging.
    type testAPI struct {
    	t         *testing.T
    	path      string // the escaped path it answers on
    	n         int
    	claim     int // if non-zero, the "total" it reports instead of n
    	failAt    int // offset at which it answers 500 (-1: never)
    	brokenAt  int // offset at which it sends malformed JSON (-1: never)
    	mu        sync.Mutex
    	offsets   []string
    	badParams []string
    }

    func (a *testAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    	if r.URL.EscapedPath() != a.path || r.Method != http.MethodGet {
    		a.t.Errorf("server received %s %s, want GET %s", r.Method, r.URL.EscapedPath(), a.path)
    		http.NotFound(w, r)
    		return
    	}
    	q := r.URL.Query()
    	a.mu.Lock()
    	a.offsets = append(a.offsets, q.Get("offset"))
    	for k := range q {
    		if k != "limit" && k != "offset" {
    			a.badParams = append(a.badParams, k)
    		}
    	}
    	a.mu.Unlock()
    	limit, err1 := strconv.Atoi(q.Get("limit"))
    	offset, err2 := strconv.Atoi(q.Get("offset"))
    	if err1 != nil || err2 != nil || limit < 1 {
    		http.Error(w, `{"error": "limit and offset are required"}`, http.StatusBadRequest)
    		return
    	}
    	if offset == a.failAt {
    		http.Error(w, `{"error": "database on fire"}`, http.StatusInternalServerError)
    		return
    	}
    	if offset == a.brokenAt {
    		io.WriteString(w, `{"issues": [{"id": 1, "title": "tru`)
    		return
    	}
    	var page struct {
    		Issues []Issue `json:"issues"`
    		Total  int     `json:"total"`
    	}
    	page.Issues = []Issue{}
    	for i := offset; i < min(offset+limit, a.n); i++ {
    		page.Issues = append(page.Issues, Issue{i + 1, fmt.Sprintf("Issue %d", i+1)})
    	}
    	page.Total = a.n
    	if a.claim != 0 {
    		page.Total = a.claim
    	}
    	w.Header().Set("Content-Type", "application/json")
    	jsonv1.NewEncoder(w).Encode(page)
    }

    func testList(t *testing.T, api *testAPI, base, project string, limit int) ([]Issue, error) {
    	t.Helper()
    	api.t = t
    	srv := httptest.NewServer(api)
    	defer srv.Close()
    	tb := &testBodyTracker{next: srv.Client().Transport}
    	got, err := listProjectIssues(t.Context(), &http.Client{Transport: tb}, srv.URL+base, project, limit)
    	if tb.opened.Load() != tb.closed.Load() {
    		t.Errorf("%d responses but %d body closes: close every body", tb.opened.Load(), tb.closed.Load())
    	}
    	if len(api.badParams) > 0 {
    		t.Errorf("requests had unexpected query parameters %v; send only limit and offset", api.badParams)
    	}
    	return got, err
    }

    func testWant(n int) []Issue {
    	want := []Issue{}
    	for i := range n {
    		want = append(want, Issue{i + 1, fmt.Sprintf("Issue %d", i+1)})
    	}
    	return want
    }

    func TestListPages(t *testing.T) {
    	tests := []struct {
    		n, limit    int
    		wantOffsets []string
    	}{
    		{5, 2, []string{"0", "2", "4"}},
    		{6, 3, []string{"0", "3"}},
    		{3, 50, []string{"0"}},
    		{50, 50, []string{"0"}},
    		{1, 1, []string{"0"}},
    		{0, 10, []string{"0"}},
    	}
    	for _, tt := range tests {
    		api := &testAPI{path: "/v1/projects/apollo/issues", n: tt.n, failAt: -1, brokenAt: -1}
    		got, err := testList(t, api, "/v1/", "apollo", tt.limit)
    		if err != nil || !slices.Equal(got, testWant(tt.n)) {
    			t.Errorf("%d issues, limit %d: listProjectIssues = %v, %v, want %v, nil", tt.n, tt.limit, got, err, testWant(tt.n))
    		}
    		if got == nil && err == nil {
    			t.Errorf("%d issues: listProjectIssues returned a nil slice, want an empty, non-nil one", tt.n)
    		}
    		if !slices.Equal(api.offsets, tt.wantOffsets) {
    			t.Errorf("%d issues, limit %d: requested offsets %v, want %v (use total to avoid asking for an empty page)", tt.n, tt.limit, api.offsets, tt.wantOffsets)
    		}
    	}
    }

    func TestListEscapesProject(t *testing.T) {
    	for _, tt := range []struct{ base, project, path string }{
    		{"", "web app", "/projects/web%20app/issues"},
    		{"/v1", "café/ui", "/v1/projects/caf%C3%A9%2Fui/issues"},
    		{"/v1/", "q&a?", "/v1/projects/q&a%3F/issues"},
    	} {
    		api := &testAPI{path: tt.path, n: 3, failAt: -1, brokenAt: -1}
    		got, err := testList(t, api, tt.base, tt.project, 2)
    		if err != nil || len(got) != 3 {
    			t.Errorf("project %q: listProjectIssues = %v, %v, want 3 issues from %s", tt.project, got, err, tt.path)
    		}
    	}
    }

    func TestListServerOvercounts(t *testing.T) {
    	api := &testAPI{path: "/v1/projects/apollo/issues", n: 4, claim: 10, failAt: -1, brokenAt: -1}
    	got, err := testList(t, api, "/v1", "apollo", 2)
    	if err != nil || !slices.Equal(got, testWant(4)) {
    		t.Errorf("server claims total 10 but has 4: listProjectIssues = %v, %v, want the 4 issues", got, err)
    	}
    	if !slices.Equal(api.offsets, []string{"0", "2", "4"}) {
    		t.Errorf("requested offsets %v, want [0 2 4] (stop at the first empty page)", api.offsets)
    	}
    }

    func TestListErrors(t *testing.T) {
    	for _, api := range []*testAPI{
    		{path: "/v1/projects/apollo/issues", n: 9, failAt: 4, brokenAt: -1},
    		{path: "/v1/projects/apollo/issues", n: 9, failAt: -1, brokenAt: 4},
    		{path: "/v1/projects/apollo/issues", n: 9, failAt: 0, brokenAt: -1},
    	} {
    		at := max(api.failAt, api.brokenAt)
    		got, err := testList(t, api, "/v1", "apollo", 2)
    		if err == nil || got != nil {
    			t.Errorf("page at offset %d fails: listProjectIssues = %v, %v, want nil and an error", at, got, err)
    			continue
    		}
    		if !strings.Contains(err.Error(), strconv.Itoa(at)) {
    			t.Errorf("page at offset %d fails: error %q should mention the offset", at, err)
    		}
    		if n := len(api.offsets); n != at/2+1 {
    			t.Errorf("page at offset %d fails: server saw %d requests, want %d (stop at the failure)", at, n, at/2+1)
    		}
    	}
    }
---

Trackr's older project endpoint pages with **offset and limit**: you ask for
`limit` issues starting at `offset`, and every page also reports the `total`
number of issues in the project.

```
GET /v1/projects/web%20app/issues?limit=2&offset=0
{"issues": [{"id": 1, "title": "Login broken"}, {"id": 2, "title": "Dark mode"}], "total": 3}
```

Complete `listProjectIssues(ctx, client, baseURL, project, limit)`. It returns
every issue in the project, in order:

- Request `<baseURL>/projects/<project>/issues?limit=<limit>&offset=<offset>`
  with `client`, starting at offset 0. Send only those two query parameters.
- The project name is a single path segment: `web app` becomes `web%20app`,
  and `café/ui` becomes `caf%C3%A9%2Fui`. `baseURL` may end in `/`.
- The next offset is the number of issues collected so far.
- Stop once you've collected `total` issues. Don't request a page you know is
  empty: 6 issues with a limit of 3 is exactly 2 requests.
- Also stop at the first empty page, so a server that over-reports `total`
  can't send you round in circles.
- A status other than `200 OK`, a failed request or malformed JSON returns `nil`
  and an error that mentions the offset. Send no more requests after it.
- An empty project returns an empty, **non-nil** slice (after one request).

## Example

```
listProjectIssues(ctx, c, "https://api.trackr.dev/v1/", "web app", 2)
// GET /v1/projects/web%20app/issues?limit=2&offset=0 -> 2 issues, total 3
// GET /v1/projects/web%20app/issues?limit=2&offset=2 -> 1 issue,  total 3
// => [{1 Login broken} {2 Dark mode} {3 Typo on home page}], nil
```

## Constraints

- The fake API checks the escaped path and the query of every request, and the
  tests check the exact list of offsets you asked for.
- Close every response body.
