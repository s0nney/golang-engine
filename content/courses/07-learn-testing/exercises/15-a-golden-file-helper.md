---
title: A Golden File Helper
difficulty: hard
after: testing-io-and-files
hints:
  - 'For `diffLines`, split both texts with `strings.Split(s, "\n")` and walk index `i` from 0 to the longer length. Skip indexes where both lines exist and are equal. Otherwise write `line N:`, then `- %q` if want has line `i` and `+ %q` if got has it. `%q` is what makes trailing spaces and tabs visible.'
  - 'Count the differing lines as you go: print the first 5, and only count the rest, then add `... and N more differing lines` at the end if N > 0. A `strings.Builder` with `fmt.Fprintf` keeps this tidy.'
  - 'In `assertGolden`: with `update`, `os.MkdirAll(filepath.Dir(path), 0o755)` then `os.WriteFile`. Otherwise read the file; `errors.Is(err, fs.ErrNotExist)` means it''s missing (`t.Fatalf`, mentioning the path and `-update`). Replace `"\r\n"` with `"\n"` in what you read, then `t.Errorf` with the path and the diff if `diffLines` returns anything.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"os"
    	"path/filepath"
    	"runtime"
    	"testing"
    )

    // maxDiffs is how many differing lines diffLines shows before summarizing.
    const maxDiffs = 5

    // diffLines compares want and got line by line and returns "" if they're
    // equal, or a readable description of the differences.
    func diffLines(want, got string) string {
    	return ""
    }

    // assertGolden compares got with the golden file at path. With update set,
    // it writes got to path instead.
    func assertGolden(t testing.TB, path, got string, update bool) {
    	t.Helper()
    }

    // ---- a pretend testing.TB so Run can show what your helper reports ----

    type demoTB struct{ testing.TB }

    func (demoTB) Helper()                   {}
    func (demoTB) Logf(f string, a ...any)   { fmt.Printf("  LOG:   "+f+"\n", a...) }
    func (demoTB) Errorf(f string, a ...any) { fmt.Printf("  ERROR: "+f+"\n", a...) }
    func (demoTB) Fatalf(f string, a ...any) {
    	fmt.Printf("  FATAL: "+f+"\n", a...)
    	runtime.Goexit()
    }

    // demo runs assertGolden in its own goroutine, so a Fatalf can stop it.
    func demo(path, got string, update bool) {
    	fmt.Printf("assertGolden(t, path, %q, %v)\n", got, update)
    	done := make(chan struct{})
    	go func() {
    		defer close(done)
    		assertGolden(demoTB{}, path, got, update)
    	}()
    	<-done
    }

    func main() {
    	fmt.Print(diffLines("rent  -$1,200.00\ncash  $50.00\n", "rent  -$1,200.00\ncash  $50.00 \nfuel  -$60.00\n"))
    	fmt.Println()

    	dir, _ := os.MkdirTemp("", "golden")
    	defer os.RemoveAll(dir)
    	path := filepath.Join(dir, "testdata", "march.golden")
    	demo(path, "old statement\n", false)
    	demo(path, "old statement\n", true)
    	demo(path, "old statement\n", false)
    	demo(path, "new statement\n", false)
    }
  solution: |
    package main

    import (
    	"errors"
    	"fmt"
    	"io/fs"
    	"os"
    	"path/filepath"
    	"runtime"
    	"strings"
    	"testing"
    )

    // maxDiffs is how many differing lines diffLines shows before summarizing.
    const maxDiffs = 5

    // diffLines compares want and got line by line and returns "" if they're
    // equal, or a readable description of the differences.
    func diffLines(want, got string) string {
    	if want == got {
    		return ""
    	}
    	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
    	var b strings.Builder
    	shown, extra := 0, 0
    	for i := range max(len(w), len(g)) {
    		if i < len(w) && i < len(g) && w[i] == g[i] {
    			continue
    		}
    		if shown == maxDiffs {
    			extra++
    			continue
    		}
    		shown++
    		fmt.Fprintf(&b, "line %d:\n", i+1)
    		if i < len(w) {
    			fmt.Fprintf(&b, "- %q\n", w[i])
    		}
    		if i < len(g) {
    			fmt.Fprintf(&b, "+ %q\n", g[i])
    		}
    	}
    	if extra > 0 {
    		fmt.Fprintf(&b, "... and %d more differing lines\n", extra)
    	}
    	return b.String()
    }

    // assertGolden compares got with the golden file at path. With update set,
    // it writes got to path instead.
    func assertGolden(t testing.TB, path, got string, update bool) {
    	t.Helper()
    	if update {
    		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
    			t.Fatalf("updating golden file: %v", err)
    		}
    		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
    			t.Fatalf("updating golden file: %v", err)
    		}
    		t.Logf("updated %s", path)
    		return
    	}
    	data, err := os.ReadFile(path)
    	if errors.Is(err, fs.ErrNotExist) {
    		t.Fatalf("golden file %s doesn't exist: run the test with -update to create it", path)
    	}
    	if err != nil {
    		t.Fatalf("reading golden file: %v", err)
    	}
    	want := strings.ReplaceAll(string(data), "\r\n", "\n")
    	if diff := diffLines(want, got); diff != "" {
    		t.Errorf("output doesn't match %s (-want +got):\n%s", path, diff)
    	}
    }

    // ---- a pretend testing.TB so Run can show what your helper reports ----

    type demoTB struct{ testing.TB }

    func (demoTB) Helper()                   {}
    func (demoTB) Logf(f string, a ...any)   { fmt.Printf("  LOG:   "+f+"\n", a...) }
    func (demoTB) Errorf(f string, a ...any) { fmt.Printf("  ERROR: "+f+"\n", a...) }
    func (demoTB) Fatalf(f string, a ...any) {
    	fmt.Printf("  FATAL: "+f+"\n", a...)
    	runtime.Goexit()
    }

    // demo runs assertGolden in its own goroutine, so a Fatalf can stop it.
    func demo(path, got string, update bool) {
    	fmt.Printf("assertGolden(t, path, %q, %v)\n", got, update)
    	done := make(chan struct{})
    	go func() {
    		defer close(done)
    		assertGolden(demoTB{}, path, got, update)
    	}()
    	<-done
    }

    func main() {
    	fmt.Print(diffLines("rent  -$1,200.00\ncash  $50.00\n", "rent  -$1,200.00\ncash  $50.00 \nfuel  -$60.00\n"))
    	fmt.Println()

    	dir, _ := os.MkdirTemp("", "golden")
    	defer os.RemoveAll(dir)
    	path := filepath.Join(dir, "testdata", "march.golden")
    	demo(path, "old statement\n", false)
    	demo(path, "old statement\n", true)
    	demo(path, "old statement\n", false)
    	demo(path, "new statement\n", false)
    }
  tests: |
    package main

    import (
    	"fmt"
    	"os"
    	"path/filepath"
    	"runtime"
    	"strings"
    	"testing"
    )

    func TestDiffLines(t *testing.T) {
    	tests := []struct {
    		name, want, got, diff string
    	}{
    		{"equal", "a\nb\n", "a\nb\n", ""},
    		{"both empty", "", "", ""},
    		{"one changed line", "a\nb\nc", "a\nX\nc", "line 2:\n- \"b\"\n+ \"X\"\n"},
    		{"trailing space", "total $5\n", "total $5 \n", "line 1:\n- \"total $5\"\n+ \"total $5 \"\n"},
    		{"tab vs spaces", "a\tb", "a    b", "line 1:\n- \"a\\tb\"\n+ \"a    b\"\n"},
    		{"got has an extra line", "a\n", "a\nb\n", "line 2:\n- \"\"\n+ \"b\"\nline 3:\n+ \"\"\n"},
    		{"got is missing a line", "a\nb", "a", "line 2:\n- \"b\"\n"},
    		{"empty want", "", "x", "line 1:\n- \"\"\n+ \"x\"\n"},
    		{"two separate changes", "a\nb\nc\nd", "A\nb\nc\nD", "line 1:\n- \"a\"\n+ \"A\"\nline 4:\n- \"d\"\n+ \"D\"\n"},
    		{"exactly five differences", "1\n2\n3\n4\n5", "a\nb\nc\nd\ne",
    			"line 1:\n- \"1\"\n+ \"a\"\nline 2:\n- \"2\"\n+ \"b\"\nline 3:\n- \"3\"\n+ \"c\"\nline 4:\n- \"4\"\n+ \"d\"\nline 5:\n- \"5\"\n+ \"e\"\n"},
    		{"eight differences", "1\n2\n3\n4\n5\n6\n7\n8", "a\nb\nc\nd\ne\nf\ng\nh",
    			"line 1:\n- \"1\"\n+ \"a\"\nline 2:\n- \"2\"\n+ \"b\"\nline 3:\n- \"3\"\n+ \"c\"\nline 4:\n- \"4\"\n+ \"d\"\nline 5:\n- \"5\"\n+ \"e\"\n... and 3 more differing lines\n"},
    		{"summary counts missing lines", "same\n1\n2\n3\n4\n5", "same\na\nb\nc\nd\ne\nf\ng",
    			"line 2:\n- \"1\"\n+ \"a\"\nline 3:\n- \"2\"\n+ \"b\"\nline 4:\n- \"3\"\n+ \"c\"\nline 5:\n- \"4\"\n+ \"d\"\nline 6:\n- \"5\"\n+ \"e\"\n... and 2 more differing lines\n"},
    	}
    	for _, tt := range tests {
    		if got := diffLines(tt.want, tt.got); got != tt.diff {
    			t.Errorf("%s: diffLines(%q, %q) returned\n%s\nwant\n%s", tt.name, tt.want, tt.got, got, tt.diff)
    		}
    	}
    }

    // watchTB wraps a real *testing.T and records what the helper reports.
    type watchTB struct {
    	*testing.T
    	helper   bool
    	failed   bool
    	fatal    bool
    	messages []string
    }

    func (w *watchTB) Helper()             { w.helper = true }
    func (w *watchTB) Log(a ...any)        {}
    func (w *watchTB) Logf(string, ...any) {}
    func (w *watchTB) record(fatal bool, msg string) {
    	w.failed = true
    	w.fatal = w.fatal || fatal
    	w.messages = append(w.messages, msg)
    }
    func (w *watchTB) Error(a ...any)            { w.record(false, fmt.Sprint(a...)) }
    func (w *watchTB) Errorf(f string, a ...any) { w.record(false, fmt.Sprintf(f, a...)) }
    func (w *watchTB) Fail()                     { w.record(false, "") }
    func (w *watchTB) Fatal(a ...any)            { w.record(true, fmt.Sprint(a...)); runtime.Goexit() }
    func (w *watchTB) Fatalf(f string, a ...any) { w.record(true, fmt.Sprintf(f, a...)); runtime.Goexit() }
    func (w *watchTB) FailNow()                  { w.record(true, ""); runtime.Goexit() }

    func call(t *testing.T, path, got string, update bool) (w *watchTB, finished bool) {
    	w = &watchTB{T: t}
    	done := make(chan struct{})
    	go func() {
    		defer close(done)
    		assertGolden(w, path, got, update)
    		finished = true
    	}()
    	<-done
    	return w, finished
    }

    func TestAssertGoldenMissingFile(t *testing.T) {
    	path := filepath.Join(t.TempDir(), "testdata", "march.golden")
    	w, finished := call(t, path, "statement\n", false)
    	if !w.fatal || finished {
    		t.Fatalf("assertGolden with a missing golden file: want it to stop the test with t.Fatalf (failed=%v, stopped=%v)", w.failed, !finished)
    	}
    	msg := strings.Join(w.messages, "\n")
    	if !strings.Contains(msg, path) || !strings.Contains(msg, "-update") {
    		t.Errorf("assertGolden with a missing golden file said %q; mention the path and -update so the reader knows how to fix it", msg)
    	}
    }

    func TestAssertGoldenUpdate(t *testing.T) {
    	path := filepath.Join(t.TempDir(), "testdata", "march.golden")
    	for _, content := range []string{"first\nversion\n", "second\n"} {
    		w, _ := call(t, path, content, true)
    		if w.failed {
    			t.Fatalf("assertGolden(t, path, %q, update=true) failed: %q", content, w.messages)
    		}
    		data, err := os.ReadFile(path)
    		if err != nil {
    			t.Fatalf("after assertGolden(t, path, %q, update=true), reading the golden file: %v (create the testdata directory if it's missing)", content, err)
    		}
    		if string(data) != content {
    			t.Errorf("after assertGolden(t, path, %q, update=true), the golden file contains %q", content, data)
    		}
    	}
    }

    func TestAssertGoldenCompares(t *testing.T) {
    	dir := t.TempDir()
    	write := func(name, content string) string {
    		p := filepath.Join(dir, name)
    		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
    			t.Fatal(err)
    		}
    		return p
    	}

    	match := write("match.golden", "rent  -$1,200.00\ncash  $50.00\n")
    	if w, _ := call(t, match, "rent  -$1,200.00\ncash  $50.00\n", false); w.failed {
    		t.Errorf("assertGolden reported %q for output that matches the golden file", w.messages)
    	} else if !w.helper {
    		t.Error("assertGolden never calls t.Helper(), so its failures point at the wrong line")
    	}

    	crlf := write("crlf.golden", "rent  -$1,200.00\r\ncash  $50.00\r\n")
    	if w, _ := call(t, crlf, "rent  -$1,200.00\ncash  $50.00\n", false); w.failed {
    		t.Errorf("assertGolden reported %q for a golden file with Windows line endings (\\r\\n) that otherwise matches; treat \\r\\n as \\n", w.messages)
    	}

    	want, got := "rent  -$1,200.00\ncash  $50.00\n", "rent  -$1,200.00\ncash  $50.00 \n"
    	mismatch := write("mismatch.golden", want)
    	w, finished := call(t, mismatch, got, false)
    	if !w.failed {
    		t.Fatal("assertGolden reported nothing for output that doesn't match the golden file")
    	}
    	if w.fatal || !finished {
    		t.Error("assertGolden stopped the test on a mismatch: use t.Errorf, so the rest of the test still runs")
    	}
    	msg := strings.Join(w.messages, "\n")
    	if !strings.Contains(msg, mismatch) {
    		t.Errorf("assertGolden's mismatch message %q doesn't mention the golden file's path", msg)
    	}
    	if d := diffLines(want, got); d == "" || !strings.Contains(msg, strings.TrimSpace(d)) {
    		t.Errorf("assertGolden's mismatch message:\n%s\nshould include diffLines' output:\n%s", msg, d)
    	}
    }
---

Ledgerly's statement tests compare whole reports with **golden files**:
the expected output lives in `testdata/*.golden`, and `go test -update`
rewrites them when the output changes on purpose. Every test copies the same
dozen lines to do that, and when a test fails it prints two 80-line blobs and
leaves you to spot the difference. Time for a proper helper.

## Part 1: diffLines

`diffLines(want, got)` returns `""` if the texts are equal. Otherwise it
compares them **line by line** (split on `"\n"`; line `i` of one against line
`i` of the other, no clever alignment) and describes each differing line:

```text
line 2:
- "cash  $50.00"
+ "cash  $50.00 "
```

- `line N:` uses 1-based line numbers.
- `- %q` shows want's line and `+ %q` shows got's line. Quoting makes
  invisible differences (trailing spaces, tabs, empty lines) visible.
- If one text has no line `N` at all, leave its row out. So
  `diffLines("a\nb", "a")` is `line 2:\n- "b"\n`.
- Show at most **5** differing lines. If there are more, end with
  `... and 3 more differing lines` (with the right count).
- Every line of the result ends with `"\n"`.

## Part 2: assertGolden

`assertGolden(t, path, got, update)` is the helper tests call:

- It marks itself with `t.Helper()`.
- With `update` set, it writes `got` to `path`, creating the parent directory
  if needed, and doesn't compare anything.
- Otherwise it reads the golden file. If it doesn't exist, it stops the test
  with `t.Fatalf`, mentioning the path and that `-update` creates it.
- It treats `\r\n` in the golden file as `\n`, because Git on Windows may
  check files out that way.
- On a mismatch it calls `t.Errorf` (not `Fatalf`, so the rest of the test
  still runs) with the path and the output of `diffLines(want, got)`.

## Example

In a real project the flag and a test look like this:

```go
var update = flag.Bool("update", false, "rewrite golden files")

func TestMarchStatement(t *testing.T) {
	got := Statement(march)
	assertGolden(t, filepath.Join("testdata", "march.golden"), got, *update)
}
```

**Run** shows a diff and then walks `assertGolden` through a missing file,
an update, a match and a mismatch in a temporary directory.

## Constraints

- The grader checks `diffLines`' output character for character, and runs
  `assertGolden` against real files in a temporary directory, with a fake
  `testing.TB` that records what it reports.
