---
title: The HTTP Methods
quiz:
  - question: '`trackr close 42` marks issue 42 as closed and changes nothing else. Which request fits best?'
    options:
      - text: '`GET /issues/42?state=closed`'
      - text: '`DELETE /issues/42`'
      - text: '`PATCH /issues/42` with body `{"state": "closed"}`'
        correct: true
      - text: '`POST /issues` with body `{"id": 42, "state": "closed"}`'
    explanation: |
      PATCH changes *part* of an existing resource. GET must never change anything,
      DELETE would remove the issue, and POST to the collection creates a new one.
  - question: Which method creates a new issue when you don't know its ID yet?
    options:
      - text: '`POST /projects/apollo/issues`'
        correct: true
      - text: '`GET /projects/apollo/issues/new`'
      - text: '`PUT /projects/apollo/issues`'
    explanation: |
      You POST to the collection, and the server picks the ID and returns the new
      issue (usually with `201 Created`).
---

So far `trackr` has only *read* data with `GET`. To create, change and delete issues it
needs the other **HTTP methods** (sometimes called verbs). The method is the first
word of every request:

```
PATCH /issues/42 HTTP/1.1
Host: api.trackr.dev
Content-Type: application/json

{"state": "closed"}
```

## The five you'll use

| Method | Meaning | Trackr example | Typical success |
| --- | --- | --- | --- |
| `GET` | read a resource | `GET /issues/42` | `200 OK` + the issue |
| `POST` | create something (or run an action) | `POST /projects/apollo/issues` | `201 Created` + the new issue |
| `PUT` | replace a resource completely | `PUT /issues/42` + the whole issue | `200 OK` or `204 No Content` |
| `PATCH` | change part of a resource | `PATCH /issues/42` + `{"state": "closed"}` | `200 OK` + the updated issue |
| `DELETE` | remove a resource | `DELETE /issues/42` | `204 No Content` |

Together they cover **CRUD**: Create (POST), Read (GET), Update (PUT/PATCH) and Delete
(DELETE). The URL names the *thing*, and the method says what to *do* with it. That's
why REST APIs have `/issues/42`, not `/getIssue?id=42` and `/deleteIssue?id=42`.

There are a few more (`HEAD` is a GET without the body, `OPTIONS` asks what's
allowed), but the five above do nearly all the work.

## Constants in net/http

Go names them so you can't typo a method:

```go
http.MethodGet    // "GET"
http.MethodPost   // "POST"
http.MethodPut    // "PUT"
http.MethodPatch  // "PATCH"
http.MethodDelete // "DELETE"
```

Methods are case-sensitive. `"get"` isn't `"GET"`, and many servers will reject it.

## PUT vs PATCH

The difference is what happens to fields you **don't** send.

```
PUT /issues/42      {"title": "Login broken", "state": "closed"}
PATCH /issues/42    {"state": "closed"}
```

With `PUT`, the body *is* the new issue. Leave out `labels` and a strict server clears
them. With `PATCH`, the body lists only the changes, and everything else stays as it
was. `trackr`'s `close`, `assign` and `label` commands are all PATCHes. (That's where
`omitzero` from the last chapter shines: unset fields simply aren't sent.)

## Which ones have a body?

- `POST`, `PUT` and `PATCH` send a request body, usually JSON.
- `GET` and `DELETE` normally don't. Put filters in the query string instead.

Responses can have bodies for any method, except that `204 No Content` never does.

## Beyond the convenience functions

`http.Get` is fine for quick reads, and there's an `http.Post`, but there's no
`http.Patch`, `http.Put` or `http.Delete`. For those, and for any request that needs
headers or a context, you build a `*http.Request` yourself. That's the lesson after
next. First, one more idea that decides which requests are safe to retry.

## Further reading

- MDN, "HTTP request methods": https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Methods
