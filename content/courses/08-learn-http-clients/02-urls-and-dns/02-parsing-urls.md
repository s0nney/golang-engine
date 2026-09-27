---
title: Parsing URLs
quiz:
  - question: |
      What does this print?

      ```go
      u, _ := url.Parse("https://api.trackr.dev/issues?label=bug&label=ui")
      q := u.Query()
      fmt.Println(q.Get("label"), len(q["label"]))
      ```
    options:
      - text: '`bug 2`'
        correct: true
      - text: '`ui 2`'
      - text: '`bug,ui 1`'
      - text: '`bug 1`'
    explanation: |
      `url.Values` is a `map[string][]string`. `Get` returns only the *first* value for a
      key, while indexing the map gives you all of them.
  - question: |
      What's in `u.Host` here?

      ```go
      u, err := url.Parse("localhost:8080/issues")
      ```
    options:
      - text: '`"localhost:8080"`'
      - text: '`"localhost"`'
      - text: 'Nothing: `err` is non-nil'
      - text: '`""`, because Go parsed `localhost` as the scheme'
        correct: true
    explanation: |
      Without `http://`, the text before the first colon looks like a scheme, so you get
      `Scheme: "localhost"` and an empty host, with no error. Always include the scheme,
      and check `u.Scheme` and `u.Host` when a URL comes from user input.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"net/url"
    )

    // parseIssueLink takes an issue link a user pasted, such as
    //
    //	https://trackr.dev/projects/apollo/issues/42?tab=comments
    //
    // and returns the project ("apollo") and issue number (42).
    // It returns an error if the link doesn't parse, isn't http or https,
    // has no host, or its path isn't /projects/<project>/issues/<positive id>.
    func parseIssueLink(link string) (project string, id int, err error) {
    	u, err := url.Parse(link)
    	if err != nil {
    		return "", 0, err
    	}
    	// ?
    	_ = u
    	return "", 0, nil
    }

    func main() {
    	for _, link := range []string{
    		"https://trackr.dev/projects/apollo/issues/42?tab=comments",
    		"http://localhost:8080/projects/web%20app/issues/7/",
    		"trackr.dev/projects/apollo/issues/42",
    		"https://trackr.dev/projects/apollo/issues/latest",
    	} {
    		project, id, err := parseIssueLink(link)
    		fmt.Printf("%q %d %v\n", project, id, err)
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"net/url"
    	"strconv"
    	"strings"
    )

    func parseIssueLink(link string) (project string, id int, err error) {
    	u, err := url.Parse(link)
    	if err != nil {
    		return "", 0, err
    	}
    	if u.Scheme != "http" && u.Scheme != "https" {
    		return "", 0, fmt.Errorf("issue link %q: scheme must be http or https", link)
    	}
    	if u.Host == "" {
    		return "", 0, fmt.Errorf("issue link %q: missing host", link)
    	}

    	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
    	if len(parts) != 4 || parts[0] != "projects" || parts[1] == "" || parts[2] != "issues" {
    		return "", 0, fmt.Errorf("issue link %q: want a path like /projects/<project>/issues/<id>", link)
    	}
    	id, err = strconv.Atoi(parts[3])
    	if err != nil || id <= 0 {
    		return "", 0, fmt.Errorf("issue link %q: bad issue number %q", link, parts[3])
    	}
    	return parts[1], id, nil
    }

    func main() {
    	for _, link := range []string{
    		"https://trackr.dev/projects/apollo/issues/42?tab=comments",
    		"http://localhost:8080/projects/web%20app/issues/7/",
    		"trackr.dev/projects/apollo/issues/42",
    		"https://trackr.dev/projects/apollo/issues/latest",
    	} {
    		project, id, err := parseIssueLink(link)
    		fmt.Printf("%q %d %v\n", project, id, err)
    	}
    }
  tests: |
    package main

    import "testing"

    func TestParseIssueLink(t *testing.T) {
    	tests := []struct {
    		link    string
    		project string
    		id      int
    	}{
    		{"https://trackr.dev/projects/apollo/issues/42", "apollo", 42},
    		{"https://trackr.dev/projects/apollo/issues/42?tab=comments#c7", "apollo", 42},
    		{"http://localhost:8080/projects/web%20app/issues/7/", "web app", 7},
    		{"https://TRACKR.dev/projects/r2-d2/issues/1001", "r2-d2", 1001},
    	}
    	for _, tt := range tests {
    		project, id, err := parseIssueLink(tt.link)
    		if err != nil {
    			t.Errorf("parseIssueLink(%q) returned error %v, want (%q, %d)", tt.link, err, tt.project, tt.id)
    			continue
    		}
    		if project != tt.project || id != tt.id {
    			t.Errorf("parseIssueLink(%q) = (%q, %d), want (%q, %d)", tt.link, project, id, tt.project, tt.id)
    		}
    	}
    }

    func TestParseIssueLinkErrors(t *testing.T) {
    	for _, link := range []string{
    		"http://[::1",                           // doesn't parse
    		"trackr.dev/projects/apollo/issues/42",  // no scheme, so no host
    		"ftp://trackr.dev/projects/a/issues/1",  // wrong scheme
    		"https:///projects/apollo/issues/1",     // no host
    		"https://trackr.dev/projects/apollo",    // too short
    		"https://trackr.dev/users/ana/issues/4", // not a project
    		"https://trackr.dev/projects/apollo/pulls/4",
    		"https://trackr.dev/projects/apollo/issues/latest",
    		"https://trackr.dev/projects/apollo/issues/-3",
    		"https://trackr.dev/projects/apollo/issues/0",
    		"https://trackr.dev/projects/apollo/issues/42/comments",
    		"https://trackr.dev/projects//issues/42",
    	} {
    		project, id, err := parseIssueLink(link)
    		if err == nil {
    			t.Errorf("parseIssueLink(%q) = (%q, %d, nil), want an error", link, project, id)
    		}
    	}
    }
---

`trackr` reads a base URL from its config file, and users paste issue links into it.
Pulling those apart with `strings.Split` is a bug waiting to happen. Use `net/url`.

## url.Parse

`url.Parse` turns a string into a `*url.URL` struct with one field per part:

```go
package main

import (
	"fmt"
	"net/url"
)

func main() {
	u, err := url.Parse("https://ana:s3cret@api.trackr.dev:8443/projects/apollo/issues?state=open&label=bug&label=ui#comments")
	if err != nil {
		fmt.Println("bad URL:", err)
		return
	}
	fmt.Println("scheme:  ", u.Scheme)
	fmt.Println("user:    ", u.User.Username())
	fmt.Println("host:    ", u.Host)
	fmt.Println("hostname:", u.Hostname())
	fmt.Println("port:    ", u.Port())
	fmt.Println("path:    ", u.Path)
	fmt.Println("rawquery:", u.RawQuery)
	fmt.Println("fragment:", u.Fragment)

	q := u.Query()
	fmt.Println("state:   ", q.Get("state"))
	fmt.Println("labels:  ", q["label"])
	fmt.Println("missing: ", q.Get("assignee") == "")
	fmt.Println("redacted:", u.Redacted())
}
```

Output:

```
scheme:   https
user:     ana
host:     api.trackr.dev:8443
hostname: api.trackr.dev
port:     8443
path:     /projects/apollo/issues
rawquery: state=open&label=bug&label=ui
fragment: comments
state:    open
labels:   [bug ui]
missing:  true
redacted: https://ana:xxxxx@api.trackr.dev:8443/projects/apollo/issues?state=open&label=bug&label=ui#comments
```

A few things to notice:

- `Host` includes the port. `Hostname()` and `Port()` split it for you, and they handle
  IPv6 addresses like `[::1]:8080` correctly, which a naive split on `:` would not.
- `Port()` is empty when the URL doesn't name one. It doesn't fill in the default 443.
- `Path` is **decoded**: `%20` becomes a space. The original encoded form is kept in
  `RawPath` when it matters.
- `Redacted()` hides the password. Handy for logs.

## Query values

`u.Query()` parses `RawQuery` into a `url.Values`, which is just a
`map[string][]string`, because a key can repeat.

```go
q := u.Query()
q.Get("state")  // first value, or "" if missing
q["label"]      // all values: []string{"bug", "ui"}
q.Has("sort")   // is the key there at all?
```

`Get` returning `""` for a missing key means you can't tell "missing" from "empty"
(`?assignee=`). Use `Has` when that difference matters.

`Query()` parses a fresh copy every time you call it. Changing that copy does **not**
change `u`. You'll see how to write changes back in the next lesson.

## What counts as an error?

`url.Parse` is surprisingly forgiving. It only fails on things that can't be a URL at all:

```go
url.Parse("http://localhost:1:2") // error: invalid port ":1:2" after host
url.Parse("api.trackr.dev/issues") // no error! Host "", Path "api.trackr.dev/issues"
url.Parse("localhost:8080/issues") // no error! Scheme "localhost", Host ""
```

The last two are relative references or odd schemes, and they're valid URLs, just not
what you meant. When a user types a base URL into `trackr`'s config, validate it:

```go
func parseBaseURL(s string) (*url.URL, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("base URL %q: scheme must be http or https", s)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("base URL %q: missing host", s)
	}
	return u, nil
}
```

## Turning it back into a string

`u.String()` reassembles the URL, re-escaping as needed. `fmt.Println(u)` calls it for
you. And that's exactly what `http.Get` and friends need.

## Your turn

People paste issue links into `trackr` all the time: `trackr show
https://trackr.dev/projects/apollo/issues/42?tab=comments`. Complete `parseIssueLink` so
it pulls out the project and the issue number:

1. The starter already calls `url.Parse` and returns its error.
2. Return an error unless the scheme is `http` or `https` and the host isn't empty.
3. The path must be exactly `/projects/<project>/issues/<id>`. A trailing slash is fine
   (`strings.Trim(u.Path, "/")` and `strings.Split` help). The query and fragment
   don't matter.
4. `<project>` must not be empty, and `<id>` must be a positive whole number
   (`strconv.Atoi`).
5. Return the project as it appears in `u.Path`, which is already decoded, so
   `web%20app` comes back as `web app`.
