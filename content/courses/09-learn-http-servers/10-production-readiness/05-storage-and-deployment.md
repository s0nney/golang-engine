---
title: Deploying Squeak
quiz:
  - question: |
      At shutdown, in what order should Squeak stop its parts?
    options:
      - text: Close the database, then stop the HTTP server, then flush the logs
      - text: Stop accepting requests and drain them, stop background jobs, then close the database
        correct: true
      - text: Close everything at once in separate goroutines
      - text: The order doesn't matter, because the process is exiting anyway
    explanation: |
      Shut down in the reverse order of startup. In-flight requests and background jobs
      still use the database, so closing it first would turn a graceful drain into a
      burst of 500s. Close each thing only after everything that uses it has stopped.
  - question: |
      Squeak runs behind a load balancer that sets `X-Forwarded-For`. The rate limiter
      keys on the leftmost address in that header, taken from any request. What's the problem?
    options:
      - text: '`X-Forwarded-For` is always empty'
      - text: Any client can send its own `X-Forwarded-For` header, pick a new "IP" per request, and never hit the limit
        correct: true
      - text: Load balancers strip all headers
      - text: There's no problem; the header can't be forged
    explanation: |
      Headers are just text the client controls. Only trust forwarding headers set by
      your own proxy: take the address your proxy appended (the rightmost entries it
      controls), and ignore the header on requests that didn't come through it.
---

Squeak routes, validates, authenticates, stores, tests and shuts down gracefully. The
last step is running it somewhere other people can reach. Deployment platforms vary
wildly (a VPS with systemd, a container platform, Kubernetes), but the checklist for the
program itself is surprisingly short.

## Startup order

`main` builds Squeak from the bottom up, and each step can fail loudly *before* the
server accepts traffic:

1. **Load configuration** and validate it (two lessons ago). A typo in an
   environment variable should stop the deploy, not surface at 3 a.m.
2. **Set up logging**, so everything after this can report problems.
3. **Open storage.** For the SQL store from chapter 9: `sql.Open`, a bounded
   `PingContext` to prove the database is reachable, pool limits such as
   `db.SetMaxOpenConns` sized for what the database can handle, and migrations.
4. **Assemble the router and middleware** with the store and config.
5. **Start serving**, and only then mark the instance ready (last lesson).

## Shutdown order

Shutdown runs the list backwards: mark not-ready, stop accepting requests and drain the
ones in flight (`Shutdown`), stop background jobs and wait for them, and only then
close the database. Close a resource only after everything that uses it has stopped.

## Behind a proxy

Most deployments put Squeak behind a load balancer or reverse proxy that terminates
HTTPS and forwards plain HTTP to your process. Two consequences:

- **HTTPS still matters.** Passwords and bearer tokens must never cross a public network
  in plain text. Make sure the proxy (or Squeak itself, with `ListenAndServeTLS`)
  handles TLS, and redirect or reject plain HTTP.
- **`r.RemoteAddr` is now the proxy's address.** The real client IP arrives in a header
  such as `X-Forwarded-For`, which any client can also send. Trust it only when the
  request really came from your proxy, and only the part your proxy wrote. Otherwise a
  client can pick a fresh "IP" for every request and walk straight past the rate limiter.

## The pre-flight checklist

Before real users arrive:

- **Tests pass with `-race`**, and `go vet` is clean.
- **Build one static binary**: `CGO_ENABLED=0 go build -o squeak ./cmd/squeak`. Go
  binaries need no runtime installed, which makes tiny container images easy.
- **Server timeouts are set** (chapter 1), and **request bodies are size-limited**
  (chapter 4).
- **Secrets come from the environment** or a secret manager, never from the image,
  the repository or the logs.
- **The process receives `SIGTERM`**, and the platform waits longer than your drain
  timeout before killing it.
- **Health checks** point at `/livez` and `/readyz`.
- **Backups** of the database exist, and you've **restored one**. A backup you've never
  restored is a hope, not a backup.

## Rehearse

Finally, try the failure paths on purpose, in a staging environment: restart Squeak
while a request is in flight and check the squeak survives; stop the database and watch
readiness flip; send `SIGTERM` and read the shutdown logs. Every surprise you find here
is one your users won't.

Further reading: [Managing database connections](https://go.dev/doc/database/manage-connections)
and the [net/http server docs](https://pkg.go.dev/net/http#Server).
