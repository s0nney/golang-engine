---
title: Page Links
difficulty: easy
after: urls-and-dns
hints:
  - 'Parse the string with `url.Parse`, and get the query as a `url.Values` map with `u.Query()`. Editing the map doesn''t change `u` until you write it back.'
  - '`q.Set("page", ...)` replaces every existing value, and `q.Del("page")` removes it. Then store `q.Encode()` in `u.RawQuery` and return `u.String()`.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"net/url"
    )

    // withPage returns rawURL with its "page" query parameter set to page.
    //
    //   - Any existing page value (even several) is replaced.
    //   - page <= 1 means the first page, which is the default, so the page
    //     parameter is removed instead.
    //   - Every other parameter, the path and the fragment are kept. The query
    //     is re-encoded with url.Values.Encode, so keys come out sorted.
    //   - If rawURL doesn't parse, return "" and the error.
    func withPage(rawURL string, page int) (string, error) {
    	// 1. Parse rawURL with url.Parse.
    	// 2. Get the query with u.Query(), then Set or Del "page".
    	// 3. Put it back with u.RawQuery = q.Encode() and return u.String().
    	_ = url.Parse
    	return "", nil
    }

    func main() {
    	fmt.Println(withPage("https://api.trackr.dev/v1/issues?state=open", 3))
    	// want: https://api.trackr.dev/v1/issues?page=3&state=open <nil>
    	fmt.Println(withPage("https://api.trackr.dev/v1/issues?page=3&state=open", 1))
    	// want: https://api.trackr.dev/v1/issues?state=open <nil>
    }
  solution: |
    package main

    import (
    	"fmt"
    	"net/url"
    	"strconv"
    )

    func withPage(rawURL string, page int) (string, error) {
    	u, err := url.Parse(rawURL)
    	if err != nil {
    		return "", err
    	}
    	q := u.Query()
    	if page <= 1 {
    		q.Del("page")
    	} else {
    		q.Set("page", strconv.Itoa(page))
    	}
    	u.RawQuery = q.Encode()
    	return u.String(), nil
    }

    func main() {
    	fmt.Println(withPage("https://api.trackr.dev/v1/issues?state=open", 3))
    	fmt.Println(withPage("https://api.trackr.dev/v1/issues?page=3&state=open", 1))
    }
  tests: |
    package main

    import (
    	"net/http"
    	"net/http/httptest"
    	"testing"
    )

    func TestWithPage(t *testing.T) {
    	tests := []struct {
    		raw  string
    		page int
    		want string
    	}{
    		{"https://api.trackr.dev/v1/issues", 2, "https://api.trackr.dev/v1/issues?page=2"},
    		{"https://api.trackr.dev/v1/issues?state=open", 3, "https://api.trackr.dev/v1/issues?page=3&state=open"},
    		{"https://api.trackr.dev/v1/issues?page=3&state=open", 4, "https://api.trackr.dev/v1/issues?page=4&state=open"},
    		{"https://api.trackr.dev/v1/issues?page=1&page=9", 5, "https://api.trackr.dev/v1/issues?page=5"},
    		{"https://api.trackr.dev/v1/issues?page=3&state=open", 1, "https://api.trackr.dev/v1/issues?state=open"},
    		{"https://api.trackr.dev/v1/issues?page=3", 0, "https://api.trackr.dev/v1/issues"},
    		{"https://api.trackr.dev/v1/issues?page=3", -2, "https://api.trackr.dev/v1/issues"},
    		{"https://api.trackr.dev/v1/issues?label=bug&label=ui", 2, "https://api.trackr.dev/v1/issues?label=bug&label=ui&page=2"},
    		{"https://api.trackr.dev/v1/search?q=caf%C3%A9+cr%C3%A8me", 2, "https://api.trackr.dev/v1/search?page=2&q=caf%C3%A9+cr%C3%A8me"},
    		{"https://api.trackr.dev/v1/search?q=a%26b", 12, "https://api.trackr.dev/v1/search?page=12&q=a%26b"},
    		{"https://trackr.dev/projects/web%20app/issues#top", 2, "https://trackr.dev/projects/web%20app/issues?page=2#top"},
    		{"http://127.0.0.1:8080/issues?", 2, "http://127.0.0.1:8080/issues?page=2"},
    	}
    	for _, tt := range tests {
    		got, err := withPage(tt.raw, tt.page)
    		if err != nil || got != tt.want {
    			t.Errorf("withPage(%q, %d) = %q, %v\n\twant %q, <nil>", tt.raw, tt.page, got, err, tt.want)
    		}
    	}
    }

    func TestWithPageBadURL(t *testing.T) {
    	for _, raw := range []string{"http://[::1", "https://api.trackr.dev/%zz", ":no-scheme"} {
    		got, err := withPage(raw, 2)
    		if err == nil || got != "" {
    			t.Errorf("withPage(%q, 2) = %q, %v, want \"\" and an error", raw, got, err)
    		}
    	}
    }

    func TestWithPageReachesServer(t *testing.T) {
    	var gotPage, gotQ string
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		gotPage, gotQ = r.URL.Query().Get("page"), r.URL.Query().Get("q")
    	}))
    	defer srv.Close()

    	link, err := withPage(srv.URL+"/search?q=%E6%97%A5%E6%9C%AC+%26+more", 7)
    	if err != nil {
    		t.Fatalf("withPage: %v", err)
    	}
    	if link == "" {
    		t.Fatal("withPage returned an empty URL")
    	}
    	resp, err := srv.Client().Get(link)
    	if err != nil {
    		t.Fatalf("GET %s: %v", link, err)
    	}
    	resp.Body.Close()
    	if gotPage != "7" || gotQ != "日本 & more" {
    		t.Errorf("server received page=%q q=%q, want page=%q q=%q", gotPage, gotQ, "7", "日本 & more")
    	}
    }
---

`trackr list` shows 50 issues at a time and prints a "next page" link at the
bottom. The link is the current URL with a different `page` parameter, and every
other filter the user typed has to survive the trip.

Complete `withPage(rawURL, page)`:

- Set the `page` query parameter to `page`, replacing any value (or values) it
  already had.
- If `page <= 1`, **remove** `page` instead: page 1 is the default, so the link
  stays clean.
- Keep the scheme, host, path, fragment and every other query parameter,
  including repeated ones like `label=bug&label=ui`.
- Re-encode the query with `url.Values.Encode`, so keys come out sorted.
- If `rawURL` doesn't parse, return `""` and the error.

## Examples

```
withPage("https://api.trackr.dev/v1/issues?state=open", 3)
// "https://api.trackr.dev/v1/issues?page=3&state=open"

withPage("https://api.trackr.dev/v1/issues?page=3&state=open", 1)
// "https://api.trackr.dev/v1/issues?state=open"

withPage("https://api.trackr.dev/v1/search?q=caf%C3%A9", 2)
// "https://api.trackr.dev/v1/search?page=2&q=caf%C3%A9"
```

## Constraints

- Don't build the query with string concatenation. Search terms can hold `&`,
  `=` and non-ASCII letters, and they must reach the server intact. One test sends
  your link to a fake API and checks what it received.
- `add` vs `set` matters: a URL that already has `page=3` must not end up with
  two `page` values.
