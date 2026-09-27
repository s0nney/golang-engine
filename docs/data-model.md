# Data model

Everything lives in one SQLite file (default `goland.db`, plus `-wal`/`-shm` files because
the database runs in WAL mode). The schema is defined in Go, in `internal/store`; there
are no migration files.

There are two kinds of table:

- **Content tables** are a copy of the embedded course files. `Store.Seed` drops and
  recreates them in one transaction on every start, so they always match the binary.
- **Learner tables** (`progress`, `activity`) are created with `CREATE TABLE IF NOT EXISTS`
  and never dropped. They reference content by **slug**, not by row ID, because row IDs
  change on every reseed.

## Content tables (`store.go`)

```
courses    (id, slug UNIQUE, title, description, position)
chapters   (id, course_id → courses, slug, title, description, position, UNIQUE(course_id, slug))
lessons    (id, chapter_id → chapters, slug, title, html, position, UNIQUE(chapter_id, slug))
questions  (id, lesson_id → lessons, prompt, explanation, position)
options    (id, question_id → questions, text, correct, position)
exercises  (lesson_id PK → lessons, starter, tests, expected_output)
```

- `position` preserves the `NN-` ordering from the file names.
- `html`, `prompt`, `explanation` and option `text` hold **rendered HTML**; Markdown is
  rendered once, at load time.
- `exercises` deliberately has no `solution` column: the reference solution only exists
  in the content files and is used by `cmd/validate`, never served.

## Progress (`progress.go`)

```sql
CREATE TABLE progress (
  session, course, chapter, lesson,       -- PRIMARY KEY
  quiz_passed     INTEGER DEFAULT 0,
  exercise_passed INTEGER DEFAULT 0,
  code            TEXT    DEFAULT '',     -- last code run or submitted
  updated_at      TEXT    DEFAULT CURRENT_TIMESTAMP
);
```

- `session` is the UUID from the `goland_session` cookie.
- A lesson is **complete** when every part it has is passed: the quiz if it has one and the
  exercise if it has one (`completeLessons` query). Lessons that no longer exist in the
  content simply drop out of the join.
- `exercise_passed` is sticky: `max(old, new)`, so a later failing Run doesn't un-complete
  a lesson.
- Run, Submit and Reset all save `code`; Reset saves `''`, which makes the page fall back to
  the starter.

## Activity (`activity.go`)

```sql
CREATE TABLE activity (
  session, day,              -- day is 'YYYY-MM-DD' in the learner's timezone
  kind  CHECK (kind IN ('lesson', 'exercise')),
  course, chapter, lesson,
  PRIMARY KEY (session, day, kind, course, chapter, lesson)
);
```

One row per passed quiz (`lesson`) or passed exercise (`exercise`) per lesson per day.
The primary key plus `ON CONFLICT DO NOTHING` makes repeat passes on the same day free,
while a review on a later day adds a new row. Rows are written in the **same transaction**
as the progress update, so progress and activity can't disagree. See
[activity-and-streaks.md](activity-and-streaks.md) for how rows turn into the heatmap.

## Reset

`POST /progress/reset` deletes the session's rows from both `progress` and `activity`
in one transaction.

## Operational notes

- Deleting `goland.db*` loses all learner progress but nothing else; content is reseeded
  on the next start.
- Renaming a course, chapter or lesson slug (the part after `NN-`) orphans progress on it.
  Renumbering (`NN-`) is safe.
- Back up with `sqlite3 goland.db ".backup copy.db"` (safe while the server runs in WAL
  mode), or stop the server and copy all three files.
