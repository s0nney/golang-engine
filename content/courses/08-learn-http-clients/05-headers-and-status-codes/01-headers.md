---
title: Headers
quiz:
  - question: |
      What does this print?

      ```go
      h := http.Header{}
      h.Set("x-trackr-client", "cli/1.4")
      fmt.Println(h.Get("X-TRACKR-CLIENT"), h["x-trackr-client"] == nil)
      ```
    options:
      - text: '`cli/1.4 false`'
      - text: '`cli/1.4 true`'
        correct: true
      - text: '` true` (an empty string, then true)'
    explanation: |
      `Set` and `Get` canonicalize the key to `X-Trackr-Client`, so `Get` with any
      capitalization finds it. Indexing the map directly skips canonicalization, and
      the lowercase key isn't there.
  - question: 'A response has two `Set-Cookie` headers. How do you get both?'
    options:
      - text: '`resp.Header.Get("Set-Cookie")`'
      - text: '`resp.Header.Values("Set-Cookie")`'
        correct: true
      - text: '`strings.Split(resp.Header.Get("Set-Cookie"), ",")`'
    explanation: |
      `http.Header` is a `map[string][]string` because a header can repeat. `Get` returns
      only the first value, and `Values` returns them all.
---

Every request and response carries **headers**: metadata lines before the body, like
`Content-Type: application/json`. They say what the body is, who's asking, whether
you're allowed in, how many requests you have left, and much more.

```
GET /issues/42 HTTP/1.1
Host: api.trackr.dev
User-Agent: trackr/1.4
Accept: application/json
Authorization: Bearer tk_live_...
```

## http.Header

In Go, both `req.Header` and `resp.Header` are an `http.Header`, which is just:

```go
type Header map[string][]string
```

A slice, because a header name can appear more than once. Use the methods rather than
the raw map:

```go
h := http.Header{}
h.Set("x-trackr-client", "cli/1.4")    // replace any existing values
h.Add("accept", "application/json")     // append one more value
h.Add("Accept", "text/plain")
fmt.Println(h)
// map[Accept:[application/json text/plain] X-Trackr-Client:[cli/1.4]]

fmt.Println(h.Get("X-TRACKR-CLIENT"))  // cli/1.4 (first value, or "")
fmt.Println(h.Values("ACCEPT"))        // [application/json text/plain]
h.Del("Accept")                        // remove every value
```

## Case doesn't matter (if you use the methods)

Header names are case-insensitive in HTTP. Go handles that by **canonicalizing** every
key the methods touch: first letter and every letter after a hyphen uppercase, the rest
lowercase. `x-ratelimit-remaining` becomes `X-Ratelimit-Remaining`
(`http.CanonicalHeaderKey` shows you the result).

The gotcha is the raw map. It's an ordinary Go map with no magic:

```go
fmt.Println(h["x-trackr-client"]) // [] : wrong case, not found
h["x-raw"] = []string{"1"}
fmt.Println(h.Get("X-Raw"))       // "" : Get looks for "X-Raw", map has "x-raw"
```

Stick to `Set`, `Add`, `Get`, `Values` and `Del`.

## Setting request headers

Change `req.Header` after building the request and before `Do`:

```go
req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
if err != nil {
	return err
}
req.Header.Set("Accept", "application/json")
req.Header.Set("User-Agent", "trackr/1.4")
```

Some headers are filled in for you if you don't set them:

| Header | Default from Go |
| --- | --- |
| `User-Agent` | `Go-http-client/1.1` |
| `Host` | taken from the URL |
| `Content-Length` | from a `bytes`/`strings` body |
| `Accept-Encoding` | `gzip`, and Go transparently unzips the response |

A descriptive `User-Agent` like `trackr/1.4` is polite: it helps API operators see
who's calling and contact you if your client misbehaves.

## Reading response headers

```go
fmt.Println(resp.Header.Get("Content-Type"))          // application/json; charset=utf-8
fmt.Println(resp.Header.Get("X-Ratelimit-Remaining")) // 4999
fmt.Println(resp.Header.Values("Set-Cookie"))         // [a=1 b=2]
```

Headers you'll meet in this course:

| Header | Direction | Meaning |
| --- | --- | --- |
| `Content-Type` | both | the body's format |
| `Accept` | request | formats the client wants back |
| `Authorization` | request | credentials |
| `Location` | response | URL of a created or moved resource |
| `Retry-After` | response | how long to wait before trying again |
| `Link` | response | next and previous pages |
| `X-Ratelimit-*` | response | rate-limit budget (not standard, but common) |

Headers starting with `X-` are conventionally custom ones. Each API documents its own.

## Further reading

- MDN, "HTTP headers": https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers
