---
title: Choosing a Pipeline Error Contract
quiz:
- question: A stage returns separate unbuffered value and error channels. The caller drains all values before reading errors. What can happen?
  options:
  - text: The runtime merges the channels
  - text: The stage can block sending an error while the caller waits for more values
    correct: true
  - text: Errors are dropped automatically
  explanation: The producer cannot continue or close the value channel until its error send completes. Consume both channels concurrently or choose a different reporting contract.
---

Dispatchly imports menus from several restaurants. One unavailable restaurant might
invalidate the entire import, or you might keep the other menus and report partial
failure. Decide that policy before choosing your channel types.

## Stop on the first failure

For all-or-nothing work, use the error group you built earlier. A failing stage records
an error and cancels the shared context. Every upstream producer and downstream stage
must observe cancellation at blocking sends and receives. Returning an error doesn't
magically stop another goroutine blocked on a send.

Another possible API returns a stream and a buffered terminal-error channel. The stage
reports exactly one final error, possibly nil, and closes its stream. A one-slot buffer
lets that terminal report complete even when the caller is still draining values.
The caller still needs cancellation if it abandons the stream early.

## Keep a result for each input

For independent restaurants, pair each value with its error:

```go
type MenuResult struct {
	Restaurant string
	Items      int
	Err        error
}

func summarize(in <-chan MenuResult) (int, error) {
	items := 0
	var failures []error
	for result := range in {
		if result.Err != nil {
			failures = append(failures,
				fmt.Errorf("%s: %w", result.Restaurant, result.Err))
			continue
		}
		items += result.Items
	}
	return items, errors.Join(failures...)
}
```

This fragment uses `fmt` and `errors`. Each input produces either a success or failure
record, and the restaurant name preserves identity even when workers finish out of
order. The caller receives useful partial results alongside a combined error. Document
that explicitly so callers don't throw away successful work whenever `err != nil`.

For very large imports, retaining every error can itself consume excessive memory.
Choose a bounded error summary or stream errors to a sink when appropriate.

## Separate value and error streams

Independent channels can work when different components consume successes and failures.
However, a single consumer must usually `select` over both. When one closes, set its
local channel variable to nil and continue draining the other. Remember the nil-channel
lesson: repeatedly receiving from an already closed channel creates a busy loop.

Define who closes each channel. With several workers, a coordinator can wait for all
senders before closing their shared outputs. Receivers mustn't close a channel while
workers may still send to it.

Choose the simplest contract that matches the job: cancellation for an unusable batch,
result records for per-item outcomes, or separate streams when consumers truly differ.
In every case, test a failure while another stage is blocked and verify every goroutine
can exit.

Further reading: [Go blog: Pipelines and cancellation](https://go.dev/blog/pipelines).
