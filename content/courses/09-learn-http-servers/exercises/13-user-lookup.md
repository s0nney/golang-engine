---
title: User Lookup
difficulty: medium
after: sql-and-databases
hints:
  - 'The handle is a value, so it travels as an argument: `s.db.QueryRowContext(ctx, "SELECT id, handle, display_name, bio, created_at FROM users WHERE handle = ?", strings.ToLower(handle))`. Nothing from the request is ever pasted into the SQL text.'
  - '`bio` can be `NULL`, and scanning `NULL` into a `*string` is an error. Scan it into a `sql.NullString` and copy `.String` (which is `""` when the column is `NULL`) into the user.'
  - '`QueryRowContext` reports "no row" from `Scan` as `sql.ErrNoRows`. Turn that into `ErrUserNotFound`, and wrap anything else with `%w`. In the handler, `errors.Is(err, ErrUserNotFound)` is a 404. Anything else gets logged with `cfg.logger.Error(..., "error", err)` and answered with a fixed 500 message.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"database/sql"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"log/slog"
    	"net/http"
    	"net/http/httptest"
    	"os"
    	"time"
    	"uuid"
    )

    type User struct {
    	ID          uuid.UUID `json:"id"`
    	Handle      string    `json:"handle"`
    	DisplayName string    `json:"display_name"`
    	Bio         string    `json:"bio"`
    	CreatedAt   time.Time `json:"created_at"`
    }

    var ErrUserNotFound = errors.New("user not found")

    // RowScanner is what *sql.Row gives you.
    type RowScanner interface {
    	Scan(dest ...any) error
    }

    // DB is the one method of *sql.DB this store needs. (A one-line adapter
    // makes *sql.DB fit; the tests use a fake.)
    type DB interface {
    	QueryRowContext(ctx context.Context, query string, args ...any) RowScanner
    }

    type UserStore struct {
    	db DB
    }

    func (s *UserStore) GetUserByHandle(ctx context.Context, handle string) (User, error) {
    	return User{}, nil
    }

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, err := json.Marshal(payload)
    	if err != nil {
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

    type apiConfig struct {
    	users  *UserStore
    	logger *slog.Logger
    }

    func (cfg *apiConfig) handleGetUser(w http.ResponseWriter, r *http.Request) {
    	respondWithError(w, http.StatusNotImplemented, "not implemented")
    }

    // demoDB knows one user, pip, whose bio is NULL.
    type demoDB struct{}

    type demoRow struct{ args []any }

    func (demoDB) QueryRowContext(ctx context.Context, query string, args ...any) RowScanner {
    	fmt.Printf("query: %s  args: %v\n", query, args)
    	return demoRow{args}
    }

    func (r demoRow) Scan(dest ...any) error {
    	if len(r.args) != 1 || r.args[0] != "pip" {
    		return sql.ErrNoRows
    	}
    	if len(dest) != 5 {
    		return fmt.Errorf("demo db: the query returns 5 columns, Scan got %d", len(dest))
    	}
    	bio, ok := dest[3].(*sql.NullString)
    	if !ok {
    		return fmt.Errorf("demo db: converting NULL to %T is unsupported", dest[3])
    	}
    	*bio = sql.NullString{}
    	*dest[0].(*uuid.UUID) = uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	*dest[1].(*string) = "pip"
    	*dest[2].(*string) = "Pip Squeakson"
    	*dest[4].(*time.Time) = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
    	return nil
    }

    func main() {
    	cfg := &apiConfig{users: &UserStore{db: demoDB{}}, logger: slog.New(slog.NewJSONHandler(os.Stdout, nil))}
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/users/{handle}", cfg.handleGetUser)
    	for _, handle := range []string{"Pip", "nobody"} {
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/users/"+handle, nil))
    		fmt.Println(rec.Code, rec.Body.String())
    	}
    }
  solution: |
    package main

    import (
    	"context"
    	"database/sql"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"log/slog"
    	"net/http"
    	"net/http/httptest"
    	"os"
    	"strings"
    	"time"
    	"uuid"
    )

    type User struct {
    	ID          uuid.UUID `json:"id"`
    	Handle      string    `json:"handle"`
    	DisplayName string    `json:"display_name"`
    	Bio         string    `json:"bio"`
    	CreatedAt   time.Time `json:"created_at"`
    }

    var ErrUserNotFound = errors.New("user not found")

    type RowScanner interface {
    	Scan(dest ...any) error
    }

    type DB interface {
    	QueryRowContext(ctx context.Context, query string, args ...any) RowScanner
    }

    type UserStore struct {
    	db DB
    }

    const selectUserByHandle = "SELECT id, handle, display_name, bio, created_at FROM users WHERE handle = ?"

    func (s *UserStore) GetUserByHandle(ctx context.Context, handle string) (User, error) {
    	var (
    		u   User
    		bio sql.NullString
    	)
    	row := s.db.QueryRowContext(ctx, selectUserByHandle, strings.ToLower(handle))
    	err := row.Scan(&u.ID, &u.Handle, &u.DisplayName, &bio, &u.CreatedAt)
    	if errors.Is(err, sql.ErrNoRows) {
    		return User{}, ErrUserNotFound
    	}
    	if err != nil {
    		return User{}, fmt.Errorf("getting user by handle: %w", err)
    	}
    	u.Bio = bio.String
    	return u, nil
    }

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, err := json.Marshal(payload)
    	if err != nil {
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

    type apiConfig struct {
    	users  *UserStore
    	logger *slog.Logger
    }

    func (cfg *apiConfig) handleGetUser(w http.ResponseWriter, r *http.Request) {
    	u, err := cfg.users.GetUserByHandle(r.Context(), r.PathValue("handle"))
    	if errors.Is(err, ErrUserNotFound) {
    		respondWithError(w, http.StatusNotFound, "user not found")
    		return
    	}
    	if err != nil {
    		cfg.logger.ErrorContext(r.Context(), "loading user", "error", err)
    		respondWithError(w, http.StatusInternalServerError, "couldn't load user")
    		return
    	}
    	respondWithJSON(w, http.StatusOK, u)
    }

    type demoDB struct{}

    type demoRow struct{ args []any }

    func (demoDB) QueryRowContext(ctx context.Context, query string, args ...any) RowScanner {
    	fmt.Printf("query: %s  args: %v\n", query, args)
    	return demoRow{args}
    }

    func (r demoRow) Scan(dest ...any) error {
    	if len(r.args) != 1 || r.args[0] != "pip" {
    		return sql.ErrNoRows
    	}
    	if len(dest) != 5 {
    		return fmt.Errorf("demo db: the query returns 5 columns, Scan got %d", len(dest))
    	}
    	bio, ok := dest[3].(*sql.NullString)
    	if !ok {
    		return fmt.Errorf("demo db: converting NULL to %T is unsupported", dest[3])
    	}
    	*bio = sql.NullString{}
    	*dest[0].(*uuid.UUID) = uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	*dest[1].(*string) = "pip"
    	*dest[2].(*string) = "Pip Squeakson"
    	*dest[4].(*time.Time) = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
    	return nil
    }

    func main() {
    	cfg := &apiConfig{users: &UserStore{db: demoDB{}}, logger: slog.New(slog.NewJSONHandler(os.Stdout, nil))}
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/users/{handle}", cfg.handleGetUser)
    	for _, handle := range []string{"Pip", "nobody"} {
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/users/"+handle, nil))
    		fmt.Println(rec.Code, rec.Body.String())
    	}
    }
  tests: |
    package main

    import (
    	"bytes"
    	"context"
    	"database/sql"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"log/slog"
    	"net/http"
    	"net/http/httptest"
    	"net/url"
    	"strings"
    	"testing"
    	"time"
    	"uuid"
    )

    var (
    	pipID   = uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	breeID  = uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-000000000b1e")
    	joined  = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
    	errConn = errors.New("dial tcp 10.0.3.7:5432: connection refused")
    )

    // fakeDB holds rows of users by handle, as the real table would:
    // id, handle, display_name, bio (nil for NULL), created_at.
    type fakeDB struct {
    	rows    map[string][]any
    	err     error
    	queries []string
    	args    [][]any
    }

    type fakeRow struct {
    	values []any
    	err    error
    }

    func (db *fakeDB) QueryRowContext(ctx context.Context, query string, args ...any) RowScanner {
    	db.queries = append(db.queries, query)
    	db.args = append(db.args, args)
    	if db.err != nil {
    		return fakeRow{err: db.err}
    	}
    	if !strings.Contains(query, "?") || len(args) != 1 {
    		return fakeRow{err: fmt.Errorf("fake db: expected one ? placeholder and one argument, got %q with %d args", query, len(args))}
    	}
    	handle, _ := args[0].(string)
    	values, ok := db.rows[handle]
    	if !ok {
    		return fakeRow{err: sql.ErrNoRows}
    	}
    	return fakeRow{values: values}
    }

    func (r fakeRow) Scan(dest ...any) error {
    	if r.err != nil {
    		return r.err
    	}
    	if len(dest) != len(r.values) {
    		return fmt.Errorf("fake db: Scan got %d destinations, the query returns %d columns", len(dest), len(r.values))
    	}
    	for i, v := range r.values {
    		switch d := dest[i].(type) {
    		case *uuid.UUID:
    			id, ok := v.(uuid.UUID)
    			if !ok {
    				return fmt.Errorf("fake db: column %d: can't scan %T into *uuid.UUID", i, v)
    			}
    			*d = id
    		case *string:
    			s, ok := v.(string)
    			if !ok {
    				return fmt.Errorf("fake db: column %d: converting %v (%T) to string is unsupported", i, v, v)
    			}
    			*d = s
    		case *sql.NullString:
    			if v == nil {
    				*d = sql.NullString{}
    			} else if s, ok := v.(string); ok {
    				*d = sql.NullString{String: s, Valid: true}
    			} else {
    				return fmt.Errorf("fake db: column %d: can't scan %T into *sql.NullString", i, v)
    			}
    		case *time.Time:
    			tm, ok := v.(time.Time)
    			if !ok {
    				return fmt.Errorf("fake db: column %d: can't scan %T into *time.Time", i, v)
    			}
    			*d = tm
    		default:
    			return fmt.Errorf("fake db: column %d: unsupported destination %T", i, dest[i])
    		}
    	}
    	return nil
    }

    func newFake() *fakeDB {
    	return &fakeDB{rows: map[string][]any{
    		"pip":  {pipID, "pip", "Pip Squeakson", nil, joined},
    		"bree": {breeID, "bree", "Bree", "cheese critic", joined.Add(time.Hour)},
    	}}
    }

    func TestGetUserByHandle(t *testing.T) {
    	db := newFake()
    	s := &UserStore{db: db}
    	u, err := s.GetUserByHandle(t.Context(), "bree")
    	want := User{ID: breeID, Handle: "bree", DisplayName: "Bree", Bio: "cheese critic", CreatedAt: joined.Add(time.Hour)}
    	if err != nil || u != want {
    		t.Errorf("GetUserByHandle(\"bree\") = %+v, %v, want %+v, nil", u, err, want)
    	}
    	u, err = s.GetUserByHandle(t.Context(), "PiP")
    	want = User{ID: pipID, Handle: "pip", DisplayName: "Pip Squeakson", Bio: "", CreatedAt: joined}
    	if err != nil || u != want {
    		t.Errorf("GetUserByHandle(\"PiP\") = %+v, %v, want %+v, nil (handles are case-insensitive, and a NULL bio is \"\")", u, err, want)
    	}
    	if len(db.args) > 0 && len(db.args[len(db.args)-1]) == 1 && db.args[len(db.args)-1][0] != "pip" {
    		t.Errorf("GetUserByHandle(\"PiP\") passed %v as the argument, want \"pip\" (lowercase it)", db.args[len(db.args)-1])
    	}
    }

    func TestNotFoundAndErrors(t *testing.T) {
    	db := newFake()
    	s := &UserStore{db: db}
    	if _, err := s.GetUserByHandle(t.Context(), "nobody"); !errors.Is(err, ErrUserNotFound) {
    		t.Errorf("unknown handle: err = %v, want ErrUserNotFound", err)
    	}
    	db.err = errConn
    	_, err := s.GetUserByHandle(t.Context(), "pip")
    	if !errors.Is(err, errConn) {
    		t.Errorf("database failure: err = %v, want an error wrapping the database's error (use %%w)", err)
    	}
    	if errors.Is(err, ErrUserNotFound) || errors.Is(err, sql.ErrNoRows) {
    		t.Errorf("database failure reported as not found: %v", err)
    	}
    }

    func TestPlaceholders(t *testing.T) {
    	db := newFake()
    	s := &UserStore{db: db}
    	evil := "x' OR '1'='1"
    	if _, err := s.GetUserByHandle(t.Context(), evil); !errors.Is(err, ErrUserNotFound) {
    		t.Errorf("GetUserByHandle(%q) = %v, want ErrUserNotFound", evil, err)
    	}
    	for i, q := range db.queries {
    		if strings.Contains(q, "OR '1'") || strings.Contains(strings.ToLower(q), "x'") {
    			t.Errorf("query %d = %q contains the handle: pass it as an argument, never in the SQL text", i, q)
    		}
    		if !strings.Contains(q, "FROM users") {
    			t.Errorf("query %d = %q, want a SELECT ... FROM users", i, q)
    		}
    	}
    }

    func serve(cfg *apiConfig, handle string) *httptest.ResponseRecorder {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/users/{handle}", cfg.handleGetUser)
    	rec := httptest.NewRecorder()
    	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/users/"+url.PathEscape(handle), nil))
    	return rec
    }

    func TestHandler(t *testing.T) {
    	var logs bytes.Buffer
    	db := newFake()
    	cfg := &apiConfig{users: &UserStore{db: db}, logger: slog.New(slog.NewJSONHandler(&logs, nil))}

    	rec := serve(cfg, "Bree")
    	if rec.Code != 200 || rec.Header().Get("Content-Type") != "application/json" {
    		t.Fatalf("GET /api/users/Bree: %d %q %s, want 200 application/json", rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
    	}
    	var got map[string]any
    	json.Unmarshal(rec.Body.Bytes(), &got)
    	if got["handle"] != "bree" || got["display_name"] != "Bree" || got["bio"] != "cheese critic" || got["id"] != breeID.String() {
    		t.Errorf("GET /api/users/Bree body = %s, want bree's profile", rec.Body.String())
    	}

    	rec = serve(cfg, "nobody")
    	var e map[string]string
    	if rec.Code != 404 || json.Unmarshal(rec.Body.Bytes(), &e) != nil || e["error"] == "" {
    		t.Errorf("GET /api/users/nobody: %d %s, want 404 with a JSON error", rec.Code, rec.Body.String())
    	}
    	if logs.Len() != 0 {
    		t.Errorf("a 404 was logged: %s(an unknown user is routine, not an error)", logs.String())
    	}

    	db.err = errConn
    	rec = serve(cfg, "pip")
    	if rec.Code != 500 || json.Unmarshal(rec.Body.Bytes(), &e) != nil || e["error"] == "" {
    		t.Errorf("database failure: %d %s, want 500 with a JSON error", rec.Code, rec.Body.String())
    	}
    	if strings.Contains(rec.Body.String(), "10.0.3.7") || strings.Contains(rec.Body.String(), "dial tcp") {
    		t.Errorf("database failure: response %s leaks the database error", rec.Body.String())
    	}
    	var entry map[string]any
    	line, _, _ := strings.Cut(logs.String(), "\n")
    	if err := json.Unmarshal([]byte(line), &entry); err != nil || entry["level"] != "ERROR" || !strings.Contains(fmt.Sprint(entry["error"]), "10.0.3.7") {
    		t.Errorf("database failure: log = %q, want one ERROR entry with the database error in an \"error\" attribute", logs.String())
    	}
    }
---

Squeak's users now live in a SQL `users` table:

| column | type |
|---|---|
| `id` | UUID |
| `handle` | text, stored lowercase, unique |
| `display_name` | text |
| `bio` | text, **nullable** |
| `created_at` | timestamp |

Profile pages call `GET /api/users/{handle}`. Implement both layers:

**`(s *UserStore) GetUserByHandle(ctx, handle)`** runs one query through
`s.db.QueryRowContext` that selects `id, handle, display_name, bio, created_at`
(in that order) `FROM users` where the handle matches:

- Handles are case-insensitive: look up `strings.ToLower(handle)`.
- The handle must be passed as a **`?` placeholder argument**, never pasted into
  the SQL.
- A `NULL` bio becomes `""`.
- No such row (`sql.ErrNoRows`): return `ErrUserNotFound`. Any other error:
  return it wrapped, so `errors.Is` still finds the original.

**`cfg.handleGetUser`** answers:

- `200` with the user as JSON on success.
- `404` with a JSON error for `ErrUserNotFound`. Don't log it: unknown users are
  routine.
- `500` with a fixed JSON error for anything else, after logging the real error
  with `cfg.logger` at **error** level, as an `"error"` attribute.

## Example

```
GET /api/users/Pip
query: SELECT id, handle, display_name, bio, created_at FROM users WHERE handle = ?  args: [pip]
200 {"id":"0192f1e2-…","handle":"pip","display_name":"Pip Squeakson","bio":"","created_at":"2026-01-02T03:04:05Z"}

GET /api/users/nobody
404 {"error":"user not found"}
```

## Constraints

- No real database: the tests use a fake `DB` whose `Scan` behaves like
  `database/sql`'s, including refusing to put `NULL` into a `*string`.
- The 500 response must not contain the database's error text. The log should.
