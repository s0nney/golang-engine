---
title: SQL Injection and Placeholders
quiz:
  - question: |
      A login lookup is built like this, and someone submits the email
      `x' OR '1'='1`. What SQL does the database run?

      ```go
      query := "SELECT id FROM users WHERE email = '" + email + "'"
      ```
    options:
      - text: '`SELECT id FROM users WHERE email = ''x'' OR ''1''=''1''`, which matches every user'
        correct: true
      - text: A query looking for a user whose email is literally `x' OR '1'='1`
      - text: Nothing, because Go escapes quotes in string concatenation
      - text: A syntax error, so it's harmless
    explanation: |
      The quote in the input closes the string literal early, and the rest becomes SQL.
      `'1'='1'` is always true, so the `WHERE` matches every row. That's SQL injection.
      With a placeholder (`WHERE email = ?`), the whole input is one value and can never
      become SQL.
  - question: Which of these can **not** be passed as a `?` placeholder argument?
    options:
      - text: A search string typed by the user
      - text: The user ID from the access token
      - text: The name of the column to sort by, taken from `?sort=` in the URL
        correct: true
      - text: The page size from `?limit=`
    explanation: |
      Placeholders stand for *values* only. Table names, column names, `ASC`/`DESC` and
      other keywords are part of the SQL itself. When they depend on input, choose them
      from a fixed allowlist in Go (for example a `map[string]string` from `"newest"` to
      `"created_at DESC"`), and never paste the input itself.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    	"time"
    	"uuid"
    )

    const (
    	defaultLimit = 20
    	maxLimit     = 100
    )

    // ListFilter describes which squeaks GET /api/squeaks should return.
    type ListFilter struct {
    	AuthorID uuid.UUID // uuid.Nil(): squeaks by every author
    	Search   string    // "": no search; otherwise the body must contain it
    	Before   time.Time // zero: no cursor; otherwise only squeaks created before it
    	Limit    int       // <= 0: defaultLimit; capped at maxLimit
    }

    // escapeLike escapes LIKE's wildcards (% and _) and the escape character
    // itself, so user input matches literally.
    func escapeLike(s string) string {
    	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
    	return r.Replace(s)
    }

    // buildListQuery returns the SQL and the arguments for its ? placeholders.
    // This version pastes user input straight into the SQL. Fix it.
    func buildListQuery(f ListFilter) (string, []any) {
    	var where []string
    	if f.AuthorID != uuid.Nil() {
    		where = append(where, fmt.Sprintf("author_id = '%s'", f.AuthorID))
    	}
    	if f.Search != "" {
    		where = append(where, fmt.Sprintf("body LIKE '%%%s%%'", f.Search))
    	}
    	if !f.Before.IsZero() {
    		where = append(where, fmt.Sprintf("created_at < '%s'", f.Before.Format(time.RFC3339Nano)))
    	}

    	query := "SELECT id, author_id, body, created_at FROM squeaks"
    	if len(where) > 0 {
    		query += " WHERE " + strings.Join(where, " AND ")
    	}
    	query += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT %d", f.Limit)
    	return query, nil
    }

    func main() {
    	pip := uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	for _, f := range []ListFilter{
    		{},
    		{AuthorID: pip, Limit: 5},
    		{Search: "cheese"},
    		{Search: "'; DROP TABLE squeaks; --"},
    	} {
    		query, args := buildListQuery(f)
    		fmt.Println(query)
    		fmt.Printf("  args: %v\n", args)
    	}
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    	"time"
    	"uuid"
    )

    const (
    	defaultLimit = 20
    	maxLimit     = 100
    )

    // ListFilter describes which squeaks GET /api/squeaks should return.
    type ListFilter struct {
    	AuthorID uuid.UUID // uuid.Nil(): squeaks by every author
    	Search   string    // "": no search; otherwise the body must contain it
    	Before   time.Time // zero: no cursor; otherwise only squeaks created before it
    	Limit    int       // <= 0: defaultLimit; capped at maxLimit
    }

    // escapeLike escapes LIKE's wildcards (% and _) and the escape character
    // itself, so user input matches literally.
    func escapeLike(s string) string {
    	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
    	return r.Replace(s)
    }

    // buildListQuery returns the SQL and the arguments for its ? placeholders.
    func buildListQuery(f ListFilter) (string, []any) {
    	var where []string
    	var args []any
    	if f.AuthorID != uuid.Nil() {
    		where = append(where, "author_id = ?")
    		args = append(args, f.AuthorID)
    	}
    	if f.Search != "" {
    		where = append(where, `body LIKE ? ESCAPE '\'`)
    		args = append(args, "%"+escapeLike(f.Search)+"%")
    	}
    	if !f.Before.IsZero() {
    		where = append(where, "created_at < ?")
    		args = append(args, f.Before)
    	}

    	query := "SELECT id, author_id, body, created_at FROM squeaks"
    	if len(where) > 0 {
    		query += " WHERE " + strings.Join(where, " AND ")
    	}
    	query += " ORDER BY created_at DESC, id DESC LIMIT ?"

    	limit := f.Limit
    	if limit <= 0 {
    		limit = defaultLimit
    	}
    	args = append(args, min(limit, maxLimit))
    	return query, args
    }

    func main() {
    	pip := uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	for _, f := range []ListFilter{
    		{},
    		{AuthorID: pip, Limit: 5},
    		{Search: "cheese"},
    		{Search: "'; DROP TABLE squeaks; --"},
    	} {
    		query, args := buildListQuery(f)
    		fmt.Println(query)
    		fmt.Printf("  args: %v\n", args)
    	}
    }
  tests: |
    package main

    import (
    	"fmt"
    	"reflect"
    	"strings"
    	"testing"
    	"time"
    	"uuid"
    )

    var (
    	testAuthor = uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	testBefore = time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
    )

    const (
    	base  = "SELECT id, author_id, body, created_at FROM squeaks"
    	order = " ORDER BY created_at DESC, id DESC LIMIT ?"
    	like  = `body LIKE ? ESCAPE '\'`
    )

    // squash collapses runs of whitespace, so spacing differences don't matter.
    func squash(s string) string { return strings.Join(strings.Fields(s), " ") }

    func TestBuildListQuery(t *testing.T) {
    	for _, tt := range []struct {
    		name      string
    		filter    ListFilter
    		wantQuery string
    		wantArgs  []any
    	}{
    		{"no filters", ListFilter{}, base + order, []any{20}},
    		{"author", ListFilter{AuthorID: testAuthor, Limit: 5},
    			base + " WHERE author_id = ?" + order, []any{testAuthor, 5}},
    		{"search", ListFilter{Search: "cheese"},
    			base + " WHERE " + like + order, []any{"%cheese%", 20}},
    		{"before", ListFilter{Before: testBefore, Limit: 50},
    			base + " WHERE created_at < ?" + order, []any{testBefore, 50}},
    		{"everything", ListFilter{AuthorID: testAuthor, Search: "brie", Before: testBefore, Limit: 10},
    			base + " WHERE author_id = ? AND " + like + " AND created_at < ?" + order,
    			[]any{testAuthor, "%brie%", testBefore, 10}},
    		{"limit too big", ListFilter{Limit: 5000}, base + order, []any{100}},
    		{"negative limit", ListFilter{Limit: -3}, base + order, []any{20}},
    		{"wildcards in search", ListFilter{Search: "100%_off"},
    			base + " WHERE " + like + order, []any{`%100\%\_off%`, 20}},
    	} {
    		t.Run(tt.name, func(t *testing.T) {
    			query, args := buildListQuery(tt.filter)
    			if squash(query) != squash(tt.wantQuery) {
    				t.Errorf("buildListQuery(%+v) query:\n got  %s\n want %s", tt.filter, query, tt.wantQuery)
    			}
    			if !reflect.DeepEqual(args, tt.wantArgs) {
    				t.Errorf("buildListQuery(%+v) args = %#v, want %#v", tt.filter, args, tt.wantArgs)
    			}
    		})
    	}
    }

    func TestNoUserInputInSQL(t *testing.T) {
    	for _, evil := range []string{
    		"'; DROP TABLE squeaks; --",
    		"x' OR '1'='1",
    		`nibble" OR 1=1 --`,
    	} {
    		query, args := buildListQuery(ListFilter{AuthorID: testAuthor, Search: evil, Before: testBefore})
    		for _, bad := range []string{evil, testAuthor.String(), "2026"} {
    			if strings.Contains(query, bad) {
    				t.Errorf("query contains user input %q; pass it as an argument instead:\n%s", bad, query)
    			}
    		}
    		if n := strings.Count(query, "?"); n != len(args) {
    			t.Errorf("query has %d placeholders but %d args:\n%s\nargs: %s", n, len(args), query, fmt.Sprint(args...))
    		}
    	}
    }
---

Squeak's timeline endpoint is growing options: `GET /api/squeaks?author=...&q=cheese&before=...&limit=20`.
Each one adds a condition to the SQL. The obvious way to build it is to glue strings
together. The obvious way is also how databases get stolen.

## SQL injection

```go
query := "SELECT id, body FROM squeaks WHERE body LIKE '%" + search + "%'"
```

Try it with `search` set to `cheese`. Fine. Now set it to:

```
'; DROP TABLE squeaks; --
```

and the database receives:

```sql
SELECT id, body FROM squeaks WHERE body LIKE '%'; DROP TABLE squeaks; --%'
```

The quote in the input *ended the string*, the semicolon ended the statement, a second
statement deletes the table, and `--` comments out the leftovers. The user's text
escaped from being data and became **code**. That's **SQL injection**, and it has been
near the top of every web security list for over twenty years.

Quieter versions are worse, because nobody notices: `x' OR '1'='1` turns a lookup for one
row into a query that matches all of them, and a `UNION SELECT` can read other tables.

Escaping quotes by hand is not the fix. There are too many edge cases (backslashes,
encodings, different databases). The fix is to never mix code and data in the first place.

## Placeholders

Write the SQL with **placeholders** where values go, and pass the values separately:

```go
rows, err := db.QueryContext(ctx,
	"SELECT id, body FROM squeaks WHERE author_id = ? AND body LIKE ?",
	authorID, "%"+search+"%")
```

The SQL text and the values travel to the database separately. The database parses the
SQL first, so its structure is fixed before any value arrives, and each value is only
ever treated as a value. `'; DROP TABLE squeaks; --` is just a strange thing to search for.

The placeholder syntax depends on the database: SQLite and MySQL use `?`, and PostgreSQL
uses numbered ones: `$1`, `$2`. Everything else works the same.

The rule: **every value that comes from outside your code goes through a placeholder.**
Request bodies, path values, query parameters, headers, even data you read back from
the database. No exceptions for "it's just a number" or "it's a UUID, it can't contain
quotes", because the next refactor might change that.

## What placeholders can't do

Placeholders stand for values, never for pieces of SQL: not table names, not column
names, not `ASC` or `DESC`. If a client may choose the sort order, map their choice
through an **allowlist**:

```go
var sortOrders = map[string]string{
	"newest": "created_at DESC, id DESC",
	"oldest": "created_at, id",
}

order, ok := sortOrders[r.URL.Query().Get("sort")]
if !ok {
	order = sortOrders["newest"]
}
query += " ORDER BY " + order // only ever one of our own strings
```

Concatenating SQL is fine when every piece is a constant from your own code. The danger is
only ever *input*.

## LIKE has its own wildcards

Placeholders stop injection, but inside a `LIKE` pattern, `%` and `_` are still
wildcards. A search for `100%` would match "100 percent" too. To search literally,
escape them and tell the database which escape character you used:

```go
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

// ... WHERE body LIKE ? ESCAPE '\'   with the argument "%" + escapeLike(search) + "%"
```

## Your task

`buildListQuery` turns a `ListFilter` into SQL plus its arguments, ready for
`db.QueryContext(ctx, query, args...)`. The starter pastes the values into the SQL with
`fmt.Sprintf`, so **Run** shows the `DROP TABLE` trick working. Rewrite it with
placeholders so it returns exactly this shape:

```sql
SELECT id, author_id, body, created_at FROM squeaks
WHERE author_id = ? AND body LIKE ? ESCAPE '\' AND created_at < ?
ORDER BY created_at DESC, id DESC LIMIT ?
```

- Only include the conditions that apply (a nil `AuthorID`, empty `Search` or zero
  `Before` means "don't filter on this"), in the order shown, joined with `AND`. Leave
  out `WHERE` entirely when there are none. (The tests ignore differences in spacing.)
- The arguments go in the same order as their placeholders:
  - the `uuid.UUID` itself for the author (Go 1.27's `database/sql` knows how to send
    a UUID),
  - `"%" + escapeLike(f.Search) + "%"` for the search,
  - the `time.Time` for `Before`,
  - and the limit, as an `int`, **last**.
- A `Limit` of 0 or less means `defaultLimit`, and anything above `maxLimit` is capped
  at `maxLimit`. The limit is a placeholder too.

The tests check the SQL and the arguments for a table of filters, and then throw hostile
input at it: none of it may appear in the SQL, and the number of `?`s must match the
number of arguments.

Further reading: the Go docs on
[avoiding SQL injection risk](https://go.dev/doc/database/sql-injection).
