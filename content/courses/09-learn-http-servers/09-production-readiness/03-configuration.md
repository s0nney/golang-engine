---
title: Configuration at Startup
quiz:
  - question: |
      SQUEAK_DRAIN_TIMEOUT is set to nonsense. What should startup do?
    options:
      - text: 'Ignore it and use ten seconds'
      - text: 'Return a clear configuration error before serving traffic'
        correct: true
      - text: 'Wait until the first shutdown to parse it'
    explanation: |
      An explicitly supplied invalid value is a deployment mistake. Reject it before accepting requests; use defaults only when the setting is absent.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    	"time"
    )

    type Config struct {
    	Addr         string
    	DrainTimeout time.Duration
    }

    func loadConfig(lookup func(string) (string, bool)) (Config, error) {
    	// Parse and validate both settings.
    	_ = strings.TrimSpace
    	return Config{}, nil
    }

    func main() {
    	env := map[string]string{"SQUEAK_ADDR": ":9090", "SQUEAK_DRAIN_TIMEOUT": "5s"}
    	cfg, err := loadConfig(func(key string) (string, bool) { v, ok := env[key]; return v, ok })
    	fmt.Printf("addr=%s drain=%s error=%v\n", cfg.Addr, cfg.DrainTimeout, err)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    	"time"
    )

    type Config struct {
    	Addr         string
    	DrainTimeout time.Duration
    }

    func loadConfig(lookup func(string) (string, bool)) (Config, error) {
    	cfg := Config{Addr: "127.0.0.1:8080", DrainTimeout: 10 * time.Second}
    	if value, ok := lookup("SQUEAK_ADDR"); ok {
    		cfg.Addr = strings.TrimSpace(value)
    		if cfg.Addr == "" {
    			return Config{}, fmt.Errorf("SQUEAK_ADDR must not be empty")
    		}
    	}
    	if value, ok := lookup("SQUEAK_DRAIN_TIMEOUT"); ok {
    		d, err := time.ParseDuration(value)
    		if err != nil || d <= 0 || d > time.Minute {
    			return Config{}, fmt.Errorf("SQUEAK_DRAIN_TIMEOUT must be a duration greater than zero and at most 1m")
    		}
    		cfg.DrainTimeout = d
    	}
    	return cfg, nil
    }

    func main() {
    	env := map[string]string{"SQUEAK_ADDR": ":9090", "SQUEAK_DRAIN_TIMEOUT": "5s"}
    	cfg, err := loadConfig(func(key string) (string, bool) { v, ok := env[key]; return v, ok })
    	fmt.Printf("addr=%s drain=%s error=%v\n", cfg.Addr, cfg.DrainTimeout, err)
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    	"time"
    )

    func TestLoadConfig(t *testing.T) {
    	for _, tc := range []struct {
    		name string
    		env  map[string]string
    		want Config
    		bad  string
    	}{
    		{"defaults", nil, Config{"127.0.0.1:8080", 10 * time.Second}, ""},
    		{"address only", map[string]string{"SQUEAK_ADDR": " :9090 "}, Config{":9090", 10 * time.Second}, ""},
    		{"duration only", map[string]string{"SQUEAK_DRAIN_TIMEOUT": "250ms"}, Config{"127.0.0.1:8080", 250 * time.Millisecond}, ""},
    		{"upper boundary", map[string]string{"SQUEAK_DRAIN_TIMEOUT": "1m"}, Config{"127.0.0.1:8080", time.Minute}, ""},
    		{"both", map[string]string{"SQUEAK_ADDR": ":9000", "SQUEAK_DRAIN_TIMEOUT": "1ns"}, Config{":9000", time.Nanosecond}, ""},
    		{"empty address", map[string]string{"SQUEAK_ADDR": ""}, Config{}, "SQUEAK_ADDR"},
    		{"blank address", map[string]string{"SQUEAK_ADDR": " \t "}, Config{}, "SQUEAK_ADDR"},
    		{"empty duration", map[string]string{"SQUEAK_DRAIN_TIMEOUT": ""}, Config{}, "SQUEAK_DRAIN_TIMEOUT"},
    		{"invalid duration", map[string]string{"SQUEAK_DRAIN_TIMEOUT": "nonsense"}, Config{}, "SQUEAK_DRAIN_TIMEOUT"},
    		{"unit required", map[string]string{"SQUEAK_DRAIN_TIMEOUT": "10"}, Config{}, "SQUEAK_DRAIN_TIMEOUT"},
    		{"zero", map[string]string{"SQUEAK_DRAIN_TIMEOUT": "0s"}, Config{}, "SQUEAK_DRAIN_TIMEOUT"},
    		{"negative", map[string]string{"SQUEAK_DRAIN_TIMEOUT": "-1s"}, Config{}, "SQUEAK_DRAIN_TIMEOUT"},
    		{"too long", map[string]string{"SQUEAK_DRAIN_TIMEOUT": "60.000000001s"}, Config{}, "SQUEAK_DRAIN_TIMEOUT"},
    	} {
    		t.Run(tc.name, func(t *testing.T) {
    			got, err := loadConfig(func(key string) (string, bool) { v, ok := tc.env[key]; return v, ok })
    			if tc.bad != "" {
    				if err == nil || !strings.Contains(err.Error(), tc.bad) {
    					t.Fatalf("error = %v, want error naming %s", err, tc.bad)
    				}
    				return
    			}
    			if err != nil || got != tc.want {
    				t.Fatalf("loadConfig = %+v, %v; want %+v, nil", got, err, tc.want)
    			}
    		})
    	}
    }

---

Your laptop and a deployed Squeak instance need different settings. Hard-coding a
port into every handler makes that difference painful. Read configuration once at
startup, validate it, and pass typed values to the components that need them.

## Strings at the edge, types inside

Environment variables are strings. Convert them before building your server:

```go
type Config struct {
	Addr         string
	DrainTimeout time.Duration
}
```

Now the shutdown code receives a duration, not a string it might fail to parse during
an incident. Keep defaults in one place. A missing setting may use a default; an
explicitly empty or malformed setting should produce a useful error.

`os.LookupEnv` returns both the value and whether it exists. That distinction is lost
with `os.Getenv`, which returns an empty string for both missing and empty values.
For a testable loader, accept a function with the same signature:

```go
cfg, err := loadConfig(os.LookupEnv)
if err != nil {
	return err
}
```

Tests can supply a map lookup. They don't need to mutate the process environment,
and parallel tests can't accidentally change one another's settings.

## Your task

Complete `loadConfig` in the editor:

1. If `SQUEAK_ADDR` is absent, use `127.0.0.1:8080`. If present, trim surrounding
   whitespace and reject an empty result. This exercise leaves address syntax checking
   to the server when it binds the listener.
2. If `SQUEAK_DRAIN_TIMEOUT` is absent, use ten seconds. If present, parse it with
   `time.ParseDuration` and require a strictly positive value no greater than one minute.
3. Return errors that name the setting responsible. Don't silently replace an invalid
   supplied value with a default.

The main function uses a fixed example environment so Run always shows the same result.
Tests cover absent values, overrides, whitespace, malformed durations, and boundaries.

## Keep secrets separate from diagnostics

Configuration can eventually include a signing key or database credentials. Validate
required secrets at startup too, but never print the entire configuration struct or
include secret values in error messages. An error like `SQUEAK_SIGNING_KEY is required`
is enough to locate the mistake.

Environment variables are a delivery mechanism, not encryption. Your deployment system
must control who can read them. A local `.env` file also isn't loaded automatically by
Go; this course uses the process environment directly.

Once startup succeeds, treat configuration as immutable. A single snapshot is much
easier to reason about than handlers reading changing environment values on each request.
