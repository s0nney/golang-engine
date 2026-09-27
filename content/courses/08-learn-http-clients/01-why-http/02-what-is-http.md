---
title: What Is HTTP?
quiz:
  - question: |
      Which part of this request tells the server **which resource** you want?

      ```
      GET /projects/apollo/issues?state=open HTTP/1.1
      Host: api.trackr.dev
      Accept: application/json
      ```
    options:
      - text: '`GET`'
      - text: '`/projects/apollo/issues?state=open`'
        correct: true
      - text: '`HTTP/1.1`'
      - text: '`Accept: application/json`'
    explanation: |
      The first line is *method*, *target*, *version*. The target (path plus query string)
      names the resource. `GET` says what to do with it, and `Accept` is a header saying
      which format you'd like back.
  - question: How does an HTTP/1.1 message mark the end of its headers?
    options:
      - text: A line containing only `END`
      - text: A closing brace `}`
      - text: An empty line
        correct: true
      - text: The connection closes
    explanation: |
      Headers are one per line, and a blank line (`\r\n` on its own) separates them from
      the body. Everything after that blank line is the body.
---

**HTTP** (HyperText Transfer Protocol) is the set of rules clients and servers use to
talk on the web. It was built for web pages, but today it carries nearly every API
call too, including the ones `trackr` makes.

## A request is just text

HTTP/1.1 is a plain-text protocol. You can watch the bytes Go would send by building a
request and dumping it with `net/http/httputil`. Nothing touches the network here:

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

func main() {
	req, err := http.NewRequest("GET", "https://api.trackr.dev/projects/apollo/issues?state=open", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept", "application/json")

	dump, err := httputil.DumpRequestOut(req, false)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s", dump)
}
```

It prints:

```
GET /projects/apollo/issues?state=open HTTP/1.1
Host: api.trackr.dev
User-Agent: Go-http-client/1.1
Accept: application/json
Accept-Encoding: gzip

```

Every request has the same shape:

1. **Request line**: the **method** (`GET`), the **target** (path and query) and the
   protocol version.
2. **Headers**: `Name: value` pairs with extra information. `Host` says which site you
   want, since one server can host many. Go adds `User-Agent` and `Accept-Encoding` for you.
3. **A blank line**, which ends the headers.
4. An optional **body**. A `GET` usually has none. A `POST` that creates an issue
   carries the new issue's JSON here.

## A response has the same shape

```
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 58

[{"id":1,"title":"Login button is blue"},{"id":2,"title":"…"}]
```

1. **Status line**: the version, a **status code** (`200`) and a reason phrase (`OK`).
2. **Headers**, such as `Content-Type`, which says what the body is.
3. **A blank line**.
4. **The body**: here, a JSON array of issues.

You'll spend whole chapters on methods, headers, status codes and JSON bodies. For now,
remember the shape: *start line, headers, blank line, body*.

## What about HTTP/2 and HTTP/3?

Newer versions send the same information in a compact binary format and can run many
requests over one connection at once (HTTP/3 even swaps TCP for QUIC). The *meaning*
doesn't change: there's still a method, a path, headers, a status and a body. Go's
`net/http` negotiates HTTP/2 over HTTPS automatically, and your code looks exactly the
same either way.

## HTTP sits on top of other layers

HTTP doesn't move bytes by itself. It rides on lower layers:

```
HTTP     "GET /projects/apollo/issues"
TLS      encryption (for https:// URLs)
TCP      reliable, ordered stream of bytes
IP       delivers packets between machines
```

You'll meet TLS properly in the HTTPS chapter. The nice part is that `net/http` handles
every layer below HTTP for you.

## Further reading

- MDN, "An overview of HTTP": https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/Overview
