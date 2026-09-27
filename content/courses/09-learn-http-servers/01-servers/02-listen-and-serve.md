---
title: ListenAndServe
quiz:
  - question: |
      What's wrong with this `main`?

      ```go
      func main() {
      	mux := http.NewServeMux()
      	mux.HandleFunc("/", home)
      	http.ListenAndServe(":8080", mux)
      	fmt.Println("server started on :8080")
      }
      ```
    options:
      - text: '`ListenAndServe` needs a full URL like `"http://localhost:8080"`'
      - text: The message prints only after the server stops, and the returned error is thrown away
        correct: true
      - text: '`HandleFunc` must be called after `ListenAndServe`'
      - text: Nothing, it prints the message and then serves
    explanation: |
      `ListenAndServe` blocks for as long as the server runs, so the `Println` only runs
      once it has *failed*. Log before you start, and always handle the error it
      returns: `log.Fatal(http.ListenAndServe(":8080", mux))`.
  - question: Why do Go developers usually avoid registering routes on `http.DefaultServeMux`?
    options:
      - text: It's slower than a mux from `http.NewServeMux`
      - text: It doesn't support method patterns like `"GET /squeaks"`
      - text: It's a global that any imported package can add handlers to
        correct: true
      - text: It was removed in Go 1.22
    explanation: |
      `http.HandleFunc` and `http.Handle` register on a package-level global. Any
      dependency can quietly add routes there (the classic example is `net/http/pprof`
      exposing debug pages). Your own `http.NewServeMux()` contains exactly what you put in it.
---

Time to put a handler on a real port. The quickest way is `http.ListenAndServe`:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Squeak is alive!")
	})

	log.Println("listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

Run it on your machine with `go run .`, then in a second terminal:

```
$ curl localhost:8080
Squeak is alive!
```

Press Ctrl+C to stop it.

## The pieces

- **`http.NewServeMux()`** creates a *router* (Go calls it a multiplexer, or mux). It
  looks at each request's method and path and passes it to the matching handler. It
  gets a whole chapter of its own next.
- **`mux.HandleFunc(pattern, fn)`** registers a function for a pattern. `"/"` matches
  every path that nothing more specific claims.
- **`http.ListenAndServe(addr, handler)`** opens a TCP listener on `addr` and serves
  every connection with `handler`. The mux *is* a handler, so it can be passed straight in.

## The address

`":8080"` means "port 8080 on every network interface". `"localhost:8080"` would only
accept connections from the same machine, which is handy in development. Ports below
1024 usually need root, which is why dev servers love 8080 and friends.

## It blocks, and it only returns errors

`ListenAndServe` runs its accept loop on the calling goroutine and doesn't return while
the server is healthy. It returns only on failure, for example when the port is already
in use:

```
listen tcp :8080: bind: address already in use
```

That's why the idiom wraps it in `log.Fatal`: if it ever returns, print why and exit.
Anything written after it in `main` won't run until the server dies. In the last
chapter you'll replace this with a graceful shutdown that returns cleanly.

## Every request gets a goroutine

Each connection is served on its own goroutine. Two mice posting at the same instant
run your handler code *simultaneously*. That's fantastic for throughput and it means
any shared state (a counter, a map of squeaks) must be protected, a theme that will keep
coming back.

## Avoid the default mux

You'll see tutorials write `http.HandleFunc("/", ...)` and pass `nil` to
`ListenAndServe`. That uses `http.DefaultServeMux`, a package-level global. Any
package you import can register routes on it behind your back. Creating your own mux
with `http.NewServeMux()` keeps your routes explicit, and it's what you'll do for the
rest of the course.

## Don't ship this version

`http.ListenAndServe` creates a server with *no timeouts at all*. A client can open a
connection and send its headers one byte per minute, holding a goroutine and a socket
hostage forever. Do that a few thousand times and your server falls over. The next
lesson fixes it with `http.Server`.
