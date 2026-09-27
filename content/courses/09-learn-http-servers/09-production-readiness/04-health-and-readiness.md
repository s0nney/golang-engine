---
title: Health and Readiness
quiz:
  - question: |
      The database is temporarily unavailable but Squeak can still answer HTTP. Which probe should fail?
    options:
      - text: 'Liveness, so every instance restarts immediately'
      - text: 'Neither probe, because HTTP is working'
      - text: 'Readiness, so traffic can stop until the dependency recovers'
        correct: true
    explanation: |
      Liveness asks whether the process is responsive. Readiness asks whether it can serve its workload. Restarting every instance rarely repairs an unavailable shared database.
---

A load balancer needs to know whether a Squeak instance can take requests. A process
supervisor needs to know whether the process is responsive. Those are different questions.

**Liveness** is a cheap indication that the process can answer. **Readiness** says the
instance is prepared to serve its workload. Keep the endpoints small: no passwords,
connection strings, internal stack traces, or expensive full-system scans in the response.

## A readiness gate

A flag shared between the shutdown code and handlers must be synchronized. With
`sync/atomic`, `net/http`, and `fmt`, the gate can look like this:

```go
func probeHandler(ready *atomic.Bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "alive")
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !ready.Load() {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprintln(w, "ready")
	})
	return mux
}
```

The zero value is false. Mark the instance ready after configuration and required
initialization succeed. This gate only tracks lifecycle state; it doesn't prove a
database is reachable. If database access is essential, a readiness policy can also
use a bounded `PingContext`, or consult a recently updated dependency status.

Be careful with probes that make their own network requests. Set a short deadline,
limit their frequency, and avoid turning a database slowdown into thousands of extra
queries. Liveness generally shouldn't depend on that shared database: restarting all
your healthy processes won't bring the database back.

## Leaving the load balancer

During shutdown, mark readiness false before draining. Routing changes take time to
propagate. Your deployment environment may provide a drain interval before signalling
the process, or your command may need an explicit, bounded interval while it still
serves existing traffic. A readiness flag alone doesn't guarantee the balancer has
stopped sending requests.

Then call `Shutdown` and wait for it, as in the first lesson. Keep probes fast and
independent from login middleware so a monitoring system doesn't need a user account.
They still belong behind appropriate deployment access controls.

## Test transitions

Use `httptest.NewRecorder` to call `/readyz` before and after `ready.Store(true)`.
Expect 503, then 200, then 503 after setting it false again. Assert `/livez` still returns
200 in each state. You're testing the contract a deployment system relies on, not
whether a particular orchestrator happens to be installed on your laptop.
