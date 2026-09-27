---
title: Assemble Squeak
quiz:
  - question: |
      Squeak's router is built as `withRequestID(recoverPanics(mux))`. A request comes
      in for a URL that matches no route. Does the 404 response get an `X-Request-ID`?
    options:
      - text: No, because middleware only runs for registered routes
      - text: Yes, because the middleware wraps the whole mux, and the mux's own 404 is written inside it
        correct: true
      - text: Only if the request had an `Authorization` header
      - text: No, because `recoverPanics` stops the request first
    explanation: |
      Middleware wrapped around the mux runs for *every* request the server receives,
      matched or not. That's the place for things every response needs. Middleware
      wrapped around one handler (like `requireAuth`) only runs for that route.
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"context"
    	"crypto/hmac"
    	"crypto/rand"
    	"crypto/sha256"
    	"encoding/base64"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"log"
    	"maps"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    	"strings"
    	"sync"
    	"time"
    	"unicode/utf8"
    	"uuid"
    )

    // ======== storage (chapter 5) ========

    var ErrNotFound = errors.New("not found")

    type Squeak struct {
    	ID        uuid.UUID `json:"id"`
    	AuthorID  uuid.UUID `json:"author_id"`
    	Body      string    `json:"body"`
    	CreatedAt time.Time `json:"created_at"`
    }

    // SqueakStore is implemented by MemoryStore here, and by a SQLStore in production.
    type SqueakStore interface {
    	CreateSqueak(ctx context.Context, authorID uuid.UUID, body string) (Squeak, error)
    	GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error)
    	ListSqueaks(ctx context.Context, authorID uuid.UUID) ([]Squeak, error)
    	DeleteSqueak(ctx context.Context, id uuid.UUID) error
    }

    type MemoryStore struct {
    	mu      sync.RWMutex
    	squeaks map[uuid.UUID]Squeak
    }

    func NewMemoryStore() *MemoryStore {
    	return &MemoryStore{squeaks: make(map[uuid.UUID]Squeak)}
    }

    func (s *MemoryStore) CreateSqueak(ctx context.Context, authorID uuid.UUID, body string) (Squeak, error) {
    	sq := Squeak{ID: uuid.NewV7(), AuthorID: authorID, Body: body, CreatedAt: time.Now().UTC()}
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	s.squeaks[sq.ID] = sq
    	return sq, nil
    }

    func (s *MemoryStore) GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error) {
    	s.mu.RLock()
    	defer s.mu.RUnlock()
    	sq, ok := s.squeaks[id]
    	if !ok {
    		return Squeak{}, ErrNotFound
    	}
    	return sq, nil
    }

    func (s *MemoryStore) ListSqueaks(ctx context.Context, authorID uuid.UUID) ([]Squeak, error) {
    	s.mu.RLock()
    	list := slices.Collect(maps.Values(s.squeaks))
    	s.mu.RUnlock()
    	if authorID != uuid.Nil() {
    		list = slices.DeleteFunc(list, func(sq Squeak) bool { return sq.AuthorID != authorID })
    	}
    	slices.SortFunc(list, func(a, b Squeak) int {
    		return cmp.Or(a.CreatedAt.Compare(b.CreatedAt), a.ID.Compare(b.ID))
    	})
    	return list, nil
    }

    func (s *MemoryStore) DeleteSqueak(ctx context.Context, id uuid.UUID) error {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	if _, ok := s.squeaks[id]; !ok {
    		return ErrNotFound
    	}
    	delete(s.squeaks, id)
    	return nil
    }

    // ======== JSON helpers (chapter 4) ========

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, err := json.Marshal(payload)
    	if err != nil {
    		log.Printf("encoding response: %v", err)
    		w.WriteHeader(http.StatusInternalServerError)
    		return
    	}
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	respondWithJSON(w, code, map[string]string{"error": msg})
    }

    // ======== access tokens (chapter 6) ========

    var errInvalidToken = errors.New("invalid token")

    var b64 = base64.RawURLEncoding

    type claims struct {
    	Issuer    string `json:"iss"`
    	Subject   string `json:"sub"`
    	IssuedAt  int64  `json:"iat"`
    	ExpiresAt int64  `json:"exp"`
    }

    const tokenHeader = `{"alg":"HS256","typ":"JWT"}`

    func sign(secret []byte, msg string) []byte {
    	mac := hmac.New(sha256.New, secret)
    	mac.Write([]byte(msg))
    	return mac.Sum(nil)
    }

    func makeToken(userID uuid.UUID, secret []byte, now time.Time, ttl time.Duration) (string, error) {
    	payload, err := json.Marshal(claims{Issuer: "squeak", Subject: userID.String(), IssuedAt: now.Unix(), ExpiresAt: now.Add(ttl).Unix()})
    	if err != nil {
    		return "", err
    	}
    	unsigned := b64.EncodeToString([]byte(tokenHeader)) + "." + b64.EncodeToString(payload)
    	return unsigned + "." + b64.EncodeToString(sign(secret, unsigned)), nil
    }

    func validateToken(token string, secret []byte, now time.Time) (uuid.UUID, error) {
    	parts := strings.Split(token, ".")
    	if len(parts) != 3 {
    		return uuid.Nil(), errInvalidToken
    	}
    	header, err := b64.DecodeString(parts[0])
    	if err != nil || string(header) != tokenHeader {
    		return uuid.Nil(), errInvalidToken
    	}
    	sig, err := b64.DecodeString(parts[2])
    	if err != nil || !hmac.Equal(sig, sign(secret, parts[0]+"."+parts[1])) {
    		return uuid.Nil(), errInvalidToken
    	}
    	payload, err := b64.DecodeString(parts[1])
    	if err != nil {
    		return uuid.Nil(), errInvalidToken
    	}
    	var c claims
    	if err := json.Unmarshal(payload, &c); err != nil || c.Issuer != "squeak" || !now.Before(time.Unix(c.ExpiresAt, 0)) {
    		return uuid.Nil(), errInvalidToken
    	}
    	return uuid.Parse(c.Subject)
    }

    // ======== the app ========

    type apiConfig struct {
    	squeaks   SqueakStore
    	jwtSecret []byte
    }

    type userIDKey struct{}

    func userIDFrom(ctx context.Context) (uuid.UUID, bool) {
    	id, ok := ctx.Value(userIDKey{}).(uuid.UUID)
    	return id, ok
    }

    // requireAuth (chapter 6) lets a request through only with a valid bearer token.
    func (cfg *apiConfig) requireAuth(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
    		if !ok || !strings.EqualFold(scheme, "Bearer") {
    			w.Header().Set("WWW-Authenticate", `Bearer realm="squeak"`)
    			respondWithError(w, http.StatusUnauthorized, "missing bearer token")
    			return
    		}
    		userID, err := validateToken(strings.TrimSpace(token), cfg.jwtSecret, time.Now())
    		if err != nil {
    			w.Header().Set("WWW-Authenticate", `Bearer realm="squeak"`)
    			respondWithError(w, http.StatusUnauthorized, "invalid or expired token")
    			return
    		}
    		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey{}, userID)))
    	})
    }

    // withRequestID (chapter 3) gives every response an X-Request-ID header.
    func withRequestID(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.Header().Set("X-Request-ID", rand.Text())
    		next.ServeHTTP(w, r)
    	})
    }

    // recoverPanics (chapter 3) turns a panicking handler into a 500.
    func recoverPanics(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		defer func() {
    			v := recover()
    			if v == nil {
    				return
    			}
    			if v == http.ErrAbortHandler {
    				panic(v)
    			}
    			log.Printf("panic: %v", v)
    			respondWithError(w, http.StatusInternalServerError, "internal server error")
    		}()
    		next.ServeHTTP(w, r)
    	})
    }

    func handleHealthz(w http.ResponseWriter, r *http.Request) {
    	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    	w.Write([]byte("OK"))
    }

    // handleListSqueaks: GET /api/squeaks, or ?author=<uuid> for one mouse's squeaks.
    func (cfg *apiConfig) handleListSqueaks(w http.ResponseWriter, r *http.Request) {
    	authorID := uuid.Nil()
    	if s := r.URL.Query().Get("author"); s != "" {
    		id, err := uuid.Parse(s)
    		if err != nil {
    			respondWithError(w, http.StatusBadRequest, "invalid author id")
    			return
    		}
    		authorID = id
    	}
    	list, err := cfg.squeaks.ListSqueaks(r.Context(), authorID)
    	if err != nil {
    		log.Printf("listing squeaks: %v", err)
    		respondWithError(w, http.StatusInternalServerError, "couldn't list squeaks")
    		return
    	}
    	respondWithJSON(w, http.StatusOK, list)
    }

    func (cfg *apiConfig) handleGetSqueak(w http.ResponseWriter, r *http.Request) {
    	id, err := uuid.Parse(r.PathValue("id"))
    	if err != nil {
    		respondWithError(w, http.StatusBadRequest, "invalid squeak id")
    		return
    	}
    	sq, err := cfg.squeaks.GetSqueak(r.Context(), id)
    	if errors.Is(err, ErrNotFound) {
    		respondWithError(w, http.StatusNotFound, "squeak not found")
    		return
    	}
    	if err != nil {
    		log.Printf("getting squeak %s: %v", id, err)
    		respondWithError(w, http.StatusInternalServerError, "couldn't get squeak")
    		return
    	}
    	respondWithJSON(w, http.StatusOK, sq)
    }

    func (cfg *apiConfig) handleCreateSqueak(w http.ResponseWriter, r *http.Request) {
    	userID, ok := userIDFrom(r.Context())
    	if !ok {
    		respondWithError(w, http.StatusUnauthorized, "not logged in")
    		return
    	}
    	var params struct {
    		Body string `json:"body"`
    	}
    	r.Body = http.MaxBytesReader(w, r.Body, 1<<12)
    	if err := json.UnmarshalRead(r.Body, &params); err != nil {
    		respondWithError(w, http.StatusBadRequest, "invalid JSON body")
    		return
    	}
    	body := strings.TrimSpace(params.Body)
    	if body == "" || utf8.RuneCountInString(body) > 140 {
    		respondWithError(w, http.StatusBadRequest, "a squeak must be 1 to 140 characters")
    		return
    	}
    	sq, err := cfg.squeaks.CreateSqueak(r.Context(), userID, body)
    	if err != nil {
    		log.Printf("creating squeak: %v", err)
    		respondWithError(w, http.StatusInternalServerError, "couldn't create squeak")
    		return
    	}
    	respondWithJSON(w, http.StatusCreated, sq)
    }

    func (cfg *apiConfig) handleDeleteSqueak(w http.ResponseWriter, r *http.Request) {
    	userID, ok := userIDFrom(r.Context())
    	if !ok {
    		respondWithError(w, http.StatusUnauthorized, "not logged in")
    		return
    	}
    	id, err := uuid.Parse(r.PathValue("id"))
    	if err != nil {
    		respondWithError(w, http.StatusBadRequest, "invalid squeak id")
    		return
    	}
    	sq, err := cfg.squeaks.GetSqueak(r.Context(), id)
    	if errors.Is(err, ErrNotFound) {
    		respondWithError(w, http.StatusNotFound, "squeak not found")
    		return
    	}
    	if err != nil {
    		log.Printf("getting squeak %s: %v", id, err)
    		respondWithError(w, http.StatusInternalServerError, "couldn't delete squeak")
    		return
    	}
    	if sq.AuthorID != userID {
    		respondWithError(w, http.StatusForbidden, "you can only delete your own squeaks")
    		return
    	}
    	if err := cfg.squeaks.DeleteSqueak(r.Context(), id); err != nil {
    		if errors.Is(err, ErrNotFound) {
    			respondWithError(w, http.StatusNotFound, "squeak not found")
    			return
    		}
    		log.Printf("deleting squeak %s: %v", id, err)
    		respondWithError(w, http.StatusInternalServerError, "couldn't delete squeak")
    		return
    	}
    	w.WriteHeader(http.StatusNoContent)
    }

    func (cfg *apiConfig) handleMe(w http.ResponseWriter, r *http.Request) {
    	userID, ok := userIDFrom(r.Context())
    	if !ok {
    		respondWithError(w, http.StatusUnauthorized, "not logged in")
    		return
    	}
    	respondWithJSON(w, http.StatusOK, map[string]string{"id": userID.String()})
    }

    // ======== your part ========

    // newRouter assembles Squeak: every route, the auth middleware on the
    // protected ones, and request IDs and panic recovery around everything.
    func newRouter(cfg *apiConfig) http.Handler {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/healthz", handleHealthz)
    	// ?
    	return mux
    }

    func main() {
    	log.SetFlags(0)
    	cfg := &apiConfig{squeaks: NewMemoryStore(), jwtSecret: []byte("squeak-dev-secret-change-me-please")}
    	router := newRouter(cfg)

    	pip, whiskers := uuid.NewV7(), uuid.NewV7()
    	pipToken, _ := makeToken(pip, cfg.jwtSecret, time.Now(), time.Hour)
    	whiskersToken, _ := makeToken(whiskers, cfg.jwtSecret, time.Now(), time.Hour)

    	send := func(label, method, path, token, body string) *httptest.ResponseRecorder {
    		req := httptest.NewRequest(method, path, strings.NewReader(body))
    		if token != "" {
    			req.Header.Set("Authorization", "Bearer "+token)
    		}
    		rec := httptest.NewRecorder()
    		router.ServeHTTP(rec, req)
    		hasID := rec.Header().Get("X-Request-ID") != ""
    		fmt.Printf("%-30s -> %d (request ID: %v)\n", label, rec.Code, hasID)
    		return rec
    	}

    	send("health check", "GET", "/api/healthz", "", "")
    	send("anonymous squeak", "POST", "/api/squeaks", "", `{"body":"who am I?"}`)
    	rec := send("pip squeaks", "POST", "/api/squeaks", pipToken, `{"body":"squeak is assembled!"}`)
    	var created Squeak
    	json.Unmarshal(rec.Body.Bytes(), &created)
    	path := "/api/squeaks/" + created.ID.String()
    	send("list squeaks", "GET", "/api/squeaks", "", "")
    	send("get pip's squeak", "GET", path, "", "")
    	send("pip asks who they are", "GET", "/api/users/me", pipToken, "")
    	send("whiskers deletes pip's squeak", "DELETE", path, whiskersToken, "")
    	send("pip deletes pip's squeak", "DELETE", path, pipToken, "")
    }
  solution: |
    package main

    import (
    	"cmp"
    	"context"
    	"crypto/hmac"
    	"crypto/rand"
    	"crypto/sha256"
    	"encoding/base64"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"log"
    	"maps"
    	"net/http"
    	"net/http/httptest"
    	"slices"
    	"strings"
    	"sync"
    	"time"
    	"unicode/utf8"
    	"uuid"
    )

    // ======== storage (chapter 5) ========

    var ErrNotFound = errors.New("not found")

    type Squeak struct {
    	ID        uuid.UUID `json:"id"`
    	AuthorID  uuid.UUID `json:"author_id"`
    	Body      string    `json:"body"`
    	CreatedAt time.Time `json:"created_at"`
    }

    // SqueakStore is implemented by MemoryStore here, and by a SQLStore in production.
    type SqueakStore interface {
    	CreateSqueak(ctx context.Context, authorID uuid.UUID, body string) (Squeak, error)
    	GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error)
    	ListSqueaks(ctx context.Context, authorID uuid.UUID) ([]Squeak, error)
    	DeleteSqueak(ctx context.Context, id uuid.UUID) error
    }

    type MemoryStore struct {
    	mu      sync.RWMutex
    	squeaks map[uuid.UUID]Squeak
    }

    func NewMemoryStore() *MemoryStore {
    	return &MemoryStore{squeaks: make(map[uuid.UUID]Squeak)}
    }

    func (s *MemoryStore) CreateSqueak(ctx context.Context, authorID uuid.UUID, body string) (Squeak, error) {
    	sq := Squeak{ID: uuid.NewV7(), AuthorID: authorID, Body: body, CreatedAt: time.Now().UTC()}
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	s.squeaks[sq.ID] = sq
    	return sq, nil
    }

    func (s *MemoryStore) GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error) {
    	s.mu.RLock()
    	defer s.mu.RUnlock()
    	sq, ok := s.squeaks[id]
    	if !ok {
    		return Squeak{}, ErrNotFound
    	}
    	return sq, nil
    }

    func (s *MemoryStore) ListSqueaks(ctx context.Context, authorID uuid.UUID) ([]Squeak, error) {
    	s.mu.RLock()
    	list := slices.Collect(maps.Values(s.squeaks))
    	s.mu.RUnlock()
    	if authorID != uuid.Nil() {
    		list = slices.DeleteFunc(list, func(sq Squeak) bool { return sq.AuthorID != authorID })
    	}
    	slices.SortFunc(list, func(a, b Squeak) int {
    		return cmp.Or(a.CreatedAt.Compare(b.CreatedAt), a.ID.Compare(b.ID))
    	})
    	return list, nil
    }

    func (s *MemoryStore) DeleteSqueak(ctx context.Context, id uuid.UUID) error {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	if _, ok := s.squeaks[id]; !ok {
    		return ErrNotFound
    	}
    	delete(s.squeaks, id)
    	return nil
    }

    // ======== JSON helpers (chapter 4) ========

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, err := json.Marshal(payload)
    	if err != nil {
    		log.Printf("encoding response: %v", err)
    		w.WriteHeader(http.StatusInternalServerError)
    		return
    	}
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	respondWithJSON(w, code, map[string]string{"error": msg})
    }

    // ======== access tokens (chapter 6) ========

    var errInvalidToken = errors.New("invalid token")

    var b64 = base64.RawURLEncoding

    type claims struct {
    	Issuer    string `json:"iss"`
    	Subject   string `json:"sub"`
    	IssuedAt  int64  `json:"iat"`
    	ExpiresAt int64  `json:"exp"`
    }

    const tokenHeader = `{"alg":"HS256","typ":"JWT"}`

    func sign(secret []byte, msg string) []byte {
    	mac := hmac.New(sha256.New, secret)
    	mac.Write([]byte(msg))
    	return mac.Sum(nil)
    }

    func makeToken(userID uuid.UUID, secret []byte, now time.Time, ttl time.Duration) (string, error) {
    	payload, err := json.Marshal(claims{Issuer: "squeak", Subject: userID.String(), IssuedAt: now.Unix(), ExpiresAt: now.Add(ttl).Unix()})
    	if err != nil {
    		return "", err
    	}
    	unsigned := b64.EncodeToString([]byte(tokenHeader)) + "." + b64.EncodeToString(payload)
    	return unsigned + "." + b64.EncodeToString(sign(secret, unsigned)), nil
    }

    func validateToken(token string, secret []byte, now time.Time) (uuid.UUID, error) {
    	parts := strings.Split(token, ".")
    	if len(parts) != 3 {
    		return uuid.Nil(), errInvalidToken
    	}
    	header, err := b64.DecodeString(parts[0])
    	if err != nil || string(header) != tokenHeader {
    		return uuid.Nil(), errInvalidToken
    	}
    	sig, err := b64.DecodeString(parts[2])
    	if err != nil || !hmac.Equal(sig, sign(secret, parts[0]+"."+parts[1])) {
    		return uuid.Nil(), errInvalidToken
    	}
    	payload, err := b64.DecodeString(parts[1])
    	if err != nil {
    		return uuid.Nil(), errInvalidToken
    	}
    	var c claims
    	if err := json.Unmarshal(payload, &c); err != nil || c.Issuer != "squeak" || !now.Before(time.Unix(c.ExpiresAt, 0)) {
    		return uuid.Nil(), errInvalidToken
    	}
    	return uuid.Parse(c.Subject)
    }

    // ======== the app ========

    type apiConfig struct {
    	squeaks   SqueakStore
    	jwtSecret []byte
    }

    type userIDKey struct{}

    func userIDFrom(ctx context.Context) (uuid.UUID, bool) {
    	id, ok := ctx.Value(userIDKey{}).(uuid.UUID)
    	return id, ok
    }

    // requireAuth (chapter 6) lets a request through only with a valid bearer token.
    func (cfg *apiConfig) requireAuth(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
    		if !ok || !strings.EqualFold(scheme, "Bearer") {
    			w.Header().Set("WWW-Authenticate", `Bearer realm="squeak"`)
    			respondWithError(w, http.StatusUnauthorized, "missing bearer token")
    			return
    		}
    		userID, err := validateToken(strings.TrimSpace(token), cfg.jwtSecret, time.Now())
    		if err != nil {
    			w.Header().Set("WWW-Authenticate", `Bearer realm="squeak"`)
    			respondWithError(w, http.StatusUnauthorized, "invalid or expired token")
    			return
    		}
    		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey{}, userID)))
    	})
    }

    // withRequestID (chapter 3) gives every response an X-Request-ID header.
    func withRequestID(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		w.Header().Set("X-Request-ID", rand.Text())
    		next.ServeHTTP(w, r)
    	})
    }

    // recoverPanics (chapter 3) turns a panicking handler into a 500.
    func recoverPanics(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		defer func() {
    			v := recover()
    			if v == nil {
    				return
    			}
    			if v == http.ErrAbortHandler {
    				panic(v)
    			}
    			log.Printf("panic: %v", v)
    			respondWithError(w, http.StatusInternalServerError, "internal server error")
    		}()
    		next.ServeHTTP(w, r)
    	})
    }

    func handleHealthz(w http.ResponseWriter, r *http.Request) {
    	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    	w.Write([]byte("OK"))
    }

    // handleListSqueaks: GET /api/squeaks, or ?author=<uuid> for one mouse's squeaks.
    func (cfg *apiConfig) handleListSqueaks(w http.ResponseWriter, r *http.Request) {
    	authorID := uuid.Nil()
    	if s := r.URL.Query().Get("author"); s != "" {
    		id, err := uuid.Parse(s)
    		if err != nil {
    			respondWithError(w, http.StatusBadRequest, "invalid author id")
    			return
    		}
    		authorID = id
    	}
    	list, err := cfg.squeaks.ListSqueaks(r.Context(), authorID)
    	if err != nil {
    		log.Printf("listing squeaks: %v", err)
    		respondWithError(w, http.StatusInternalServerError, "couldn't list squeaks")
    		return
    	}
    	respondWithJSON(w, http.StatusOK, list)
    }

    func (cfg *apiConfig) handleGetSqueak(w http.ResponseWriter, r *http.Request) {
    	id, err := uuid.Parse(r.PathValue("id"))
    	if err != nil {
    		respondWithError(w, http.StatusBadRequest, "invalid squeak id")
    		return
    	}
    	sq, err := cfg.squeaks.GetSqueak(r.Context(), id)
    	if errors.Is(err, ErrNotFound) {
    		respondWithError(w, http.StatusNotFound, "squeak not found")
    		return
    	}
    	if err != nil {
    		log.Printf("getting squeak %s: %v", id, err)
    		respondWithError(w, http.StatusInternalServerError, "couldn't get squeak")
    		return
    	}
    	respondWithJSON(w, http.StatusOK, sq)
    }

    func (cfg *apiConfig) handleCreateSqueak(w http.ResponseWriter, r *http.Request) {
    	userID, ok := userIDFrom(r.Context())
    	if !ok {
    		respondWithError(w, http.StatusUnauthorized, "not logged in")
    		return
    	}
    	var params struct {
    		Body string `json:"body"`
    	}
    	r.Body = http.MaxBytesReader(w, r.Body, 1<<12)
    	if err := json.UnmarshalRead(r.Body, &params); err != nil {
    		respondWithError(w, http.StatusBadRequest, "invalid JSON body")
    		return
    	}
    	body := strings.TrimSpace(params.Body)
    	if body == "" || utf8.RuneCountInString(body) > 140 {
    		respondWithError(w, http.StatusBadRequest, "a squeak must be 1 to 140 characters")
    		return
    	}
    	sq, err := cfg.squeaks.CreateSqueak(r.Context(), userID, body)
    	if err != nil {
    		log.Printf("creating squeak: %v", err)
    		respondWithError(w, http.StatusInternalServerError, "couldn't create squeak")
    		return
    	}
    	respondWithJSON(w, http.StatusCreated, sq)
    }

    func (cfg *apiConfig) handleDeleteSqueak(w http.ResponseWriter, r *http.Request) {
    	userID, ok := userIDFrom(r.Context())
    	if !ok {
    		respondWithError(w, http.StatusUnauthorized, "not logged in")
    		return
    	}
    	id, err := uuid.Parse(r.PathValue("id"))
    	if err != nil {
    		respondWithError(w, http.StatusBadRequest, "invalid squeak id")
    		return
    	}
    	sq, err := cfg.squeaks.GetSqueak(r.Context(), id)
    	if errors.Is(err, ErrNotFound) {
    		respondWithError(w, http.StatusNotFound, "squeak not found")
    		return
    	}
    	if err != nil {
    		log.Printf("getting squeak %s: %v", id, err)
    		respondWithError(w, http.StatusInternalServerError, "couldn't delete squeak")
    		return
    	}
    	if sq.AuthorID != userID {
    		respondWithError(w, http.StatusForbidden, "you can only delete your own squeaks")
    		return
    	}
    	if err := cfg.squeaks.DeleteSqueak(r.Context(), id); err != nil {
    		if errors.Is(err, ErrNotFound) {
    			respondWithError(w, http.StatusNotFound, "squeak not found")
    			return
    		}
    		log.Printf("deleting squeak %s: %v", id, err)
    		respondWithError(w, http.StatusInternalServerError, "couldn't delete squeak")
    		return
    	}
    	w.WriteHeader(http.StatusNoContent)
    }

    func (cfg *apiConfig) handleMe(w http.ResponseWriter, r *http.Request) {
    	userID, ok := userIDFrom(r.Context())
    	if !ok {
    		respondWithError(w, http.StatusUnauthorized, "not logged in")
    		return
    	}
    	respondWithJSON(w, http.StatusOK, map[string]string{"id": userID.String()})
    }

    // ======== your part ========

    // newRouter assembles Squeak: every route, the auth middleware on the
    // protected ones, and request IDs and panic recovery around everything.
    func newRouter(cfg *apiConfig) http.Handler {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/healthz", handleHealthz)
    	mux.HandleFunc("GET /api/squeaks", cfg.handleListSqueaks)
    	mux.HandleFunc("GET /api/squeaks/{id}", cfg.handleGetSqueak)

    	auth := cfg.requireAuth
    	mux.Handle("POST /api/squeaks", auth(http.HandlerFunc(cfg.handleCreateSqueak)))
    	mux.Handle("DELETE /api/squeaks/{id}", auth(http.HandlerFunc(cfg.handleDeleteSqueak)))
    	mux.Handle("GET /api/users/me", auth(http.HandlerFunc(cfg.handleMe)))

    	return withRequestID(recoverPanics(mux))
    }

    func main() {
    	log.SetFlags(0)
    	cfg := &apiConfig{squeaks: NewMemoryStore(), jwtSecret: []byte("squeak-dev-secret-change-me-please")}
    	router := newRouter(cfg)

    	pip, whiskers := uuid.NewV7(), uuid.NewV7()
    	pipToken, _ := makeToken(pip, cfg.jwtSecret, time.Now(), time.Hour)
    	whiskersToken, _ := makeToken(whiskers, cfg.jwtSecret, time.Now(), time.Hour)

    	send := func(label, method, path, token, body string) *httptest.ResponseRecorder {
    		req := httptest.NewRequest(method, path, strings.NewReader(body))
    		if token != "" {
    			req.Header.Set("Authorization", "Bearer "+token)
    		}
    		rec := httptest.NewRecorder()
    		router.ServeHTTP(rec, req)
    		hasID := rec.Header().Get("X-Request-ID") != ""
    		fmt.Printf("%-30s -> %d (request ID: %v)\n", label, rec.Code, hasID)
    		return rec
    	}

    	send("health check", "GET", "/api/healthz", "", "")
    	send("anonymous squeak", "POST", "/api/squeaks", "", `{"body":"who am I?"}`)
    	rec := send("pip squeaks", "POST", "/api/squeaks", pipToken, `{"body":"squeak is assembled!"}`)
    	var created Squeak
    	json.Unmarshal(rec.Body.Bytes(), &created)
    	path := "/api/squeaks/" + created.ID.String()
    	send("list squeaks", "GET", "/api/squeaks", "", "")
    	send("get pip's squeak", "GET", path, "", "")
    	send("pip asks who they are", "GET", "/api/users/me", pipToken, "")
    	send("whiskers deletes pip's squeak", "DELETE", path, whiskersToken, "")
    	send("pip deletes pip's squeak", "DELETE", path, pipToken, "")
    }
  tests: |
    package main

    import (
    	"bytes"
    	"context"
    	"encoding/json/v2"
    	"io"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"os"
    	"strings"
    	"testing"
    	"time"
    	"uuid"
    )

    var testSecret = []byte("test-secret-for-the-assembled-squeak")

    type app struct {
    	t      *testing.T
    	store  *MemoryStore
    	client *http.Client
    	pip    string // pip's token
    	whisk  string // whiskers' token
    	pipID  uuid.UUID
    }

    func newApp(t *testing.T, store SqueakStore) *app {
    	t.Helper()
    	mem, _ := store.(*MemoryStore)
    	cfg := &apiConfig{squeaks: store, jwtSecret: testSecret}
    	srv := httptest.NewTestServer(t, newRouter(cfg))
    	pipID, whiskID := uuid.NewV7(), uuid.NewV7()
    	pip, _ := makeToken(pipID, testSecret, time.Now(), time.Hour)
    	whisk, _ := makeToken(whiskID, testSecret, time.Now(), time.Hour)
    	return &app{t: t, store: mem, client: srv.Client(), pip: pip, whisk: whisk, pipID: pipID}
    }

    func (a *app) do(method, path, token, body string) (*http.Response, string) {
    	a.t.Helper()
    	req, err := http.NewRequestWithContext(a.t.Context(), method, "http://squeak.test"+path, strings.NewReader(body))
    	if err != nil {
    		a.t.Fatal(err)
    	}
    	if token != "" {
    		req.Header.Set("Authorization", "Bearer "+token)
    	}
    	resp, err := a.client.Do(req)
    	if err != nil {
    		a.t.Fatalf("%s %s: %v", method, path, err)
    	}
    	defer resp.Body.Close()
    	data, _ := io.ReadAll(resp.Body)
    	return resp, string(data)
    }

    func TestRoutes(t *testing.T) {
    	a := newApp(t, NewMemoryStore())
    	existing, _ := a.store.CreateSqueak(t.Context(), a.pipID, "pip was here")
    	path := "/api/squeaks/" + existing.ID.String()

    	for _, tt := range []struct {
    		name, method, path, token, body string
    		want                            int
    	}{
    		{"health check is public", "GET", "/api/healthz", "", "", 200},
    		{"listing is public", "GET", "/api/squeaks", "", "", 200},
    		{"getting one squeak is public", "GET", path, "", "", 200},
    		{"unknown squeak", "GET", "/api/squeaks/" + uuid.NewV7().String(), "", "", 404},
    		{"posting needs a token", "POST", "/api/squeaks", "", `{"body":"sneaky"}`, 401},
    		{"posting with a forged token", "POST", "/api/squeaks", "not.a.token", `{"body":"sneaky"}`, 401},
    		{"posting with a token", "POST", "/api/squeaks", a.pip, `{"body":"hello"}`, 201},
    		{"/api/users/me needs a token", "GET", "/api/users/me", "", "", 401},
    		{"/api/users/me with a token", "GET", "/api/users/me", a.pip, "", 200},
    		{"deleting needs a token", "DELETE", path, "", "", 401},
    		{"whiskers can't delete pip's squeak", "DELETE", path, a.whisk, "", 403},
    		{"pip deletes pip's squeak", "DELETE", path, a.pip, "", 204},
    		{"wrong method", "PUT", "/api/squeaks", a.pip, "", 405},
    	} {
    		t.Run(tt.name, func(t *testing.T) {
    			resp, body := a.do(tt.method, tt.path, tt.token, tt.body)
    			if resp.StatusCode != tt.want {
    				t.Errorf("%s %s: status %d, want %d (body %q)", tt.method, tt.path, resp.StatusCode, tt.want, body)
    			}
    		})
    	}
    }

    func TestUnauthenticatedPostStoresNothing(t *testing.T) {
    	a := newApp(t, NewMemoryStore())
    	a.do("POST", "/api/squeaks", "", `{"body":"sneaky"}`)
    	if list, _ := a.store.ListSqueaks(t.Context(), uuid.Nil()); len(list) != 0 {
    		t.Errorf("a POST without a token stored %d squeak(s); wrap the handler in cfg.requireAuth", len(list))
    	}
    }

    func TestCreatedSqueakBelongsToCaller(t *testing.T) {
    	a := newApp(t, NewMemoryStore())
    	_, body := a.do("POST", "/api/squeaks", a.pip, `{"body":"mine"}`)
    	var sq Squeak
    	if err := json.Unmarshal([]byte(body), &sq); err != nil {
    		t.Fatalf("POST /api/squeaks body %q isn't a squeak: %v", body, err)
    	}
    	if sq.AuthorID != a.pipID {
    		t.Errorf("new squeak's author = %s, want the caller %s", sq.AuthorID, a.pipID)
    	}
    }

    func TestEveryResponseHasRequestID(t *testing.T) {
    	a := newApp(t, NewMemoryStore())
    	for _, r := range []struct{ method, path, token string }{
    		{"GET", "/api/healthz", ""},
    		{"GET", "/api/squeaks", ""},
    		{"POST", "/api/squeaks", ""},
    		{"GET", "/api/users/me", a.pip},
    		{"GET", "/no/such/route", ""},
    	} {
    		resp, _ := a.do(r.method, r.path, r.token, "")
    		if resp.Header.Get("X-Request-ID") == "" {
    			t.Errorf("%s %s (status %d) has no X-Request-ID header; wrap the whole mux in withRequestID", r.method, r.path, resp.StatusCode)
    		}
    	}
    }

    // panicStore blows up on every call, like a bug deep in a real store.
    type panicStore struct{ SqueakStore }

    func (panicStore) ListSqueaks(ctx context.Context, authorID uuid.UUID) ([]Squeak, error) {
    	panic("nil pointer somewhere in the store")
    }

    func TestPanicsAreRecovered(t *testing.T) {
    	var logs bytes.Buffer
    	log.SetOutput(&logs)
    	t.Cleanup(func() { log.SetOutput(os.Stderr) })

    	a := newApp(t, panicStore{})
    	resp, body := a.do("GET", "/api/squeaks", "", "")
    	if resp.StatusCode != http.StatusInternalServerError {
    		t.Fatalf("GET /api/squeaks with a panicking store: status %d, want 500 (wrap the mux in recoverPanics)", resp.StatusCode)
    	}
    	if strings.Contains(body, "nil pointer") {
    		t.Errorf("500 body %q leaks the panic message", body)
    	}
    	if resp, _ := a.do("GET", "/api/healthz", "", ""); resp.StatusCode != 200 {
    		t.Errorf("after a panic, GET /api/healthz = %d, want 200", resp.StatusCode)
    	}
    }
---

Nine courses ago you printed your first line of Go. Now you're going to wire up a
complete, tested JSON API from parts you built yourself.

## The pieces

Everything in the editor comes from an earlier chapter:

| Piece | From |
|---|---|
| `Squeak`, `SqueakStore`, `MemoryStore` | chapter 5, storage |
| `respondWithJSON`, `respondWithError` | chapter 4, JSON APIs |
| `makeToken`, `validateToken`, `requireAuth`, `userIDFrom` | chapter 6, authentication |
| `withRequestID`, `recoverPanics` | chapter 3, middleware |
| `handleCreateSqueak`, `handleGetSqueak`, `handleListSqueaks` | chapters 4 and 5 |
| `handleDeleteSqueak` with its ownership check | chapter 7, authorization |
| `handleHealthz`, `handleMe` | chapters 1 and 6 |

In a real repository these would live in separate files, maybe separate packages,
with a `main.go` that does little more than this:

```go
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig(os.LookupEnv) // configuration
	if err != nil {
		log.Fatal(err)
	}
	db, err := openDatabase(ctx, cfg.DatabaseURL) // sql.Open + PingContext + migrations
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := &apiConfig{squeaks: &SQLStore{db: db}, jwtSecret: cfg.JWTSecret}
	srv := &http.Server{Addr: cfg.Addr, Handler: newRouter(app), ReadHeaderTimeout: 5 * time.Second}
	if err := serve(ctx, srv); err != nil { // graceful shutdown
		log.Fatal(err)
	}
}
```

Every line is a chapter of this course. The one piece still missing is in the middle:
`newRouter`, which decides which URL goes where and which middleware guards it.

## Your task

Finish `newRouter(cfg)` so that it registers:

| Route | Handler | Who can call it |
|---|---|---|
| `GET /api/healthz` | `handleHealthz` | anyone |
| `GET /api/squeaks` | `cfg.handleListSqueaks` | anyone |
| `GET /api/squeaks/{id}` | `cfg.handleGetSqueak` | anyone |
| `POST /api/squeaks` | `cfg.handleCreateSqueak` | logged in (`cfg.requireAuth`) |
| `DELETE /api/squeaks/{id}` | `cfg.handleDeleteSqueak` | logged in (`cfg.requireAuth`) |
| `GET /api/users/me` | `cfg.handleMe` | logged in (`cfg.requireAuth`) |

and then wraps the **whole mux** in `recoverPanics` and `withRequestID`, so every
response, even a 404 for an unknown URL or a 500 from a panic, carries an
`X-Request-ID` header.

`requireAuth` takes an `http.Handler`, so convert a method value with
`http.HandlerFunc(cfg.handleCreateSqueak)` and register the result with `mux.Handle`.

The tests start your router with `httptest.NewTestServer` and play through a whole
session: public reads, rejected anonymous and forged posts, Pip squeaking, Whiskers
failing to delete Pip's squeak, and a store that panics mid-request.

## What you built

Look back at what Squeak does now:

- **Routing** with method patterns and wildcards, and exact knowledge of which pattern wins.
- **Middleware** for request IDs, logging, metrics and panic recovery, composed in a
  deliberate order.
- **JSON** in and out, with size limits, validation and consistent error bodies.
- **Storage** behind an interface: a concurrency-safe in-memory store, and the SQL to
  back it with a real database.
- **Authentication** with properly hashed passwords, signed access tokens and revocable
  refresh tokens. **Authorization** that stops IDORs. **Webhooks** that are verified and
  idempotent.
- **Tests** at every level, including time-dependent behaviour tested in milliseconds.
- **Production habits**: graceful shutdown, structured logs, validated configuration
  and health checks.

## The end of the core path

This is the last lesson of the last course on the core path. Across the nine courses you've
gone from variables and loops, through interfaces, closures, generics, algorithms,
data structures, goroutines and channels, testing, and both sides of HTTP. That's the
toolkit working Go developers use every day.

Where to go from here is up to you. Some real next steps:

1. **Take the next courses.** Two courses go beyond the core:
   [Learn Generics and Advanced Types](/courses/learn-advanced-types) goes deep on Go's
   type system, interfaces and reflection, and
   [Learn Cryptography](/courses/learn-cryptography) explains the hashing, signing and
   encryption you've been using in Squeak's authentication and webhooks.
2. **Ship Squeak for real.** Create a module on your machine, split the code into
   files, add `modernc.org/sqlite` or `pgx` with a `SQLStore` and migrations, and
   deploy it somewhere that gives you a URL. Nothing teaches like a server you have
   to keep running.
3. **Grow it.** Likes (a join table and a `UNIQUE` constraint), follows and a
   personalised timeline (joins, indexes, cursor pagination), editing (`PATCH`,
   `edited_at`), per-user rate limits on login, and an OpenAPI description of the API.
4. **Build something of your own.** A CLI tool, a bot, a small game server, or an API
   for a hobby. Pick something you'd actually use, because you'll finish it.
5. **Read good Go.** The standard library is written in clear, idiomatic Go. Start with
   `net/http`'s `ServeMux`: you know its routing rules, so its source is a great first read.
6. **Keep up with the language.** Read the release notes when a new Go version comes
   out every six months, and the [Go blog](https://go.dev/blog/) for the reasoning
   behind new features.
7. **Contribute.** Fix a documentation typo or a small bug in an open-source Go project
   you use. Code review from experienced maintainers is the fastest way to level up.

Revisit any of the earlier courses when a topic comes up for real. It will click
differently now that you've built something with it, whether that's
[concurrency](/courses/learn-concurrency), [testing](/courses/learn-testing), or
[HTTP clients](/courses/learn-http-clients) for the other side of every request Squeak
answers.

Congratulations, and happy squeaking.
