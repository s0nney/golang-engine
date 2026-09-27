---
title: Rendezvous and Barriers
quiz:
- question: Four workers call Done and then Wait on an arrival group initialized to four. When may any worker pass Wait?
  options:
  - text: After its own Done call
  - text: After all four workers have arrived
    correct: true
  - text: After the first worker starts
  explanation: The group reaches zero only after all four arrivals. Use a separate group to wait for the workers to finish the phase after the barrier.
---

Dispatchly is preparing a delivery simulation. Each worker loads one zone, but no
worker may begin matching orders until every zone is loaded. That is a **barrier**:
all participants reach a point before any proceeds. With two participants, the same
idea is called a **rendezvous**.

A semaphore answers "how many may work at once?" A barrier answers "has everyone
finished preparing?" They solve different coordination problems.

## A single phase

Use one wait group for arrival and another for final completion:

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	zones := make([]string, 4)
	observed := make([]int, len(zones))
	var arrived, finished sync.WaitGroup
	arrived.Add(len(zones))
	for i := range zones {
		finished.Go(func() {
			zones[i] = fmt.Sprintf("zone-%d", i)
			arrived.Done()
			arrived.Wait()
			for _, zone := range zones {
				if zone != "" {
					observed[i]++
				}
			}
		})
	}
	finished.Wait()
	fmt.Println(observed)
}
```

The output is `[4 4 4 4]`. Each worker writes a different slice element before arriving.
The barrier makes those writes visible before any worker reads the whole slice. The
`finished` group keeps `main` alive until the reads and result writes finish too.
Multiple goroutines may call `Wait` on the same group.

## Arrivals must be real

Initialize the arrival count before starting workers. If it includes a worker that
never starts or never calls `Done`, everyone else waits forever. Conversely, calling
`Done` before the zone is prepared lets other workers observe incomplete state.
A barrier is only as sound as its arrival protocol.

This example is deliberately a single phase. Don't reset the arrival counter for
another round while previous waiters may still be returning from `Wait`. For several
known phases, use a distinct group per phase. A reusable barrier needs generation
tracking so arrivals from different rounds cannot mix.

`WaitGroup.Wait` isn't cancellable. If a preparation can fail, prefer having workers
report results to a coordinator that either closes a proceed channel or cancels the
whole operation. Every participant must agree on the failure path. A missing arrival
must become an explicit failure, not an unexplained hang in production.

Further reading: [sync.WaitGroup](https://pkg.go.dev/sync#WaitGroup).
