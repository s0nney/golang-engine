---
title: Copying URLs with Clone
quiz:
  - question: |
      What does this print?

      ```go
      base, _ := url.Parse("https://api.trackr.dev/v1")
      u := base
      u.Path = "/v1/issues"
      fmt.Println(base)
      ```
    options:
      - text: '`https://api.trackr.dev/v1`'
      - text: '`https://api.trackr.dev/v1/issues`'
        correct: true
      - text: It doesn't compile, because `u` is a pointer
    explanation: |
      `url.Parse` returns a `*url.URL`, so `u := base` copies the *pointer*. Both
      variables point at the same struct, and changing `u.Path` changes `base` too.
      Use `base.Clone()` to get an independent copy.
  - question: Which of these gives you a URL you can change without touching `base`?
    options:
      - text: '`u := base`'
      - text: '`u := &base`'
      - text: '`u := base.Clone()`'
        correct: true
      - text: '`u := base.Query()`'
    explanation: |
      `Clone` (new in Go 1.27) allocates a fresh `url.URL` and deep-copies its fields,
      including the `User` info. The first two options share the original, and `Query`
      returns the query parameters, not a URL.
---

`trackr` parses its base URL **once**, at startup, and keeps it around as a
`*url.URL`. Every request then starts from that base and adds a path and a query. That
design has a trap in it.

## The shared-pointer bug

`url.Parse` returns a pointer. Copying a pointer doesn't copy the URL:

```go
type Client struct {
	base *url.URL
}

// issueURL is BUGGY: it edits the client's base URL in place.
func (c *Client) issueURL(id int) string {
	u := c.base // same struct as c.base!
	u.Path = u.Path + "/issues/" + strconv.Itoa(id)
	return u.String()
}
```

Call it twice and the paths pile up:

```
https://api.trackr.dev/v1/issues/1
https://api.trackr.dev/v1/issues/1/issues/2
```

Worse, if two goroutines build URLs at once, they race on the same struct. Run that
under `go test -race` and the race detector will complain loudly.

## Clone makes a real copy

Go 1.27 added `(*url.URL).Clone`, which returns a brand new `*url.URL` with every field
copied, and `url.Values.Clone` for query maps:

```go
package main

import (
	"fmt"
	"net/url"
)

func main() {
	base, _ := url.Parse("https://api.trackr.dev/v1?tenant=acme")

	u := base.Clone()
	u.Path = "/v1/projects"
	q := u.Query()
	q.Set("state", "open")
	u.RawQuery = q.Encode()

	fmt.Println("base: ", base)
	fmt.Println("clone:", u)

	labels := url.Values{"label": {"bug"}}
	more := labels.Clone()
	more.Add("label", "ui")
	fmt.Println(labels, more)
}
```

Output:

```
base:  https://api.trackr.dev/v1?tenant=acme
clone: https://api.trackr.dev/v1/projects?state=open&tenant=acme
map[label:[bug]] map[label:[bug ui]]
```

The base is untouched. `Clone` on a nil `*url.URL` returns nil rather than panicking.

## Isn't `*base` a copy?

Before Go 1.27 people wrote `u := *base`, which copies the struct. It mostly works, but
it's a *shallow* copy: the `User` field is a pointer and stays shared. `Clone` also
copies the `User` info, and it says what you mean. Reach for `Clone`.

`url.Values` is a map, so plain assignment shares it too:

```go
alias := labels
alias.Add("label", "p0") // labels now has p0 as well
```

## What about JoinPath?

`(*url.URL).JoinPath` already returns a new URL and leaves the original alone. So this
is safe without `Clone`:

```go
u := c.base.JoinPath("issues", strconv.Itoa(id))
```

But `JoinPath` copies the query along with everything else, and if you then *change*
`u` in other ways, you want to be sure it's yours. The rule of thumb for `trackr`:

> Treat the stored base URL as read-only. Start every request with `Clone` or
> `JoinPath`, and only change the copy.

## Fixed

```go
func (c *Client) issueURL(id int) string {
	u := c.base.Clone()
	u.Path = u.Path + "/issues/" + strconv.Itoa(id)
	return u.String()
}
```

Now each call starts fresh, and concurrent callers never touch shared state.
