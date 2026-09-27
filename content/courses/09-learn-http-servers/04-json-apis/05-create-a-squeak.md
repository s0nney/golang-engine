---
title: Create a Squeak
quiz:
  - question: |
      A handler has this failure path. What does the client receive for an empty body?

      ```go
      body, err := cleanSqueak(params.Body)
      if err != nil {
      	respondWithError(w, http.StatusBadRequest, err.Error())
      }
      respondWithJSON(w, http.StatusCreated, Squeak{ID: 1, Body: body})
      ```
    options:
      - text: A clean 400 with the error JSON
      - text: A 400 status, but the error JSON *and* the squeak JSON glued together in the body
        correct: true
      - text: A 201 with the squeak
      - text: A 500, because the handler wrote twice
    explanation: |
      Without a `return`, the handler carries on. The second `WriteHeader(201)` is ignored
      (the 400 was already sent) with a "superfluous WriteHeader" log line, but the
      second body *is* written, so the client gets two JSON documents back to back.
exercise:
  starter: |
    package main

    import (
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync/atomic"
    	"unicode/utf8"
    )

    const (
    	maxBodyBytes = 1024 // limit for the whole request body
    	maxSqueakLen = 140  // limit for the squeak text, in characters
    )

    var bannedWords = map[string]bool{"trap": true, "cat": true, "owl": true}

    type Squeak struct {
    	ID     int64  `json:"id"`
    	Author string `json:"author"`
    	Body   string `json:"body"`
    }

    type createSqueakParams struct {
    	Author string `json:"author"`
    	Body   string `json:"body"`
    }

    type apiConfig struct {
    	lastID atomic.Int64
    }

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

    // respondWithError responds with {"error": msg} and the given status.
    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	// ?
    }

    // cleanSqueak trims body, rejects it if it's empty ("squeak body is required")
    // or longer than maxSqueakLen characters ("squeak is too long"), and
    // replaces banned words (any capitalisation, split on spaces) with "****".
    func cleanSqueak(body string) (string, error) {
    	// ?
    	return body, nil
    }

    // handleCreateSqueak handles POST /api/squeaks. See the lesson for the rules.
    func (cfg *apiConfig) handleCreateSqueak(w http.ResponseWriter, r *http.Request) {
    	var params createSqueakParams
    	json.UnmarshalRead(r.Body, &params)
    	// ?
    	respondWithJSON(w, http.StatusOK, Squeak{Author: params.Author, Body: params.Body})
    }

    func main() {
    	cfg := &apiConfig{}
    	mux := http.NewServeMux()
    	mux.HandleFunc("POST /api/squeaks", cfg.handleCreateSqueak)

    	for _, body := range []string{
    		`{"author":"pip","body":"  beware the CAT  "}`,
    		`{"author":"pip","body":""}`,
    		`{"author":"pip","body":"hi","mood":"sneaky"}`,
    		`{"body":"who am i"}`,
    		`{"author":"pip","body":"` + strings.Repeat("s", 2000) + `"}`,
    	} {
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest("POST", "/api/squeaks", strings.NewReader(body)))
    		fmt.Printf("%d Location=%q %s\n", rec.Code, rec.Header().Get("Location"), rec.Body.String())
    	}
    	_, _ = errors.New, utf8.RuneCountInString
    }
  solution: |
    package main

    import (
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync/atomic"
    	"unicode/utf8"
    )

    const (
    	maxBodyBytes = 1024
    	maxSqueakLen = 140
    )

    var bannedWords = map[string]bool{"trap": true, "cat": true, "owl": true}

    type Squeak struct {
    	ID     int64  `json:"id"`
    	Author string `json:"author"`
    	Body   string `json:"body"`
    }

    type createSqueakParams struct {
    	Author string `json:"author"`
    	Body   string `json:"body"`
    }

    type apiConfig struct {
    	lastID atomic.Int64
    }

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

    type errorResponse struct {
    	Error string `json:"error"`
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	respondWithJSON(w, code, errorResponse{Error: msg})
    }

    func cleanSqueak(body string) (string, error) {
    	body = strings.TrimSpace(body)
    	if body == "" {
    		return "", errors.New("squeak body is required")
    	}
    	if utf8.RuneCountInString(body) > maxSqueakLen {
    		return "", errors.New("squeak is too long")
    	}
    	words := strings.Split(body, " ")
    	for i, word := range words {
    		if bannedWords[strings.ToLower(word)] {
    			words[i] = "****"
    		}
    	}
    	return strings.Join(words, " "), nil
    }

    func (cfg *apiConfig) handleCreateSqueak(w http.ResponseWriter, r *http.Request) {
    	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
    	var params createSqueakParams
    	err := json.UnmarshalRead(r.Body, &params, json.RejectUnknownMembers(true))
    	if err != nil {
    		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
    			respondWithError(w, http.StatusRequestEntityTooLarge, "request body too large")
    			return
    		}
    		respondWithError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
    		return
    	}
    	author := strings.TrimSpace(params.Author)
    	if author == "" {
    		respondWithError(w, http.StatusBadRequest, "author is required")
    		return
    	}
    	body, err := cleanSqueak(params.Body)
    	if err != nil {
    		respondWithError(w, http.StatusBadRequest, err.Error())
    		return
    	}
    	squeak := Squeak{ID: cfg.lastID.Add(1), Author: author, Body: body}
    	w.Header().Set("Location", fmt.Sprintf("/api/squeaks/%d", squeak.ID))
    	respondWithJSON(w, http.StatusCreated, squeak)
    }

    func main() {
    	cfg := &apiConfig{}
    	mux := http.NewServeMux()
    	mux.HandleFunc("POST /api/squeaks", cfg.handleCreateSqueak)

    	for _, body := range []string{
    		`{"author":"pip","body":"  beware the CAT  "}`,
    		`{"author":"pip","body":""}`,
    		`{"author":"pip","body":"hi","mood":"sneaky"}`,
    		`{"body":"who am i"}`,
    		`{"author":"pip","body":"` + strings.Repeat("s", 2000) + `"}`,
    	} {
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest("POST", "/api/squeaks", strings.NewReader(body)))
    		fmt.Printf("%d Location=%q %s\n", rec.Code, rec.Header().Get("Location"), rec.Body.String())
    	}
    }
  tests: |
    package main

    import (
    	"encoding/json/v2"
    	"net/http"
    	"net/http/httptest"
    	"strconv"
    	"strings"
    	"testing"
    )

    func post(t *testing.T, cfg *apiConfig, body string) *httptest.ResponseRecorder {
    	t.Helper()
    	rec := httptest.NewRecorder()
    	cfg.handleCreateSqueak(rec, httptest.NewRequest("POST", "/api/squeaks", strings.NewReader(body)))
    	return rec
    }

    func wantError(t *testing.T, rec *httptest.ResponseRecorder, body string, wantCode int, wantMsg string) {
    	t.Helper()
    	if rec.Code != wantCode {
    		t.Errorf("POST %.60s: status = %d, want %d", body, rec.Code, wantCode)
    		return
    	}
    	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
    		t.Errorf("POST %.60s: Content-Type = %q, want application/json", body, ct)
    	}
    	var e struct {
    		Error string `json:"error"`
    	}
    	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
    		t.Errorf("POST %.60s: response %q is not a single JSON object like {\"error\":\"...\"}: %v", body, rec.Body.String(), err)
    		return
    	}
    	if e.Error == "" {
    		t.Errorf("POST %.60s: response %q has an empty \"error\" field", body, rec.Body.String())
    	}
    	if wantMsg != "" && !strings.Contains(e.Error, wantMsg) {
    		t.Errorf("POST %.60s: error = %q, want it to contain %q", body, e.Error, wantMsg)
    	}
    }

    func TestCreateSqueak(t *testing.T) {
    	cfg := &apiConfig{}
    	for i, tt := range []struct {
    		body, wantBody string
    	}{
    		{`{"author":"pip","body":"cheese at noon"}`, "cheese at noon"},
    		{`{"author":"whiskers","body":"  beware the CAT by the Trap  "}`, "beware the **** by the ****"},
    		{`{"author":"pip","body":"` + strings.Repeat("é", 140) + `"}`, strings.Repeat("é", 140)},
    	} {
    		rec := post(t, cfg, tt.body)
    		if rec.Code != http.StatusCreated {
    			t.Fatalf("POST %.60s: status = %d, want 201 (body %q)", tt.body, rec.Code, rec.Body.String())
    		}
    		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
    			t.Errorf("Content-Type = %q, want application/json", ct)
    		}
    		var got Squeak
    		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
    			t.Fatalf("response %q is not a squeak: %v", rec.Body.String(), err)
    		}
    		wantID := int64(i + 1)
    		if got.ID != wantID {
    			t.Errorf("squeak #%d: id = %d, want %d (IDs count up from 1)", i+1, got.ID, wantID)
    		}
    		if got.Body != tt.wantBody {
    			t.Errorf("POST %.60s: body = %q, want %q", tt.body, got.Body, tt.wantBody)
    		}
    		if got.Author == "" {
    			t.Errorf("POST %.60s: author is empty in the response", tt.body)
    		}
    		wantLoc := "/api/squeaks/" + strconv.FormatInt(wantID, 10)
    		if loc := rec.Header().Get("Location"); loc != wantLoc {
    			t.Errorf("squeak #%d: Location = %q, want %q", i+1, loc, wantLoc)
    		}
    	}
    }

    func TestCreateSqueakErrors(t *testing.T) {
    	cfg := &apiConfig{}
    	for _, tt := range []struct {
    		body     string
    		wantCode int
    		wantMsg  string
    	}{
    		{`{"author":"pip","body":""}`, 400, "squeak body is required"},
    		{`{"author":"pip","body":"     "}`, 400, "squeak body is required"},
    		{`{"author":"pip","body":"` + strings.Repeat("s", 141) + `"}`, 400, "too long"},
    		{`{"body":"who am i"}`, 400, "author is required"},
    		{`{"author":"  ","body":"who am i"}`, 400, "author is required"},
    		{`{"author":"pip","body":"hi","mood":"sneaky"}`, 400, ""},
    		{`{"author":"pip","body":`, 400, ""},
    		{`{"author":"pip","body":42}`, 400, ""},
    		{`not json at all`, 400, ""},
    		{`{"author":"pip","body":"` + strings.Repeat("s", 2000) + `"}`, 413, ""},
    	} {
    		wantError(t, post(t, cfg, tt.body), tt.body, tt.wantCode, tt.wantMsg)
    	}
    	if got := cfg.lastID.Load(); got != 0 {
    		t.Errorf("rejected squeaks used up IDs: lastID = %d, want 0", got)
    	}
    }
---

Time to write Squeak's most important endpoint: `POST /api/squeaks`. It pulls together
everything in this chapter: a size limit, strict decoding, validation, consistent errors
and a proper `201 Created`.

Real storage arrives next chapter. For now the handler just hands out IDs from an
`atomic.Int64` and echoes the squeak back.

## Your task

1. **`respondWithError`**: respond with the JSON `{"error": msg}` and the given
   status. Build it on the provided `respondWithJSON`.
2. **`cleanSqueak`**: trim surrounding whitespace, then
   - return an error containing `squeak body is required` if nothing is left,
   - return an error containing `squeak is too long` if it's over `maxSqueakLen`
     **characters** (count runes, not bytes),
   - replace each banned word (split on single spaces, compared in lowercase) with
     `****`.
3. **`handleCreateSqueak`**:
   - limit the body to `maxBodyBytes` with `http.MaxBytesReader`,
   - decode with `json.UnmarshalRead` and `json.RejectUnknownMembers(true)`,
   - body too large: **413** with any error message,
   - any other decode error: **400** with any error message,
   - trimmed `author` empty: **400** with `author is required`,
   - `cleanSqueak` fails: **400** with its error message,
   - otherwise take the next ID with `cfg.lastID.Add(1)` (so IDs go 1, 2, 3...), set
     `Location: /api/squeaks/<id>`, and respond **201** with the squeak as JSON (the
     trimmed author and the cleaned body).

Rejected requests must **not** use up an ID, so only call `Add` once everything is
valid. And every error response must be a single JSON object, so `return` after each
one.

The starter imports `errors` and `unicode/utf8` for you and keeps them "used" with a
throwaway line at the end of `main`. Delete that line once your code uses them.
