---
title: Relational Databases
quiz:
  - question: |
      With foreign keys enforced, what happens here?

      ```sql
      CREATE TABLE squeaks (
          id        TEXT PRIMARY KEY,
          author_id TEXT NOT NULL REFERENCES users (id),
          body      TEXT NOT NULL
      );
      INSERT INTO squeaks (id, author_id, body)
      VALUES ('s1', 'no-such-user', 'orphan squeak');
      ```
    options:
      - text: The row is inserted, and `author_id` is set to `NULL`
      - text: The row is inserted, and a matching user is created automatically
      - text: 'The insert fails with a foreign key constraint error'
        correct: true
      - text: The insert succeeds, but `SELECT` won't return the row
    explanation: |
      `REFERENCES users (id)` promises that every `author_id` matches an existing user.
      The database checks that on every insert and update and rejects the row that
      would break it. (SQLite only enforces foreign keys after
      `PRAGMA foreign_keys = ON`. PostgreSQL always does.)
  - question: |
      Squeak's register handler already calls `GetUserByEmail` and answers 409 when the
      email is taken. Why still put `UNIQUE` on the `email` column?
    options:
      - text: '`UNIQUE` makes lookups by email case-insensitive'
      - text: Two sign-ups at the same moment can both pass the check before either inserts; only the database can reject the second one atomically
        correct: true
      - text: Without it, the database stores emails in random order
      - text: It's required for every `TEXT` column
    explanation: |
      Check-then-insert in application code is a race, just like the map check you
      protected with a mutex in chapter 5. But now there may be several Squeak servers,
      so no mutex covers them all. A constraint is checked by the database itself, so
      exactly one insert wins and the other gets an error you can turn into a 409.
  - question: What is a table's **primary key**?
    options:
      - text: The first column in the table, whatever it holds
      - text: A column (or set of columns) whose value is unique and never null, identifying exactly one row
        correct: true
      - text: The password used to open the table
      - text: A column that other tables are not allowed to reference
    explanation: |
      The primary key is a row's identity. Squeak uses its UUIDs, so `id TEXT PRIMARY KEY`
      means no two squeaks can share an ID. Foreign keys in other tables point at it.
---

Squeak's `MemoryStore` has served you well, and it has one fatal flaw: restart the
server and every squeak is gone. Real apps keep their data in a **database**, a separate
program built to store data safely on disk, let many clients use it at once, and answer
questions about it quickly.

The most common kind is the **relational database**: PostgreSQL, MySQL, SQLite and
friends. You talk to all of them in **SQL** (Structured Query Language), which is old,
everywhere, and well worth learning properly.

## Tables, rows and columns

A relational database stores data in **tables**. A table is like a spreadsheet with
strict rules: every **row** is one record, and every **column** has a name and a type.

```
squeaks
 id  | author_id  | body            | created_at
-----+------------+-----------------+---------------------
 s1  | u-pip      | first squeak!   | 2026-09-01 09:30:00
 s2  | u-whiskers | cheese at noon? | 2026-09-01 09:45:00
 s3  | u-pip      | yes please      | 2026-09-01 10:00:00
```

That's Squeak's `[]Squeak`, laid out flat. (Real IDs are UUIDs; they're shortened here
so the tables fit on the page.)

## Schemas

The **schema** is the description of your tables: their names, columns, types and
rules. You create it with `CREATE TABLE`:

```sql
CREATE TABLE users (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMP NOT NULL
);

CREATE TABLE squeaks (
    id         TEXT PRIMARY KEY,
    author_id  TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    body       TEXT NOT NULL CHECK (length(body) BETWEEN 1 AND 140),
    created_at TIMESTAMP NOT NULL
);
```

SQL keywords aren't case-sensitive. Writing them in capitals is a convention that makes
queries easier to scan. `--` starts a comment.

## Constraints: rules the database enforces

The words after each type are **constraints**, and they're the reason to love a
database. Your Go code can have bugs. The constraints hold anyway.

- **`PRIMARY KEY`**: this column identifies the row. Unique, and never empty.
- **`NOT NULL`**: the column must have a value. SQL's `NULL` means "no value", a bit
  like Go's `nil`. Without `NOT NULL`, any column may hold it.
- **`UNIQUE`**: no two rows may share this value. Two mice can't register
  `pip@squeak.dev`, even if two sign-up requests arrive in the same millisecond.
- **`REFERENCES users (id)`**: a **foreign key**. Every squeak's `author_id` must be
  the `id` of a real user. `ON DELETE CASCADE` says "when a user is deleted, delete
  their squeaks too".
- **`CHECK (...)`**: any condition you like. Here, squeaks are 1 to 140 characters.

Break a rule and the statement fails with an error such as
`UNIQUE constraint failed: users.email`. Your store turns that into `ErrEmailTaken`, and
the handler answers 409.

## Relations

Notice that a squeak doesn't *contain* its author. It holds the author's ID, and the
user lives in exactly one place, the `users` table. That's the "relational" part: data
is split into tables that point at each other through keys, so each fact is stored once.
Change Pip's email and every squeak "sees" the new one, because none of them copied it.
You'll put the pieces back together with a `JOIN` in a couple of lessons.

## Which database?

- **SQLite** is a library, not a server: the whole database is a single file, and your
  program reads and writes it directly. It's brilliant for small apps, tests and tools.
- **PostgreSQL** is a server that many app instances connect to over the network. It's
  the usual choice for a web API that runs on several machines.

The SQL in this chapter works in both, apart from small differences that are pointed
out along the way. Go talks to either through the same standard-library package,
`database/sql`, plus a **driver** for the specific database.

Squeak's handlers won't notice any of this. They talk to the `SqueakStore` interface
from chapter 5, and by the end of this chapter you'll know how to build a SQL-backed
implementation of it.

Further reading: [SQLite's SQL reference](https://www.sqlite.org/lang.html) and the
[PostgreSQL tutorial](https://www.postgresql.org/docs/current/tutorial.html).
