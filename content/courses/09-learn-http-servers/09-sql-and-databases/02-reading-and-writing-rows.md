---
title: Reading and Writing Rows
quiz:
  - question: |
      Using the `squeaks` table from this lesson, what does this query return?

      ```sql
      SELECT body FROM squeaks
      WHERE author_id = 'u-pip'
      ORDER BY created_at DESC
      LIMIT 2;
      ```
    options:
      - text: '`first squeak!`, `yes please`'
      - text: '`bring crackers`, `yes please`'
        correct: true
      - text: '`yes please`, `bring crackers`'
      - text: '`cheese at noon?`, `bring crackers`'
    explanation: |
      `WHERE` keeps Pip's three squeaks, `ORDER BY created_at DESC` puts the newest first
      (10:15, 10:00, 09:30), and `LIMIT 2` keeps the first two of those.
  - question: What does `UPDATE squeaks SET body = 'edited';` do?
    options:
      - text: Changes the most recently created squeak
      - text: Fails, because `UPDATE` needs a `WHERE`
      - text: Changes the body of **every** squeak in the table
        correct: true
      - text: Nothing until you add `LIMIT 1`
    explanation: |
      `WHERE` is optional, and without it `UPDATE` and `DELETE` apply to every row. The
      database won't ask "are you sure?". Always write the `WHERE` first.
  - question: |
      None of the squeaks has been edited, so `edited_at` is `NULL` in every row. How many
      rows does this count?

      ```sql
      SELECT count(*) FROM squeaks WHERE edited_at = NULL;
      ```
    options:
      - text: All of them
      - text: '0'
        correct: true
      - text: It's a syntax error
      - text: '1'
    explanation: |
      `NULL` means "unknown", and comparing anything with an unknown gives unknown, not
      true, even `NULL = NULL`. `WHERE` only keeps rows where the condition is true, so
      nothing matches. Test for missing values with `IS NULL` (or `IS NOT NULL`).
---

SQL has four statements for everyday data work, often called **CRUD**: Create
(`INSERT`), Read (`SELECT`), Update (`UPDATE`) and Delete (`DELETE`). Every store method
Squeak has maps onto one of them.

The examples use this data:

```
users                              squeaks
 id         | email                 id | author_id  | body            | created_at
------------+---------------------  ---+------------+-----------------+-----------
 u-pip      | pip@squeak.dev        s1 | u-pip      | first squeak!   | 09:30
 u-whiskers | whiskers@squeak.dev   s2 | u-whiskers | cheese at noon? | 09:45
 u-brie     | brie@squeak.dev       s3 | u-pip      | yes please      | 10:00
                                    s4 | u-pip      | bring crackers  | 10:15
```

## INSERT: CreateSqueak

```sql
INSERT INTO squeaks (id, author_id, body, created_at)
VALUES ('s5', 'u-brie', 'hello, world', '2026-09-01 10:30:00');
```

List the columns, then the values in the same order. Text goes in **single** quotes.
(Double quotes are for names of tables and columns, a classic source of confusion.)

## SELECT: GetSqueak and ListSqueaks

```sql
SELECT id, body FROM squeaks WHERE id = 's1';
```

A `SELECT` reads like a sentence: *select these columns from this table where this is
true*. The pieces come in a fixed order:

```sql
SELECT id, author_id, body, created_at  -- which columns (* means all of them)
FROM squeaks                            -- which table
WHERE author_id = 'u-pip'               -- which rows
ORDER BY created_at, id                 -- in what order
LIMIT 20;                               -- how many at most
```

- **`WHERE`** filters rows. Combine conditions with `AND`, `OR` and `NOT`, and compare
  with `=`, `<>` (not equal), `<`, `>`, `<=`, `>=`. `body LIKE '%cheese%'` matches
  text containing "cheese": `%` is "any characters" and `_` is "exactly one".
- **`ORDER BY`** sorts, ascending by default, or `DESC` for newest first. Without it,
  the database may return rows in **any order**, just like ranging over a Go map. List
  a unique column last (`created_at, id`) so ties always come out the same way.
- **`LIMIT`** caps the number of rows, which is how you build pages of results.

Prefer naming columns over `SELECT *`. Your Go code scans columns by position, so if
someone adds a column to the table, `*` quietly changes what your query returns.

Functions like `count(*)` summarise rows instead of returning them:

```sql
SELECT count(*) FROM squeaks WHERE author_id = 'u-pip';  -- 3
```

## UPDATE: editing a squeak

```sql
UPDATE squeaks
SET body = 'yes please!', edited_at = '2026-09-01 10:05:00'
WHERE id = 's3' AND author_id = 'u-pip';
```

`SET` changes columns, and `WHERE` picks the rows to change. Adding
`author_id = 'u-pip'` to the `WHERE` is an ownership check done by the database: if
Whiskers tries it, **zero rows** match and nothing changes. The database reports how many
rows a statement affected, and your Go code can check that number to answer 403 or 404.

## DELETE: DeleteSqueak

```sql
DELETE FROM squeaks WHERE id = 's4';
```

## The two scariest statements in SQL

```sql
UPDATE squeaks SET body = 'edited';   -- every squeak now says "edited"
DELETE FROM squeaks;                  -- every squeak is gone
```

Both are perfectly valid. Without a `WHERE`, `UPDATE` and `DELETE` affect **every row**,
and there's no undo. Write the `WHERE` clause first, then fill in the rest.

## NULL is not a value

`NULL` means "missing" or "unknown". Comparisons with it are never true:
`edited_at = NULL` matches nothing, and so does `edited_at <> NULL`. Use `IS NULL` and
`IS NOT NULL`:

```sql
SELECT count(*) FROM squeaks WHERE edited_at IS NULL;  -- 4 (nobody has edited yet)
```

In Go, a nullable column needs a type that can say "no value": `sql.Null[time.Time]`,
or a pointer like `*time.Time`. Scanning `NULL` into a plain `time.Time` is an error.
That's a good reason to make columns `NOT NULL` unless "missing" really means something.

## RETURNING

PostgreSQL and SQLite (3.35+) can hand back the rows a statement changed, which saves a
second query:

```sql
UPDATE squeaks SET body = 'yes please!' WHERE id = 's3' RETURNING id, body, created_at;
```

Squeak's stores generate IDs and timestamps in Go, so they mostly don't need it. It's
handy when the *database* fills in a value, such as a default or an auto-incrementing ID.

Further reading: [SQLBolt](https://sqlbolt.com/) has interactive exercises for
everything in this lesson.
