---
title: Reading Runtime Metrics
quiz:
- question: The goroutine count increased after traffic doubled. What does that prove?
  options:
  - text: There is definitely a goroutine leak
  - text: Nothing by itself; compare workload, trends, and goroutine stacks
    correct: true
  - text: GOMAXPROCS must be doubled
  explanation: A higher count can reflect legitimate work or stuck goroutines. A metric is a signal to investigate, not a diagnosis.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"runtime/metrics"
    )

    func schedulerCounts() (map[string]uint64, error) {
    	_ = metrics.KindUint64 // remove when you use metrics
    	return map[string]uint64{}, nil
    }

    func main() {
    	counts, err := schedulerCounts()
    	fmt.Println("goroutines:", counts["/sched/goroutines:goroutines"])
    	fmt.Println("GOMAXPROCS:", counts["/sched/gomaxprocs:threads"])
    	fmt.Println("error:", err)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"runtime/metrics"
    )

    func schedulerCounts() (map[string]uint64, error) {
    	samples := []metrics.Sample{
    		{Name: "/sched/goroutines:goroutines"},
    		{Name: "/sched/gomaxprocs:threads"},
    	}
    	metrics.Read(samples)
    	counts := make(map[string]uint64, len(samples))
    	for _, s := range samples {
    		if s.Value.Kind() != metrics.KindUint64 {
    			return nil, fmt.Errorf("unexpected metric kind for %s: %v", s.Name, s.Value.Kind())
    		}
    		counts[s.Name] = s.Value.Uint64()
    	}
    	return counts, nil
    }

    func main() {
    	counts, err := schedulerCounts()
    	fmt.Println("goroutines:", counts["/sched/goroutines:goroutines"])
    	fmt.Println("GOMAXPROCS:", counts["/sched/gomaxprocs:threads"])
    	fmt.Println("error:", err)
    }
  tests: |
    package main

    import (
    	"runtime"
    	"testing"
    )

    func TestSchedulerCounts(t *testing.T) {
    	previous := runtime.GOMAXPROCS(1)
    	defer runtime.GOMAXPROCS(previous)
    	for _, n := range []int{1, 2} {
    		runtime.GOMAXPROCS(n)
    		counts, err := schedulerCounts()
    		if err != nil {
    			t.Fatal(err)
    		}
    		if len(counts) != 2 {
    			t.Fatalf("got %d metrics, want 2", len(counts))
    		}
    		if got := counts["/sched/gomaxprocs:threads"]; got != uint64(n) {
    			t.Errorf("GOMAXPROCS metric = %d, want %d", got, n)
    		}
    		if got := runtime.GOMAXPROCS(0); got != n {
    			t.Errorf("schedulerCounts changed GOMAXPROCS to %d", got)
    		}
    		if got := counts["/sched/goroutines:goroutines"]; got == 0 {
    			t.Error("missing or zero goroutine count")
    		}
    	}
    }

---

Dispatchly slows down under lunchtime traffic. Before changing worker counts, ask
what the runtime is doing. `runtime/metrics` exposes measurements such as live
goroutines, scheduler latency, allocation, and garbage-collection activity.

## Read named samples

Each name includes a unit. Values have a kind: unsigned integer, floating-point number,
or histogram. Check the kind before calling its matching accessor. An unknown metric
produces `KindBad`; calling the wrong accessor panics.

```go
package main

import (
	"fmt"
	"runtime/metrics"
)

func main() {
	samples := []metrics.Sample{
		{Name: "/sched/goroutines:goroutines"},
		{Name: "/sched/gomaxprocs:threads"},
	}
	metrics.Read(samples)
	for _, sample := range samples {
		if sample.Value.Kind() == metrics.KindUint64 {
			fmt.Println(sample.Name, sample.Value.Uint64())
		}
	}
}
```

The counts depend on the machine and workload, so don't hard-code expected numbers.
`metrics.All()` describes available names, units, kinds, and whether a measurement is
cumulative. Discover metrics from the running toolchain rather than assuming every Go
version exposes the same list. Histogram values contain bucket boundaries and counts;
a bucket count isn't a duration or a ready-made percentile.

## Interpret changes in context

A growing goroutine count might mean more active deliveries, blocked downstream calls,
or a leak. If traffic subsides but the count keeps climbing, inspect goroutine stacks.
Scheduler latency describes time goroutines spend runnable before execution; it isn't
the full duration of a customer request. Correlate it with CPU saturation, latency,
and your queue depth before changing `GOMAXPROCS`.

Use repeated measurements with the same workload. For cumulative counters, compare
deltas over an interval and account for process restarts. Metrics summarize behavior;
they don't identify the exact line causing it. Goroutine dumps and profiles do; [Learn Testing](/courses/learn-testing/coverage-and-tooling/profiling-with-pprof) shows how to collect and read them with `pprof`.

## Your task

Implement `schedulerCounts`, which reads the two metric names from the example and
returns a map from each name to its unsigned count. Return a descriptive error if
either value has an unexpected kind. Keep the function read-only: it mustn't change
`GOMAXPROCS` to make a test pass.

The tests check the names and compare the configured thread limit while holding it
fixed. They only require the goroutine count to be positive because the test harness
also runs goroutines. Production exporters can collect these measurements with
Prometheus or OpenTelemetry; the standard-library exercise teaches what those
integrations are reporting.

Further reading: [runtime/metrics](https://pkg.go.dev/runtime/metrics).
