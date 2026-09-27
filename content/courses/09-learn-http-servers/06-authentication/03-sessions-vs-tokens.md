---
title: Sessions vs Tokens
quiz:
  - question: 'A JWT''s payload is `eyJzdWIiOiJwaXAiLCJhZG1pbiI6ZmFsc2V9`. Who can read it?'
    options:
      - text: Only the server, because it's encrypted with the secret
      - text: Anyone who has the token; it's just base64url-encoded JSON
        correct: true
      - text: Nobody, it's a hash
      - text: Only the client that received it
    explanation: |
      A signed JWT is *signed*, not encrypted. Anyone can base64-decode the header and
      payload. The signature only proves nobody *changed* them. Never put secrets such as
      passwords or personal data you wouldn't show the user into a JWT.
  - question: What's the main drawback of stateless signed tokens compared with server-side sessions?
    options:
      - text: They need a database lookup on every request
      - text: They can't be revoked before they expire without adding back some server-side state
        correct: true
      - text: They only work in browsers
      - text: They can't carry a user ID
    explanation: |
      The server accepts any correctly signed, unexpired token, so a stolen token works
      until `exp`. That's why access tokens are short-lived and paired with revocable refresh
      tokens, which do live on the server.
  - question: An attacker changes a JWT header to `{"alg":"none"}` and removes the signature. What should a correct server do?
    options:
      - text: Accept it, since the token says no signature is needed
      - text: Reject it; the server decides which algorithm it expects and never lets the token choose
        correct: true
      - text: Re-sign it with the server's secret
      - text: Return a 500
    explanation: |
      Early JWT libraries trusted the `alg` field and accepted unsigned tokens. The fix is
      to verify with the algorithm *you* expect (here HMAC-SHA256) and reject anything else.
---

HTTP is **stateless**: every request stands alone, and the server doesn't remember that
this connection logged in a minute ago. After login, the client needs something it can
send with each request to say "it's me again". There are two big families.

## Sessions: the server remembers

1. At login, the server generates a long random **session ID** (`crypto/rand.Text()`),
   stores `sessionID -> userID, expiry` in its database, and sends the ID to the client,
   usually in a cookie.
2. On every request, the client sends the ID back. The server looks it up.
3. Logging out deletes the row. The session is dead instantly.

The ID means nothing by itself. It's a random key into the server's table.

**Pros:** trivial to revoke (log out everywhere, ban a user), and the token reveals
nothing. **Cons:** every request needs a lookup in shared storage, which all your
servers must reach.

## Tokens: the client carries proof

1. At login, the server builds a small statement ("this is user `0192...`, valid until
   10:15") and **signs** it with a secret key only the server knows.
2. The client sends the whole signed token with every request, usually in an
   `Authorization: Bearer <token>` header.
3. The server checks the signature and the expiry. **No lookup needed.** If the
   signature is valid, the server itself must have issued it.

**Pros:** stateless and fast, and any server with the key can verify. **Cons:** you
can't easily revoke one. A stolen token works until it expires. The usual fix is
covered at the end of this chapter.

## JWTs

The most common token format is the **JSON Web Token** (JWT, often said "jot"). It's
three base64url-encoded parts joined by dots:

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzcXVlYWsiLCJzdWIiOiIwMTky...In0.3q2-7wQ...
└───────────── header ──────────────┘ └─────────────── payload ──────────────┘ └ signature ┘
```

- **Header:** JSON describing the token, e.g. `{"alg":"HS256","typ":"JWT"}`.
  `HS256` means HMAC with SHA-256.
- **Payload:** JSON "claims". Registered claim names include:
  - `iss` (issuer): who made it, e.g. `"squeak"`,
  - `sub` (subject): who it's about, e.g. the user ID,
  - `iat` (issued at) and `exp` (expires): Unix timestamps in seconds.
- **Signature:** `HMAC-SHA256(secret, header + "." + payload)`, base64url-encoded.

Two facts people get wrong:

1. **It's signed, not encrypted.** Anyone can decode the payload. The signature proves
   it hasn't been *modified*, not that it's *private*. Keep secrets out of it.
2. **Never let the token choose the algorithm.** A token claiming `"alg":"none"` must be
   rejected. The server knows it uses HS256 and verifies exactly that.

## HMAC in one paragraph

An HMAC is a hash keyed with a secret: `hmac.New(sha256.New, secret)`. Without the
secret, you can't produce a valid MAC for a modified message, and with it, verifying is
just recomputing and comparing, with `hmac.Equal` so the comparison runs in constant time.
Everything you need is in the standard library: `crypto/hmac`, `crypto/sha256` and
`encoding/base64` (with `base64.RawURLEncoding`, the unpadded URL-safe alphabet JWTs use).

In production you might use a well-reviewed JWT library. Here you'll build the HS256
format by hand in the next exercise, which is the best way to understand what those
libraries do.

## Which should Squeak use?

Both are fine and plenty of production systems use each. Browser apps often prefer
session cookies (`HttpOnly` cookies can't be read by JavaScript, so XSS can't steal them).
APIs for mobile apps and other services often use bearer tokens. Squeak follows the
common API pattern:

- a short-lived **access token** (a JWT, one hour) sent as `Authorization: Bearer ...`,
- a long-lived **refresh token** (random, stored server-side, revocable) used only to get
  new access tokens.

So Squeak gets the speed of tokens for everyday requests *and* the revocability of
sessions where it matters.
