---
title: Recent Activity
difficulty: hard
after: paging-and-errors
hints:
  - 'Split off `fetchPage(ctx, client, pageURL) (issues []Issue, next string, err error)`. For `next`, split the Link header on `,`, then each part on `;`: the URL is between `<` and `>`, and one of the params is `rel="next"` or `rel=next`. Resolve it with `pageURL.Parse(next)` (a `*url.URL` method) so relative links work.'
  - 'The iterator is `return func(yield func(Issue, error) bool) { ... }`. Inside: a `seen := map[string]bool{}` of page URLs, and a loop that fetches, yields and moves on. Every `yield` that returns false must `return` at once, and so must every error you yield.'
  - 'The early stop is just another `return`: the first issue with `UpdatedAt.Before(since)` ends the whole iteration, before `next` is ever fetched. Compare timestamps with `Before`, never `==`.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"iter"
    	"net/http"
    	"net/http/httptest"
    	"time"
    )

    // Issue is one entry in the activity feed.
    type Issue struct {
    	ID        int       `json:"id"`
    	Title     string    `json:"title"`
    	UpdatedAt time.Time `json:"updated_at"`
    }

    // ErrPageLoop means the server's next links went round in a circle.
    var ErrPageLoop = errors.New("pagination loop")

    func recentIssues(ctx context.Context, client *http.Client, firstURL string, since time.Time) iter.Seq2[Issue, error] {
    	return func(yield func(Issue, error) bool) {
    	}
    }

    func main() {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /activity", func(w http.ResponseWriter, r *http.Request) {
    		if r.URL.Query().Get("page") == "2" {
    			fmt.Fprint(w, `[{"id": 3, "title": "Typo", "updated_at": "2026-09-25T08:00:00Z"},
    				{"id": 9, "title": "Old news", "updated_at": "2026-08-01T00:00:00Z"}]`)
    			return
    		}
    		w.Header().Set("Link", `</activity?page=2>; rel="next"`)
    		fmt.Fprint(w, `[{"id": 7, "title": "Dark mode", "updated_at": "2026-09-27T09:30:00Z"},
    			{"id": 4, "title": "Login broken", "updated_at": "2026-09-26T17:00:00Z"}]`)
    	})
    	srv := httptest.NewServer(mux)
    	defer srv.Close()

    	since := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
    	for iss, err := range recentIssues(context.Background(), srv.Client(), srv.URL+"/activity", since) {
    		if err != nil {
    			fmt.Println("error:", err)
    			break
    		}
    		fmt.Println(iss.ID, iss.Title)
    	}
    	fmt.Println("(end of recent activity)")
    	// want:
    	// 7 Dark mode
    	// 4 Login broken
    	// 3 Typo
    	// (end of recent activity)
    }
  solution: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"iter"
    	"net/http"
    	"net/http/httptest"
    	"net/url"
    	"slices"
    	"strings"
    	"time"
    )

    // Issue is one entry in the activity feed.
    type Issue struct {
    	ID        int       `json:"id"`
    	Title     string    `json:"title"`
    	UpdatedAt time.Time `json:"updated_at"`
    }

    // ErrPageLoop means the server's next links went round in a circle.
    var ErrPageLoop = errors.New("pagination loop")

    func recentIssues(ctx context.Context, client *http.Client, firstURL string, since time.Time) iter.Seq2[Issue, error] {
    	return func(yield func(Issue, error) bool) {
    		pageURL, err := url.Parse(firstURL)
    		if err != nil {
    			yield(Issue{}, err)
    			return
    		}
    		seen := map[string]bool{}
    		for pageURL != nil {
    			if seen[pageURL.String()] {
    				yield(Issue{}, fmt.Errorf("%w: %s requested twice", ErrPageLoop, pageURL))
    				return
    			}
    			seen[pageURL.String()] = true
    			issues, next, err := fetchActivityPage(ctx, client, pageURL)
    			if err != nil {
    				yield(Issue{}, err)
    				return
    			}
    			for _, iss := range issues {
    				if iss.UpdatedAt.Before(since) {
    					return // newest first: everything after this is older still
    				}
    				if !yield(iss, nil) {
    					return
    				}
    			}
    			pageURL = next
    		}
    	}
    }

    // fetchActivityPage GETs one page. next is nil on the last page.
    func fetchActivityPage(ctx context.Context, client *http.Client, pageURL *url.URL) ([]Issue, *url.URL, error) {
    	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL.String(), nil)
    	if err != nil {
    		return nil, nil, err
    	}
    	resp, err := client.Do(req)
    	if err != nil {
    		return nil, nil, err
    	}
    	defer resp.Body.Close()
    	if resp.StatusCode != http.StatusOK {
    		return nil, nil, fmt.Errorf("GET %s: %s", pageURL, resp.Status)
    	}
    	var issues []Issue
    	if err := json.UnmarshalRead(resp.Body, &issues); err != nil {
    		return nil, nil, fmt.Errorf("GET %s: %w", pageURL, err)
    	}
    	var next *url.URL
    	if link, ok := nextLink(resp.Header.Get("Link")); ok {
    		if next, err = pageURL.Parse(link); err != nil {
    			return nil, nil, fmt.Errorf("bad next link %q: %w", link, err)
    		}
    	}
    	return issues, next, nil
    }

    // nextLink finds the rel="next" URL in a Link header.
    func nextLink(header string) (string, bool) {
    	for part := range strings.SplitSeq(header, ",") {
    		target, params, ok := strings.Cut(part, ";")
    		if !ok {
    			continue
    		}
    		target = strings.TrimSpace(target)
    		if !strings.HasPrefix(target, "<") || !strings.HasSuffix(target, ">") {
    			continue
    		}
    		for p := range strings.SplitSeq(params, ";") {
    			k, v, _ := strings.Cut(strings.TrimSpace(p), "=")
    			rels := strings.Fields(strings.Trim(v, `"`)) // rel can list several: rel="next last"
    			if strings.EqualFold(k, "rel") && slices.ContainsFunc(rels, isNext) {
    				return target[1 : len(target)-1], true
    			}
    		}
    	}
    	return "", false
    }

    func isNext(rel string) bool { return strings.EqualFold(rel, "next") }

    func main() {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /activity", func(w http.ResponseWriter, r *http.Request) {
    		if r.URL.Query().Get("page") == "2" {
    			fmt.Fprint(w, `[{"id": 3, "title": "Typo", "updated_at": "2026-09-25T08:00:00Z"},
    				{"id": 9, "title": "Old news", "updated_at": "2026-08-01T00:00:00Z"}]`)
    			return
    		}
    		w.Header().Set("Link", `</activity?page=2>; rel="next"`)
    		fmt.Fprint(w, `[{"id": 7, "title": "Dark mode", "updated_at": "2026-09-27T09:30:00Z"},
    			{"id": 4, "title": "Login broken", "updated_at": "2026-09-26T17:00:00Z"}]`)
    	})
    	srv := httptest.NewServer(mux)
    	defer srv.Close()

    	since := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
    	for iss, err := range recentIssues(context.Background(), srv.Client(), srv.URL+"/activity", since) {
    		if err != nil {
    			fmt.Println("error:", err)
    			break
    		}
    		fmt.Println(iss.ID, iss.Title)
    	}
    	fmt.Println("(end of recent activity)")
    }
  tests: |
    package main

    import (
    	"context"
    	"errors"
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    	"slices"
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

    // testDay returns midnight UTC on day d of September 2026.
    func testDay(d int) time.Time { return time.Date(2026, 9, d, 0, 0, 0, 0, time.UTC) }

    type testPage struct {
    	ids    []int  // issue id i was updated on day i of September (ids are newest first)
    	link   string // Link header; "{srv}" is replaced by the server's URL
    	status int
    	body   string // raw body instead of ids
    }

    // testFeed serves pages keyed by path+"?"+query and records what was asked for.
    type testFeed struct {
    	pages map[string]testPage
    	url   string
    	mu    sync.Mutex
    	asked []string
    }

    func (f *testFeed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    	key := r.URL.Path
    	if r.URL.RawQuery != "" {
    		key += "?" + r.URL.RawQuery
    	}
    	f.mu.Lock()
    	f.asked = append(f.asked, key)
    	f.mu.Unlock()
    	p, ok := f.pages[key]
    	if !ok || r.Method != http.MethodGet {
    		http.Error(w, "no such page: "+r.Method+" "+key, http.StatusNotFound)
    		return
    	}
    	if p.link != "" {
    		w.Header().Set("Link", strings.ReplaceAll(p.link, "{srv}", f.url))
    	}
    	w.Header().Set("Content-Type", "application/json")
    	if p.status != 0 {
    		w.WriteHeader(p.status)
    	}
    	if p.body != "" {
    		io.WriteString(w, p.body)
    		return
    	}
    	var items []string
    	for _, id := range p.ids {
    		items = append(items, fmt.Sprintf(`{"id": %d, "title": "Issue %d", "updated_at": %q}`, id, id, testDay(id).Format(time.RFC3339)))
    	}
    	io.WriteString(w, "["+strings.Join(items, ", ")+"]")
    }

    func (f *testFeed) requests() []string {
    	f.mu.Lock()
    	defer f.mu.Unlock()
    	return slices.Clone(f.asked)
    }

    // testCollect iterates recentIssues, stopping after stopAfter issues (0 = never),
    // and returns the ids, the errors and the pages the server was asked for.
    func testCollect(t *testing.T, pages map[string]testPage, since time.Time, stopAfter int) (ids []int, errs []error, asked []string) {
    	t.Helper()
    	f := &testFeed{pages: pages}
    	srv := httptest.NewServer(f)
    	defer srv.Close()
    	f.url = srv.URL
    	tb := &testBodyTracker{next: srv.Client().Transport}
    	seq := recentIssues(t.Context(), &http.Client{Transport: tb}, srv.URL+"/v1/activity", since)
    	if n := len(f.requests()); n != 0 {
    		t.Errorf("recentIssues sent %d request(s) before iteration started; fetch lazily", n)
    	}
    	for iss, err := range seq {
    		if err != nil {
    			errs = append(errs, err)
    			if iss != (Issue{}) {
    				t.Errorf("error %v was yielded with issue %+v, want Issue{}", err, iss)
    			}
    			continue // a correct iterator stops by itself after an error
    		}
    		if iss.Title != fmt.Sprintf("Issue %d", iss.ID) || !iss.UpdatedAt.Equal(testDay(iss.ID)) {
    			t.Errorf("yielded %+v, want issue %d with title %q updated %v", iss, iss.ID, fmt.Sprintf("Issue %d", iss.ID), testDay(iss.ID))
    		}
    		ids = append(ids, iss.ID)
    		if len(ids) == stopAfter {
    			break
    		}
    		if len(ids) > 100 {
    			t.Fatal("more than 100 issues yielded: is the iterator looping?")
    		}
    	}
    	if tb.opened.Load() != tb.closed.Load() {
    		t.Errorf("%d responses but %d body closes: close every body", tb.opened.Load(), tb.closed.Load())
    	}
    	return ids, errs, f.requests()
    }

    var testThreePages = map[string]testPage{
    	"/v1/activity":        {ids: []int{27, 26, 25}, link: `</v1/activity?page=2>; rel="next"`},
    	"/v1/activity?page=2": {ids: []int{22, 20, 19}, link: `<{srv}/v1/activity?page=3>; rel="next", </v1/activity>; rel="prev"`},
    	"/v1/activity?page=3": {ids: []int{12, 3}, link: `</v1/activity?page=2>; rel="prev"`},
    }

    func TestRecentAll(t *testing.T) {
    	ids, errs, asked := testCollect(t, testThreePages, testDay(1), 0)
    	if want := []int{27, 26, 25, 22, 20, 19, 12, 3}; !slices.Equal(ids, want) || errs != nil {
    		t.Errorf("since Sep 1: got ids %v and errors %v, want %v and none", ids, errs, want)
    	}
    	if want := []string{"/v1/activity", "/v1/activity?page=2", "/v1/activity?page=3"}; !slices.Equal(asked, want) {
    		t.Errorf("server was asked for %v, want %v (follow relative and absolute next links)", asked, want)
    	}
    }

    func TestRecentStopsAtOlderIssue(t *testing.T) {
    	tests := []struct {
    		since     time.Time
    		wantIDs   []int
    		wantPages int
    	}{
    		{testDay(21), []int{27, 26, 25, 22}, 2},
    		{testDay(20), []int{27, 26, 25, 22, 20}, 2}, // updated exactly at since: included
    		{testDay(20).Add(time.Second), []int{27, 26, 25, 22}, 2},
    		{testDay(25), []int{27, 26, 25}, 2}, // page 1 ends exactly at since; page 2 starts older
    		{testDay(26), []int{27, 26}, 1},
    		{testDay(28), nil, 1},
    		{testDay(12), []int{27, 26, 25, 22, 20, 19, 12}, 3},
    	}
    	for _, tt := range tests {
    		ids, errs, asked := testCollect(t, testThreePages, tt.since, 0)
    		if !slices.Equal(ids, tt.wantIDs) || errs != nil {
    			t.Errorf("since %v: got ids %v and errors %v, want %v and none", tt.since.Format(time.DateTime), ids, errs, tt.wantIDs)
    		}
    		if len(asked) != tt.wantPages {
    			t.Errorf("since %v: fetched %d pages %v, want %d (stop at the first older issue, without fetching more)", tt.since.Format(time.DateTime), len(asked), asked, tt.wantPages)
    		}
    	}
    }

    func TestRecentConsumerBreaks(t *testing.T) {
    	for _, tt := range []struct{ stop, wantPages int }{{1, 1}, {3, 1}, {4, 2}, {6, 2}} {
    		ids, _, asked := testCollect(t, testThreePages, testDay(1), tt.stop)
    		if len(ids) != tt.stop {
    			t.Errorf("break after %d: got %d issues", tt.stop, len(ids))
    		}
    		if len(asked) != tt.wantPages {
    			t.Errorf("break after %d issues: fetched %d pages %v, want %d (don't fetch pages nobody asked for)", tt.stop, len(asked), asked, tt.wantPages)
    		}
    	}
    }

    func TestRecentEmptyAndLinkFormats(t *testing.T) {
    	ids, errs, asked := testCollect(t, map[string]testPage{"/v1/activity": {ids: nil, body: `[]`}}, testDay(1), 0)
    	if ids != nil || errs != nil || len(asked) != 1 {
    		t.Errorf("empty feed: got ids %v, errors %v after %d requests, want nothing after 1", ids, errs, len(asked))
    	}

    	pages := map[string]testPage{
    		"/v1/activity":           {ids: []int{27}, link: `</v1/activity?c=b%3D%3D>; rel=next`},
    		"/v1/activity?c=b%3D%3D": {body: `[]`, link: `<page-3>; title="the end"; rel="last next"`},
    		"/v1/page-3":             {ids: []int{9}, link: `</v1/activity>; rel="first"`},
    	}
    	ids, errs, asked = testCollect(t, pages, testDay(1), 0)
    	if !slices.Equal(ids, []int{27, 9}) || errs != nil {
    		t.Errorf("unquoted rel, an empty middle page, rel=\"last next\" and a path-relative link: got ids %v, errors %v (asked for %v), want [27 9] and none", ids, errs, asked)
    	}
    }

    func TestRecentErrors(t *testing.T) {
    	tests := []struct {
    		why     string
    		page3   testPage
    		wantIDs []int
    	}{
    		{"page 3 answers 500", testPage{status: 500, body: `{"error": "oops"}`}, []int{27, 26, 25, 22, 20, 19}},
    		{"page 3 is malformed", testPage{body: `[{"id": 12, "title": "Iss`}, []int{27, 26, 25, 22, 20, 19}},
    		{"page 3 has a bad timestamp", testPage{body: `[{"id": 12, "title": "Issue 12", "updated_at": "yesterday"}]`}, []int{27, 26, 25, 22, 20, 19}},
    	}
    	for _, tt := range tests {
    		pages := map[string]testPage{
    			"/v1/activity":        testThreePages["/v1/activity"],
    			"/v1/activity?page=2": testThreePages["/v1/activity?page=2"],
    			"/v1/activity?page=3": tt.page3,
    		}
    		ids, errs, asked := testCollect(t, pages, testDay(1), 0)
    		if !slices.Equal(ids, tt.wantIDs) || len(errs) != 1 {
    			t.Errorf("%s: got ids %v and %d error(s) %v, want %v and exactly one error, then stop", tt.why, ids, len(errs), errs, tt.wantIDs)
    		}
    		if len(asked) != 3 {
    			t.Errorf("%s: fetched %v, want 3 pages and no more", tt.why, asked)
    		}
    	}
    }

    func TestRecentLoop(t *testing.T) {
    	pages := map[string]testPage{
    		"/v1/activity":        {ids: []int{27, 26}, link: `</v1/activity?page=2>; rel="next"`},
    		"/v1/activity?page=2": {ids: []int{25}, link: `</v1/activity?page=3>; rel="next"`},
    		"/v1/activity?page=3": {ids: []int{24}, link: `<{srv}/v1/activity?page=2>; rel="next"`},
    	}
    	ids, errs, asked := testCollect(t, pages, testDay(1), 0)
    	if !slices.Equal(ids, []int{27, 26, 25, 24}) || len(errs) != 1 || !errors.Is(errs[0], ErrPageLoop) {
    		t.Errorf("page 3 links back to page 2: got ids %v and errors %v, want [27 26 25 24] and one error wrapping ErrPageLoop", ids, errs)
    	}
    	if len(asked) != 3 {
    		t.Errorf("page 3 links back to page 2: fetched %v, want 3 pages (don't fetch page 2 again)", asked)
    	}
    }

    func TestRecentCancelled(t *testing.T) {
    	f := &testFeed{pages: testThreePages}
    	srv := httptest.NewServer(f)
    	defer srv.Close()
    	ctx, cancel := context.WithCancel(t.Context())
    	cancel()
    	var errs []error
    	for _, err := range recentIssues(ctx, srv.Client(), srv.URL+"/v1/activity", testDay(1)) {
    		errs = append(errs, err)
    		if len(errs) > 5 {
    			break
    		}
    	}
    	if len(errs) != 1 || !errors.Is(errs[0], context.Canceled) {
    		t.Errorf("cancelled context: yielded errors %v, want exactly one, wrapping context.Canceled", errs)
    	}
    }
---

`trackr activity --since 2026-09-20` lists recently updated issues. The
activity feed is sorted **newest first** and paged with `Link` headers:

```
GET /v1/activity
Link: </v1/activity?page=2>; rel="next"

[{"id": 7, "title": "Dark mode", "updated_at": "2026-09-27T09:30:00Z"}, ...]
```

The feed goes back years, but `trackr` only needs the last week. Because it's
sorted, the first issue older than `--since` means every later issue is older
too, so there's no reason to fetch another page.

Complete `recentIssues(ctx, client, firstURL, since)`. It returns an
`iter.Seq2[Issue, error]` that:

- Fetches **lazily**: nothing is requested until the caller starts ranging,
  and each page only when the previous one has been used up.
- Starts at `firstURL` and follows the `rel="next"` link of each page until a
  page has none. Links may be absolute, root-relative (`/v1/...`) or
  path-relative (`page-3`), and are resolved against the URL of the page they
  came from.
- Yields each issue updated **at or after** `since`, in feed order. At the
  first issue updated before `since`, it stops: that issue isn't yielded and no
  more pages are fetched.
- Stops without fetching more if the caller breaks out of the loop.
- On a failed request, a status other than `200 OK` or malformed JSON, yields
  `(Issue{}, err)` **once** and stops.
- If a next link points to a page it already fetched, yields an error wrapping
  `ErrPageLoop` once and stops, instead of going round forever.

## Parsing Link

A `Link` header holds comma-separated entries like
`<url>; rel="next"`. An entry can have several parameters in any order, `rel`
may be quoted or not, and it may list several relations (`rel="last next"`).
An entry is the next link if one of its relations is `next`. The tests use all
of these forms. (You can assume URLs in `Link` never contain commas.)

## Constraints

- Pages can be empty (`[]`) and still have a next link.
- Compare times with `Before`/`After`, not `==`.
- The tests record every page the fake API was asked for, and count response
  bodies: close every one.
