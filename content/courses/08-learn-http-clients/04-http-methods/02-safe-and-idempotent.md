---
title: Safe and Idempotent
quiz:
  - question: 'A `POST /issues` times out, so you don''t know if the server got it. Why is blindly retrying risky?'
    options:
      - text: POST requests can't be sent twice on the same connection
      - text: POST isn't idempotent, so if the first attempt did reach the server you may create a duplicate issue
        correct: true
      - text: The server will reject the second POST with 409 automatically
    explanation: |
      A timeout doesn't tell you whether the server acted. Repeating an idempotent
      request is harmless, but repeating a POST can create a second issue.
  - question: Which request is idempotent but **not** safe?
    options:
      - text: '`GET /issues/42`'
      - text: '`DELETE /issues/42`'
        correct: true
      - text: '`POST /issues`'
    explanation: |
      DELETE changes the server (so it isn't safe), but deleting twice leaves the same
      result as deleting once (so it's idempotent). GET is both, and POST is neither.
  - question: 'Is `PATCH /issues/42` with body `{"points_delta": 1}` idempotent?'
    options:
      - text: Yes, PATCH is always idempotent
      - text: No, each request adds one more point
        correct: true
    explanation: |
      PATCH carries no guarantee. `{"state": "closed"}` gives the same result every
      time, but "add one" changes the result on every repeat.
---

Networks fail in an annoying way: sometimes you **don't know** whether your request
arrived. The connection drops after you sent it, or the response times out. Did the
server create the issue or not? Two properties tell you whether trying again is OK.

## Safe: read-only

A **safe** method doesn't change anything on the server. Calling it is like looking.

- `GET`, `HEAD` and `OPTIONS` are safe.

Safe requests can be retried, cached, prefetched and logged without a second thought.
It's also why a `GET` must never change data. If `GET /issues/42/close` closed an
issue, a link previewer or a crawler could close it by accident.

## Idempotent: repeat-proof

An **idempotent** request has the same effect whether you send it once or ten times.

- `GET`, `HEAD`, `OPTIONS`: nothing changes, so trivially idempotent.
- `PUT /issues/42` with a full issue: the issue ends up as that body, however many
  times you send it.
- `DELETE /issues/42`: after the first one it's gone, and after the fifth it's still
  gone. (The later ones might return `404`, but the *server state* is the same.)
- `POST` is **not** idempotent. Each `POST /issues` creates another issue.
- `PATCH` is **not guaranteed** to be. `{"state": "closed"}` is repeat-proof in
  practice, `{"points_delta": 1}` is not.

| Method | Safe | Idempotent |
| --- | --- | --- |
| GET | yes | yes |
| PUT | no | yes |
| DELETE | no | yes |
| PATCH | no | not guaranteed |
| POST | no | no |

## Why a client cares

Chapter 6 adds automatic retries to `trackr`. The rule it follows:

> Retry idempotent requests on network errors and 5xx responses. Don't
> automatically retry a POST unless you have a way to make it idempotent.

Go's own transport follows the same idea. If a reused keep-alive connection turns out
to be dead before any response arrives, `net/http` quietly retries the request on a
fresh connection, but only for `GET`, `HEAD`, `OPTIONS` and `TRACE`, or for a request
that carries an `Idempotency-Key` header (more on that below). A plain POST is never
replayed behind your back.

## Making POST safe to retry

Many APIs accept an **idempotency key**: a unique ID the client generates and sends
in a header. If the server sees the same key twice, it returns the original result
instead of creating a duplicate.

```go
req.Header.Set("Idempotency-Key", uuid.New().String())
```

The `uuid` package is new in Go 1.27's standard library. Generate the key **once per
logical operation**, before the first attempt, and reuse it for every retry. A new key
per attempt defeats the point.

Not every API supports this, and the header name varies (check the docs), but when
it's there, it turns a risky retry into a safe one.

## Further reading

- MDN, "Idempotent": https://developer.mozilla.org/en-US/docs/Glossary/Idempotent
