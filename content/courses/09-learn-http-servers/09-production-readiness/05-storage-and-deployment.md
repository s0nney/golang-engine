---
title: From Memory to Deployment
quiz:
  - question: |
      Why should a SQL-backed Squeak store reuse one long-lived *sql.DB?
    options:
      - text: 'It represents a concurrency-safe connection pool'
        correct: true
      - text: 'It guarantees every query uses one permanent connection'
      - text: 'It stores all rows in Go memory'
    explanation: |
      sql.DB manages a pool and is safe for concurrent use. Open it during startup, reuse it across requests, and close it after requests and background work have stopped.
---

Squeak now routes, validates, authenticates, and tests requests. Its in-memory store
still loses every squeak when the process exits. That is useful for exercises and tests,
but persistence is the next boundary to cross before real users depend on it.

## Keep the handler contract

The repository interface from chapter 5 is the seam. A SQL store implements the same
methods while keeping SQL and driver details out of handlers. For example, an adapter
can translate the database's missing-row error:

```go
func storageError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
```

A real adapter also maps rows into `Squeak`, passes the request context to queries,
and uses parameterized statements. Never build SQL by concatenating a user's body or ID.
Changes that must succeed together belong in a transaction, such as creating a resource
and recording an idempotency result. Database constraints should enforce uniqueness;
an earlier "does it exist?" query cannot prevent two concurrent inserts.

`database/sql` supplies the common API, but drivers are separate dependencies. The
browser exercises stay standard-library-only, so adding a concrete database belongs
in your own application project. Choose a driver, create versioned schema migrations,
and implement the existing repository contract there.

## Own the pool

Open a long-lived `*sql.DB` at startup, configure pool limits for the database's capacity,
and verify connectivity with a bounded `PingContext`. Opening the handle alone doesn't
prove the database is reachable. Reuse that handle across requests; don't open and close
a pool per handler call. Close it after HTTP requests and background jobs have drained.

Integration tests should exercise migrations and repository behavior against the chosen
database. Keep the fast handler tests with fakes too: a database test doesn't replace
checking that a missing squeak becomes a 404.

## Assemble and ship

Your command now has a clear order: load configuration, initialize logging and storage,
assemble the router and middleware, then start serving. Shutdown reverses the dependencies:
stop traffic, drain handlers, stop jobs, then close storage.

Before a deployment, run your tests and race checks, build a binary, and give it the
configuration it expects. Arrange HTTPS at the server or a trusted proxy. If a proxy
supplies client IP headers, trust them only from that proxy; arbitrary clients must not
choose their own rate-limit identity. Set body limits and server timeouts, keep credentials
out of the image, and make sure the process receives termination signals.

Finally, rehearse a restart: create a squeak, stop gracefully, start again, and retrieve
it. Try a failed dependency and observe readiness. Verify backups can actually be restored.
These checks turn deployment assumptions into behavior you can see.

You've built the pieces of a Go HTTP service. The next project is assembling them around
persistent storage, with the same small interfaces and tests that made Squeak understandable.

Further reading: [Managing database connections](https://go.dev/doc/database/manage-connections).
