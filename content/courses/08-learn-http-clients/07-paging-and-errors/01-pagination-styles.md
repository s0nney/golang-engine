---
title: Pagination Styles
quiz:
  - question: '`trackr` is on page 3 (`?offset=40&limit=20`) when someone deletes issue #5, which was on page 1. What happens when it fetches page 4?'
    options:
      - text: Nothing unusual
      - text: One issue is skipped, because everything after the deleted one shifted up by one position
        correct: true
      - text: The server returns an error because the data changed
    explanation: |
      Offsets count positions, not items. After a deletion, the item that was at
      position 60 moves to 59, which you already "passed", so you never see it.
      Cursor pagination doesn't have this problem.
  - question: What should a client do with a cursor like `eyJpZCI6NDJ9`?
    options:
      - text: Decode it (it's base64) and build the next one itself
      - text: Treat it as opaque and send it back exactly as received
        correct: true
      - text: Increment it by the page size
    explanation: |
      A cursor is the server's private bookmark. Its format can change at any time, and
      only the server knows how to interpret it.
---

A project with 12,000 issues isn't sent in one response. That would be slow, huge,
and hard on the server. APIs send data in **pages**, and the client asks for one page
at a time. There are a few common styles.

## Offset and limit

```
GET /issues?offset=0&limit=50     issues 1-50
GET /issues?offset=50&limit=50    issues 51-100
GET /issues?offset=100&limit=50   issues 101-150
```

Or the same idea with page numbers: `?page=3&per_page=50`. The response often includes
a total:

```json
{"issues": [...], "total": 12034}
```

**Pros:** simple, and you can jump straight to page 80 or show "page 3 of 241".

**Cons:**

- **Shifting data.** If issues are added or deleted while you page, items move between
  pages. You skip some or see others twice.
- **Slow deep pages.** For `offset=100000`, many databases have to walk past 100,000
  rows before returning any.

## Cursor-based

The server hands you an opaque **cursor** (a bookmark) with each page, and you send
it back to get the next one:

```
GET /issues?limit=50
  -> {"issues": [...50 issues...], "next_cursor": "eyJpZCI6NTB9"}
GET /issues?limit=50&cursor=eyJpZCI6NTB9
  -> {"issues": [...], "next_cursor": "eyJpZCI6MTAwfQ"}
...
GET /issues?limit=50&cursor=eyJpZCI6MTIwMDB9
  -> {"issues": [...34 issues...], "next_cursor": ""}
```

An empty (or missing, or `null`) cursor means **there are no more pages**.

Under the hood the cursor usually encodes "the last ID you saw", so the server can
query "the next 50 after ID 50" quickly, however deep you are.

**Pros:** stable when data changes, and fast at any depth.
**Cons:** no jumping to page 80, and usually no total count.

**Treat cursors as opaque.** They often look like base64, and you could decode them,
but the format is the server's business and may change without notice. Store them and
send them back, byte for byte.

## Link headers

Some APIs put the navigation in a response **header** instead of the body:

```
Link: <https://api.trackr.dev/issues?page=3>; rel="next",
      <https://api.trackr.dev/issues?page=9>; rel="last"
```

You follow the `next` URL until there isn't one. This works with either offsets or
cursors underneath, and the client doesn't need to know which. Next lesson.

## Which does Trackr use?

Cursors, like most modern APIs:

```go
type issuePage struct {
	Issues     []Issue `json:"issues"`
	NextCursor string  `json:"next_cursor"`
}
```

## The basic loop

Whatever the style, the client loop has the same shape:

```go
cursor := ""
for {
	p, err := fetchPage(ctx, client, baseURL, cursor)
	if err != nil {
		return err
	}
	for _, iss := range p.Issues {
		fmt.Println(iss.ID, iss.Title)
	}
	if p.NextCursor == "" {
		break // last page
	}
	cursor = p.NextCursor
}
```

It works, but the paging logic is tangled up with what you do with each issue.
Every command that lists things would copy this loop. In lesson 3 you'll turn it into
an iterator, so callers can just write `for iss, err := range allIssues(...)`.

## Guard against loops

A buggy server could return the same cursor forever. A defensive client stops if the
next cursor equals the current one, or after a sanity limit on pages.
