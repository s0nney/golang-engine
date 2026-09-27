---
title: Building URLs
quiz:
  - question: |
      What does this print?

      ```go
      u, _ := url.Parse("https://api.trackr.dev/issues")
      u.Query().Set("state", "open")
      fmt.Println(u)
      ```
    options:
      - text: '`https://api.trackr.dev/issues?state=open`'
      - text: '`https://api.trackr.dev/issues`'
        correct: true
      - text: '`https://api.trackr.dev/issues?state=`'
      - text: It doesn't compile
    explanation: |
      `Query()` returns a freshly parsed *copy*. Setting a key on that copy changes
      nothing in `u`. Keep the copy in a variable, change it, then write it back with
      `u.RawQuery = q.Encode()`.
  - question: |
      What does `url.JoinPath("https://api.trackr.dev/v1/", "/projects", "apollo")` return?
    options:
      - text: '`https://api.trackr.dev/v1//projects/apollo`'
      - text: '`https://api.trackr.dev/projects/apollo`'
      - text: '`https://api.trackr.dev/v1/projects/apollo`'
        correct: true
    explanation: |
      `JoinPath` adds the elements to the *existing* path and cleans the result, so
      duplicate slashes collapse to one. The base path `/v1` is kept.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"net/url"
    )

    // Filter narrows down an issue listing.
    type Filter struct {
    	State  string   // "open", "closed" or "" for any
    	Labels []string // match any of these labels
    	Page   int      // 0 means "don't send a page"
    }

    // issuesURL returns the URL for listing a project's issues:
    //
    //	<baseURL>/projects/<project>/issues?label=...&page=...&state=...
    //
    // The project name must be path-escaped, any path or query already in
    // baseURL must be kept, and empty filter fields must be left out.
    func issuesURL(baseURL, project string, f Filter) (string, error) {
    	u, err := url.Parse(baseURL)
    	if err != nil {
    		return "", err
    	}
    	// ?
    	return u.String(), nil
    }

    func main() {
    	s, err := issuesURL("https://api.trackr.dev/v1", "web app", Filter{
    		State:  "open",
    		Labels: []string{"bug", "good first issue"},
    		Page:   2,
    	})
    	fmt.Println(s, err)
    	// want: https://api.trackr.dev/v1/projects/web%20app/issues?label=bug&label=good+first+issue&page=2&state=open
    }
  solution: |
    package main

    import (
    	"fmt"
    	"net/url"
    	"strconv"
    )

    type Filter struct {
    	State  string
    	Labels []string
    	Page   int
    }

    func issuesURL(baseURL, project string, f Filter) (string, error) {
    	u, err := url.Parse(baseURL)
    	if err != nil {
    		return "", err
    	}
    	u = u.JoinPath("projects", url.PathEscape(project), "issues")

    	q := u.Query()
    	if f.State != "" {
    		q.Set("state", f.State)
    	}
    	for _, l := range f.Labels {
    		q.Add("label", l)
    	}
    	if f.Page > 0 {
    		q.Set("page", strconv.Itoa(f.Page))
    	}
    	u.RawQuery = q.Encode()
    	return u.String(), nil
    }

    func main() {
    	s, err := issuesURL("https://api.trackr.dev/v1", "web app", Filter{
    		State:  "open",
    		Labels: []string{"bug", "good first issue"},
    		Page:   2,
    	})
    	fmt.Println(s, err)
    }
  tests: |
    package main

    import (
    	"net/http"
    	"net/http/httptest"
    	"slices"
    	"testing"
    )

    func TestIssuesURL(t *testing.T) {
    	tests := []struct {
    		base, project string
    		f             Filter
    		want          string
    	}{
    		{"https://api.trackr.dev", "apollo", Filter{},
    			"https://api.trackr.dev/projects/apollo/issues"},
    		{"https://api.trackr.dev/v1/", "apollo", Filter{State: "closed"},
    			"https://api.trackr.dev/v1/projects/apollo/issues?state=closed"},
    		{"https://api.trackr.dev/v1", "web app", Filter{State: "open", Labels: []string{"bug", "good first issue"}, Page: 2},
    			"https://api.trackr.dev/v1/projects/web%20app/issues?label=bug&label=good+first+issue&page=2&state=open"},
    		{"https://api.trackr.dev", "r&d/tools", Filter{Labels: []string{"a&b"}},
    			"https://api.trackr.dev/projects/r&d%2Ftools/issues?label=a%26b"},
    		{"https://api.trackr.dev/v1?tenant=acme", "apollo", Filter{Page: 3},
    			"https://api.trackr.dev/v1/projects/apollo/issues?page=3&tenant=acme"},
    	}
    	for _, tt := range tests {
    		got, err := issuesURL(tt.base, tt.project, tt.f)
    		if err != nil {
    			t.Errorf("issuesURL(%q, %q, %+v) returned error %v", tt.base, tt.project, tt.f, err)
    			continue
    		}
    		if got != tt.want {
    			t.Errorf("issuesURL(%q, %q, %+v)\n got  %s\n want %s", tt.base, tt.project, tt.f, got, tt.want)
    		}
    	}
    }

    func TestIssuesURLBadBase(t *testing.T) {
    	if got, err := issuesURL("http://[::1", "apollo", Filter{}); err == nil {
    		t.Errorf("issuesURL(%q, ...) = %q, nil; want a parse error", "http://[::1", got)
    	}
    }

    func TestIssuesURLAgainstServer(t *testing.T) {
    	var gotPath string
    	var gotLabels []string
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		gotPath = r.URL.EscapedPath()
    		gotLabels = r.URL.Query()["label"]
    		w.Write([]byte("[]"))
    	}))
    	defer srv.Close()

    	u, err := issuesURL(srv.URL+"/v2", "mobile/ios", Filter{Labels: []string{"crash", "p0"}})
    	if err != nil {
    		t.Fatalf("issuesURL returned error %v", err)
    	}
    	resp, err := http.Get(u)
    	if err != nil {
    		t.Fatalf("GET %s: %v", u, err)
    	}
    	resp.Body.Close()
    	if want := "/v2/projects/mobile%2Fios/issues"; gotPath != want {
    		t.Errorf("server saw path %q, want %q", gotPath, want)
    	}
    	if want := []string{"crash", "p0"}; !slices.Equal(gotLabels, want) {
    		t.Errorf("server saw labels %q, want %q", gotLabels, want)
    	}
    }
---

Parsing is half the job. `trackr` also has to *build* URLs from a base URL in its
config, a project name the user typed and a bunch of filter flags. Gluing strings
together looks easy:

```go
u := base + "/projects/" + project + "/issues?state=" + state // don't
```

It breaks as soon as the base ends in `/`, the project is called `web app`, or a label
contains `&`. `net/url` has a tool for each piece.

## Paths: JoinPath

`url.JoinPath` (a function taking a string) and `(*url.URL).JoinPath` (a method) add
segments to the existing path and clean up stray slashes:

```go
s, err := url.JoinPath("https://api.trackr.dev/v1/", "projects", "apollo", "issues")
// s == "https://api.trackr.dev/v1/projects/apollo/issues"

base, _ := url.Parse("https://api.trackr.dev/v1")
u := base.JoinPath("projects", "apollo") // a new *url.URL; base is unchanged
```

One catch: elements are treated as **already escaped** path text. A `/` inside a
project name would become a separator, and `JoinPath` *resolves* `.` and `..`
segments, so a project called `../admin` would climb right out of `/projects`:

```go
base.JoinPath("projects", "../admin")
// https://api.trackr.dev/v1/admin   (oops)
```

Escape user-supplied segments yourself with `url.PathEscape`, which turns `/` into
`%2F` so the name stays one segment:

```go
base.JoinPath("projects", url.PathEscape("r&d/tools"), "issues")
// https://api.trackr.dev/v1/projects/r&d%2Ftools/issues
```

(A project literally named `..` would still be resolved, so a careful client rejects
names like `.` and `..` outright.)

## Queries: url.Values

`url.Values` (a `map[string][]string`) builds query strings. `Set` replaces a key's
values, `Add` appends one more, `Del` removes the key, and `Encode` escapes everything
and joins it with `&`:

```go
q := url.Values{}
q.Set("state", "open")
q.Add("label", "bug")
q.Add("label", "good first issue")
q.Set("assignee", "ana@trackr.dev")
fmt.Println(q.Encode())
// assignee=ana%40trackr.dev&label=bug&label=good+first+issue&state=open
```

`Encode` **sorts by key**. Servers don't care about order, and sorted output makes
URLs stable and easy to test.

## Putting it together

To add parameters to a URL that might already have some, read the query, change it,
and write it back:

```go
u := base.JoinPath("projects", url.PathEscape(project), "issues")
q := u.Query()      // a copy of any existing parameters
q.Set("state", "open")
u.RawQuery = q.Encode() // write the copy back
resp, err := http.Get(u.String())
```

## Gotcha: Query returns a copy

`u.Query()` parses `u.RawQuery` every time and hands you a *new* map. Calling
`u.Query().Set(...)` in one line does nothing useful, because the modified map is thrown
away straight after. Always go through a variable and assign `RawQuery`.

## Which escape?

| Function | Use for | Space becomes |
| --- | --- | --- |
| `url.PathEscape` | one path segment | `%20` |
| `url.QueryEscape` | one query key or value | `+` |
| `Values.Encode` | a whole query string | `+` |

## Your turn

Complete `issuesURL` so it returns the URL for listing a project's issues:

1. Add the path segments `projects`, the **path-escaped** project name, and `issues`
   to whatever path `baseURL` already has (`/v1`, `/v1/` or nothing).
2. Keep any query parameters already in `baseURL`.
3. Add `state` only if `f.State` isn't empty, one `label` for each entry in `f.Labels`,
   and `page` only if `f.Page > 0` (use `strconv.Itoa`).
4. Return the finished URL as a string. The query must be built with `Values.Encode`,
   so keys come out sorted.

If `baseURL` doesn't parse, return the error. The starter already does that part.
