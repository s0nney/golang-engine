# Activity heatmap and streaks

The roadmap page (`/`) shows a 26-week activity heatmap, the current streak, the best
streak and today's count, in the style of boot.dev and GitHub contribution graphs.

## What counts as activity

| Event | Counts? |
|---|---|
| Passing a lesson's quiz (all questions right) | Yes, once per lesson per day (`kind = 'lesson'`) |
| Passing a coding exercise with Submit | Yes, once per lesson per day (`kind = 'exercise'`) |
| Passing the same quiz/exercise again later the same day | No |
| Passing it again on a later day (a review) | Yes |
| Run, failed Submit, Reset code | No |

A lesson with both a quiz and an exercise can therefore add two completions in a day.
Rows are inserted in the same transaction as the progress update (`recordActivity` in
`internal/store/activity.go`), with `ON CONFLICT DO NOTHING` on the primary key
`(session, day, kind, course, chapter, lesson)` doing the once-per-day deduplication.

## Dates and timezones

A day is a **calendar date in the learner's timezone**, not a 24-hour window:

1. `static/activity.js` stores the browser's IANA timezone name
   (`Intl.DateTimeFormat().resolvedOptions().timeZone`) in the `goland_timezone` cookie.
2. `browserNow` in `internal/web/web.go` loads that location with `time.LoadLocation`
   (the tz database is embedded via `time/tzdata`) and returns `time.Now().In(loc)`.
   Missing, invalid, `Local`, or over-long (>100 bytes) values fall back to **UTC**, as does
   running with JavaScript disabled.
3. The date is stored as `YYYY-MM-DD`. Past rows keep the date they were recorded with,
   even if the learner later changes timezone.

The browser supplies only a zone name, never a date or timestamp, so the client can't
forge activity for arbitrary dates. Changing zones can shift the recorded calendar day;
these streaks are a self-study aid, not an anti-cheating or attendance system.

Streak arithmetic works on date values in UTC, so DST transitions never add or remove a
day.

## Streak rules (`summarizeActivity`)

- **Today** is the number of completions dated today.
- **Current streak** counts consecutive active days ending **today, or yesterday** if
  nothing is done yet today, so the streak isn't shown as broken while there's still
  time to keep it alive.
- **Best streak** is the longest run of consecutive active days across all recorded
  history (not only the visible 26 weeks). Runs are measured from their first day, so the
  work is linear in the number of active days.
- Days dated after today (possible after travelling west) are ignored until they arrive.

## Heatmap

- 26 columns (weeks) × 7 rows (Sunday to Saturday). The grid starts on the Sunday 25 weeks
  before the current week; days after today in the current week are rendered as hidden
  placeholders.
- Shade levels: 0, 1–2, 3–5, 6–9, 10+ completions (`heatmapClass`). Light and dark themes
  have separate colour palettes, and every cell
  also carries a text label ("2026-09-27: 3 completed quizzes and exercises"), shown as a
  hover tooltip and exposed as the list item's `aria-label`.
- The grid is one focusable, horizontally scrollable region; individual days are not tab
  stops, so keyboard users aren't forced through 180 cells to reach the course list.
- Under the grid: total completions and active days in the period, a legend, and a nudge
  ("Keep it going…" when the streak is alive but nothing is done today; "Your first
  completion starts your streak" for a learner with no history).

## Identity, persistence and reset

- Activity belongs to the anonymous `goland_session` cookie, like progress. Clearing
  cookies or switching browsers starts a new, empty history.
- The `activity` table is never dropped by reseeding, so it survives restarts and content
  updates.
- **Reset my progress** deletes the session's progress and activity in one transaction.
- Activity started being recorded when the feature shipped; older progress was not
  backfilled, because `progress.updated_at` also changes on plain code saves and so
  doesn't mean "completed on that day".

## Concurrency

Two tabs, or a double-clicked submit, can write at the same moment. The SQLite DSN sets
`busy_timeout(5000)` and `_txlock=immediate`, so writers queue for the lock instead of
failing with `SQLITE_BUSY` (covered by `TestConcurrentWrites`).

## Tests

- `internal/store/activity_test.go`: deduplication, reviews, reset, calendar streaks
  (yesterday grace, broken streaks, best streak, year boundary, leap day, both DST
  transitions), local-date recording, concurrent writes.
- `internal/web/web_test.go`: `TestActivityDashboard` (rendering) and `TestBrowserTimezone`
  (cookie parsing and fallbacks).
