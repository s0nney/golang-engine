---
title: DNS and Ports
quiz:
  - question: '`trackr` requests `https://api.trackr.dev/issues`. Which address and port does it connect to?'
    options:
      - text: Port 80 on whatever IP `api.trackr.dev` resolves to
      - text: Port 443 on whatever IP `api.trackr.dev` resolves to
        correct: true
      - text: Port 443 on `api.trackr.dev`, with no lookup needed
      - text: Port 8080, the standard API port
    explanation: |
      The URL names no port, so the scheme's default applies: 443 for `https`, 80 for
      `http`. Before connecting, the client asks DNS to turn the hostname into an IP
      address, because TCP connects to IPs, not names.
  - question: |
      What does this print?

      ```go
      u, _ := url.Parse("https://api.trackr.dev/issues")
      fmt.Printf("%q\n", u.Port())
      ```
    options:
      - text: '`"443"`'
      - text: '`""`'
        correct: true
      - text: '`"https"`'
    explanation: |
      `Port()` only reports a port that's written in the URL. The default port is
      filled in later, by the transport, when it dials.
---

A URL says `api.trackr.dev`, but the network doesn't route names. It routes **IP
addresses** like `203.0.113.7` or `2001:db8::7`. Two things happen before the first
byte of your request is sent: a name lookup, then a connection to a port.

## DNS: the internet's phone book

The **Domain Name System** turns hostnames into IP addresses. At a high level:

1. Your program asks the operating system's resolver: "what's `api.trackr.dev`?"
2. The resolver checks its cache and the local hosts file (`/etc/hosts`, which is
   why `localhost` always works).
3. If it still doesn't know, it asks a DNS server, which may ask others (the root,
   then `.dev`, then `trackr.dev`'s own servers) until one gives an answer.
4. The answer comes back with a **TTL** (time to live), so it can be cached.

`net/http` does all of this for you. You can also do it yourself:

```go
addrs, err := net.DefaultResolver.LookupHost(ctx, "localhost")
fmt.Println(addrs, err) // [::1 127.0.0.1] <nil>
```

A name can resolve to several addresses (IPv4 and IPv6, or many servers behind one
name). The dialer tries them in turn.

### When DNS fails

A typo in the hostname fails at the lookup step, before any HTTP happens:

```go
_, err := http.Get("http://api.trackr.invalid/issues")
fmt.Println(err)
// Get "http://api.trackr.invalid/issues": dial tcp: lookup api.trackr.invalid: no such host

if dnsErr, ok := errors.AsType[*net.DNSError](err); ok {
	fmt.Println("DNS lookup failed for", dnsErr.Name, "not found:", dnsErr.IsNotFound)
}
// DNS lookup failed for api.trackr.invalid not found: true
```

`http.Get` wraps the underlying `*net.DNSError`, and `errors.AsType` digs it out. A
nice CLI turns that into "can't find host api.trackr.invalid, check your config"
instead of a wall of text.

## Ports: which door to knock on

One machine runs many network programs. A **port** (a number from 1 to 65535) picks
which one you're talking to. A few defaults to know:

| Port | Used for |
| --- | --- |
| 80 | `http` |
| 443 | `https` |
| 8080, 8000, 3000 | common choices for local development servers |

If the URL has no port, the scheme decides. `u.Port()` returns `""` in that case,
because the port isn't written in the URL. Here's how a client works out what to dial:

```go
for _, raw := range []string{
	"https://api.trackr.dev/issues",
	"http://api.trackr.dev/issues",
	"http://localhost:8080/issues",
	"http://[::1]:9000/",
} {
	u, _ := url.Parse(raw)
	port := u.Port()
	if port == "" {
		port = map[string]string{"http": "80", "https": "443"}[u.Scheme]
	}
	fmt.Printf("%-32s dial %s\n", raw, net.JoinHostPort(u.Hostname(), port))
}
```

Output:

```
https://api.trackr.dev/issues    dial api.trackr.dev:443
http://api.trackr.dev/issues     dial api.trackr.dev:80
http://localhost:8080/issues     dial localhost:8080
http://[::1]:9000/               dial [::1]:9000
```

`net.JoinHostPort` adds the square brackets IPv6 needs. Don't glue host and port
together with `+ ":" +` yourself.

## Nobody home: connection refused

If DNS works but nothing is listening on that port, the dial fails straight away:

```
Get "http://127.0.0.1:1/issues": dial tcp 127.0.0.1:1: connect: connection refused
```

That's the error you'll see when `trackr` points at a local dev server you forgot to
start.

## Why httptest URLs look odd

`httptest.NewServer` listens on `127.0.0.1` (the loopback address, "this machine") and
asks the OS for any free port, so `srv.URL` looks like `http://127.0.0.1:41823`. No DNS
lookup is needed for an IP address, and the port changes every run. That's why every
exercise takes a `baseURL` instead of hard-coding one.
