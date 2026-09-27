---
title: The database/sql Package
quiz:
  - question: Where should Squeak call `sql.Open`?
    options:
      - text: At the start of every handler, closing the `*sql.DB` when the handler returns
      - text: Once at startup, sharing the one `*sql.DB` across all requests
        correct: true
      - text: Once per user, when they log in
      - text: Never; `database/sql` opens connections by itself
    explanation: |
      A `*sql.DB` isn't a single connection. It's a pool of them, safe for concurrent
      use, that opens and reuses connections as needed. Open it once, store it in your
      store, and close it at shutdown. Opening a pool per request throws all that reuse away.
  - question: |
      What's wrong with this code?

      ```go
      rows, err := db.QueryContext(ctx, "SELECT id FROM squeaks")
      if err != nil {
      	return err
      }
      for rows.Next() {
      	// scan...
      }
      return nil
      ```
    options:
      - text: '`QueryContext` should be `QueryRowContext`'
      - text: It never calls `rows.Close()` and never checks `rows.Err()`, so it can leak a connection and silently return a partial list
        correct: true
      - text: '`rows.Next()` must be called before the error check'
      - text: Nothing, `rows` closes itself when the loop starts
    explanation: |
      `rows` holds a pool connection until it's closed. It does close itself when `Next`
      runs out, but not if you return from the loop early, so `defer rows.Close()` right
      after the error check. And `Next` returns `false` both at the end *and* when
      something failed halfway, so check `rows.Err()` after the loop.
  - question: '`db.QueryRowContext(ctx, q, id).Scan(&sq.ID, &sq.Body)` finds no matching row. What happens?'
    options:
      - text: '`QueryRowContext` returns a nil `*sql.Row`, and `Scan` panics'
      - text: '`Scan` returns `sql.ErrNoRows`'
        correct: true
      - text: '`Scan` succeeds and leaves the fields at their zero values'
      - text: '`Scan` returns `ErrNotFound`'
    explanation: |
      `QueryRowContext` never returns an error itself. Any error, including "no rows",
      is saved and returned by `Scan`. `sql.ErrNoRows` is the database package's own
      sentinel. Translating it to Squeak's `ErrNotFound` is your store's job.
exercise:
  starter: |
    package main

    import (
    	"database/sql"
    	"errors"
    	"fmt"
    	"time"
    	"uuid"
    )

    var ErrNotFound = errors.New("not found")

    type Squeak struct {
    	ID        uuid.UUID `json:"id"`
    	AuthorID  uuid.UUID `json:"author_id"`
    	Body      string    `json:"body"`
    	CreatedAt time.Time `json:"created_at"`
    }

    // rowScanner is the method of *sql.Row (and *sql.Rows) that scanSqueak needs.
    type rowScanner interface {
    	Scan(dest ...any) error
    }

    // squeakRows is the part of *sql.Rows that collectSqueaks needs.
    type squeakRows interface {
    	Next() bool
    	Scan(dest ...any) error
    	Err() error
    	Close() error
    }

    // Every query that scanSqueak reads must select exactly these columns, in this order.
    const selectSqueaks = "SELECT id, author_id, body, created_at FROM squeaks"

    // scanSqueak reads one squeak from row. A missing row is ErrNotFound.
    func scanSqueak(row rowScanner) (Squeak, error) {
    	var sq Squeak
    	err := row.Scan(&sq.ID, &sq.AuthorID, &sq.Body, &sq.CreatedAt)
    	return sq, err
    }

    // collectSqueaks reads every row, always closes rows, and reports any
    // error that ended the iteration early.
    func collectSqueaks(rows squeakRows) ([]Squeak, error) {
    	var list []Squeak
    	for rows.Next() {
    		sq, _ := scanSqueak(rows)
    		list = append(list, sq)
    	}
    	return list, nil
    }

    // ---- in-memory stand-ins for *sql.Row and *sql.Rows, used by main ----

    type memRow struct {
    	values []any // id, author_id, body, created_at
    	err    error // returned by Scan instead, like sql.ErrNoRows
    }

    func (r memRow) Scan(dest ...any) error {
    	if r.err != nil {
    		return r.err
    	}
    	if len(dest) != len(r.values) {
    		return fmt.Errorf("sql: expected %d destination arguments in Scan, not %d", len(r.values), len(dest))
    	}
    	for i, v := range r.values {
    		switch d := dest[i].(type) {
    		case *uuid.UUID:
    			*d = v.(uuid.UUID)
    		case *string:
    			*d = v.(string)
    		case *time.Time:
    			*d = v.(time.Time)
    		default:
    			return fmt.Errorf("sql: can't scan column %d into %T", i, dest[i])
    		}
    	}
    	return nil
    }

    type memRows struct {
    	rows    []memRow
    	next    int
    	err     error // returned by Err once the rows run out
    	closed  bool
    	current memRow
    }

    func (r *memRows) Next() bool {
    	if r.closed || r.next >= len(r.rows) {
    		return false
    	}
    	r.current = r.rows[r.next]
    	r.next++
    	return true
    }

    func (r *memRows) Scan(dest ...any) error { return r.current.Scan(dest...) }
    func (r *memRows) Err() error             { return r.err }
    func (r *memRows) Close() error           { r.closed = true; return nil }

    func main() {
    	pip := uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	at := time.Date(2026, 9, 1, 9, 30, 0, 0, time.UTC)
    	row := func(id, body string, minutes int) memRow {
    		return memRow{values: []any{uuid.MustParse(id), pip, body, at.Add(time.Duration(minutes) * time.Minute)}}
    	}

    	sq, err := scanSqueak(row("0192f1e3-0000-7000-8000-000000000001", "first squeak!", 0))
    	fmt.Printf("found:   %q %v\n", sq.Body, err)
    	_, err = scanSqueak(memRow{err: sql.ErrNoRows})
    	fmt.Printf("missing: is ErrNotFound? %v\n", errors.Is(err, ErrNotFound))

    	rows := &memRows{rows: []memRow{
    		row("0192f1e3-0000-7000-8000-000000000001", "first squeak!", 0),
    		row("0192f1e3-0000-7000-8000-000000000002", "yes please", 30),
    	}}
    	list, err := collectSqueaks(rows)
    	fmt.Printf("list:    %d squeaks, err %v, closed %v\n", len(list), err, rows.closed)

    	broken := &memRows{rows: rows.rows, err: errors.New("connection reset by peer")}
    	list, err = collectSqueaks(broken)
    	fmt.Printf("broken:  %d squeaks, err %v, closed %v\n", len(list), err, broken.closed)
    }
  solution: |
    package main

    import (
    	"database/sql"
    	"errors"
    	"fmt"
    	"time"
    	"uuid"
    )

    var ErrNotFound = errors.New("not found")

    type Squeak struct {
    	ID        uuid.UUID `json:"id"`
    	AuthorID  uuid.UUID `json:"author_id"`
    	Body      string    `json:"body"`
    	CreatedAt time.Time `json:"created_at"`
    }

    // rowScanner is the method of *sql.Row (and *sql.Rows) that scanSqueak needs.
    type rowScanner interface {
    	Scan(dest ...any) error
    }

    // squeakRows is the part of *sql.Rows that collectSqueaks needs.
    type squeakRows interface {
    	Next() bool
    	Scan(dest ...any) error
    	Err() error
    	Close() error
    }

    // Every query that scanSqueak reads must select exactly these columns, in this order.
    const selectSqueaks = "SELECT id, author_id, body, created_at FROM squeaks"

    // scanSqueak reads one squeak from row. A missing row is ErrNotFound.
    func scanSqueak(row rowScanner) (Squeak, error) {
    	var sq Squeak
    	err := row.Scan(&sq.ID, &sq.AuthorID, &sq.Body, &sq.CreatedAt)
    	if errors.Is(err, sql.ErrNoRows) {
    		return Squeak{}, ErrNotFound
    	}
    	if err != nil {
    		return Squeak{}, fmt.Errorf("scanning squeak: %w", err)
    	}
    	return sq, nil
    }

    // collectSqueaks reads every row, always closes rows, and reports any
    // error that ended the iteration early.
    func collectSqueaks(rows squeakRows) ([]Squeak, error) {
    	defer rows.Close()
    	var list []Squeak
    	for rows.Next() {
    		sq, err := scanSqueak(rows)
    		if err != nil {
    			return nil, err
    		}
    		list = append(list, sq)
    	}
    	if err := rows.Err(); err != nil {
    		return nil, fmt.Errorf("listing squeaks: %w", err)
    	}
    	return list, nil
    }

    // ---- in-memory stand-ins for *sql.Row and *sql.Rows, used by main ----

    type memRow struct {
    	values []any // id, author_id, body, created_at
    	err    error // returned by Scan instead, like sql.ErrNoRows
    }

    func (r memRow) Scan(dest ...any) error {
    	if r.err != nil {
    		return r.err
    	}
    	if len(dest) != len(r.values) {
    		return fmt.Errorf("sql: expected %d destination arguments in Scan, not %d", len(r.values), len(dest))
    	}
    	for i, v := range r.values {
    		switch d := dest[i].(type) {
    		case *uuid.UUID:
    			*d = v.(uuid.UUID)
    		case *string:
    			*d = v.(string)
    		case *time.Time:
    			*d = v.(time.Time)
    		default:
    			return fmt.Errorf("sql: can't scan column %d into %T", i, dest[i])
    		}
    	}
    	return nil
    }

    type memRows struct {
    	rows    []memRow
    	next    int
    	err     error // returned by Err once the rows run out
    	closed  bool
    	current memRow
    }

    func (r *memRows) Next() bool {
    	if r.closed || r.next >= len(r.rows) {
    		return false
    	}
    	r.current = r.rows[r.next]
    	r.next++
    	return true
    }

    func (r *memRows) Scan(dest ...any) error { return r.current.Scan(dest...) }
    func (r *memRows) Err() error             { return r.err }
    func (r *memRows) Close() error           { r.closed = true; return nil }

    func main() {
    	pip := uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	at := time.Date(2026, 9, 1, 9, 30, 0, 0, time.UTC)
    	row := func(id, body string, minutes int) memRow {
    		return memRow{values: []any{uuid.MustParse(id), pip, body, at.Add(time.Duration(minutes) * time.Minute)}}
    	}

    	sq, err := scanSqueak(row("0192f1e3-0000-7000-8000-000000000001", "first squeak!", 0))
    	fmt.Printf("found:   %q %v\n", sq.Body, err)
    	_, err = scanSqueak(memRow{err: sql.ErrNoRows})
    	fmt.Printf("missing: is ErrNotFound? %v\n", errors.Is(err, ErrNotFound))

    	rows := &memRows{rows: []memRow{
    		row("0192f1e3-0000-7000-8000-000000000001", "first squeak!", 0),
    		row("0192f1e3-0000-7000-8000-000000000002", "yes please", 30),
    	}}
    	list, err := collectSqueaks(rows)
    	fmt.Printf("list:    %d squeaks, err %v, closed %v\n", len(list), err, rows.closed)

    	broken := &memRows{rows: rows.rows, err: errors.New("connection reset by peer")}
    	list, err = collectSqueaks(broken)
    	fmt.Printf("broken:  %d squeaks, err %v, closed %v\n", len(list), err, broken.closed)
    }
  tests: |
    package main

    import (
    	"database/sql"
    	"errors"
    	"fmt"
    	"testing"
    	"time"
    	"uuid"
    )

    var (
    	testAuthor = uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	testTime   = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
    )

    // testRow checks the destinations scanSqueak passes, like database/sql does.
    type testRow struct {
    	sq  Squeak
    	err error
    }

    func (r testRow) Scan(dest ...any) error {
    	if r.err != nil {
    		return r.err
    	}
    	if len(dest) != 4 {
    		return fmt.Errorf("Scan got %d destinations; the query selects 4 columns (id, author_id, body, created_at)", len(dest))
    	}
    	id, ok1 := dest[0].(*uuid.UUID)
    	author, ok2 := dest[1].(*uuid.UUID)
    	body, ok3 := dest[2].(*string)
    	created, ok4 := dest[3].(*time.Time)
    	if !ok1 || !ok2 || !ok3 || !ok4 {
    		return fmt.Errorf("Scan destinations are %T, %T, %T, %T; want *uuid.UUID, *uuid.UUID, *string, *time.Time", dest[0], dest[1], dest[2], dest[3])
    	}
    	*id, *author, *body, *created = r.sq.ID, r.sq.AuthorID, r.sq.Body, r.sq.CreatedAt
    	return nil
    }

    type testRows struct {
    	rows   []testRow
    	i      int
    	err    error
    	closed int
    }

    func (r *testRows) Next() bool {
    	if r.closed > 0 || r.i >= len(r.rows) {
    		return false
    	}
    	r.i++
    	return true
    }

    func (r *testRows) Scan(dest ...any) error {
    	return r.rows[r.i-1].Scan(dest...)
    }

    func (r *testRows) Err() error   { return r.err }
    func (r *testRows) Close() error { r.closed++; return nil }

    func squeak(n int) Squeak {
    	return Squeak{
    		ID:        uuid.MustParse(fmt.Sprintf("0192f1e3-0000-7000-8000-%012d", n)),
    		AuthorID:  testAuthor,
    		Body:      fmt.Sprintf("squeak #%d", n),
    		CreatedAt: testTime.Add(time.Duration(n) * time.Minute),
    	}
    }

    func TestScanSqueak(t *testing.T) {
    	want := squeak(1)
    	got, err := scanSqueak(testRow{sq: want})
    	if err != nil {
    		t.Fatalf("scanSqueak(valid row) error = %v", err)
    	}
    	if got != want {
    		t.Errorf("scanSqueak = %+v, want %+v", got, want)
    	}
    }

    func TestScanSqueakNoRows(t *testing.T) {
    	_, err := scanSqueak(testRow{err: sql.ErrNoRows})
    	if !errors.Is(err, ErrNotFound) {
    		t.Fatalf("scanSqueak(sql.ErrNoRows) error = %v, want ErrNotFound so the handler can answer 404", err)
    	}
    	if errors.Is(err, sql.ErrNoRows) {
    		t.Errorf("scanSqueak(sql.ErrNoRows) error still matches sql.ErrNoRows; translate it so handlers never see database errors")
    	}
    }

    func TestScanSqueakOtherError(t *testing.T) {
    	boom := errors.New("database is locked")
    	sq, err := scanSqueak(testRow{err: boom})
    	if !errors.Is(err, boom) {
    		t.Errorf("scanSqueak(failing row) error = %v, want it to wrap %v", err, boom)
    	}
    	if errors.Is(err, ErrNotFound) {
    		t.Errorf("a database failure must not look like ErrNotFound (that would be a 404 instead of a 500)")
    	}
    	if sq != (Squeak{}) {
    		t.Errorf("scanSqueak returned %+v alongside an error, want the zero Squeak", sq)
    	}
    }

    func TestCollectSqueaks(t *testing.T) {
    	rows := &testRows{rows: []testRow{{sq: squeak(1)}, {sq: squeak(2)}, {sq: squeak(3)}}}
    	list, err := collectSqueaks(rows)
    	if err != nil {
    		t.Fatalf("collectSqueaks error = %v", err)
    	}
    	if len(list) != 3 {
    		t.Fatalf("collectSqueaks returned %d squeaks, want 3", len(list))
    	}
    	for i, sq := range list {
    		if sq != squeak(i+1) {
    			t.Errorf("squeak %d = %+v, want %+v", i, sq, squeak(i+1))
    		}
    	}
    	if rows.closed == 0 {
    		t.Error("collectSqueaks didn't call rows.Close(); an unclosed *sql.Rows keeps its connection busy")
    	}
    }

    func TestCollectSqueaksEmpty(t *testing.T) {
    	rows := &testRows{}
    	list, err := collectSqueaks(rows)
    	if err != nil || len(list) != 0 {
    		t.Errorf("collectSqueaks(no rows) = %d squeaks, %v; want 0 squeaks and no error (an empty list isn't \"not found\")", len(list), err)
    	}
    	if rows.closed == 0 {
    		t.Error("collectSqueaks didn't call rows.Close()")
    	}
    }

    func TestCollectSqueaksIterationError(t *testing.T) {
    	boom := errors.New("connection reset by peer")
    	rows := &testRows{rows: []testRow{{sq: squeak(1)}}, err: boom}
    	list, err := collectSqueaks(rows)
    	if !errors.Is(err, boom) {
    		t.Errorf("collectSqueaks error = %v, want it to wrap rows.Err() (%v); Next returning false doesn't mean every row was read", err, boom)
    	}
    	if list != nil {
    		t.Errorf("collectSqueaks returned %d squeaks alongside an error, want nil", len(list))
    	}
    	if rows.closed == 0 {
    		t.Error("collectSqueaks didn't call rows.Close() when iteration failed")
    	}
    }

    func TestCollectSqueaksScanError(t *testing.T) {
    	boom := errors.New("converting column 3: bad timestamp")
    	rows := &testRows{rows: []testRow{{sq: squeak(1)}, {err: boom}, {sq: squeak(3)}}}
    	list, err := collectSqueaks(rows)
    	if !errors.Is(err, boom) {
    		t.Errorf("collectSqueaks error = %v, want it to wrap the Scan error (%v)", err, boom)
    	}
    	if list != nil {
    		t.Errorf("collectSqueaks returned %d squeaks alongside an error, want nil", len(list))
    	}
    	if rows.closed == 0 {
    		t.Error("collectSqueaks didn't call rows.Close() after a Scan error; use defer")
    	}
    }
---

Go talks to SQL databases through one standard-library package, `database/sql`. It
knows nothing about any particular database. A **driver** package does the talking,
and `database/sql` gives you the same API whichever driver you use.

## Opening a database

Drivers are third-party modules. For SQLite, a popular pure-Go one is
`modernc.org/sqlite`. PostgreSQL users often pick `github.com/jackc/pgx`. You import the
driver only for its side effect of registering itself, then open it by name:

```go
import (
	"database/sql"

	_ "modernc.org/sqlite" // registers the "sqlite" driver
)

db, err := sql.Open("sqlite", "file:squeak.db?_pragma=foreign_keys(1)")
if err != nil {
	return err
}
defer db.Close()

if err := db.PingContext(ctx); err != nil {
	return fmt.Errorf("connecting to database: %w", err)
}
```

(The `_pragma` part switches on SQLite's foreign key checks, which are off by default.)

Two things surprise people:

- **`sql.Open` doesn't connect.** It only checks its arguments. `PingContext` is what
  proves the database is actually there, so call it at startup.
- **`*sql.DB` is a connection pool**, not a connection. It's safe for concurrent use,
  opens connections as requests need them, and reuses them afterwards. Open it **once**
  at startup and share it, exactly like Squeak's `MemoryStore`.

The exercises in this course only use the standard library, so they can't import a
driver. The code in this lesson was tested with `modernc.org/sqlite`, and you can run it
in your own project after `go get modernc.org/sqlite`.

## The three methods you'll use most

All of them take a `context.Context` first, so a cancelled request (the client hung up)
also cancels its query. Pass `r.Context()` from the handler all the way down, which is
why `SqueakStore`'s methods took a `ctx` from day one.

| Method | For | Returns |
|---|---|---|
| `ExecContext` | `INSERT`, `UPDATE`, `DELETE` | `sql.Result` (rows affected) |
| `QueryRowContext` | a `SELECT` that returns at most one row | `*sql.Row` |
| `QueryContext` | a `SELECT` that returns many rows | `*sql.Rows` |

Go 1.27's `database/sql` understands the new `uuid.UUID` type: pass one as an argument
and it's sent as its string form, and `Scan` can read a UUID column straight into a
`*uuid.UUID`. `time.Time` works in both directions too.

## A SQL-backed SqueakStore

Here's most of a `SqueakStore` backed by SQL. It satisfies the same interface as
`MemoryStore`, so none of Squeak's handlers change.

```go
type SQLStore struct {
	db *sql.DB
}

const selectSqueaks = "SELECT id, author_id, body, created_at FROM squeaks"

func (s *SQLStore) CreateSqueak(ctx context.Context, authorID uuid.UUID, body string) (Squeak, error) {
	sq := Squeak{ID: uuid.NewV7(), AuthorID: authorID, Body: body, CreatedAt: time.Now().UTC()}
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO squeaks (id, author_id, body, created_at) VALUES (?, ?, ?, ?)",
		sq.ID, sq.AuthorID, sq.Body, sq.CreatedAt)
	if err != nil {
		return Squeak{}, fmt.Errorf("inserting squeak: %w", err)
	}
	return sq, nil
}

func (s *SQLStore) GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error) {
	return scanSqueak(s.db.QueryRowContext(ctx, selectSqueaks+" WHERE id = ?", id))
}

func (s *SQLStore) ListSqueaks(ctx context.Context, authorID uuid.UUID) ([]Squeak, error) {
	query, args := selectSqueaks, []any{}
	if authorID != uuid.Nil() {
		query += " WHERE author_id = ?"
		args = append(args, authorID)
	}
	rows, err := s.db.QueryContext(ctx, query+" ORDER BY created_at, id", args...)
	if err != nil {
		return nil, fmt.Errorf("listing squeaks: %w", err)
	}
	return collectSqueaks(rows)
}

func (s *SQLStore) DeleteSqueak(ctx context.Context, id uuid.UUID) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM squeaks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting squeak: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("deleting squeak: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
```

Notice how `DeleteSqueak` finds out whether the squeak existed: a `DELETE` that matches
nothing isn't an error in SQL, so it asks the `sql.Result` how many rows were affected.

## Scanning and its rules

The two helpers, `scanSqueak` and `collectSqueaks`, hold the rules that trip everyone up:

1. **`Scan` destinations match the selected columns, in order.** `Scan` fills pointers
   by position, not by name. That's why every query shares the `selectSqueaks` prefix.
2. **`QueryRowContext` reports errors from `Scan`.** A missing row comes back as
   **`sql.ErrNoRows`** when you call `Scan`. The store translates it into `ErrNotFound`,
   so handlers never learn which database you use.
3. **Close your rows.** A `*sql.Rows` holds one of the pool's connections until it's
   closed. Forget, and the pool keeps opening new connections until the database
   refuses them (or, if you've capped the pool with `db.SetMaxOpenConns`, until every
   query waits forever). `defer rows.Close()` straight after checking the error from
   `QueryContext`.
4. **Check `rows.Err()` after the loop.** `rows.Next()` returns `false` when the rows
   run out *and* when the connection fails halfway through. Only `rows.Err()` can tell
   those apart. Skip it and you'll serve half a timeline as if it were all of it.

`*sql.Row` and `*sql.Rows` both have a `Scan(dest ...any) error` method, so one small
interface lets a single `scanSqueak` read from either, and lets tests feed it fakes.

## Your task

Write the two helpers against small interfaces, so they work with the real `*sql.Row`
and `*sql.Rows` and with the in-memory stand-ins the tests use:

1. **`scanSqueak(row)`** scans `id`, `author_id`, `body` and `created_at` (in that
   order) straight into a `Squeak`'s fields. `sql.ErrNoRows` becomes `ErrNotFound`,
   with no trace of `sql.ErrNoRows` left in it. Any other error is returned wrapped
   (use `%w`) with the zero `Squeak`.
2. **`collectSqueaks(rows)`** always closes `rows` (use `defer`), scans every row with
   `scanSqueak`, stops at the first error, and after the loop returns `rows.Err()`
   (wrapped) if it isn't nil. When there's an error, return a nil slice with it.

No rows at all is not an error: a mouse who hasn't squeaked has an empty profile,
not a missing one.

**Run** shows the starter's bugs: a missing squeak that isn't `ErrNotFound`, rows that
never get closed, and a broken connection that looks like success.

Further reading: [Accessing relational databases](https://go.dev/doc/database/) in the
Go documentation.
