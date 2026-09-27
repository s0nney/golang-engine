---
title: Validation
quiz:
  - question: |
      Squeaks may be at most 140 characters. What does this print?

      ```go
      body := strings.Repeat("é", 100)
      fmt.Println(len(body) > 140, utf8.RuneCountInString(body) > 140)
      ```
    options:
      - text: '`false false`'
      - text: '`true true`'
      - text: '`true false`'
        correct: true
      - text: '`false true`'
    explanation: |
      `len` counts **bytes**, and `é` takes 2 bytes in UTF-8, so `len(body)` is 200.
      There are only 100 characters (runes). Checking `len` would wrongly reject this
      squeak, so count runes for user-facing limits.
  - question: Where should validation of a new squeak happen?
    options:
      - text: Only in the client app, so the server stays fast
      - text: On the server, after decoding and before saving, even if the client validates too
        correct: true
      - text: Only in the database
      - text: Inside the JSON decoder, using struct tags
    explanation: |
      Clients are convenient for instant feedback, but anyone can call your API with
      `curl`. The server is the only place you control, so it must enforce every rule.
      Go's JSON package has no validation tags, so you write the checks yourself.
---

`json.UnmarshalRead` told you the body was *well-formed*. It didn't tell you it made
sense. These all decode perfectly into `createSqueakParams`:

```json
{"body": ""}
{"body": "                "}
{"body": "a squeak that goes on and on and on ... for 5000 characters"}
{}
```

**Validation** is checking the decoded values against your rules and refusing politely
when they don't hold. Never trust the client to have done it, because anyone can call
your API with `curl`.

## Squeak's rules

1. The body is required, and whitespace-only doesn't count.
2. At most **140 characters**.
3. Squeak is a family-friendly app. The words `trap`, `cat` and `owl` (any
   capitalisation) are replaced with `****`.

Rules 1 and 2 *reject* the request. Rule 3 *cleans* it: the squeak is still accepted.
Both are common, and it's worth being clear which is which.

## Characters, not bytes

Go strings are bytes. `len("fromage")` is 7, but `len("affiné")` is 7 too, for 6
characters, because `é` is two bytes in UTF-8. A limit you show to users as
"140 characters" must count **runes**:

```go
if utf8.RuneCountInString(body) > 140 { ... }
```

(Runes aren't quite what humans call characters either: some emoji are several runes
long. For a squeak limit, runes are close enough and what most APIs use.)

## A validate function

Keep validation in plain functions that take values and return errors. They don't need
`http` at all, which makes them trivial to test:

```go
package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const maxSqueakLen = 140

var bannedWords = map[string]bool{"trap": true, "cat": true, "owl": true}

var (
	errEmptySqueak   = errors.New("squeak body is required")
	errSqueakTooLong = fmt.Errorf("squeak is too long (max %d characters)", maxSqueakLen)
)

// cleanSqueak validates body and returns the version to store.
func cleanSqueak(body string) (string, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return "", errEmptySqueak
	}
	if utf8.RuneCountInString(body) > maxSqueakLen {
		return "", errSqueakTooLong
	}

	words := strings.Split(body, " ")
	for i, word := range words {
		if bannedWords[strings.ToLower(word)] {
			words[i] = "****"
		}
	}
	return strings.Join(words, " "), nil
}

func main() {
	for _, body := range []string{
		"  cheese at noon  ",
		"beware the Cat by the door",
		"   ",
		strings.Repeat("é", 140),
		strings.Repeat("squeak ", 30),
	} {
		cleaned, err := cleanSqueak(body)
		if err != nil {
			fmt.Println("rejected:", err)
			continue
		}
		fmt.Printf("accepted: %.40q\n", cleaned)
	}
}
```

```
accepted: "cheese at noon"
accepted: "beware the **** by the door"
rejected: squeak body is required
accepted: "éééééééééééééééééééééééééééééééééééééééé"
rejected: squeak is too long (max 140 characters)
```

(`%.40q` trims the printed string to 40 characters so the long one fits on a line.)

Notes on the choices:

- **Trim first**, then check emptiness, so `"   "` is rejected.
- **Sentinel errors** (`errEmptySqueak`) let handlers and tests use `errors.Is` rather
  than comparing message strings.
- The word filter splits on single spaces, so `"cat!"` slips through. Real content
  filters get complicated fast. This one is deliberately simple.

## Validation in the handler

The handler stitches decode, validate and respond together:

```go
func (cfg *apiConfig) handleCreateSqueak(w http.ResponseWriter, r *http.Request) {
	var params createSqueakParams
	if code, err := decodeJSON(w, r, &params); err != nil {
		respondWithError(w, code, err.Error())
		return
	}
	body, err := cleanSqueak(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	// ... save the squeak, then respondWithJSON(w, http.StatusCreated, squeak)
}
```

Every failure path **returns** right after responding. Forgetting a `return` is the
number one handler bug: the code carries on and writes a second response.

## 400 or 422?

Some APIs split failures in two: `400 Bad Request` when the body can't be parsed at all,
and `422 Unprocessable Content` when it parses but breaks a rule. Others use 400 for
everything. Both are fine. What matters is consistency and a clear error message.
Squeak uses **400** for all bad input.

`respondWithError` doesn't exist yet. Consistent error responses are next.
