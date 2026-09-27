---
title: Link Headers
quiz:
  - question: |
      Given this header, what URL should `trackr` request next?

      ```
      Link: <https://api.trackr.dev/issues?page=1>; rel="first",
            <https://api.trackr.dev/issues?page=4>; rel="next",
            <https://api.trackr.dev/issues?page=2>; rel="prev"
      ```
    options:
      - text: '`https://api.trackr.dev/issues?page=1`'
      - text: '`https://api.trackr.dev/issues?page=2`'
      - text: '`https://api.trackr.dev/issues?page=4`'
        correct: true
    explanation: |
      Pick the link by its `rel`, not by its position. The order of links in the
      header means nothing.
  - question: How does a client know it has reached the last page with Link-header pagination?
    options:
      - text: The response is empty
      - text: There's no link with `rel="next"`
        correct: true
      - text: The status code is 204
    explanation: |
      The server simply leaves out the `next` link on the last page. An empty page
      isn't a reliable signal: the final page can be full.
exercise:
  starter: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strconv"
    	"strings"
    )

    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    }

    // parseLinks parses a Link header into a map from rel to URL.
    func parseLinks(h http.Header) map[string]string {
    	links := map[string]string{}
    	for _, v := range h.Values("Link") {
    		for part := range strings.SplitSeq(v, ",") {
    			target, params, ok := strings.Cut(part, ";")
    			if !ok {
    				continue
    			}
    			target = strings.TrimSpace(target)
    			if !strings.HasPrefix(target, "<") || !strings.HasSuffix(target, ">") {
    				continue
    			}
    			target = target[1 : len(target)-1]
    			for p := range strings.SplitSeq(params, ";") {
    				key, val, _ := strings.Cut(strings.TrimSpace(p), "=")
    				if strings.EqualFold(key, "rel") {
    					for rel := range strings.FieldsSeq(strings.Trim(val, `"`)) {
    						links[strings.ToLower(rel)] = target
    					}
    				}
    			}
    		}
    	}
    	return links
    }

    // listAll GETs firstURL, then keeps following the rel="next" link of each
    // response until there isn't one, and returns every issue from every page.
    // Relative links are resolved against the URL of the page they came from.
    // Any status other than 200 OK is an error.
    func listAll(ctx context.Context, client *http.Client, firstURL string) ([]Issue, error) {
    	var all []Issue
    	next := firstURL
    	// ?
    	_ = next
    	return all, nil
    }

    func main() {
    	// A fake API: 5 issues, 2 per page. Its next links are relative.
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    		page = max(page, 1)
    		fmt.Printf("  (server: %s)\n", r.URL.RequestURI())
    		var issues []Issue
    		for id := page*2 - 1; id <= min(page*2, 5); id++ {
    			issues = append(issues, Issue{ID: id, Title: fmt.Sprint("Issue ", id)})
    		}
    		if page*2 < 5 {
    			w.Header().Set("Link", fmt.Sprintf(`</issues?page=%d>; rel="next"`, page+1))
    		}
    		json.MarshalWrite(w, issues)
    	}))
    	defer srv.Close()

    	issues, err := listAll(context.Background(), srv.Client(), srv.URL+"/issues")
    	fmt.Println(len(issues), "issues", err)
    	for _, iss := range issues {
    		fmt.Printf("#%d %s\n", iss.ID, iss.Title)
    	}
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
    	"strings"
    )

    type Issue struct {
    	ID    int    `json:"id"`
    	Title string `json:"title"`
    }

    // parseLinks parses a Link header into a map from rel to URL.
    func parseLinks(h http.Header) map[string]string {
    	links := map[string]string{}
    	for _, v := range h.Values("Link") {
    		for part := range strings.SplitSeq(v, ",") {
    			target, params, ok := strings.Cut(part, ";")
    			if !ok {
    				continue
    			}
    			target = strings.TrimSpace(target)
    			if !strings.HasPrefix(target, "<") || !strings.HasSuffix(target, ">") {
    				continue
    			}
    			target = target[1 : len(target)-1]
    			for p := range strings.SplitSeq(params, ";") {
    				key, val, _ := strings.Cut(strings.TrimSpace(p), "=")
    				if strings.EqualFold(key, "rel") {
    					for rel := range strings.FieldsSeq(strings.Trim(val, `"`)) {
    						links[strings.ToLower(rel)] = target
    					}
    				}
    			}
    		}
    	}
    	return links
    }

    func listAll(ctx context.Context, client *http.Client, firstURL string) ([]Issue, error) {
    	var all []Issue
    	next := firstURL
    	for next != "" {
    		issues, link, err := getPage(ctx, client, next)
    		if err != nil {
    			return nil, err
    		}
    		all = append(all, issues...)
    		next = link
    	}
    	return all, nil
    }

    // getPage fetches one page and returns its issues and the absolute URL of the
    // next page ("" on the last page).
    func getPage(ctx context.Context, client *http.Client, pageURL string) ([]Issue, string, error) {
    	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
    	if err != nil {
    		return nil, "", err
    	}
    	resp, err := client.Do(req)
    	if err != nil {
    		return nil, "", err
    	}
    	defer resp.Body.Close()

    	if resp.StatusCode != http.StatusOK {
    		return nil, "", fmt.Errorf("GET %s: %s", pageURL, resp.Status)
    	}
    	var issues []Issue
    	if err := json.UnmarshalRead(resp.Body, &issues); err != nil {
    		return nil, "", fmt.Errorf("decoding %s: %w", pageURL, err)
    	}

    	next := parseLinks(resp.Header)["next"]
    	if next == "" {
    		return issues, "", nil
    	}
    	ref, err := url.Parse(next)
    	if err != nil {
    		return nil, "", err
    	}
    	return issues, resp.Request.URL.ResolveReference(ref).String(), nil
    }

    func main() {
    	// A fake API: 5 issues, 2 per page. Its next links are relative.
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    		page = max(page, 1)
    		fmt.Printf("  (server: %s)\n", r.URL.RequestURI())
    		var issues []Issue
    		for id := page*2 - 1; id <= min(page*2, 5); id++ {
    			issues = append(issues, Issue{ID: id, Title: fmt.Sprint("Issue ", id)})
    		}
    		if page*2 < 5 {
    			w.Header().Set("Link", fmt.Sprintf(`</issues?page=%d>; rel="next"`, page+1))
    		}
    		json.MarshalWrite(w, issues)
    	}))
    	defer srv.Close()

    	issues, err := listAll(context.Background(), srv.Client(), srv.URL+"/issues")
    	fmt.Println(len(issues), "issues", err)
    	for _, iss := range issues {
    		fmt.Printf("#%d %s\n", iss.ID, iss.Title)
    	}
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
    	"strconv"
    	"testing"
    )

    // linkAPI serves issues 1..total, per per page, at /v2/items?page=N.
    // absolute chooses absolute or relative next links. failPage, if > 0, gets a 500.
    func linkAPI(total, per, failPage int, absolute bool, seen *[]string) *httptest.Server {
    	var srv *httptest.Server
    	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		*seen = append(*seen, r.URL.RequestURI())
    		if r.URL.Path != "/v2/items" {
    			http.NotFound(w, r)
    			return
    		}
    		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    		page = max(page, 1)
    		if page == failPage {
    			http.Error(w, `{"error":"db timeout"}`, http.StatusInternalServerError)
    			return
    		}
    		issues := []Issue{}
    		for id := (page-1)*per + 1; id <= min(page*per, total); id++ {
    			issues = append(issues, Issue{ID: id, Title: fmt.Sprint("issue ", id)})
    		}
    		if page*per < total {
    			next := fmt.Sprintf("/v2/items?page=%d&per_page=%d", page+1, per)
    			if absolute {
    				next = srv.URL + next
    			}
    			w.Header().Add("Link", fmt.Sprintf(`<%s/v2/items?page=1>; rel="first"`, srv.URL))
    			w.Header().Add("Link", fmt.Sprintf(`<%s>; rel="next"`, next))
    		}
    		json.MarshalWrite(w, issues)
    	}))
    	return srv
    }

    func ids(issues []Issue) []int {
    	var out []int
    	for _, iss := range issues {
    		out = append(out, iss.ID)
    	}
    	return out
    }

    func TestListAllAbsoluteLinks(t *testing.T) {
    	var seen []string
    	srv := linkAPI(7, 3, 0, true, &seen)
    	defer srv.Close()

    	got, err := listAll(context.Background(), srv.Client(), srv.URL+"/v2/items")
    	if err != nil {
    		t.Fatalf("listAll returned error %v", err)
    	}
    	if want := []int{1, 2, 3, 4, 5, 6, 7}; !slices.Equal(ids(got), want) {
    		t.Errorf("listAll returned IDs %v, want %v", ids(got), want)
    	}
    	if want := []string{"/v2/items", "/v2/items?page=2&per_page=3", "/v2/items?page=3&per_page=3"}; !slices.Equal(seen, want) {
    		t.Errorf("server saw %q, want %q (follow rel=\"next\" exactly as given)", seen, want)
    	}
    }

    func TestListAllRelativeLinks(t *testing.T) {
    	var seen []string
    	srv := linkAPI(5, 2, 0, false, &seen)
    	defer srv.Close()

    	got, err := listAll(context.Background(), srv.Client(), srv.URL+"/v2/items")
    	if err != nil {
    		t.Fatalf("with relative next links: listAll returned error %v (resolve them with resp.Request.URL.ResolveReference)", err)
    	}
    	if want := []int{1, 2, 3, 4, 5}; !slices.Equal(ids(got), want) {
    		t.Errorf("with relative next links: listAll returned IDs %v, want %v", ids(got), want)
    	}
    }

    func TestListAllSinglePage(t *testing.T) {
    	var seen []string
    	srv := linkAPI(2, 10, 0, true, &seen)
    	defer srv.Close()

    	got, err := listAll(context.Background(), srv.Client(), srv.URL+"/v2/items")
    	if err != nil || !slices.Equal(ids(got), []int{1, 2}) {
    		t.Errorf("listAll = %v, %v; want [1 2], nil", ids(got), err)
    	}
    	if len(seen) != 1 {
    		t.Errorf("made %d requests for a single page, want 1 (stop when there's no next link)", len(seen))
    	}
    }

    func TestListAllFailingPage(t *testing.T) {
    	var seen []string
    	srv := linkAPI(20, 3, 2, true, &seen)
    	defer srv.Close()

    	got, err := listAll(context.Background(), srv.Client(), srv.URL+"/v2/items")
    	if err == nil {
    		t.Errorf("page 2 returned 500: listAll = (%v, nil), want an error", ids(got))
    	}
    	if len(seen) != 2 {
    		t.Errorf("server saw %d requests, want 2 (stop at the failing page)", len(seen))
    	}
    }
---

Some APIs (GitHub's is the famous one) put pagination in a `Link` response header.
The body is just the array of items, and the header says where to go next:

```
Link: <https://api.trackr.dev/issues?page=3&per_page=50>; rel="next",
      <https://api.trackr.dev/issues?page=1&per_page=50>; rel="prev first",
      <https://api.trackr.dev/issues?page=9&per_page=50>; rel="last"
```

(It arrives as one long line; it's wrapped here to fit.)

## The format

The header is a comma-separated list of links. Each link is:

- a URL in angle brackets: `<https://...>`;
- then `;`-separated parameters, most importantly `rel`, the **relation**. One link
  can have several space-separated rels (`rel="prev first"`).

The common rels are `next`, `prev`, `first` and `last`. For paging you only need
`next`: follow it until it's gone.

## Why a client should love it

With Link headers the client doesn't build page URLs at all. It doesn't need to know
whether the server uses offsets, page numbers or cursors, or what the parameters are
called. It just follows `next`. The server can change its paging scheme and clients
keep working.

Always use the URL **as given**. Don't rebuild it from pieces, or you might drop a
parameter the server needs.

## Parsing it

The standard library doesn't have a Link-header parser, but a small one covers what
real APIs send:

```go
package main

import (
	"fmt"
	"net/http"
	"strings"
)

// parseLinks parses a Link header into a map from rel to URL.
func parseLinks(h http.Header) map[string]string {
	links := map[string]string{}
	for _, v := range h.Values("Link") {
		for part := range strings.SplitSeq(v, ",") {
			target, params, ok := strings.Cut(part, ";")
			if !ok {
				continue
			}
			target = strings.TrimSpace(target)
			if !strings.HasPrefix(target, "<") || !strings.HasSuffix(target, ">") {
				continue
			}
			target = target[1 : len(target)-1]
			for p := range strings.SplitSeq(params, ";") {
				key, val, _ := strings.Cut(strings.TrimSpace(p), "=")
				if strings.EqualFold(key, "rel") {
					for rel := range strings.FieldsSeq(strings.Trim(val, `"`)) {
						links[strings.ToLower(rel)] = target
					}
				}
			}
		}
	}
	return links
}

func main() {
	h := http.Header{}
	h.Set("Link", `<https://api.trackr.dev/issues?page=3&per_page=50>; rel="next", `+
		`<https://api.trackr.dev/issues?page=1&per_page=50>; rel="prev first", `+
		`<https://api.trackr.dev/issues?page=9&per_page=50>; rel="last"`)
	links := parseLinks(h)
	for _, rel := range []string{"first", "prev", "next", "last"} {
		fmt.Printf("%-5s %s\n", rel, links[rel])
	}
}
```

Output:

```
first https://api.trackr.dev/issues?page=1&per_page=50
prev  https://api.trackr.dev/issues?page=1&per_page=50
next  https://api.trackr.dev/issues?page=3&per_page=50
last  https://api.trackr.dev/issues?page=9&per_page=50
```

A few details:

- `h.Values("Link")` handles servers that send several `Link` headers instead of one
  comma-separated header.
- `strings.SplitSeq` and `strings.FieldsSeq` (Go 1.24) return iterators, so there are no
  intermediate slices.
- Rel names are case-insensitive, so they're lowercased.
- This simple split breaks if a URL itself contains a comma. That's rare in practice
  (commas in URLs are usually escaped as `%2C`), but a fully spec-compliant parser has
  to handle quoting properly.

## The loop

```go
next := baseURL + "/issues?per_page=100"
for next != "" {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, next, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	var issues []Issue
	err = checkResponse(resp) // the *APIError helper from chapter 5
	if err == nil {
		err = json.UnmarshalRead(resp.Body, &issues)
	}
	resp.Body.Close()
	if err != nil {
		return err
	}
	// ... use issues ...
	next = parseLinks(resp.Header)["next"] // "" when there's no next page
}
```

The loop ends when the map has no `next` key, because a missing key gives the zero
value, `""`. Notice the body is closed inside the loop, not with `defer`. A `defer` in
a loop only runs when the whole function returns, so every page's connection would stay
busy until the end.

## Relative links

The spec allows relative URLs, like `</issues?page=3>; rel="next"`. Resolve them
against the URL of the request that returned them:

```go
ref, err := url.Parse(next)
if err != nil {
	return err
}
next = resp.Request.URL.ResolveReference(ref).String()
```

`resp.Request` is the request that produced this response (after any redirects), so
it's the right base.

## Your turn

Trackr's older `/v2` endpoints page with Link headers. `parseLinks` is written for you.
Complete `listAll(ctx, client, firstURL)` so it returns every issue from every page:

1. Start at `firstURL`. For each page, build a request with `ctx`, send it with
   `client`, and close the body before moving on (a helper that fetches **one** page
   and uses `defer` keeps this tidy).
2. Return an error for any status other than `200 OK`, and for decode errors.
3. Append the page's issues, then follow the `rel="next"` link.
4. The fake API in `main` sends **relative** links such as `</issues?page=2>`, so
   resolve each one against `resp.Request.URL` (add the `net/url` import).
5. Stop when a page has no `next` link.
