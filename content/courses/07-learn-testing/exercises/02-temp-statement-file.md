---
title: A Temp Statement File
difficulty: easy
after: the-testing-package
hints:
  - '`t.TempDir()` gives you a fresh, empty directory and registers a cleanup that deletes it (and everything in it) when the test finishes. `filepath.Join(dir, name)` builds the file''s path.'
  - 'If `os.WriteFile` fails, the test can''t go on without its file, so stop it with `t.Fatalf`, not `t.Errorf`. And call `t.Helper()` first thing, so the failure is reported at the line that called `writeStatement`.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"os"
    	"testing"
    )

    // writeStatement is a test helper. It writes csv to a file called name
    // inside a temporary directory that is deleted when the test ends, and
    // returns the file's full path.
    //
    // If the file can't be written it stops the test with t.Fatalf.
    func writeStatement(t testing.TB, name, csv string) string {
    	// 1. Mark this function as a helper.
    	// 2. Get a temporary directory from t.
    	// 3. Write csv to name inside it (permissions 0o644).
    	// 4. On error, t.Fatalf with the path and the error.
    	// 5. Return the path.
    	return ""
    }

    // ---- a pretend testing.TB so Run can show what your helper does ----

    type demoTB struct {
    	testing.TB
    	cleanups []func()
    }

    func (d *demoTB) Helper() {}
    func (d *demoTB) TempDir() string {
    	dir, err := os.MkdirTemp("", "demo")
    	if err != nil {
    		panic(err)
    	}
    	d.cleanups = append(d.cleanups, func() { os.RemoveAll(dir) })
    	return dir
    }
    func (d *demoTB) Cleanup(f func())                  { d.cleanups = append(d.cleanups, f) }
    func (d *demoTB) Fatalf(format string, args ...any) { fmt.Printf("FATAL: "+format+"\n", args...) }
    func (d *demoTB) Errorf(format string, args ...any) { fmt.Printf("ERROR: "+format+"\n", args...) }

    func main() {
    	tb := &demoTB{}
    	path := writeStatement(tb, "march.csv", "2026-03-01,rent,-120000\n")
    	data, err := os.ReadFile(path)
    	fmt.Printf("path %q\ncontents %q, error %v\n", path, data, err)
    	for _, f := range tb.cleanups {
    		f()
    	}
    	_, err = os.Stat(path)
    	fmt.Println("deleted after the test:", os.IsNotExist(err))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"os"
    	"path/filepath"
    	"testing"
    )

    // writeStatement is a test helper. It writes csv to a file called name
    // inside a temporary directory that is deleted when the test ends, and
    // returns the file's full path.
    //
    // If the file can't be written it stops the test with t.Fatalf.
    func writeStatement(t testing.TB, name, csv string) string {
    	t.Helper()
    	path := filepath.Join(t.TempDir(), name)
    	if err := os.WriteFile(path, []byte(csv), 0o644); err != nil {
    		t.Fatalf("writing %s: %v", path, err)
    	}
    	return path
    }

    // ---- a pretend testing.TB so Run can show what your helper does ----

    type demoTB struct {
    	testing.TB
    	cleanups []func()
    }

    func (d *demoTB) Helper() {}
    func (d *demoTB) TempDir() string {
    	dir, err := os.MkdirTemp("", "demo")
    	if err != nil {
    		panic(err)
    	}
    	d.cleanups = append(d.cleanups, func() { os.RemoveAll(dir) })
    	return dir
    }
    func (d *demoTB) Cleanup(f func())                  { d.cleanups = append(d.cleanups, f) }
    func (d *demoTB) Fatalf(format string, args ...any) { fmt.Printf("FATAL: "+format+"\n", args...) }
    func (d *demoTB) Errorf(format string, args ...any) { fmt.Printf("ERROR: "+format+"\n", args...) }

    func main() {
    	tb := &demoTB{}
    	path := writeStatement(tb, "march.csv", "2026-03-01,rent,-120000\n")
    	data, err := os.ReadFile(path)
    	fmt.Printf("path %q\ncontents %q, error %v\n", path, data, err)
    	for _, f := range tb.cleanups {
    		f()
    	}
    	_, err = os.Stat(path)
    	fmt.Println("deleted after the test:", os.IsNotExist(err))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"os"
    	"path/filepath"
    	"runtime"
    	"testing"
    )

    // watchTB wraps a real *testing.T. It records Helper calls and failures
    // instead of failing the real test.
    type watchTB struct {
    	*testing.T
    	helper  bool
    	failed  bool
    	fatal   bool
    	message string
    }

    func (w *watchTB) Helper()        { w.helper = true }
    func (w *watchTB) Error(a ...any) { w.failed, w.message = true, fmt.Sprint(a...) }
    func (w *watchTB) Errorf(f string, a ...any) {
    	w.failed, w.message = true, fmt.Sprintf(f, a...)
    }
    func (w *watchTB) Fatal(a ...any) {
    	w.failed, w.fatal, w.message = true, true, fmt.Sprint(a...)
    	runtime.Goexit()
    }
    func (w *watchTB) Fatalf(f string, a ...any) {
    	w.failed, w.fatal, w.message = true, true, fmt.Sprintf(f, a...)
    	runtime.Goexit()
    }
    func (w *watchTB) FailNow() { w.failed, w.fatal = true, true; runtime.Goexit() }
    func (w *watchTB) Fail()    { w.failed = true }

    // call runs writeStatement in its own goroutine, so a Fatalf can stop it.
    func call(t *testing.T, name, csv string) (w *watchTB, path string, finished bool) {
    	w = &watchTB{T: t}
    	done := make(chan struct{})
    	go func() {
    		defer close(done)
    		path = writeStatement(w, name, csv)
    		finished = true
    	}()
    	<-done
    	return w, path, finished
    }

    func TestWritesTheFile(t *testing.T) {
    	const csv = "2026-03-01,rent,-120000\n2026-03-02,cash,5000\n"
    	var path string
    	t.Run("inner", func(t *testing.T) {
    		var w *watchTB
    		w, path, _ = call(t, "march.csv", csv)
    		if w.failed {
    			t.Fatalf("writeStatement(t, %q, ...) reported a failure: %s", "march.csv", w.message)
    		}
    		if path == "" {
    			t.Fatal(`writeStatement returned "", want the path of the file it wrote`)
    		}
    		if filepath.Base(path) != "march.csv" {
    			t.Errorf("writeStatement returned %q, want a path ending in march.csv", path)
    		}
    		data, err := os.ReadFile(path)
    		if err != nil {
    			t.Fatalf("reading the returned path: %v", err)
    		}
    		if string(data) != csv {
    			t.Errorf("file contains %q, want %q", data, csv)
    		}
    		if !w.helper {
    			t.Error("writeStatement never calls t.Helper(), so failures inside it point at the wrong line")
    		}
    	})
    	if path == "" {
    		return
    	}
    	if _, err := os.Stat(path); !os.IsNotExist(err) {
    		os.RemoveAll(filepath.Dir(path))
    		t.Errorf("%s still exists after the test finished: use t.TempDir(), which cleans up for you", path)
    	}
    }

    func TestSeparateDirectories(t *testing.T) {
    	_, a, _ := call(t, "a.csv", "x")
    	_, b, _ := call(t, "a.csv", "y")
    	if a == "" || a == b {
    		t.Errorf("two calls returned %q and %q; each call should get its own temp directory", a, b)
    	}
    }

    func TestStopsOnWriteError(t *testing.T) {
    	w, _, finished := call(t, filepath.Join("no-such-dir", "march.csv"), "x")
    	if !w.failed {
    		t.Fatal(`writeStatement(t, "no-such-dir/march.csv", ...) can't write the file (the directory doesn't exist) but reported no failure`)
    	}
    	if !w.fatal || finished {
    		t.Error("writeStatement reported the write error but didn't stop the test: use t.Fatalf, not t.Errorf")
    	}
    	if w.message == "" {
    		t.Error("writeStatement stopped the test without a message: include the path and the error")
    	}
    }
---

Tests that read statements from disk need a file to read. Creating one in
every test is repetitive, and forgetting to delete it litters the machine, so
Ledgerly's tests use a helper.

Write `writeStatement(t, name, csv)`. It:

- writes `csv` to a file called `name` inside a **temporary directory** that
  the testing package deletes when the test ends (`t.TempDir()`);
- returns the full path to that file;
- if the file can't be written, stops the test with `t.Fatalf`, mentioning the
  path and the error;
- marks itself as a helper, so failures are reported at the caller's line.

## Example

```go
func TestImportMarch(t *testing.T) {
	path := writeStatement(t, "march.csv", "2026-03-01,rent,-120000\n")
	txns, err := ImportFile(path)
	// ...
} // the directory holding march.csv is deleted here
```

**Run** calls your helper with a pretend `testing.TB`, reads the file back,
runs the cleanups and checks the file is gone.

## Constraints

- Each call gets its own directory, so two calls with the same `name` don't
  collide.
- Accept `testing.TB`, not `*testing.T`, so benchmarks can use the helper too.
