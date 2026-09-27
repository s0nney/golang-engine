---
title: Your Own http.Client
quiz:
  - question: What is `http.DefaultClient.Timeout`?
    options:
      - text: 30 seconds
      - text: 10 seconds
      - text: Zero, which means no timeout at all
        correct: true
    explanation: |
      The default client will wait forever for a server that accepts the connection
      and then never answers. That's the main reason not to use it in real programs.
  - question: 'How many `*http.Client` values should `trackr` create?'
    options:
      - text: A new one for every request
      - text: One, created at startup and shared by every request
        correct: true
      - text: One per goroutine, because clients aren't safe for concurrent use
    explanation: |
      A client (really, its `Transport`) holds the pool of open connections. Sharing
      one lets requests reuse them, and clients are safe for concurrent use.
---

`http.Get`, `http.Post` and `http.DefaultClient.Do` all use one package-level client,
`http.DefaultClient`. It's convenient, and it's a trap for anything long-running.

## The problem with DefaultClient

```go
var DefaultClient = &Client{}
```

That's its whole definition. A zero `Client` has:

- **No timeout.** A server that accepts your connection and then hangs will block
  `trackr` forever.
- **Global, shared state.** Any package in your program (or a library you import) can
  change `http.DefaultClient` or `http.DefaultTransport`, and that affects you.

Fine for a quick script, not for a tool people rely on.

## Make your own

`http.Client` is a struct with a handful of fields:

```go
client := &http.Client{
	Timeout: 10 * time.Second,
}
resp, err := client.Do(req)
```

| Field | What it does | Zero value |
| --- | --- | --- |
| `Timeout` | limit for the whole request, including reading the body | no limit |
| `Transport` | the `http.RoundTripper` that actually sends requests | `http.DefaultTransport` |
| `CheckRedirect` | decides whether to follow a redirect | follow up to 10 |
| `Jar` | stores cookies between requests | cookies ignored |

The methods are the ones you know: `Do`, `Get`, `Post`, `Head`, `PostForm`.

## Create once, share everywhere

A `Client` keeps a pool of open connections (inside its `Transport`) so later requests
skip the slow connection setup. It's also **safe for concurrent use** by many
goroutines. So create it once and share it:

```go
type Client struct {
	baseURL *url.URL
	token   string
	http    *http.Client
}

func New(baseURL *url.URL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}
```

Making a new `http.Client` per request isn't a disaster (they all share
`DefaultTransport` unless you give them their own), but creating a new **Transport**
per request is. Every request would pay for a fresh connection, and the old idle
ones would pile up.

## Controlling redirects

By default a client follows up to 10 redirects. `CheckRedirect` lets you change that.
Returning `http.ErrUseLastResponse` stops and hands you the redirect response itself:

```go
client := &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if req.URL.Host != via[0].URL.Host {
			return fmt.Errorf("refusing redirect to another host: %s", req.URL.Host)
		}
		if len(via) >= 3 {
			return errors.New("too many redirects")
		}
		return nil
	},
}
```

`via` holds the requests made so far, oldest first. This one refuses to leave the
Trackr host, which keeps custom auth headers like `X-Api-Key` from leaking (see the
last chapter).

## Let callers inject it

Libraries shouldn't hard-code their client. Accept one, and fall back to a sensible
default:

```go
func New(baseURL *url.URL, token string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{baseURL: baseURL, token: token, http: httpClient}
}
```

Now callers can tune timeouts or add a proxy, and tests can pass in a client wired
to a fake server (chapter 9). Exercises from here on take a `*http.Client` for
exactly this reason.

## Further reading

- The `net/http` docs on Clients and Transports: https://pkg.go.dev/net/http#hdr-Clients_and_Transports
