---
title: Clients and Servers
quiz:
  - question: In the Trackr setup, which program is the **client**?
    options:
      - text: The Trackr API that stores the issues
      - text: The database behind the Trackr API
      - text: The `trackr` command-line tool you type commands into
        correct: true
      - text: Whichever program started first
    explanation: |
      The client is the side that *starts* the conversation by sending a request.
      Your `trackr` CLI asks for issues; the API server waits for requests and answers them.
      Which one started first doesn't matter.
  - question: A server gets two requests, one after the other, from the same client. What does it know about the first request while it handles the second?
    options:
      - text: Everything, because the connection remembers it
      - text: Nothing, unless the client sends that context again (for example a token) or the server stored it somewhere
        correct: true
      - text: Only the URL of the first request
    explanation: |
      HTTP is **stateless**: each request stands alone. That's why clients send things
      like an `Authorization` header on *every* request instead of logging in once.
---

Welcome to Trackr! Over this course you'll build the brains of `trackr`, a
command-line client for a project-tracking API. It lists issues, creates new ones,
closes old ones and looks up users. None of that lives on your laptop. It lives on
a **server** somewhere, and your program talks to it over the network.

## Two roles

Almost every networked program plays one of two roles:

- A **client** starts the conversation. It sends a **request**: "give me the open
  issues in project `apollo`".
- A **server** waits for requests and sends back a **response**: "here are 12 issues".

```
 trackr (client)                         api.trackr.dev (server)
 ───────────────                         ───────────────────────
       │  request: GET /projects/apollo/issues   │
       │ ──────────────────────────────────────▶ │
       │                                          │  looks up issues
       │  response: 200 OK + JSON list            │
       │ ◀────────────────────────────────────── │
```

The same program can be both. The Trackr API is a server to your CLI, but it might be
a *client* of a database or an email service. The role belongs to a single
conversation, not to the machine.

## Why not just share a database?

You could let every `trackr` user connect straight to the database. Please don't.
An API in the middle:

- **controls access**, so users only see the projects they're allowed to see;
- **validates input**, so nobody creates an issue with a 40 MB title;
- **hides the storage**, so the team can swap databases without breaking every client;
- **speaks a common language**, so a Go CLI, a web app and a phone app can all use it.

That common language is almost always **HTTP**, and the data inside is usually **JSON**.
Both are covered in this course.

## Stateless by design

HTTP servers treat each request on its own. If `trackr` fetches issues and then
fetches users, the server doesn't "remember" the first call. Anything it needs, such as
who you are, has to be in each request. That sounds wasteful, but it's what lets a
busy API spread requests across many machines: any of them can answer any request.

## What you'll build

Chapter by chapter, `trackr` grows from a single `http.Get` into a proper client that:

1. builds URLs safely and decodes JSON responses;
2. creates, updates and deletes issues with the right HTTP methods;
3. authenticates, and turns error responses into Go errors;
4. uses timeouts, retries, backoff and rate limits;
5. pages through thousands of issues with an iterator;
6. talks HTTPS safely without leaking secrets;
7. is fully tested against fake servers, with no internet required.

You already know Go, concurrency, `context` and testing. This course puts them to work
on the network.

## A note on the network

The exercise runner has no internet access. That's fine, because every exercise starts a
tiny fake Trackr API on your own machine with `net/http/httptest` and hands your code
its address. By the last chapter you'll be writing those fake servers yourself.
