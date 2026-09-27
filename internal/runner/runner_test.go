//go:build unix

package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckDoesNotRunOrDependOnParentGit(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles Go programs")
	}
	dir := t.TempDir()
	// A broken checkout above the scratch module must not break exercise builds.
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", dir)
	r, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Check(t.Context(), "package main\nfunc main() { panic(\"must not run\") }"); !got.OK {
		t.Fatalf("compile-only check failed: %s", got.Output)
	}
	if got := r.Run(t.Context(), "package main\nimport \"fmt\"\nfunc main() { fmt.Println(42) }"); !got.OK || !SameOutput(got.Output, "42") {
		t.Fatalf("Run depended on VCS metadata: %+v", got)
	}
}

const tests = `package main

import "testing"

func TestAdd(t *testing.T) {
	if got := add(2, 3); got != 5 {
		t.Errorf("add(2, 3) = %d, want 5", got)
	}
}
`

func TestRunner(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles Go programs")
	}
	r, err := New()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name    string
		code    string
		tests   string
		ok      bool
		wantOut string
	}{
		{"prints", "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"hi\") }", "", true, "hi\n"},
		{"compile error", "package main\nfunc main() { x := 1 }", "", false, "declared and not used"},
		{"panic", "package main\nfunc main() { panic(\"boom\") }", "", false, "panic: boom"},
		{"infinite loop", "package main\nfunc main() { for {} }", "", false, "Killed: took longer"},
		{"output flood", "package main\nimport \"fmt\"\nfunc main() { for { fmt.Println(\"spam spam spam\") } }", "", false, "(output truncated)"},
		{"tests pass", "package main\nfunc add(a, b int) int { return a + b }\nfunc main() {}", tests, true, "--- PASS: TestAdd"},
		{"tests fail", "package main\nfunc add(a, b int) int { return a - b }\nfunc main() {}", tests, false, "add(2, 3) = -1, want 5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var res Result
			if tc.tests != "" {
				res = r.Test(t.Context(), tc.code, tc.tests)
			} else {
				res = r.Run(t.Context(), tc.code)
			}
			if res.OK != tc.ok || !strings.Contains(res.Output, tc.wantOut) {
				t.Errorf("OK = %v, want %v; output should contain %q:\n%s", res.OK, tc.ok, tc.wantOut, res.Output)
			}
		})
	}
}
