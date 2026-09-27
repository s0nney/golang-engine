---
title: Anatomy of a URL
quiz:
  - question: |
      In this URL, what is the **path**?

      ```
      https://api.trackr.dev:8443/projects/apollo/issues?state=open#top
      ```
    options:
      - text: '`api.trackr.dev:8443/projects/apollo/issues`'
      - text: '`/projects/apollo/issues`'
        correct: true
      - text: '`/projects/apollo/issues?state=open`'
      - text: '`/projects/apollo/issues?state=open#top`'
    explanation: |
      The path starts at the first `/` after the host and stops at `?` (the query) or `#`
      (the fragment). `state=open` is the query, and `top` is the fragment.
  - question: Which part of a URL is **never sent** to the server?
    options:
      - text: The query string
      - text: The port
      - text: The fragment (after `#`)
        correct: true
      - text: The path
    explanation: |
      The fragment is for the client only. Browsers use it to scroll to a spot on the page.
      An HTTP client strips it before sending the request.
---

Every request `trackr` makes starts with a **URL** (Uniform Resource Locator). It
answers three questions at once: *how* to talk (the scheme), *who* to talk to (the
host) and *what* to ask for (the path and query).

## The parts

```
  https://ana:s3cret@api.trackr.dev:8443/projects/apollo/issues?state=open&label=bug#comments
  └─┬─┘   └───┬────┘ └──────┬─────┘ └┬─┘└──────────┬─────────┘ └─────────┬────────┘ └───┬───┘
 scheme   userinfo       host      port          path                  query          fragment
```

| Part | Example | What it's for |
| --- | --- | --- |
| **scheme** | `https` | The protocol. `http` or `https` for us. |
| **userinfo** | `ana:s3cret` | Optional username and password. Rare, and risky (see below). |
| **host** | `api.trackr.dev` | The machine to connect to: a domain name or an IP address. |
| **port** | `8443` | Which "door" on that machine. Optional, since each scheme has a default. |
| **path** | `/projects/apollo/issues` | Which resource on the server. |
| **query** | `state=open&label=bug` | Extra parameters as `key=value` pairs joined by `&`. |
| **fragment** | `comments` | A spot inside the resource. Only the client uses it. |

## Paths name resources

REST-style APIs like Trackr use the path as an address for a *thing*:

```
/projects                      every project
/projects/apollo               one project
/projects/apollo/issues        the issues in that project
/issues/42                     issue number 42
/users/ana                     a user
```

Plural nouns name collections, and an ID after them picks one item. The HTTP method
(next chapters) says what to *do* with that thing.

## Queries filter and tweak

The query string holds options that don't identify a new resource but change what you
get back: filters, sorting and paging.

```
/projects/apollo/issues?state=open&label=bug&label=ui&sort=-created&page=2
```

Notice that `label` appears twice. That's allowed, and it usually means "any of these".
Order doesn't matter to most servers.

## Special characters must be escaped

URLs can only contain a limited set of characters. Anything else, such as a space, `&`
inside a value, or non-ASCII letters, is **percent-encoded**: a `%` followed by the
byte in hex.

```
title "login & signup"   →   title=login+%26+signup   (in a query)
project "web app"        →   /projects/web%20app      (in a path)
```

An unescaped `&` in a value would split it into two parameters, and an unescaped `/` in
a project name would look like an extra path segment. Building URLs by gluing strings
together gets this wrong sooner or later, which is why the next two lessons use
`net/url` to do it properly.

## Don't put secrets in URLs

`https://ana:s3cret@...` works, and so does `?api_key=...`, but URLs get written to
server logs, proxy logs, browser history and error messages. Tokens belong in
**headers** instead. Chapter 5 shows how, and chapter 8 covers keeping them out of logs.

## Further reading

- MDN, "What is a URL?": https://developer.mozilla.org/en-US/docs/Learn_web_development/Howto/Web_mechanics/What_is_a_URL
