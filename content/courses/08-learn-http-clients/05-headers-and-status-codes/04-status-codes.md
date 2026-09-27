---
title: Status Codes
quiz:
  - question: |
      What does this print for a `204 No Content` response?

      ```go
      ok := resp.StatusCode >= 200 && resp.StatusCode < 300
      fmt.Println(ok, resp.StatusCode == http.StatusOK)
      ```
    options:
      - text: '`true true`'
      - text: '`true false`'
        correct: true
      - text: '`false false`'
    explanation: |
      204 is in the 2xx success class, but it isn't 200. Checking
      `StatusCode == http.StatusOK` treats perfectly good 201s and 204s as failures.
  - question: '`trackr` gets `429 Too Many Requests`. Whose "fault" is it, and is retrying reasonable?'
    options:
      - text: A server bug, so retry immediately
      - text: The client is sending too fast. Wait (see `Retry-After`) and then retry.
        correct: true
      - text: The request is malformed. Never retry.
    explanation: |
      429 is a 4xx, so the problem is on the client's side, but it's temporary: slow
      down and try again later. Chapter 6 handles it properly.
  - question: Which of these is worth retrying automatically?
    options:
      - text: '`400 Bad Request`'
      - text: '`404 Not Found`'
      - text: '`503 Service Unavailable`'
        correct: true
      - text: '`401 Unauthorized`'
    explanation: |
      503 means the server is temporarily unable to cope, and a later attempt may well
      succeed. The others will fail the same way until the request (or the token)
      changes.
---

Every response starts with a three-digit **status code**. The first digit tells you
the class, and that's often all your code needs to look at.

## The five classes

| Class | Meaning | Whose problem? |
| --- | --- | --- |
| **1xx** informational | "keep going" (you'll rarely see these) | nobody's |
| **2xx** success | it worked | nobody's |
| **3xx** redirection | look somewhere else | Go follows these for you |
| **4xx** client error | *your* request is wrong | fix the request |
| **5xx** server error | the server failed | maybe retry later |

## The ones `trackr` meets

| Code | Constant | When |
| --- | --- | --- |
| 200 | `http.StatusOK` | GET or PATCH succeeded |
| 201 | `http.StatusCreated` | POST created an issue |
| 204 | `http.StatusNoContent` | DELETE succeeded, no body |
| 301/302/307/308 | `http.StatusMovedPermanently`, ... | redirected (handled by `Client`) |
| 304 | `http.StatusNotModified` | your cached copy is still good |
| 400 | `http.StatusBadRequest` | malformed request |
| 401 | `http.StatusUnauthorized` | missing or invalid token |
| 403 | `http.StatusForbidden` | valid token, but not allowed |
| 404 | `http.StatusNotFound` | no such issue |
| 409 | `http.StatusConflict` | e.g. project slug already taken |
| 422 | `http.StatusUnprocessableEntity` | well-formed but invalid, e.g. empty title |
| 429 | `http.StatusTooManyRequests` | slow down (rate limited) |
| 500 | `http.StatusInternalServerError` | server bug |
| 502 | `http.StatusBadGateway` | a proxy couldn't reach the API |
| 503 | `http.StatusServiceUnavailable` | overloaded or in maintenance |
| 504 | `http.StatusGatewayTimeout` | a proxy gave up waiting |

`http.StatusText(code)` gives the standard phrase, e.g.
`http.StatusText(429)` is `"Too Many Requests"`. `resp.Status` already contains
code and phrase together: `"429 Too Many Requests"`.

## Check the class, not one code

For "did it work?", check the whole 2xx range:

```go
if resp.StatusCode < 200 || resp.StatusCode > 299 {
	return fmt.Errorf("unexpected status: %s", resp.Status)
}
```

Checking `== http.StatusOK` is only right when the API promises exactly 200. Plenty of
real bugs come from a client that treats `201 Created` as an error.

## 401 vs 403

They're easy to mix up:

- **401 Unauthorized** really means *unauthenticated*: "I don't know who you are."
  The token is missing, malformed or expired. Tell the user to log in again.
- **403 Forbidden**: "I know who you are, and you can't do that." A new token won't
  help. You need permission.

## Which errors are worth retrying?

A useful split for later chapters:

| Retry? | Codes | Why |
| --- | --- | --- |
| yes, after a delay | 429, 502, 503, 504 | temporary: overloaded, rate limited, proxy trouble |
| maybe | 500 | sometimes a one-off, sometimes a bug that fails every time |
| no | other 4xx | the same request will fail the same way |

## The error body

Non-2xx responses usually carry a body explaining what went wrong. Trackr sends:

```json
{"error": "issue 999 not found", "code": "not_found"}
```

That message is far more useful to show the user than "404 Not Found". In the next
lesson you'll decode it into a proper Go error type.

## Further reading

- MDN, "HTTP response status codes": https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Status
