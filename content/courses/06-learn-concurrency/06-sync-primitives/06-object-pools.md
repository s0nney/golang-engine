---
title: Reusing Temporary Objects with sync.Pool
quiz:
- question: After Put(buf), which assumption is safe?
  options:
  - text: The next Get returns buf
  - text: The pool keeps buf until the process exits
  - text: The caller must stop using buf, and a later Get may allocate another object
    correct: true
  explanation: Pool entries may disappear at any time. After returning an object, another goroutine may own it. Correctness must not depend on reuse.
---

Dispatchly formats thousands of delivery receipts. If profiling shows temporary
buffers dominating allocation, reusing them may reduce garbage-collection work.
`sync.Pool` is designed for these disposable scratch objects.

## Borrow, use, return

Here's a complete example. It always prints `A1: dispatched`, regardless of whether
the pool reuses a buffer:

```go
package main

import (
	"bytes"
	"fmt"
	"sync"
)

var receiptBuffers = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func receipt(id string) string {
	buf := receiptBuffers.Get().(*bytes.Buffer)
	buf.Reset()
	defer receiptBuffers.Put(buf)
	buf.WriteString(id)
	buf.WriteString(": dispatched")
	return buf.String()
}

func main() {
	fmt.Println(receipt("A1"))
}
```

`Get` obtains an available object or calls `New` when necessary. `New` is optional;
without it an empty pool can return nil. Set it during initialization and don't change
it while other goroutines call `Get`. Returning pointers avoids boxing whole buffer
values and preserves the intended ownership model.

The pool itself is safe for concurrent access. A borrowed buffer is still owned by
one caller and isn't automatically safe for concurrent use. After `Put`, relinquish
all access, including reads through aliases. Returning `buf.Bytes()` here would expose
the pooled backing array; a subsequent receipt could overwrite it. `buf.String()`
provides a string whose contents remain valid independently of future buffer writes.

## Reuse is optional

The runtime may remove pool entries at any time. Never store the only copy of a menu,
a job, or an application record in a pool. Don't assert that a particular `Get` returns
the last pointer passed to `Put`, and don't assert an exact allocation count across
machines or garbage-collection cycles.

A pool has no fixed capacity and doesn't limit concurrent work. A semaphore limits
admission; a pool reduces allocation when reuse happens. Database connections need
explicit lifetime and capacity management, so this isn't a connection pool.

## Measure before adding one

Compare representative benchmarks with `-benchmem`, then inspect a heap profile. A pool
can retain oversized buffers after an unusually large receipt; consider discarding
buffers above a chosen capacity instead of returning them. Resetting a buffer also
doesn't securely erase its backing memory, so avoid pooling sensitive data without
understanding its lifetime. Start with straightforward allocation until measurements
show that reuse solves a real cost.

Further reading: [sync.Pool](https://pkg.go.dev/sync#Pool).
