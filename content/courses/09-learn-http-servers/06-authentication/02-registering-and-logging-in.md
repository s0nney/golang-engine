---
title: Registering and Logging In
quiz:
  - question: A login attempt uses an email that isn't registered. What should Squeak answer?
    options:
      - text: '`404` with `{"error":"no user with that email"}`'
      - text: '`401` with the same generic `{"error":"incorrect email or password"}` used for a wrong password'
        correct: true
      - text: '`400` with `{"error":"email not found, please register"}`'
      - text: '`200` with an empty token'
    explanation: |
      Different answers for "unknown email" and "wrong password" let anyone check which
      emails have accounts (*account enumeration*). Answer both with the same status
      and message. (Registration can still reveal it with a 409, a trade-off most apps
      accept.)
  - question: |
      What's the risk in this type?

      ```go
      type User struct {
      	ID           uuid.UUID `json:"id"`
      	Email        string    `json:"email"`
      	PasswordHash string    `json:"password_hash"`
      }
      // ...
      respondWithJSON(w, http.StatusCreated, user)
      ```
    options:
      - text: None, the hash is safe to share
      - text: Every response that includes a `User` leaks the password hash to the client
        correct: true
      - text: '`uuid.UUID` can''t be encoded to JSON'
      - text: Struct tags must be in upper case
    explanation: |
      Hashes are slow to crack, not impossible. Never send them anywhere. Tag the field
      `json:"-"`, or better, respond with a separate response type that simply doesn't
      have the field.
---

Squeak needs mice. Two endpoints:

- `POST /api/users` registers a new user.
- `POST /api/login` checks an email and password and (in the next lessons) hands back a
  token.

## The user type and store

```go
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // never leaves the server
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
}

var ErrEmailTaken = errors.New("email already registered")

type UserStore interface {
	CreateUser(ctx context.Context, email, passwordHash string) (User, error)
	GetUser(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
}
```

The in-memory implementation follows the same pattern as the squeak store, with one
addition: a second map from email to user ID, so lookups by email are fast and
duplicates can be caught. Checking for a duplicate and inserting must happen under **one**
`Lock`, or two simultaneous sign-ups with the same email could both succeed.

`json:"-"` makes the encoder skip `PasswordHash` entirely. Even safer is a response type
without the field, so a future refactor can't expose it by accident:

```go
type userResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
```

## Registering

```go
func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if code, err := decodeJSON(w, r, &params); err != nil {
		respondWithError(w, code, err.Error())
		return
	}
	email := strings.ToLower(strings.TrimSpace(params.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid email address")
		return
	}
	if n := len(params.Password); n < 8 || n > 256 {
		respondWithError(w, http.StatusBadRequest, "password must be 8 to 256 bytes long")
		return
	}

	hash, err := hashPassword(params.Password)
	if err != nil {
		log.Printf("hashing password: %v", err)
		respondWithError(w, http.StatusInternalServerError, "couldn't create user")
		return
	}
	user, err := cfg.users.CreateUser(r.Context(), email, hash)
	if errors.Is(err, ErrEmailTaken) {
		respondWithError(w, http.StatusConflict, "email already registered")
		return
	}
	if err != nil {
		log.Printf("creating user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "couldn't create user")
		return
	}
	respondWithJSON(w, http.StatusCreated, userResponse{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt})
}
```

Things to notice:

- **Normalise the email** (trim, lowercase) before storing and before looking up, or
  `Pip@squeak.dev` and `pip@squeak.dev` become two accounts.
- **`net/mail.ParseAddress`** is a quick sanity check. The only real proof that an
  email address works is sending it a confirmation link.
- The anonymous struct for `params` keeps the request shape right next to the code that
  uses it.
- **409 Conflict** for a duplicate email.

## Logging in

```go
func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if code, err := decodeJSON(w, r, &params); err != nil {
		respondWithError(w, code, err.Error())
		return
	}
	email := strings.ToLower(strings.TrimSpace(params.Email))

	user, err := cfg.users.GetUserByEmail(r.Context(), email)
	if err != nil && !errors.Is(err, ErrNotFound) {
		log.Printf("looking up user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "couldn't log in")
		return
	}
	hash := user.PasswordHash
	if hash == "" {
		hash = dummyHash // still do the slow work, so timing doesn't reveal unknown emails
	}
	ok, err := checkPassword(params.Password, hash)
	if err != nil || !ok || user.ID == uuid.Nil() {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password")
		return
	}
	// Success: issue tokens (next lessons).
	respondWithJSON(w, http.StatusOK, userResponse{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt})
}
```

### One answer for every failure

"Unknown email" and "wrong password" get the **same** 401 and the **same** message. If
they differed, anyone could type in emails and learn who has a Squeak account.

The same leak can happen through **timing**: checking a password takes ~80 ms, and
"no such user" would take 0 ms. So when the user doesn't exist, the handler checks
against a `dummyHash` (the hash of some random password, computed at startup), and both
paths take the same time.

## What login returns

At the moment, a successful login proves who you are for exactly one request. The mouse
would have to send their password with every squeak. The next lesson looks at how servers
remember a login: sessions and tokens.
