# Concurrency coverage audit

Reference: [Go concurrency distilled](https://antonz.org/go-concurrency-distilled/)
by Anton Zhiyanov. Reviewed 2026-09-27 against the public page, including its
subsections. Scope is topic coverage of that page, not reproduction of its examples
or the paid exercises in the author's separate book.

The curriculum keeps its beginner-oriented order and Dispatchly examples. Existing
lesson filenames and chapter directories remain stable so saved progress URLs survive.
Paths below are relative to [course 06](../courses/06-learn-concurrency/).

| Reference topic | Curriculum location | Audit outcome |
| --- | --- | --- |
| Goroutines and waiting for dependent work | `02-goroutines-and-waitgroups/01-goroutines-recap.md`, `02-waitgroups.md` | Already covered; clarified concurrent waits and adding child tasks |
| Channels: output, closing, iteration, directions, buffers, nil | All five lessons in `03-channels-in-depth`; `07-concurrency-patterns/01-generators.md` | Already covered |
| Select and nonblocking operations | `04-select/01-multiplexing.md`, `02-non-blocking-operations.md` | Already covered |
| Pipeline stages and output/completion/cancel channels | `07-concurrency-patterns/02-pipelines.md`; `04-select/05-the-done-channel.md` | Added explicit cancellation versus completion ownership |
| Pipeline errors: first failure, result records, separate error stream | `07-concurrency-patterns/06-first-error-cancels.md`, `09-pipeline-errors.md` | Added comparison of all three contracts and deadlock/cleanup concerns |
| Time: After, Timer, Stop, Reset, AfterFunc, Ticker | `04-select/03-timeouts.md`, `04-tickers.md`, `06-reusable-timers.md` | Added timer/callback lesson and reset exercise |
| Context: cancellation, deadlines, causes, values, AfterFunc, WithoutCancel | All five lessons in `05-context` | Already covered |
| WaitGroup counter, Add/Done/Wait and Go, multiple waiters | `02-goroutines-and-waitgroups/02-waitgroups.md`; `07-concurrency-patterns/07-rendezvous-and-barriers.md` | Clarified lifecycle; added multi-waiter example |
| Data races and race detector | `08-race-conditions-and-deadlocks/01-data-races-vs-race-conditions.md`, `02-the-race-detector.md` | Already covered |
| Race conditions and compare-and-set | `08-race-conditions-and-deadlocks/01-data-races-vs-race-conditions.md`; `06-sync-primitives/03-atomics.md` | Already covered, including atomic operations that do not compose |
| Mutex, TryLock, RWMutex, Locker, channel as mutex | `06-sync-primitives/01-mutex-and-rwmutex.md`, `06-locker-and-trylock.md` | Added nonblocking locks, Locker/RLocker and one-slot channel comparison |
| Semaphores and weighted semaphore awareness | `07-concurrency-patterns/05-semaphores.md` | Already covered; external weighted implementation remains further reading |
| Rendezvous and barrier | `07-concurrency-patterns/07-rendezvous-and-barriers.md` | Added phase coordination, separate completion, failure and reuse constraints |
| Signaling: Cond, Signal/Broadcast, channel equivalents | `06-sync-primitives/04-sync-cond.md`; `04-select/05-the-done-channel.md` | Already covered |
| Publish/subscribe | `07-concurrency-patterns/08-publish-subscribe.md` | Added independent subscriber queues and graded delivery exercise |
| Once and OnceFunc/OnceValue/OnceValues | `06-sync-primitives/02-once.md` | Already covered with exercise |
| Object pools | `06-sync-primitives/07-object-pools.md` | Added Pool, ownership, optional reuse and allocation tradeoffs |
| Typed atomics: Load/Store/Swap/CAS/Add and compound operations | `06-sync-primitives/03-atomics.md` | Already covered |
| Testing: synchronization, synctest, durable blocking, fake time | All four lessons in `09-testing-concurrent-code` | Already covered; new timer exercise applies fake-time checks |
| Scheduling: cores, threads, goroutines, blocking, preemption, GOMAXPROCS | `01-why-concurrency/02-gomaxprocs.md`, `03-the-go-scheduler.md` | Already covered |
| Runtime metrics and discovery | `10-concurrency-diagnostics/01-runtime-metrics.md` | Added typed metric reads, interpretation and exercise |
| Profiling: CPU, heap, goroutine, block, mutex, pprof HTTP/manual collection | `10-concurrency-diagnostics/02-profiling-contention.md` | Added concurrency-specific diagnostics, extending course 07's profiling introduction |
| Execution tracing, manual/CLI collection, flight recording | `10-concurrency-diagnostics/03-traces-and-flight-recording.md` | Added recording lifecycle, labels, snapshots and retention constraints |

## Version and correctness choices

Check behavior against the installed Go 1.27.1 documentation and official references:
[sync](https://pkg.go.dev/sync), [time](https://pkg.go.dev/time),
[runtime/metrics](https://pkg.go.dev/runtime/metrics),
[runtime/trace](https://pkg.go.dev/runtime/trace), and
[diagnostics](https://go.dev/doc/diagnostics).

- Modern Go can collect unreachable timers and tickers. Stopping remains useful to
  stop activity; it does not close their channels or join an already running callback.
- Cond supports repeated broadcasts. A broadcast isn't a stored notification for
  future waiters; the protected condition determines whether a waiter should sleep.
- Pool reuse and allocation counts are not guaranteed. Correctness cannot depend on
  retaining a particular object or receiving it from the next Get.
- GOMAXPROCS limits simultaneous Go execution, not total OS threads. Waiting goroutines
  become runnable when their wait completes, not merely because they blocked.
- Atomics provide synchronization guarantees; they need not correspond to exactly one
  machine instruction. Separately atomic operations don't make a whole sequence atomic.
- A flight recorder's byte setting takes precedence over its requested age window and
  is not a hard bound on snapshot size or total runtime memory.

## Validation

Run `go run ./cmd/validate -exec content/courses/06-learn-concurrency`.
Course 06 now contains 10 chapters, 51 lessons, and 14 coding exercises.
New examples use the standard library, and the three new exercises require neither
network access nor external services. All 14 solutions passed exercise validation;
all starters failed grading as intended. The three new starters also compiled and
ran independently. The three new solutions and scratch checks for locking, pipeline
error aggregation, and flight recording passed `go test -race`. Complete new lesson
programs were run from a scratch directory and produced the documented output.
