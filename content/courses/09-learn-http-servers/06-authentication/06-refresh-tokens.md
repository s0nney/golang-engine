---
title: Refresh Tokens
quiz:
  - question: Why does Squeak store a SHA-256 hash of each refresh token instead of the token itself?
    options:
      - text: Hashes are shorter, which saves memory
      - text: If the token table leaks, the attacker gets hashes that can't be used as tokens
        correct: true
      - text: The client needs the hash to refresh
      - text: SHA-256 makes the token expire automatically
    explanation: |
      A refresh token works like a password. Storing only its hash means a database leak
      doesn't hand out working tokens. A fast hash like SHA-256 is fine here (unlike for
      passwords), because the token is 128+ random bits and can't be guessed from a dictionary.
  - question: With refresh token *rotation*, what should happen when an already-used refresh token is presented again?
    options:
      - text: Issue a new access token as normal
      - text: Treat it as theft; reject it and revoke the whole family of tokens descended from it
        correct: true
      - text: Return the same access token as last time
      - text: Extend its expiry
    explanation: |
      Each refresh token is supposed to be used once. A second use means two parties hold
      it: the real user and an attacker. You can't tell which is which, so revoking the
      whole chain forces a fresh login and locks the thief out.
---

Squeak's access tokens expire after an hour, which keeps a stolen token's damage small.
But nobody wants to type their password every hour. The standard fix is a second kind of
token.

## Two tokens, two jobs

| | Access token | Refresh token |
|---|---|---|
| Format | signed JWT | long random string |
| Lifetime | short (15 min to 1 h) | long (days to weeks) |
| Sent | on every API request | only to `POST /api/refresh` |
| Checked by | signature and `exp` (no lookup) | database lookup |
| Revocable | not really, it just expires | yes, delete or mark it revoked |

At login, Squeak returns both:

```json
{
  "id": "0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6",
  "email": "pip@squeak.dev",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "M2KJ4GBBMVQZ7DSR5XNCEYHT3A"
}
```

When the access token expires, the client sends the refresh token to `/api/refresh` and
gets a new access token. The everyday traffic stays stateless and fast. The one stateful
lookup happens once an hour per user.

## Making one

A refresh token needs no structure, just unguessable randomness. `crypto/rand.Text()`
gives you at least 128 bits of it as a base32 string:

```go
refresh := rand.Text()
```

Store it server-side with the user ID, an expiry, and a revoked-at time:

```go
type RefreshToken struct {
	Hash      [32]byte // sha256 of the token, not the token itself
	UserID    uuid.UUID
	ExpiresAt time.Time
	RevokedAt time.Time // zero while still valid
}
```

## Store a hash

A refresh token is effectively a password that lasts for weeks. If your token table
leaks, the raw tokens would let an attacker log in as everyone. So store
`sha256.Sum256([]byte(token))` and look tokens up by hash. Unlike passwords, a fast hash
is fine here, because a 128-bit random token can't be brute-forced, however fast the
hash. (`[32]byte` arrays are comparable, so they work as map keys in the in-memory store.)

## The refresh endpoint

```go
func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := getBearerToken(r.Header)
	if err != nil {
		unauthorized(w, "missing refresh token")
		return
	}
	rt, err := cfg.refreshTokens.Get(r.Context(), sha256.Sum256([]byte(token)))
	now := time.Now()
	if err != nil || !rt.RevokedAt.IsZero() || !now.Before(rt.ExpiresAt) {
		unauthorized(w, "invalid refresh token")
		return
	}
	access, err := makeToken(rt.UserID, cfg.jwtSecret, now, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't create token")
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"token": access})
}
```

And `POST /api/revoke` sets `RevokedAt` so the token can't be used again. That's
"log out", and "log out everywhere" is revoking every token for a user ID. Access tokens
already issued still work until they expire, which is the reason to keep them short.

## Rotation

A refinement many APIs use: every refresh also issues a **new refresh token** and
revokes the old one. Each refresh token is then single-use. If an old one ever comes
back, two parties have it, the real client and a thief, and you can't tell which is
which. The safe response is to revoke every token in that chain and make the user log in
again. It turns a silent, weeks-long compromise into one that's detected quickly.

## Where the client keeps them

- **Mobile and desktop apps:** the platform's secure storage (Keychain, Keystore).
- **Browsers:** refresh tokens are best kept in an `HttpOnly`, `Secure`,
  `SameSite=Strict` cookie scoped to `/api/refresh`, where JavaScript (and so XSS) can't
  reach it. Go sets one with `http.SetCookie(w, &http.Cookie{...})`.

## Recap of Squeak's auth

1. `POST /api/users`: register (password hashed with PBKDF2).
2. `POST /api/login`: check the password, return an access JWT and a refresh token.
3. Protected routes: `requireAuth` middleware verifies the JWT from `Authorization`.
4. `POST /api/refresh`: trade a valid refresh token for a new access token.
5. `POST /api/revoke`: kill a refresh token.

You now know *who* is calling. The next chapter decides *what they're allowed to do*.
