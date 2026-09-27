---
title: Graceful Shutdown
quiz:
  - question: |
      Why should shutdown use a fresh timeout context after the signal context is cancelled?
    options:
      - text: 'The cancelled signal context would immediately end the grace period'
        correct: true
      - text: 'Shutdown cannot accept a context with a deadline'
      - text: 'A fresh context cancels every request automatically'
    explanation: |
      A shutdown deadline gives active requests time to finish. Derive it from context.Background(), because the signal context has already been cancelled.
---

Pip posts a squeak just as you deploy a new version. If the old process exits
immediately, the response disappears halfway through. A graceful shutdown gives
requests already in progress a bounded chance to finish.

## Own the server's lifetime

Keep an `http.Server` value so you can stop it. The function below accepts Squeak's
assembled handler and a context cancelled by the caller when it's time to stop.
Imports are `context`, `errors`, `net/http`, and `time`.

```go
func serve(ctx context.Context, addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr: addr, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout: 60 * time.Second,
	}
	done := make(chan error, 1)
	go func() { done <- srv.ListenAndServe() }()

	select {
	case err := <-done:
		return err // startup failed, for example an occupied port
	case <-ctx.Done():
	}

	grace, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := srv.Shutdown(grace)
	if err != nil {
		_ = srv.Close() // grace period expired: close remaining connections
	}
	serveErr := <-done
	if !errors.Is(serveErr, http.ErrServerClosed) {
		return errors.Join(err, serveErr)
	}
	return err
}
```

`Shutdown` stops accepting connections and gives active requests time to complete.
Crucially, `ListenAndServe` returns before draining is finished. Keep the main goroutine
waiting for `Shutdown`; returning from `main` kills every remaining goroutine.
The buffered error channel also lets a startup failure report itself without getting stuck.

## Connect operating-system signals

In a Unix command, use `os/signal`, `os`, and `syscall`:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
if err := serve(ctx, ":8080", handler); err != nil {
	slog.Error("server stopped", "error", err)
}
```

Here `handler` is your assembled router. Keep this signal wiring at the command boundary;
tests can supply a cancelled context without sending signals to the test process.

Shutdown doesn't join arbitrary background jobs. Give those jobs their own cancellation
and wait group, then wait for them before closing shared resources. Request handlers
must also honor their contexts when waiting on storage. Closing a socket cannot stop
a goroutine that's stuck in an unrelated computation. Hijacked connections such as
WebSockets need an explicit shutdown policy too.

The deployment system's termination allowance must be longer than your drain deadline,
with room for other cleanup. Otherwise it can kill Squeak while your code is politely waiting.

Further reading: [http.Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown).
