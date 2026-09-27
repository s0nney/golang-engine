---
title: IDs with the uuid Package
quiz:
  - question: What's the main advantage of `uuid.NewV7()` over `uuid.NewV4()` for Squeak's IDs?
    options:
      - text: V7 is more random, so it's harder to guess
      - text: V7 starts with a timestamp, so IDs created later sort after earlier ones
        correct: true
      - text: V7 is shorter
      - text: V4 is deprecated
    explanation: |
      A version 7 UUID puts a millisecond timestamp in its first 48 bits, so sorting by ID
      is (nearly) sorting by creation time. Databases like that too: new rows land at the
      end of an index instead of at random places. V4 is fully random.
  - question: |
      What does this print?

      ```go
      _, err := uuid.Parse("42")
      fmt.Println(err != nil)
      var id uuid.UUID
      fmt.Println(id == uuid.Nil())
      ```
    options:
      - text: '`true`, then `true`'
        correct: true
      - text: '`false`, then `true`'
      - text: '`true`, then `false`'
      - text: It doesn't compile, because `UUID` can't be compared with `==`
    explanation: |
      `"42"` isn't a UUID, so `Parse` fails. `UUID` is a `[16]byte` array, so it's
      comparable with `==` (and usable as a map key), and its zero value is the all-zeros
      "nil UUID".
---

Squeak's first exercise used counting IDs: 1, 2, 3. They're simple, but they have
problems in a real API:

- **They leak information.** If your squeak is `/api/squeaks/5012`, anyone can guess
  how many squeaks exist and walk through all of them.
- **They need a central counter.** Two servers (or a server and a database migration)
  can't safely hand out IDs without coordinating.

The usual alternative is a **UUID**, a 128-bit ID written as 36 characters, such as
`0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6`. There are so many possible values that you can
generate them anywhere, with no coordination, and never see a collision.

## Go 1.27's uuid package

Until recently every Go project pulled in a third-party UUID module. Go 1.27 adds one to
the standard library, as a top-level package simply called `uuid`:

```go
import "uuid"
```

The API is small:

| Function | Gives you |
|---|---|
| `uuid.New()` | a new UUID from the recommended algorithm (currently V4) |
| `uuid.NewV4()` | a fully random UUID |
| `uuid.NewV7()` | a time-ordered UUID |
| `uuid.Parse(s)` | a UUID parsed from text, or an error |
| `uuid.MustParse(s)` | like `Parse`, but panics (for constants in tests) |
| `uuid.Nil()` | the all-zeros UUID |

`uuid.UUID` is a `[16]byte`. It's comparable with `==`, works as a map key, has a
`Compare` method for sorting, and implements `MarshalText`/`UnmarshalText`, so JSON
encodes it as a string.

## V4 or V7?

A **V4** UUID is 122 random bits. A **V7** UUID starts with a 48-bit millisecond
timestamp followed by at least 62 random bits. That makes V7 IDs **sort by creation
time**, which is handy for a timeline and kind to database indexes. Squeak uses V7:

```go
package main

import (
	"encoding/json/v2"
	"fmt"
	"slices"
	"uuid"
)

type Squeak struct {
	ID   uuid.UUID `json:"id"`
	Body string    `json:"body"`
}

func main() {
	var ids []uuid.UUID
	for range 5 {
		ids = append(ids, uuid.NewV7())
	}
	fmt.Println("sorted by creation:", slices.IsSortedFunc(ids, uuid.UUID.Compare))
	fmt.Println("string length:", len(ids[0].String()))
	fmt.Println("version digit:", ids[0].String()[14:15])

	id := uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
	data, _ := json.Marshal(Squeak{ID: id, Body: "sorted squeaks!"})
	fmt.Println(string(data))

	for _, s := range []string{"0192F1E2-8C3A-7B4D-9E5F-A1B2C3D4E5F6", "42"} {
		parsed, err := uuid.Parse(s)
		fmt.Println(parsed, err)
	}
}
```

```
sorted by creation: true
string length: 36
version digit: 7
{"id":"0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6","body":"sorted squeaks!"}
0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6 <nil>
00000000-0000-0000-0000-000000000000 invalid uuid
```

Notice `uuid.UUID.Compare` used as a function value: a *method expression*, whose
signature `func(a, b uuid.UUID) int` is exactly what `slices.SortFunc` and friends want.

## Parsing path values

A UUID in a URL is client input, so parse it before use and answer 400 when it's not
a UUID:

```go
id, err := uuid.Parse(r.PathValue("id"))
if err != nil {
	respondWithError(w, http.StatusBadRequest, "invalid squeak id")
	return
}
```

`Parse` accepts upper or lower case, and even braces or a `urn:uuid:` prefix, and
`String()` always gives the canonical lowercase form. Store and compare `uuid.UUID`
values, never raw strings, so `ABC...` and `abc...` can't turn into two different IDs.

## A word on guessability

V7 UUIDs contain a timestamp, so they reveal *when* something was created, and they're
less random than V4. They're still far too random to enumerate, which is all an ID
needs. But an ID is **not a secret**: never use one (V4 or V7) as a password-reset link
or API key. Those need `crypto/rand`, which you'll use in the authentication chapter.
