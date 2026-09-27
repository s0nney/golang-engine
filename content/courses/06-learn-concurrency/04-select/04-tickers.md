---
title: Periodic Work with Tickers
quiz:
  - question: A ticker fires every second, but the loop body handling each tick takes 3 seconds. What happens to the ticks?
    options:
      - text: They queue up without limit and are delivered one after another
      - text: The ticker drops ticks for the slow receiver, so the loop simply runs back to back and no backlog builds up
        correct: true
      - text: The ticker panics because its channel is full
      - text: The ticker slows down to one tick every 3 seconds permanently
    explanation: |
      A ticker never builds a backlog. The docs say it "will adjust the time
      interval or drop ticks to make up for slow receivers". Your loop just
      sees the next tick as soon as it's ready for one.
  - question: What's the difference between `time.After(d)` and `time.Tick(d)`?
    options:
      - text: There is none, they're aliases
      - text: '`time.After` fires once; `time.Tick` fires repeatedly, every `d`'
        correct: true
      - text: '`time.Tick` fires once; `time.After` fires repeatedly'
      - text: '`time.Tick` blocks the caller for `d`'
    explanation: |
      `time.After` returns a timer's channel, which receives a single value.
      `time.Tick` returns a ticker's channel, which keeps receiving a value
      every `d` until the ticker is garbage-collected.
---

Plenty of Dispatchly's work happens on a schedule: poll each courier's location every few seconds, flush metrics every minute, re-check unassigned orders every 30 seconds. A **ticker** delivers a value on a channel at a regular interval, which makes periodic work fit neatly into a `for`/`select` loop.

## time.NewTicker

`time.NewTicker(d)` returns a `*time.Ticker` whose channel `C` receives the current time every `d`:

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	shiftOver := time.After(350 * time.Millisecond)

	pings := 0
	for {
		select {
		case <-ticker.C:
			pings++
			fmt.Println("ping courier location", pings)
		case <-shiftOver:
			fmt.Println("shift over after", pings, "pings")
			return
		}
	}
}
```

```text
ping courier location 1
ping courier location 2
ping courier location 3
shift over after 3 pings
```

The first tick comes after one full interval, not immediately. If you want to act straight away as well, do the work once before the loop.

This loop combines everything so far: a ticker for the periodic work, a timer for the end of the shift, and `select` to react to whichever comes first. Swap `shiftOver` for a done channel or a context and you have the skeleton of almost every background worker in Go.

## Stop and Reset

- `ticker.Stop()` stops future ticks. It does **not** close `C`, so never `range` over a ticker's channel expecting it to end.
- `ticker.Reset(d)` stops the ticker and restarts it with a new period. Handy for backing off: if a courier's phone stops answering, poll every 30 seconds instead of every 5.

Before Go 1.23, a ticker that was never stopped was never garbage-collected, so `defer ticker.Stop()` was mandatory. Since Go 1.23 unreferenced tickers are collected, but stopping one you're finished with is still good hygiene, and it makes your intent obvious.

## time.Tick

`time.Tick(d)` is a shortcut that returns only the channel:

```go
for range time.Tick(30 * time.Second) {
	reassignStaleOrders()
}
```

Since Go 1.23 the `time` docs say there's no reason to prefer `NewTicker` when `Tick` will do. Use `NewTicker` when you need `Stop` or `Reset`. Note that `for range time.Tick(...)` loops forever: there's no way to stop it except returning from the whole function, so it suits "for the life of the program" jobs.

## Slow receivers

If your loop body takes longer than the interval, ticks don't pile up. The ticker **drops ticks** for slow receivers. A 1-second ticker whose handler takes 3 seconds doesn't build a backlog of 2 ticks per round. The loop just gets the next tick as soon as it's ready for one.

That's usually what you want for periodic work: after a slow location sync, you want *one* catch-up sync, not three in a row.

## Tickers vs sleeping in a loop

```go
for {
	syncLocations()
	time.Sleep(5 * time.Second)
}
```

This works, but it drifts: if `syncLocations` takes 2 seconds, it runs every 7. Worse, while it's sleeping the goroutine can't react to anything else, like being told to stop. A ticker keeps a steady rhythm and lives happily inside a `select` with other cases.

## Further reading

- [Go by Example: Tickers](https://gobyexample.com/tickers)
