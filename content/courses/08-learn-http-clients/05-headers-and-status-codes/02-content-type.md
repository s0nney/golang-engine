---
title: Content-Type and Accept
quiz:
  - question: |
      The server sends `Content-Type: application/json; charset=utf-8`. What's wrong with this check?

      ```go
      if resp.Header.Get("Content-Type") != "application/json" {
          return errors.New("not JSON")
      }
      ```
    options:
      - text: Nothing, it works
      - text: It rejects a valid JSON response, because the header has a `charset` parameter
        correct: true
      - text: '`Get` is case-sensitive, so it never finds the header'
    explanation: |
      Media types can carry parameters after a `;`. Parse the header with
      `mime.ParseMediaType` and compare just the media type.
  - question: Which header does a **client** send to say which response formats it can handle?
    options:
      - text: '`Content-Type`'
      - text: '`Accept`'
        correct: true
      - text: '`Accept-Content`'
    explanation: |
      `Accept` is the client's wish list. `Content-Type` describes a body that is
      actually being sent, in either direction.
---

Two headers work as a pair: one says what a body **is**, the other says what you'd
**like**.

## Content-Type: what this body is

`Content-Type` holds a **media type** (also called a MIME type): a `type/subtype`,
optionally followed by parameters.

```
Content-Type: application/json
Content-Type: application/json; charset=utf-8
Content-Type: text/html; charset=utf-8
Content-Type: application/x-www-form-urlencoded
Content-Type: multipart/form-data; boundary=----abc123
```

It travels in both directions:

- **On a request** you set it whenever you send a body (you did in chapter 4).
- **On a response** the server tells you what it sent. Checking it before decoding
  turns a baffling `invalid character '<'` into a clear "expected JSON, got text/html".

### Parse it, don't compare it

Because of parameters like `charset`, comparing the whole string is fragile. The `mime`
package splits it properly:

```go
mt, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
fmt.Println(mt, params, err)
// application/json map[charset:utf-8] <nil>
```

Media types are case-insensitive, and `ParseMediaType` lowercases the type for you.
A helper for `trackr`:

```go
func isJSON(h http.Header) bool {
	mt, _, err := mime.ParseMediaType(h.Get("Content-Type"))
	return err == nil && (mt == "application/json" || strings.HasSuffix(mt, "+json"))
}
```

The `+json` suffix covers types like `application/problem+json`, a standard format for
error bodies that many APIs use.

## Accept: what I'd like back

`Accept` lists the media types a client understands. The server picks one (this is
called **content negotiation**):

```go
req.Header.Set("Accept", "application/json")
```

Most JSON APIs only speak JSON and ignore it, but some serve HTML to browsers and
JSON to clients based on `Accept`, and some version their API through it, like
`Accept: application/vnd.trackr.v2+json`. Sending it costs nothing and makes your
intent explicit.

If a server can't produce anything you accept, it answers `406 Not Acceptable`.

## Other formats you'll bump into

| Media type | What it is |
| --- | --- |
| `application/json` | JSON |
| `application/problem+json` | a standard JSON error body (RFC 9457) |
| `text/plain` | plain text |
| `text/html` | a web page, often an error page from a proxy |
| `application/x-www-form-urlencoded` | `key=value&...`, like a URL query, from HTML forms |
| `multipart/form-data` | file uploads |
| `application/octet-stream` | "just bytes", such as a download |

For form bodies, `url.Values` from chapter 2 does the encoding:

```go
form := url.Values{"title": {"Login broken"}, "label": {"bug"}}
req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
if err != nil {
	return err
}
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
```

(Trackr's API only takes JSON, but OAuth token endpoints, for example, expect exactly
this.)

## Sniffing

`http.DetectContentType(data)` guesses a media type from the first 512 bytes. That's
useful for uploads when you don't know what a file is, but don't use it to second-guess
an API. Trust the header, and treat a mismatch as an error.
