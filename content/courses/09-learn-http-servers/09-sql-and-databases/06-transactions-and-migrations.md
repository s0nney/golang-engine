---
title: Transactions and Migrations
quiz:
  - question: |
      `deleteUser` deletes Pip's squeaks inside a transaction, then the `DELETE FROM users`
      statement fails and the function returns the error. What's left in the database?

      ```go
      tx, err := db.BeginTx(ctx, nil)
      if err != nil {
      	return err
      }
      defer tx.Rollback()
      if _, err := tx.ExecContext(ctx, "DELETE FROM squeaks WHERE author_id = ?", id); err != nil {
      	return err
      }
      if _, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id); err != nil {
      	return err
      }
      return tx.Commit()
      ```
    options:
      - text: Pip's user row, but none of Pip's squeaks
      - text: Pip's user row *and* all of Pip's squeaks, exactly as before
        correct: true
      - text: Neither; both deletes already happened
      - text: It depends on which goroutine gets there first
    explanation: |
      Nothing inside a transaction is permanent until `Commit`. The early `return` runs
      the deferred `tx.Rollback()`, which undoes the squeak deletes too. All or nothing.
      (After a successful `Commit`, the deferred `Rollback` does nothing and returns
      `sql.ErrTxDone`, which is safe to ignore.)
  - question: Inside a transaction, which call is a bug?
    options:
      - text: '`tx.ExecContext(ctx, ...)`'
      - text: '`tx.QueryRowContext(ctx, ...)`'
      - text: '`db.ExecContext(ctx, ...)`'
        correct: true
      - text: '`tx.Commit()`'
    explanation: |
      `db` hands the statement to *any* connection in the pool, outside the transaction.
      It won't be rolled back with the rest, and it may not even see the transaction's
      uncommitted changes. Every statement that belongs to the transaction must go
      through `tx`.
  - question: Migration `0003_add_edited_at.sql` has already run in production. You spot a typo in it. What should you do?
    options:
      - text: Fix the typo in `0003` and deploy again
      - text: Write a new migration, `0004`, that corrects it
        correct: true
      - text: Delete the `0003` file
      - text: Edit the production database by hand so it matches
    explanation: |
      The migration table records that `0003` has run, so an edited `0003` never runs
      again in production, while fresh databases (tests, new developers) would run the
      new version. The two would silently drift apart. Applied migrations are history.
      Move forward with a new one.
---

Two more tools turn a database from "a place to put rows" into something you can
trust: **transactions**, which make several statements act as one, and **migrations**,
which let the schema change safely over time.

## Transactions

Deleting a Squeak account takes two statements: delete the user's squeaks, then delete
the user. What if the server crashes, or the second statement fails, between them? You'd
be left with a user whose squeaks have vanished. Half-done work like that is how
databases slowly fill with nonsense.

A **transaction** groups statements so they succeed or fail **together**:

```sql
BEGIN;
DELETE FROM squeaks WHERE author_id = 'u-pip';
DELETE FROM users WHERE id = 'u-pip';
COMMIT;   -- or ROLLBACK; to undo everything since BEGIN
```

Until `COMMIT`, the changes are invisible to everyone else. `ROLLBACK` (or a crash, or
a lost connection) throws them all away. Databases promise four things about
transactions, known as **ACID**:

- **Atomic**: all of it happens, or none of it does.
- **Consistent**: constraints hold at the end, or the transaction fails.
- **Isolated**: concurrent transactions don't see each other's half-finished work.
- **Durable**: once `COMMIT` returns, the data survives a crash.

### Transactions in Go

```go
// deleteUser removes a user and everything they wrote, or nothing at all.
func (s *SQLStore) deleteUser(ctx context.Context, userID uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // does nothing once Commit has succeeded

	if _, err := tx.ExecContext(ctx, "DELETE FROM squeaks WHERE author_id = ?", userID); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound // the deferred Rollback undoes the squeak deletes
	}
	return tx.Commit()
}
```

The pattern is always the same:

1. **`BeginTx`** reserves one connection from the pool for this transaction.
2. **`defer tx.Rollback()`** straight away. Every early `return` now undoes the work,
   and after a successful `Commit` it's a harmless no-op.
3. **Use `tx`, never `db`,** for every statement that belongs to the transaction.
   `db.ExecContext` would run on some other connection, outside the transaction.
4. **`tx.Commit()`** last, and return its error: a commit can fail too.

Keep transactions **short**. They hold locks, and other requests may wait on them. Never
make an HTTP call or wait on a user while one is open.

Webhook idempotency from chapter 7 is a classic use: "record that event `evt_1` was
processed" and "extend Pip's Squeak Gold" belong in one transaction, with a `UNIQUE`
constraint on the event ID. A duplicate delivery then fails the insert and rolls the whole
thing back, even if two deliveries arrive at the same moment on two different servers.

## Migrations

Squeak will change. Next month squeaks get an `edited_at` column, and the month after,
a `likes` table. Every copy of the database (production, staging, each developer's
laptop, every test run) must go through the same changes in the same order.

A **migration** is one numbered, append-only step of schema change:

```
migrations/
  0001_create_users.sql
  0002_create_squeaks.sql
  0003_add_edited_at.sql
```

```sql
-- 0003_add_edited_at.sql
ALTER TABLE squeaks ADD COLUMN edited_at TIMESTAMP;
```

The database keeps a small table listing the migrations it has already applied. At
startup (or as a separate deploy step), a migration tool compares that list with the
files and runs the missing ones in order, each inside a transaction where the database
allows it. Popular Go tools include `goose`, `golang-migrate` and `atlas`, and the core of
one is small enough to write yourself with `embed.FS` and a loop.

The rules that keep migrations sane:

- **Never edit a migration that has run anywhere shared.** Write a new one instead.
- **Keep each one small.** One change per file is easy to review and easy to reason about.
- **Plan for the old code.** During a deploy, the old version of Squeak may still be
  running against the new schema. Adding a nullable column is safe. Renaming or dropping
  one needs several steps: add the new column, deploy code that writes both, backfill,
  then drop the old one.
- **Test them.** Run every migration from empty in CI, so a broken one fails there
  rather than in production.

## Squeak on a real database

You now have every piece of a SQL-backed Squeak: a schema, created by migrations; a
`SQLStore` that implements `SqueakStore` with placeholders and the scanning helpers you
wrote; transactions for multi-step changes; and `sql.ErrNoRows` translated into
`ErrNotFound`, so the handlers from chapters 4 to 8 don't change at all. That's the
repository interface paying off.

Keep the fast handler tests that use fakes and `MemoryStore`. Add a smaller set of
**integration tests** that run `SQLStore` against a real database (a temporary SQLite
file, or a PostgreSQL container in CI), because only a real database can tell you that
your SQL is right.

Further reading: [Executing transactions](https://go.dev/doc/database/execute-transactions)
in the Go documentation.
