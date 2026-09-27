---
title: What TLS Protects
quiz:
  - question: '`trackr` sends `GET https://api.trackr.dev/issues?q=layoffs` with a bearer token over public Wi-Fi. What can someone on the same network see?'
    options:
      - text: Everything, including the token and the query
      - text: That you connected to `api.trackr.dev` (and roughly how much data moved), but not the path, query, headers or bodies
        correct: true
      - text: Nothing at all, not even which server you connected to
    explanation: |
      TLS encrypts the whole HTTP message: path, query, headers and bodies. The server's
      IP address, the hostname (sent in the TLS handshake) and traffic sizes and timing
      are still visible.
  - question: 'Which of these does TLS **not** give you?'
    options:
      - text: Confidentiality (eavesdroppers can't read the data)
      - text: Integrity (tampering is detected)
      - text: Server authentication (you're talking to the real `api.trackr.dev`)
      - text: Protection against a buggy or malicious API server itself
        correct: true
    explanation: |
      TLS protects the data *in transit* between you and the server. Once it arrives,
      the server can do whatever it likes with it. TLS just makes sure it's the real
      server.
---

Every Trackr request carries a bearer token that can read and change your company's
issues. Over plain `http://`, that token travels as readable text through every Wi-Fi
access point, router and ISP between you and the server. **HTTPS** is HTTP inside
**TLS** (Transport Layer Security), and it fixes that.

## Three guarantees

TLS gives a connection three properties:

1. **Confidentiality.** The data is encrypted. An eavesdropper sees random-looking
   bytes.
2. **Integrity.** Every record is authenticated. If anyone flips a bit in transit,
   the receiver detects it and drops the connection. No silent tampering.
3. **Authentication.** The server proves it really is `api.trackr.dev`, using a
   certificate (next lesson). Without this, the first two are useless, because you
   could be having a perfectly encrypted conversation with an attacker.

## What stays visible

TLS wraps the entire HTTP message, so these are **hidden**:

- the method, path and query string (`/issues?q=layoffs`);
- all headers, including `Authorization` and cookies;
- request and response bodies.

These are **visible** to someone watching the network:

- the IP addresses and ports (they're needed to route packets);
- the hostname, usually, because the client sends it in the handshake so the server
  knows which certificate to present (this is called SNI);
- how much data went each way, and when.

So "someone used Trackr at 9:14" leaks, but not what they looked at. That's why tokens
belong in headers, not URLs. Even though the URL path and query are encrypted on the
wire, URLs still end up in logs at both ends.

## The handshake, roughly

Before any HTTP is sent, client and server do a **TLS handshake**:

1. The client says hello: TLS versions and algorithms it supports, the hostname it
   wants, and its half of a key exchange.
2. The server answers with its choices, its half of the key exchange and its
   **certificate chain**, and signs the handshake with its private key.
3. The client **verifies** the certificate (is it valid, for this hostname, issued by
   someone it trusts?) and the signature.
4. Both sides now derive the same secret keys, which were never sent over the wire.
   Everything after that is encrypted.

With TLS 1.3 this costs one extra network round trip, which is one more reason
connection reuse (chapter 6) matters: you pay for the handshake once per connection,
not once per request.

## Post-quantum key exchange

A future quantum computer could break today's elliptic-curve key exchange, and an
attacker could record encrypted traffic now and decrypt it later ("harvest now, decrypt
later"). Go's TLS stack defends against that by default: since Go 1.24, clients offer
the hybrid **X25519MLKEM768** key exchange, which combines classic X25519 with the
post-quantum ML-KEM algorithm, so an attacker would have to break both. Go 1.26 added
the `SecP256r1MLKEM768` and `SecP384r1MLKEM1024` hybrids to the defaults as well.

You don't have to do anything to get this. It's negotiated automatically when the
server supports it and falls back to classic key exchange when it doesn't. (Setting
`tls.Config.CurvePreferences` to your own list that leaves the hybrids out, or running
with `GODEBUG=tlsmlkem=0`, turns it off, so leave them alone unless you have a reason.) In the next lesson you'll see
`X25519MLKEM768` show up in a real connection.

## In Go, HTTPS is just a URL

```go
resp, err := client.Get("https://api.trackr.dev/issues")
```

That's it. The scheme `https` makes the transport do the TLS handshake with sensible,
secure defaults: TLS 1.2 or newer, modern ciphers, certificate verification against
the system's trusted roots and the post-quantum key exchange. The best TLS advice for
a Go client is mostly: **don't change the defaults**. The rest of this chapter is about
the one default people are tempted to change.
