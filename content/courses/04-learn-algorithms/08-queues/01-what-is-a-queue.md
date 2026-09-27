---
title: What Is a Queue?
quiz:
  - question: You enqueue `"ava"`, `"bo"` and `"cy"`, then dequeue once. What comes out?
    options:
      - text: '`"ava"`'
        correct: true
      - text: '`"cy"`'
      - text: '`"bo"`'
    explanation: |
      A queue is first in, first out. `"ava"` joined first, so she's at the
      front and leaves first. A stack would have given you `"cy"`.
  - question: Which Clout feature needs a queue rather than a stack?
    options:
      - text: Undoing the last edit to a campaign
      - text: Processing follower-count refresh jobs in the order they were requested
        correct: true
      - text: Checking that a template's brackets are balanced
    explanation: |
      Fairness means the oldest request goes first: first in, first out. Undo and
      bracket matching both need the *newest* item first, which is a stack.
---

Every night Clout refreshes follower counts by calling each social network's
API. Refresh requests pour in from all over the app: new signups, brands
checking a creator, the nightly batch. The API only lets us make a few calls
per second, so requests have to **wait their turn**. And it should be a fair
turn: whoever asked first gets served first.

That's a **queue**.

## First in, first out

A queue is a line at a coffee shop. New people join at the **back**. The
barista serves whoever is at the **front**. The first person to arrive is the
first to leave: **FIFO**, first in, first out.

| operation | what it does | ideal cost |
|-----------|--------------|------------|
| **enqueue** (push) | add an item at the back | O(1) |
| **dequeue** (pop) | remove and return the item at the front | O(1) |
| **peek** | look at the front item | O(1) |
| **size** / **is empty** | how many are waiting | O(1) |

Compare it with a stack:

```
stack (LIFO): push 1, 2, 3 → pop gives 3, 2, 1
queue (FIFO): enqueue 1, 2, 3 → dequeue gives 1, 2, 3
```

Same inputs, opposite order out. Picking the wrong one is a real bug: a
"queue" of refresh jobs that behaved like a stack would keep serving the
newest requests, and an unlucky old request might wait forever.

## Where queues show up

- **Job and task processing**: background workers take jobs from a queue.
- **Rate limiting**: requests wait in line until there's capacity.
- **Message systems**: chat messages, notifications and event streams are
  delivered in order.
- **Buffers**: keyboard input, network packets and audio samples arrive faster
  than they're processed, so they queue up.
- **Breadth-first search**: exploring a follower graph level by level (friends,
  then friends of friends) uses a queue, where depth-first used a stack.

## The implementation challenge

A stack was easy to build on a slice because both push and pop happen at the
same end, and the end of a slice is cheap to change.

A queue works at **both** ends: add at the back, remove from the front. Adding
to the back of a slice is cheap. But removing from the *front* of a slice?
That's where things get interesting.

Over the next three lessons we'll build:

1. A **naive slice queue**, and find out why its dequeue is O(n).
2. A **ring buffer**, which makes both ends O(1).
3. A quick look at **buffered channels**, Go's built-in concurrent queue.

And in the linked-lists chapter, a fourth way: a queue built from nodes and
pointers.
