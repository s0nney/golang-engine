//go:build unix

// Package runner compiles and runs learners' Go code with the local toolchain.
//
// Code runs as the server's user, limited only by timeouts, an output cap and
// a concurrency limit. It's meant for a server bound to localhost; exposing it
// publicly needs real isolation (a container or VM) around it.
//
// No shell is involved: the go tool and the built program are started
// directly with fixed argument lists, and learner code only ever reaches
// them as files on disk.
package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	buildTimeout = 60 * time.Second
	runTimeout   = 5 * time.Second
	maxOutput    = 64 << 10
)

const goMod = "module exercise\n\ngo 1.27\n"

// Result is the outcome of running or testing code.
type Result struct {
	Output string // combined build and program output
	OK     bool   // it compiled and exited 0
}

type Runner struct {
	goBin string
	sem   chan struct{}
}

// New finds the go binary and allows one concurrent job per CPU.
func New() (*Runner, error) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return nil, fmt.Errorf("runner needs the go toolchain on PATH: %w", err)
	}
	return &Runner{goBin: goBin, sem: make(chan struct{}, runtime.NumCPU())}, nil
}

// Run builds main.go and runs the program.
func (r *Runner) Run(ctx context.Context, code string) Result {
	return r.buildAndRun(ctx, map[string]string{"main.go": code}, false, false)
}

// Check compiles a standalone exercise without executing it. Authoring checks
// use this to distinguish an unfinished starter from one that cannot build.
func (r *Runner) Check(ctx context.Context, code string) Result {
	return r.buildAndRun(ctx, map[string]string{"main.go": code}, false, true)
}

// Test builds main.go with main_test.go and runs the tests verbosely.
func (r *Runner) Test(ctx context.Context, code, tests string) Result {
	return r.buildAndRun(ctx, map[string]string{"main.go": code, "main_test.go": tests}, true, false)
}

// Submit grades code. With tests it passes when they do; otherwise it
// passes when the program's output matches expected.
func (r *Runner) Submit(ctx context.Context, code, tests, expected string) (res Result, passed bool) {
	if tests != "" {
		res = r.Test(ctx, code, tests)
		return res, res.OK
	}
	res = r.Run(ctx, code)
	return res, res.OK && SameOutput(res.Output, expected)
}

// SameOutput compares program output, ignoring trailing whitespace.
func SameOutput(got, want string) bool {
	return trimLines(got) == trimLines(want)
}

func trimLines(s string) string {
	lines := strings.Split(strings.TrimRight(s, " \t\r\n"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t\r")
	}
	return strings.Join(lines, "\n")
}

func (r *Runner) buildAndRun(ctx context.Context, files map[string]string, test, buildOnly bool) Result {
	select {
	case r.sem <- struct{}{}:
		defer func() { <-r.sem }()
	case <-ctx.Done():
		return Result{Output: "Cancelled while waiting for a free runner."}
	}

	dir, err := os.MkdirTemp("", "goland-run-")
	if err != nil {
		return internalError(err)
	}
	defer os.RemoveAll(dir)
	files["go.mod"] = goMod
	for name, src := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o600); err != nil {
			return internalError(err)
		}
	}

	out := &capped{max: maxOutput}
	// Scratch modules have no VCS identity, even if TMPDIR is inside a checkout.
	buildArgs := []string{"build", "-buildvcs=false", "-o", "prog", "."}
	if test {
		buildArgs = []string{"test", "-buildvcs=false", "-c", "-o", "prog", "."}
	}
	env := append(os.Environ(),
		"GOTOOLCHAIN=local", "GOPROXY=off", "GOFLAGS=-mod=mod", "CGO_ENABLED=0", "GOWORK=off")
	if err := start(ctx, buildTimeout, dir, env, out, r.goBin, buildArgs...); err != nil {
		return Result{Output: explain(out, err, buildTimeout)}
	}
	if buildOnly {
		return Result{Output: out.String(), OK: true}
	}

	var progArgs []string
	if test {
		progArgs = []string{"-test.v", "-test.count=1"}
	}
	env = []string{"HOME=" + dir, "TMPDIR=" + dir, "PATH=/usr/bin:/bin"}
	err = start(ctx, runTimeout, dir, env, out, filepath.Join(dir, "prog"), progArgs...)
	return Result{Output: explain(out, err, runTimeout), OK: err == nil}
}

// start runs name in its own process group and waits for it. On timeout the
// whole group is killed, not just the direct child.
func start(ctx context.Context, timeout time.Duration, dir string, env []string, out *capped, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout, cmd.Stderr = out, out
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = time.Second
	err := cmd.Run()
	if ctx.Err() != nil {
		return context.DeadlineExceeded
	}
	return err
}

func explain(out *capped, err error, timeout time.Duration) string {
	s := out.String()
	if errors.Is(err, context.DeadlineExceeded) {
		s += fmt.Sprintf("\nKilled: took longer than %s. Is there an infinite loop?", timeout)
	} else if exitErr, ok := errors.AsType[*exec.ExitError](err); ok && s == "" {
		s = exitErr.Error()
	}
	if out.truncated {
		s += "\n(output truncated)"
	}
	return s
}

func internalError(err error) Result {
	return Result{Output: "Internal error: " + err.Error()}
}

// capped is a buffer that silently drops writes past max bytes.
type capped struct {
	buf       bytes.Buffer
	max       int
	truncated bool
}

func (c *capped) Write(p []byte) (int, error) {
	if room := c.max - c.buf.Len(); len(p) > room {
		c.buf.Write(p[:max(room, 0)])
		c.truncated = true
		return len(p), nil
	}
	return c.buf.Write(p)
}

func (c *capped) String() string { return c.buf.String() }
