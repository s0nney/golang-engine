---
title: Joins and Indexes
quiz:
  - question: |
      Brie is registered but hasn't squeaked yet. Which query lists Brie with a count of `0`?

      ```sql
      -- A
      SELECT u.email, count(s.id) FROM users u
      JOIN squeaks s ON s.author_id = u.id GROUP BY u.id;

      -- B
      SELECT u.email, count(s.id) FROM users u
      LEFT JOIN squeaks s ON s.author_id = u.id GROUP BY u.id;
      ```
    options:
      - text: A
      - text: B
        correct: true
      - text: Both
      - text: Neither; `count` can't return 0
    explanation: |
      A plain (inner) `JOIN` only keeps pairs that match, so a user with no squeaks
      disappears. `LEFT JOIN` keeps every row from the left table and fills the right
      side's columns with `NULL`. `count(s.id)` doesn't count `NULL`s, so Brie gets 0.
  - question: What's the cost of adding an index?
    options:
      - text: Queries that use it become slower
      - text: Every `INSERT`, `UPDATE` and `DELETE` must also update the index, and it takes up storage
        correct: true
      - text: The table can no longer have a primary key
      - text: There's no cost, so index every column
    explanation: |
      An index is a second, sorted copy of some columns that the database keeps in step
      with the table. Reads that use it get much faster, and every write gets a little
      slower. Index the columns your real queries filter and sort by, not everything.
  - question: |
      A timeline handler lists 50 squeaks with one query, then calls
      `GetUser(ctx, sq.AuthorID)` once per squeak to show each author's email. What's
      the problem?
    options:
      - text: It returns the wrong emails
      - text: It makes 51 round trips to the database where one `JOIN` would do
        correct: true
      - text: '`GetUser` can''t be called inside a loop'
      - text: Nothing, databases cache everything
    explanation: |
      This is the **N+1 query problem**: one query for the list, plus one per item.
      Each round trip costs network time, so the page gets slower as it gets longer.
      Fetch the authors in the same query with a `JOIN`.
---

Squeak's timeline shows each squeak with its author's email. The squeak row only stores
`author_id`, and the email lives in `users`. Time to put them back together.

## JOIN

```sql
SELECT u.email, s.body
FROM squeaks s
JOIN users u ON u.id = s.author_id
ORDER BY s.created_at;
```

```
pip@squeak.dev      | first squeak!
whiskers@squeak.dev | cheese at noon?
pip@squeak.dev      | yes please
pip@squeak.dev      | bring crackers
```

`JOIN users u ON u.id = s.author_id` says "for each squeak, find the user whose `id`
matches its `author_id`, and glue the two rows together". The short names `s` and `u`
are **aliases**, so you can write `s.body` rather than `squeaks.body`. You need the
prefixes whenever both tables have a column with the same name, such as `id` and
`created_at` here.

### Inner and left joins

A plain `JOIN` (an **inner join**) only keeps rows that have a partner. That's usually
what you want: every squeak has an author. But ask "how many squeaks has each user
posted?" and users with none silently vanish. A **`LEFT JOIN`** keeps every row from
the left table, and fills in `NULL` where there's no match:

```sql
SELECT u.email, count(s.id) AS squeaks
FROM users u
LEFT JOIN squeaks s ON s.author_id = u.id
GROUP BY u.id
ORDER BY u.email;
```

```
brie@squeak.dev     | 0
pip@squeak.dev      | 3
whiskers@squeak.dev | 1
```

`GROUP BY u.id` squashes each user's rows into one, and `count(s.id)` counts the
non-`NULL` squeak IDs in each group.

### The N+1 problem

Without a join, the timeline handler would list squeaks, then look up each author
separately: 1 query for the list, plus N for the authors. With an in-memory map that's
harmless. With a database, every query is a network round trip, and a 50-squeak page
makes 51 of them. When you see a store call inside a loop, ask whether a `JOIN` could
fetch everything at once.

## Indexes

How does the database run `WHERE author_id = 'u-pip'`? Without help, it reads every row
in the table and checks each one: a **full table scan**. Fine for four squeaks, painful
for forty million.

An **index** is a sorted structure (usually a B-tree) over one or more columns, pointing
back at the rows. It works like the index at the back of a book: jump straight to
"author u-pip" instead of reading every page.

```sql
CREATE INDEX squeaks_author_created ON squeaks (author_id, created_at);
```

This one is a **composite index**, sorted by author and then by time within each author.
It speeds up the profile-page query in both of its parts, finding Pip's squeaks *and*
returning them in time order:

```sql
SELECT body FROM squeaks WHERE author_id = 'u-pip' ORDER BY created_at DESC LIMIT 20;
```

Column order matters. The index helps queries that filter on `author_id`, or on
`author_id` and `created_at`, but not on `created_at` alone, just as a phone book
sorted by surname can't find everyone called "Pip".

A few facts to remember:

- **Primary keys and `UNIQUE` columns get an index automatically.** Looking up a
  squeak by `id` or a user by `email` is already fast.
- **Foreign keys usually don't.** Index `author_id` yourself if you query by it.
- **Indexes cost writes and space.** Each insert updates every index on the table.
  Add them for the queries you actually run.
- **Ask the database.** `EXPLAIN QUERY PLAN` (SQLite) or `EXPLAIN` (PostgreSQL) shows
  whether a query uses an index:

  ```
  sqlite> EXPLAIN QUERY PLAN SELECT body FROM squeaks WHERE author_id = 'u-pip';
  `--SEARCH squeaks USING INDEX squeaks_author_created (author_id=?)
  ```

  `SEARCH ... USING INDEX` is good news. `SCAN squeaks` means it read the whole table.

Further reading: [Use The Index, Luke](https://use-the-index-luke.com/), a free
book about indexes that's much more fun than it sounds.
