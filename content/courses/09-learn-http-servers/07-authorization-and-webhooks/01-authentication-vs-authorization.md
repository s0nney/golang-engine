---
title: Authentication vs Authorization
quiz:
  - question: 'Which question does **authorization** answer?'
    options:
      - text: Who is making this request?
      - text: Is this caller allowed to do this particular thing?
        correct: true
      - text: Is the password hash strong enough?
      - text: Has the access token expired?
    explanation: |
      Authentication (authn) establishes identity: "this is Pip". Authorization (authz)
      decides permissions: "Pip may delete squeak 7 because Pip wrote it". The first gives
      you a 401 when it fails, the second a 403.
  - question: |
      `DELETE /api/squeaks/{id}` is wrapped in `requireAuth`, and the handler deletes the
      squeak with that ID. What's the vulnerability?
    options:
      - text: None; only logged-in users can reach it
      - text: Any logged-in user can delete *anyone's* squeak just by changing the ID in the URL
        correct: true
      - text: The ID should be in the body, not the URL
      - text: DELETE requests can't be authenticated
    explanation: |
      This is an *insecure direct object reference* (IDOR), one of the most common API bugs.
      Being logged in isn't the same as owning the resource. The handler must check that the
      squeak's author is the caller.
---

Two words that sound alike and get mixed up constantly:

- **Authentication** (authn): *who are you?* Squeak answers it with passwords and
  tokens. Failure is **401 Unauthorized** (the name is historical: it really means
  "unauthenticated").
- **Authorization** (authz): *are you allowed to do this?* Failure is **403 Forbidden**.

The last chapter was all authentication. `requireAuth` tells a handler that the caller
is Pip. It says nothing about what Pip may do.

## The most common API bug

Look at this handler:

```go
// Wrapped in requireAuth, so the caller is logged in.
func (cfg *apiConfig) handleDeleteSqueak(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid squeak id")
		return
	}
	if err := cfg.squeaks.DeleteSqueak(r.Context(), id); err != nil {
		respondWithError(w, http.StatusNotFound, "squeak not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

It authenticates, and it never authorizes. Whiskers logs in, copies the ID of one of
Pip's squeaks, and sends `DELETE /api/squeaks/<that id>`. Gone. This class of bug is
called an **IDOR** (insecure direct object reference), and it tops API security lists
year after year: the ID in the URL is trusted without asking whose object it is.

UUIDs make IDs hard to *guess*, but they aren't secret. They appear in URLs, in shared
links, and in every `GET /api/squeaks` response. Unguessable IDs don't replace a
permission check.

## Where authorization checks live

Authentication is the same for every route, so it fits neatly into middleware.
Authorization usually depends on the **specific resource**: you can't know who owns
squeak 7 until you've loaded squeak 7. So ownership checks live in the handler (or in a
service layer the handler calls), right after loading the thing:

```go
squeak, err := cfg.squeaks.GetSqueak(ctx, id)
// ... 404 if missing ...
if squeak.AuthorID != userID {
	respondWithError(w, http.StatusForbidden, "you can only delete your own squeaks")
	return
}
```

Rules that don't depend on a resource *can* be middleware, for example "only admins may
use `/admin/`":

```go
func (cfg *apiConfig) requireAdmin(next http.Handler) http.Handler {
	return cfg.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := userIDFrom(r.Context())
		user, err := cfg.users.GetUser(r.Context(), userID)
		if err != nil || !user.IsAdmin {
			respondWithError(w, http.StatusForbidden, "admins only")
			return
		}
		next.ServeHTTP(w, r)
	}))
}
```

`requireAdmin` builds on `requireAuth`, so it's authenticate first, then authorize. The
metrics and reset endpoints from chapter 3 finally get locked down with it.

Notice that it loads the user from the store rather than trusting an `"admin": true`
claim in the token. A claim would work too, and saves a lookup, but it stays true until
the token expires, even if you demote someone.

## 403 or 404?

When Whiskers asks to delete Pip's squeak, the honest answer is **403**: it exists, and
you can't. Some APIs deliberately answer **404** instead, so outsiders can't even confirm
that a private resource exists (GitHub does this for private repositories). For Squeak,
where every squeak is public anyway, 403 is clearer.

## Deny by default

A few principles that keep authorization bugs rare:

- **Check on every request, on the server.** Hiding the delete button in the app isn't
  a permission check.
- **Default to "no".** Write checks as "allowed if the caller is the author", never
  "blocked if the caller is someone specific".
- **Test the negative cases.** It's easy to test that Pip can delete Pip's squeak. The
  test that matters is Whiskers *failing* to. You'll write exactly that next.
