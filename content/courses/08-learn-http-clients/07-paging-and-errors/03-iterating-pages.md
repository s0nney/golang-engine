---
title: Iterating Pages with iter.Seq2
quiz:
  - question: |
      What happens when the caller `break`s out of this loop after the first issue?

      ```go
      return func(yield func(Issue, error) bool) {
          for _, iss := range p.Issues {
              yield(iss, nil)
          }
      }
      ```
    options:
      - text: The loop stops cleanly
      - text: The iterator keeps calling `yield`, and the program panics
        correct: true
      - text: The remaining issues are silently dropped
    explanation: |
      After a `break`, `yield` returns false and must not be called again. Ignoring its
      result makes the runtime panic with "range function continued iteration after
      function for loop body returned false". Write `if !yield(iss, nil) { return }`.
  - question: A caller ranges over `allIssues(...)` and breaks after 5 issues. With 50 issues per page, how many pages should be fetched?
    options:
      - text: One
        correct: true
      - text: All of them, since the iterator loads everything first
      - text: Two, because the iterator prefetches the next page
    explanation: |
      A well-written iterator is lazy: it fetches a page only when the caller wants an
      item from it. Breaking early means no more requests.
exercise:
  starter: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"fmt"
    	"iter"
    	"net/http"
    	"net/http/httptest"
    	"net/url"
    	"strconv"
    )

    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    }

    // page is one page of GET /issues.
    type page struct {
    	Issues     []Issue `json:"issues"`
    	NextCursor string  `json:"next_cursor"` // "" on the last page
    }

    // fetchPage GETs baseURL/issues, adding ?cursor=... when cursor isn't empty.
    func fetchPage(ctx context.Context, client *http.Client, baseURL, cursor string) (page, error) {
    	u, err := url.Parse(baseURL + "/issues")
    	if err != nil {
    		return page{}, err
    	}
    	if cursor != "" {
    		u.RawQuery = url.Values{"cursor": {cursor}}.Encode()
    	}
    	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
    	if err != nil {
    		return page{}, err
    	}
    	resp, err := client.Do(req)
    	if err != nil {
    		return page{}, err
    	}
    	defer resp.Body.Close()
    	if resp.StatusCode != http.StatusOK {
    		return page{}, fmt.Errorf("GET %s: %s", u, resp.Status)
    	}
    	var p page
    	if err := json.UnmarshalRead(resp.Body, &p); err != nil {
    		return page{}, fmt.Errorf("decoding page: %w", err)
    	}
    	return p, nil
    }

    // allIssues yields every issue, page by page, fetching each page only when
    // it's needed. On an error it yields (Issue{}, err) once and stops. If the
    // caller stops early, no more pages are fetched.
    func allIssues(ctx context.Context, client *http.Client, baseURL string) iter.Seq2[Issue, error] {
    	return func(yield func(Issue, error) bool) {
    		// ?
    	}
    }

    // fakeAPI serves 7 issues, 3 per page, with opaque cursors.
    func fakeAPI() *httptest.Server {
    	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		start := 0
    		if c := r.URL.Query().Get("cursor"); c != "" {
    			start, _ = strconv.Atoi(c[len("c-"):])
    		}
    		var p page
    		for id := start + 1; id <= min(start+3, 7); id++ {
    			p.Issues = append(p.Issues, Issue{ID: id, Title: fmt.Sprintf("Issue %d", id)})
    		}
    		if start+3 < 7 {
    			p.NextCursor = fmt.Sprintf("c-%d", start+3)
    		}
    		fmt.Printf("  (server: page cursor=%q)\n", r.URL.Query().Get("cursor"))
    		json.MarshalWrite(w, p)
    	}))
    }

    func main() {
    	srv := fakeAPI()
    	defer srv.Close()

    	for iss, err := range allIssues(context.Background(), srv.Client(), srv.URL) {
    		if err != nil {
    			fmt.Println("error:", err)
    			break
    		}
    		fmt.Printf("#%d %s\n", iss.ID, iss.Title)
    		if iss.ID == 5 {
    			break // we only wanted the first five
    		}
    	}
    	fmt.Println("done")
    }
  solution: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"fmt"
    	"iter"
    	"net/http"
    	"net/http/httptest"
    	"net/url"
    	"strconv"
    )

    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    }

    // page is one page of GET /issues.
    type page struct {
    	Issues     []Issue `json:"issues"`
    	NextCursor string  `json:"next_cursor"` // "" on the last page
    }

    // fetchPage GETs baseURL/issues, adding ?cursor=... when cursor isn't empty.
    func fetchPage(ctx context.Context, client *http.Client, baseURL, cursor string) (page, error) {
    	u, err := url.Parse(baseURL + "/issues")
    	if err != nil {
    		return page{}, err
    	}
    	if cursor != "" {
    		u.RawQuery = url.Values{"cursor": {cursor}}.Encode()
    	}
    	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
    	if err != nil {
    		return page{}, err
    	}
    	resp, err := client.Do(req)
    	if err != nil {
    		return page{}, err
    	}
    	defer resp.Body.Close()
    	if resp.StatusCode != http.StatusOK {
    		return page{}, fmt.Errorf("GET %s: %s", u, resp.Status)
    	}
    	var p page
    	if err := json.UnmarshalRead(resp.Body, &p); err != nil {
    		return page{}, fmt.Errorf("decoding page: %w", err)
    	}
    	return p, nil
    }

    func allIssues(ctx context.Context, client *http.Client, baseURL string) iter.Seq2[Issue, error] {
    	return func(yield func(Issue, error) bool) {
    		cursor := ""
    		for {
    			p, err := fetchPage(ctx, client, baseURL, cursor)
    			if err != nil {
    				yield(Issue{}, err)
    				return
    			}
    			for _, iss := range p.Issues {
    				if !yield(iss, nil) {
    					return
    				}
    			}
    			if p.NextCursor == "" {
    				return
    			}
    			cursor = p.NextCursor
    		}
    	}
    }

    // fakeAPI serves 7 issues, 3 per page, with opaque cursors.
    func fakeAPI() *httptest.Server {
    	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		start := 0
    		if c := r.URL.Query().Get("cursor"); c != "" {
    			start, _ = strconv.Atoi(c[len("c-"):])
    		}
    		var p page
    		for id := start + 1; id <= min(start+3, 7); id++ {
    			p.Issues = append(p.Issues, Issue{ID: id, Title: fmt.Sprintf("Issue %d", id)})
    		}
    		if start+3 < 7 {
    			p.NextCursor = fmt.Sprintf("c-%d", start+3)
    		}
    		fmt.Printf("  (server: page cursor=%q)\n", r.URL.Query().Get("cursor"))
    		json.MarshalWrite(w, p)
    	}))
    }

    func main() {
    	srv := fakeAPI()
    	defer srv.Close()

    	for iss, err := range allIssues(context.Background(), srv.Client(), srv.URL) {
    		if err != nil {
    			fmt.Println("error:", err)
    			break
    		}
    		fmt.Printf("#%d %s\n", iss.ID, iss.Title)
    		if iss.ID == 5 {
    			break // we only wanted the first five
    		}
    	}
    	fmt.Println("done")
    }
  tests: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    	"testing"
    )

    // pagedAPI serves issues 1..total, per per page. Cursors are "p<N>".
    // If failOn > 0, the request for page failOn gets a 500.
    func pagedAPI(total, per, failOn int, cursors *[]string) *httptest.Server {
    	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		if r.URL.Path != "/issues" {
    			http.NotFound(w, r)
    			return
    		}
    		c := r.URL.Query().Get("cursor")
    		*cursors = append(*cursors, c)
    		n := 1
    		if c != "" {
    			fmt.Sscanf(c, "p%d", &n)
    		}
    		if n == failOn {
    			http.Error(w, `{"error":"db timeout"}`, http.StatusInternalServerError)
    			return
    		}
    		p := page{Issues: []Issue{}}
    		for id := (n-1)*per + 1; id <= min(n*per, total); id++ {
    			p.Issues = append(p.Issues, Issue{ID: id, Title: fmt.Sprint("issue ", id)})
    		}
    		if n*per < total {
    			p.NextCursor = fmt.Sprintf("p%d", n+1)
    		}
    		json.MarshalWrite(w, p)
    	}))
    }

    func collect(t *testing.T, srv *httptest.Server) ([]int, error) {
    	t.Helper()
    	var ids []int
    	for iss, err := range allIssues(context.Background(), srv.Client(), srv.URL) {
    		if err != nil {
    			return ids, err
    		}
    		ids = append(ids, iss.ID)
    		if len(ids) > 100 {
    			t.Fatal("allIssues yielded more than 100 issues; is it stuck on one page? (use NextCursor)")
    		}
    	}
    	return ids, nil
    }

    func TestAllPages(t *testing.T) {
    	var cursors []string
    	srv := pagedAPI(10, 4, 0, &cursors)
    	defer srv.Close()

    	ids, err := collect(t, srv)
    	if err != nil {
    		t.Fatalf("allIssues yielded error %v", err)
    	}
    	if want := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}; !slices.Equal(ids, want) {
    		t.Errorf("allIssues yielded IDs %v, want %v", ids, want)
    	}
    	if want := []string{"", "p2", "p3"}; !slices.Equal(cursors, want) {
    		t.Errorf("server saw cursors %q, want %q (first request without a cursor)", cursors, want)
    	}
    }

    func TestSinglePage(t *testing.T) {
    	var cursors []string
    	srv := pagedAPI(2, 5, 0, &cursors)
    	defer srv.Close()
    	ids, err := collect(t, srv)
    	if err != nil || !slices.Equal(ids, []int{1, 2}) {
    		t.Errorf("allIssues = %v, %v; want [1 2], nil", ids, err)
    	}
    	if len(cursors) != 1 {
    		t.Errorf("made %d requests for a single page, want 1 (stop when next_cursor is empty)", len(cursors))
    	}
    }

    func TestEmpty(t *testing.T) {
    	var cursors []string
    	srv := pagedAPI(0, 5, 0, &cursors)
    	defer srv.Close()
    	ids, err := collect(t, srv)
    	if err != nil || len(ids) != 0 {
    		t.Errorf("allIssues on an empty project = %v, %v; want nothing and no error", ids, err)
    	}
    }

    func TestErrorMidway(t *testing.T) {
    	var cursors []string
    	srv := pagedAPI(20, 3, 3, &cursors)
    	defer srv.Close()

    	var ids []int
    	var errs []error
    	for iss, err := range allIssues(context.Background(), srv.Client(), srv.URL) {
    		if err != nil {
    			errs = append(errs, err)
    			continue // keep ranging: the iterator itself must stop
    		}
    		ids = append(ids, iss.ID)
    	}
    	if !slices.Equal(ids, []int{1, 2, 3, 4, 5, 6}) {
    		t.Errorf("before the failing page, got IDs %v, want [1 2 3 4 5 6]", ids)
    	}
    	if len(errs) != 1 {
    		t.Errorf("got %d errors, want exactly 1: yield the error once, then return", len(errs))
    	}
    	if len(cursors) != 3 {
    		t.Errorf("server saw %d requests, want 3 (stop after the failed page)", len(cursors))
    	}
    }

    func TestStopsEarly(t *testing.T) {
    	var cursors []string
    	srv := pagedAPI(100, 5, 0, &cursors)
    	defer srv.Close()

    	var ids []int
    	for iss, err := range allIssues(context.Background(), srv.Client(), srv.URL) {
    		if err != nil {
    			t.Fatalf("unexpected error %v", err)
    		}
    		ids = append(ids, iss.ID)
    		if iss.ID == 7 {
    			break
    		}
    	}
    	if !slices.Equal(ids, []int{1, 2, 3, 4, 5, 6, 7}) {
    		t.Errorf("got IDs %v, want 1..7", ids)
    	}
    	if len(cursors) != 2 {
    		t.Errorf("breaking after issue 7 (page 2) made %d requests, want 2 (check yield's return value)", len(cursors))
    	}
    }
---

Every `trackr` command that lists things (issues, projects, users, comments) needs the
same paging loop. Copying it everywhere is how bugs multiply. Go 1.23's range-over-func
iterators let you write the loop **once** and hand callers something they can simply
`range` over.

## The shape

`iter.Seq2[K, V]` is a function type:

```go
type Seq2[K, V any] func(yield func(K, V) bool)
```

For fallible sequences, the Go convention is `iter.Seq2[T, error]`: each step yields
either an item with a nil error, or a zero item with an error. The caller writes:

```go
for iss, err := range allIssues(ctx, client, baseURL) {
	if err != nil {
		return err
	}
	fmt.Printf("#%d %s\n", iss.ID, iss.Title)
}
```

No cursors, no pages, no `NextCursor` checks. All of that lives inside `allIssues`.

## Writing it

```go
func allIssues(ctx context.Context, client *http.Client, baseURL string) iter.Seq2[Issue, error] {
	return func(yield func(Issue, error) bool) {
		// fetch a page, yield its issues, move to the next cursor, repeat
	}
}
```

Three rules make it correct:

1. **Respect `yield`'s answer.** `yield` returns false when the caller has stopped
   (with `break`, `return` or a panic). From then on you must not call it again. Just
   `return`. Otherwise the program panics.
2. **Yield an error once, then stop.** After `yield(Issue{}, err)`, return. Even if the
   caller keeps ranging, a broken page means you can't know the next cursor.
3. **Be lazy.** Fetch a page only when you need its first item. If the caller stops
   after 5 issues, you never request page 2. That saves API calls and rate-limit
   budget.

The body is closed before the next page is fetched (`fetchPage` handles one request
with its own `defer`), so the iterator never holds more than one connection.

## Why a function, not a slice?

You could return `[]Issue` with every issue from every page. But:

- 12,000 issues means 12,000 structs in memory before you print the first one.
- The user waits for **all** pages before seeing anything.
- `trackr search --first` wants one match and would download everything anyway.

The iterator streams: first results appear after one request, memory stays at one
page, and stopping early is free. When a caller *does* want a slice, that's easy to
build from the iterator (a `slices.Collect`-style helper that stops at the first
error).

## Context still matters

The iterator captures `ctx` and uses it for every page request. If the user presses
Ctrl+C halfway through, the current page request fails with `context.Canceled`, the
iterator yields that error, and the loop ends.

## Your turn

`fetchPage` is written for you. It GETs one page and decodes this shape:

```json
{"issues": [{"id": 1, "title": "..."}, ...], "next_cursor": "c-3"}
```

Complete `allIssues` so that it:

1. fetches the first page with an empty cursor;
2. yields each issue on it with a nil error, and **returns straight away** if `yield`
   returns false;
3. moves on to `NextCursor`, stopping when it's empty;
4. on a fetch error, yields `(Issue{}, err)` **once** and returns.

Run it: the starter's `main` breaks after issue 5, and the fake server logs each page
request. Only two pages should be requested.
