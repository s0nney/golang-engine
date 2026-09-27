---
title: Building Requests
quiz:
  - question: What's the difference between `http.NewRequestWithContext` and `client.Do`?
    options:
      - text: '`NewRequestWithContext` sends the request, and `Do` reads the response'
      - text: '`NewRequestWithContext` only builds a `*http.Request` value, and `Do` actually sends it'
        correct: true
      - text: They're two names for the same thing
    explanation: |
      Building a request touches no network at all. It can only fail on bad input, like
      an invalid method or URL. `Do` is what sends it and waits for a response.
  - question: |
      What does this print?

      ```go
      ctx, cancel := context.WithCancel(context.Background())
      cancel()
      req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
      _, err := http.DefaultClient.Do(req)
      fmt.Println(errors.Is(err, context.Canceled))
      ```
    options:
      - text: '`true`'
        correct: true
      - text: '`false`'
      - text: It panics
    explanation: |
      The request carries its context. `Do` sees it's already cancelled, gives up
      without sending anything, and wraps `context.Canceled` in the error it returns.
---

`http.Get` hides two steps: building a request and sending it. Once you need a
different method, a header or a context, you do the steps yourself.

## Step 1: build it

```go
req, err := http.NewRequestWithContext(ctx, http.MethodDelete, baseURL+"/issues/42", nil)
if err != nil {
	return err
}
req.Header.Set("Accept", "application/json")
```

`http.NewRequestWithContext(ctx, method, url, body)` returns a `*http.Request`:

- **`ctx`** controls cancellation and deadlines for the whole request (next section).
- **`method`** is one of the `http.Method...` constants.
- **`url`** is a string, which is where `net/url` from chapter 2 comes in.
- **`body`** is an `io.Reader`, or `nil` for no body. Next lesson.

Building is pure. No network happens, and it only fails on bad input:

```
parse "http://api.trackr.dev/%zz": invalid URL escape "%zz"
net/http: invalid method "BAD METHOD"
```

Before sending, you can change the request: set headers, adjust `req.URL`, and so on.

## Step 2: send it

```go
resp, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer resp.Body.Close()
```

`Do` sends the request and waits for the response headers. From here it's the pattern
you already know: check `err`, defer the close, check the status, read the body.
`http.DefaultClient` is a ready-made `*http.Client`. Chapter 6 explains why you'll
soon want your own.

## All together

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("server got: %s %s (Accept: %s)\n", r.Method, r.URL, r.Header.Get("Accept"))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, srv.URL+"/issues/42", nil)
	if err != nil {
		fmt.Println("building request:", err)
		return
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("client got:", resp.Status)
}
```

Output:

```
server got: DELETE /issues/42 (Accept: application/json)
client got: 204 No Content
```

## Why WithContext?

There's also `http.NewRequest(method, url, body)`, without a context. It uses
`context.Background()`, which never cancels. That's fine in a throwaway script, but in
`trackr` you want every request tied to a context so that:

- pressing Ctrl+C (which cancels the root context) aborts in-flight requests;
- a per-command deadline ("give up after 10 seconds") covers every request;
- a caller that no longer needs the answer can stop waiting.

The context covers the **whole** exchange: DNS, connecting, sending, waiting for
headers, *and* reading the body. If it's cancelled halfway through `io.ReadAll`, the
read fails with the context's error.

A cancelled request's error wraps the context error, so you can check for it:

```go
if errors.Is(err, context.DeadlineExceeded) {
	return fmt.Errorf("Trackr API didn't answer in time: %w", err)
}
```

Go's rule of thumb applies here too: functions that do I/O take a `ctx
context.Context` as their **first** parameter. From now on, every `trackr` function
that talks to the API does.

## Reusing a request?

A `*http.Request` is meant to be sent once. To send "the same" request again (say,
on a retry), build a fresh one. Chapter 6 shows how that interacts with request
bodies.
