---
title: Structured Logging with slog
quiz:
  - question: |
      Which log field is appropriate for a failed login request?
    options:
      - text: 'The Authorization header'
      - text: 'The submitted password'
      - text: 'A request ID and a fixed failure category'
        correct: true
    explanation: |
      Request IDs let you connect events without retaining credentials. Keep passwords, tokens, and raw request bodies out of logs.
---

A production bug rarely arrives with a debugger attached. You get a time, a symptom,
and perhaps a request ID. Logs should help connect those pieces.

Squeak has used formatted messages so far. `log/slog` lets you attach named fields to
an event, so a log system can filter by route, severity, or request ID without guessing
where each value sits in a sentence.

## Configure once, attach context

Create a logger during startup using `log/slog` and `os`:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))
logger = logger.With("service", "squeak")
logger.Info("starting", "addr", ":8080")
```

The handler emits one JSON object per event. Fields include the message, level and time,
plus your attributes. A text handler can be easier to read during local development;
the calls that write events stay the same.

Pass the logger into your application configuration, just like the store. In a handler,
derive a logger with the request ID supplied by your middleware:

```go
logger := cfg.logger.With("request_id", requestID)
logger.ErrorContext(r.Context(), "loading squeak", "error", err)
respondWithError(w, http.StatusInternalServerError, "couldn't get squeak")
```

These lines belong on the storage-error path. The client gets a stable, safe message;
your logs retain diagnostic detail. Passing a context doesn't automatically add its
values as attributes. Explicitly add the request ID, or write a handler that does so.

## Choose fields deliberately

For completed requests, useful attributes include the method, matched route pattern,
status, and elapsed duration. Prefer the matched pattern such as `/api/squeaks/{id}`
over recording every unique URL. Query strings may contain credentials or personal data.

Use `Info` for normal lifecycle events, `Warn` for situations needing attention, and
`Error` for failed operations. A routine missing squeak is usually a 404, not an emergency.
Don't log the same error at every layer: wrap it with context in the store, then record
it at the boundary that handles it.

Never log a password, bearer token, session cookie, or whole login body. Even an error
string from another service can contain sensitive input, so choose what you retain.
A request ID is useful precisely because you can investigate without copying credentials.

## Make logging testable

`NewJSONHandler` accepts an `io.Writer`. A test can supply a `bytes.Buffer`, decode the
JSON, and assert that `service` and `request_id` are present. Avoid comparing an entire
log line: timestamps and attribute ordering aren't the behavior you care about.

Logs complement metrics. Metrics answer how often Squeak fails; a carefully chosen
log event helps explain one of those failures.

Further reading: [log/slog](https://pkg.go.dev/log/slog).
