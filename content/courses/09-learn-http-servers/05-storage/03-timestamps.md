---
title: Timestamps
quiz:
  - question: |
      What does this print?

      ```go
      created := time.Now()
      data, _ := json.Marshal(created)
      var decoded time.Time
      json.Unmarshal(data, &decoded)
      fmt.Println(created == decoded, created.Equal(decoded))
      ```
    options:
      - text: '`true true`'
      - text: '`false true`'
        correct: true
      - text: '`false false`'
      - text: '`true false`'
    explanation: |
      `time.Now()` carries a monotonic clock reading and a `*time.Location`, and neither
      survives JSON. `==` compares the whole struct, so it's `false`. `Equal` compares
      the instant in time, so it's `true`. Always compare times with `Equal`.
  - question: Why does Squeak store `time.Now().UTC()` instead of `time.Now()`?
    options:
      - text: UTC times are more precise
      - text: So every stored timestamp has the same location, whatever the server's local time zone is
        correct: true
      - text: JSON can't encode local times
      - text: '`time.Now()` is deprecated'
    explanation: |
      A server's local zone depends on where and how it's deployed, and it can change with
      daylight saving time. Storing UTC keeps data consistent and makes the JSON end in
      `Z`. Clients convert to the user's zone for display.
---

Every squeak needs a `created_at`, and every timeline sorted by time depends on it.
Time looks simple and hides some of Go's sneakiest gotchas.

## Store UTC

```go
now := time.Now().UTC()
```

`time.Now()` returns the time in the server's **local** zone. That depends on the machine
(your laptop says `Europe/Paris`, the container says `UTC`), so the same instant would be
written in different ways depending on where Squeak runs. Convert to UTC when you store it,
and let clients convert to the user's zone when they display it.

In JSON, a UTC `time.Time` becomes an RFC 3339 string ending in `Z`:

```json
{"created_at": "2026-09-01T09:30:00.123456789Z"}
```

## Compare with Equal, never ==

`time.Time` is a struct holding the instant, a `*time.Location`, and (for values from
`time.Now()`) a hidden **monotonic clock reading** used for measuring durations. `==`
compares all of it. Two values for the *same instant* can differ in location or
monotonic reading, so `==` says they're different:

```go
package main

import (
	"encoding/json/v2"
	"fmt"
	"time"
)

func roundTrip(t time.Time) time.Time {
	data, _ := json.Marshal(t)
	var decoded time.Time
	json.Unmarshal(data, &decoded)
	return decoded
}

func main() {
	local := time.Now()
	fmt.Println("local ==    :", local == roundTrip(local))
	fmt.Println("local Equal :", local.Equal(roundTrip(local)))

	utc := time.Now().UTC()
	fmt.Println("utc ==      :", utc == roundTrip(utc))

	paris, _ := time.LoadLocation("Europe/Paris")
	fmt.Println("Paris Equal :", utc.Equal(utc.In(paris)))
	fmt.Println("Paris ==    :", utc == utc.In(paris))
}
```

```
local ==    : false
local Equal : true
utc ==      : true
Paris Equal : true
Paris ==    : false
```

The `time.Now()` value lost its monotonic reading (and possibly its location) on the
way through JSON, so `==` fails. `.UTC()` happens to strip the monotonic reading and
decoding a `Z` time gives UTC back, so that one survives `==`. But the same instant
viewed from Paris is `==`-different again. Don't try to remember which cases work. The
rule is simple: **compare times with `Equal`**, `Before` and `After`. (It's also why
`time.Time` makes a poor map key.)

(`time.LoadLocation` needs time zone data. It's on most systems, and you can embed it
with `import _ "time/tzdata"` when it isn't.)

## Sorting by time

Squeak lists squeaks oldest first. `time.Time` has a `Compare` method that fits
`slices.SortFunc`:

```go
slices.SortFunc(squeaks, func(a, b Squeak) int {
	return a.CreatedAt.Compare(b.CreatedAt)
})
```

Two squeaks created in the same nanosecond (it happens on fast machines and coarse
clocks) compare equal and could come out in either order. Because Squeak's IDs are
UUID V7, you can break ties by ID and get a stable order:

```go
slices.SortFunc(squeaks, func(a, b Squeak) int {
	return cmp.Or(
		a.CreatedAt.Compare(b.CreatedAt),
		a.ID.Compare(b.ID),
	)
})
```

`cmp.Or` returns its first non-zero argument, which makes multi-key comparisons read
nicely.

## Who decides "now"?

If the store calls `time.Now()` directly, tests can't control it. Asserting "this squeak
was created at 09:30" becomes impossible, and testing "tokens expire after an hour"
means waiting an hour. Two common fixes:

1. **Inject a clock.** Give the store a `now func() time.Time` field that defaults to
   `time.Now` and that tests replace with a fixed time:

   ```go
   type MemoryStore struct {
   	now func() time.Time
   	// ...
   }

   func NewMemoryStore() *MemoryStore {
   	return &MemoryStore{now: time.Now}
   }
   ```

2. **Use `testing/synctest`.** Inside a synctest "bubble", `time.Now()` is a fake clock
   that starts at midnight UTC on 2000-01-01 and only moves when every goroutine is
   blocked. `time.Sleep(time.Hour)` returns instantly. You'll use it to test token
   expiry and rate limits in the testing chapter.

Squeak mostly relies on option 2, which keeps production code plain.

## Times in the API

Two more habits for timestamps in JSON APIs:

- Name fields after what happened: `created_at`, `updated_at`, `expires_at`.
- Use `omitzero` for times that may not have happened yet, like `edited_at`, so the
  client sees no field instead of `"0001-01-01T00:00:00Z"`.
