---
title: Your First GET
quiz:
  - question: |
      The server answers with `404 Not Found`. What does `http.Get` return?

      ```go
      resp, err := http.Get(url)
      ```
    options:
      - text: A nil `resp` and an error mentioning 404
      - text: A `resp` with `StatusCode == 404` and a nil `err`
        correct: true
      - text: It panics
      - text: A nil `resp` and a nil `err`
    explanation: |
      To `net/http`, a 404 is a perfectly good response: the server answered. `err` is
      only non-nil when there's *no* usable response (bad URL, DNS failure, refused
      connection, timeout). Checking `StatusCode` is your job.
  - question: What type is `resp.Body`?
    options:
      - text: '`string`'
      - text: '`[]byte`'
      - text: '`io.ReadCloser`'
        correct: true
      - text: '`*bytes.Buffer`'
    explanation: |
      The body is a *stream*. It's read from the network as you consume it, which is why
      it's an `io.ReadCloser`: you `Read` it (for example with `io.ReadAll`) and then `Close` it.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    )

    func main() {
    	// The fake Trackr API. Don't change it.
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		if r.URL.Path != "/version" {
    			http.NotFound(w, r)
    			return
    		}
    		w.Header().Set("X-Trackr-Version", "2.4.1")
    		fmt.Fprint(w, "Trackr API 2.4.1 (build 7731)")
    	}))
    	defer srv.Close()

    	// ?: GET srv.URL + "/version", then print three lines:
    	//   1. resp.Status
    	//   2. the X-Trackr-Version response header
    	//   3. the body
    	fmt.Println("TODO: fetch", srv.URL+"/version")
    }
  solution: |
    package main

    import (
    	"fmt"
    	"io"
    	"net/http"
    	"net/http/httptest"
    )

    func main() {
    	// The fake Trackr API. Don't change it.
    	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		if r.URL.Path != "/version" {
    			http.NotFound(w, r)
    			return
    		}
    		w.Header().Set("X-Trackr-Version", "2.4.1")
    		fmt.Fprint(w, "Trackr API 2.4.1 (build 7731)")
    	}))
    	defer srv.Close()

    	resp, err := http.Get(srv.URL + "/version")
    	if err != nil {
    		fmt.Println("request failed:", err)
    		return
    	}
    	defer resp.Body.Close()

    	body, err := io.ReadAll(resp.Body)
    	if err != nil {
    		fmt.Println("reading body failed:", err)
    		return
    	}

    	fmt.Println(resp.Status)
    	fmt.Println(resp.Header.Get("X-Trackr-Version"))
    	fmt.Println(string(body))
    }
  expected_output: |
    200 OK
    2.4.1
    Trackr API 2.4.1 (build 7731)
---

Time to make a real request. Go's `net/http` package has everything you need, and
the simplest entry point is `http.Get`.

## A stand-in for the Trackr API

The exercise runner can't reach the internet, so in this course the "Trackr API" is
usually a small server started inside the same program with `net/http/httptest`.
`httptest.NewServer` listens on a random local port and gives you its base URL in
`srv.URL` (something like `http://127.0.0.1:41823`). Don't worry about the server code
yet. Chapter 9 is all about it. For now, treat it as "the API".

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

func main() {
	// A fake Trackr API that answers every request with a greeting.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome to Trackr!")
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/hello")
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("reading body failed:", err)
		return
	}

	fmt.Println(resp.StatusCode, resp.Status)
	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(string(body))
}
```

Output:

```
200 200 OK
text/plain; charset=utf-8
Welcome to Trackr!
```

Against the real API, the only change would be the URL:

```go
resp, err := http.Get("https://api.trackr.dev/hello") // needs internet access
```

## Step by step

1. `http.Get(url)` sends a `GET` request and waits for the response **headers**. It
   returns a `*http.Response` and an `error`.
2. If `err != nil`, there's no response to look at. The request never got an answer.
3. `resp.StatusCode` is the number (`200`), and `resp.Status` is the full text (`"200 OK"`).
4. `resp.Header` holds the response headers. `Get` looks one up.
5. `resp.Body` is an `io.ReadCloser`. `io.ReadAll` reads it to the end and gives you a `[]byte`.
6. `defer resp.Body.Close()` releases the connection. More on that in the next lesson.

## Gotcha: a 404 is not an error

This trips up everyone once. `http.Get` returns a nil error for **any** response the
server sends, including `404 Not Found` and `500 Internal Server Error`. From the
transport's point of view, the request worked: it went out and an answer came back.

```go
resp, err := http.Get(srv.URL + "/issues/999")
if err != nil {
	return err // network trouble: no response at all
}
defer resp.Body.Close()
if resp.StatusCode != http.StatusOK {
	return fmt.Errorf("unexpected status: %s", resp.Status)
}
```

Always check the status code before trusting the body. In chapter 5 you'll turn bad
statuses into proper typed errors.

## Use the constants

`net/http` names every status code: `http.StatusOK` (200), `http.StatusNotFound` (404),
`http.StatusCreated` (201) and so on. They're easier to read than magic numbers.

## Your turn

`trackr version` shows which version of the API it's talking to. The fake API in the
editor answers `GET /version` with a plain-text body and an `X-Trackr-Version` header.
Replace the `TODO` line so that `main`:

1. sends a `GET` to `srv.URL + "/version"` with `http.Get`, printing the error and
   returning if there is one;
2. defers closing the body and reads it with `io.ReadAll`;
3. prints three lines: `resp.Status`, the `X-Trackr-Version` header, and the body.

The expected output is:

```
200 OK
2.4.1
Trackr API 2.4.1 (build 7731)
```

## Further reading

- Go by Example, "HTTP Client": https://gobyexample.com/http-client
