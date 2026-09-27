package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gofiber/fiber/v3"

	"goland-engine/internal/course"
	"goland-engine/internal/runner"
	"goland-engine/internal/store"
)

const lessonMD = `---
title: Hello
quiz:
  - question: What does ` + "`fmt.Println`" + ` do?
    options:
      - text: Nothing
      - text: Prints a line
        correct: true
    explanation: It prints a line.
---
Body text.
`

const outputExercise = `---
title: Print It
exercise:
  starter: |
    package main

    func main() {}
  solution: |
    package main

    import "fmt"

    func main() { fmt.Println("hello, textio") }
  expected_output: |
    hello, textio
---
Print hello.
`

const testsExercise = `---
title: Add
exercise:
  starter: |
    package main

    func add(a, b int) int { return 0 }

    func main() {}
  solution: |
    package main

    func add(a, b int) int { return a + b }

    func main() {}
  tests: |
    package main

    import "testing"

    func TestAdd(t *testing.T) {
    	if add(2, 3) != 5 {
    		t.Fatal("add(2, 3) != 5")
    	}
    }
---
Write add.
`

func newTestApp(t *testing.T, run *runner.Runner) *fiber.App {
	t.Helper()
	fsys := fstest.MapFS{
		"courses/01-demo/course.yaml":                {Data: []byte("title: Demo\ndescription: A demo.\n")},
		"courses/01-demo/01-intro/chapter.yaml":      {Data: []byte("title: Intro\n")},
		"courses/01-demo/01-intro/01-hello.md":       {Data: []byte(lessonMD)},
		"courses/01-demo/01-intro/02-hello-again.md": {Data: []byte(strings.Replace(lessonMD, "title: Hello", "title: Hello Again", 1))},
		"courses/01-demo/01-intro/03-print-it.md":    {Data: []byte(outputExercise)},
		"courses/01-demo/01-intro/04-add.md":         {Data: []byte(testsExercise)},
	}
	courses, err := course.Load(fsys, "courses")
	if err != nil {
		t.Fatal(err)
	}
	s, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	if err := s.Seed(t.Context(), courses); err != nil {
		t.Fatal(err)
	}
	return New(s, run)
}

// browser is a minimal client that keeps the session cookie between requests.
type browser struct {
	t      *testing.T
	app    *fiber.App
	cookie *http.Cookie
}

func (b *browser) do(method, path string, form url.Values, htmx bool) (int, string) {
	b.t.Helper()
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	r := httptest.NewRequest(method, path, body)
	if form != nil {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if htmx {
		r.Header.Set("HX-Request", "true")
	}
	if b.cookie != nil {
		r.AddCookie(b.cookie)
	}
	res, err := b.app.Test(r, fiber.TestConfig{}) // no timeout: exercises compile code
	if err != nil {
		b.t.Fatal(err)
	}
	defer res.Body.Close()
	for _, c := range res.Cookies() {
		if c.Name == sessionCookie {
			b.cookie = c
		}
	}
	out, _ := io.ReadAll(res.Body)
	return res.StatusCode, string(out)
}

func TestPages(t *testing.T) {
	b := &browser{t: t, app: newTestApp(t, nil)}
	tests := []struct {
		path string
		code int
		want string
	}{
		{"/", 200, "Demo"},
		{"/courses/demo", 200, "Start course →"},
		{"/courses/demo/intro/hello", 200, "Hello Again →"},
		{"/courses/demo/intro/print-it", 200, `<textarea id="code"`},
		{"/courses/demo/intro/nope", 404, "doesn't exist"},
		{"/courses/nope", 404, "doesn't exist"},
		{"/static/style.css", 200, "--accent"},
		{"/static/htmx.min.js", 200, "htmx"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			code, body := b.do("GET", tt.path, nil, false)
			if code != tt.code || !strings.Contains(body, tt.want) {
				t.Errorf("GET %s = %d, want %d containing %q; body:\n%s", tt.path, code, tt.code, tt.want, body)
			}
		})
	}
}

func TestQuizProgress(t *testing.T) {
	b := &browser{t: t, app: newTestApp(t, nil)}
	quiz := func(answer string) string {
		_, body := b.do("POST", "/courses/demo/intro/hello", url.Values{"action": {"quiz"}, "q0": {answer}}, false)
		return body
	}
	if body := quiz("0"); !strings.Contains(body, "Not quite") || strings.Contains(body, "All correct.") {
		t.Errorf("wrong answer accepted:\n%s", body)
	}
	if body := quiz("9"); !strings.Contains(body, "Not quite") {
		t.Errorf("out of range answer accepted:\n%s", body)
	}
	if _, body := b.do("GET", "/courses/demo", nil, false); !strings.Contains(body, "Start course →") {
		t.Errorf("progress recorded for a failed quiz:\n%s", body)
	}

	if body := quiz("1"); !strings.Contains(body, "All correct.") || !strings.Contains(body, "✓ Completed") {
		t.Errorf("correct answer not accepted:\n%s", body)
	}
	if _, body := b.do("GET", "/courses/demo", nil, false); !strings.Contains(body, "Continue: Hello Again →") {
		t.Errorf("course page doesn't continue from the next lesson:\n%s", body)
	}
	if _, body := b.do("GET", "/", nil, false); !strings.Contains(body, "1 done") {
		t.Errorf("home page doesn't show progress:\n%s", body)
	}

	// A different browser has its own progress.
	other := &browser{t: t, app: b.app}
	if _, body := other.do("GET", "/courses/demo", nil, false); !strings.Contains(body, "Start course →") {
		t.Errorf("progress leaked to another session:\n%s", body)
	}

	b.do("POST", "/progress/reset", url.Values{}, false)
	if _, body := b.do("GET", "/courses/demo", nil, false); !strings.Contains(body, "Start course →") {
		t.Errorf("reset didn't clear progress:\n%s", body)
	}
}

func TestExercises(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles Go programs")
	}
	run, err := runner.New()
	if err != nil {
		t.Fatal(err)
	}
	b := &browser{t: t, app: newTestApp(t, run)}
	const printIt = "/courses/demo/intro/print-it"
	helloMain := "package main\r\n\r\nimport \"fmt\"\r\n\r\nfunc main() { fmt.Println(\"hello, textio\") }\r\n"
	wrongMain := "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"bye\") }\n"
	addMain := "package main\n\nfunc add(a, b int) int { return a + b }\n\nfunc main() {}\n"
	send := func(path, action, code string, htmx bool) string {
		_, body := b.do("POST", path, url.Values{"action": {action}, "code": {code}}, htmx)
		return body
	}

	t.Run("run is never graded", func(t *testing.T) {
		body := send(printIt, "run", helloMain, true)
		if !strings.Contains(body, "not graded") || !strings.Contains(body, "hello, textio") || strings.Contains(body, "<html") {
			t.Errorf("unexpected run fragment:\n%s", body)
		}
	})
	t.Run("wrong output shows a comparison", func(t *testing.T) {
		body := send(printIt, "submit", wrongMain, true)
		if !strings.Contains(body, "Not quite") || !strings.Contains(body, "Expected") {
			t.Errorf("wrong output accepted:\n%s", body)
		}
	})
	t.Run("saved code is restored", func(t *testing.T) {
		if _, body := b.do("GET", printIt, nil, false); !strings.Contains(body, "fmt.Println(&#34;bye&#34;)") {
			t.Errorf("last code not restored:\n%s", body)
		}
	})
	t.Run("matching output passes", func(t *testing.T) {
		body := send(printIt, "submit", helloMain, true)
		if !strings.Contains(body, "Passed.") || !strings.Contains(body, `hx-swap-oob="true"`) {
			t.Errorf("correct output rejected:\n%s", body)
		}
	})
	t.Run("tests grade submissions", func(t *testing.T) {
		if body := send("/courses/demo/intro/add", "submit", addMain, false); !strings.Contains(body, "Passed.") || !strings.Contains(body, "<html") {
			t.Errorf("full-page submit failed:\n%s", body)
		}
	})
	t.Run("reset restores the starter", func(t *testing.T) {
		body := send(printIt, "reset", "", true)
		if !strings.Contains(body, `id="exercise"`) || !strings.Contains(body, "func main() {}") {
			t.Errorf("reset didn't restore starter:\n%s", body)
		}
		if _, home := b.do("GET", "/", nil, false); !strings.Contains(home, "2 completions · 1 active day") {
			t.Error("exercise submissions should record two activities; runs, failures and resets should not")
		}
		// Passing is kept after a reset.
		if _, body := b.do("GET", "/courses/demo", nil, false); !strings.Contains(body, "Continue: Hello →") {
			t.Errorf("unexpected course page:\n%s", body)
		}
	})
}

func TestActivityDashboard(t *testing.T) {
	b := &browser{t: t, app: newTestApp(t, nil)}
	check := func(want string) {
		t.Helper()
		status, body := b.do("GET", "/", nil, false)
		if status != 200 || !strings.Contains(body, want) {
			t.Fatalf("dashboard missing %q (status %d)", want, status)
		}
	}
	check("0 completions · 0 active days")
	b.do("POST", "/courses/demo/intro/hello", url.Values{"action": {"quiz"}, "q0": {"0"}}, false)
	check("0 completions · 0 active days")
	for range 2 {
		b.do("POST", "/courses/demo/intro/hello", url.Values{"action": {"quiz"}, "q0": {"1"}}, false)
	}
	check("1 completion · 1 active day")
	check(`data-count="1"`)
	// An exercise-only lesson cannot award activity via an empty quiz.
	if status, _ := b.do("POST", "/courses/demo/intro/print-it", url.Values{"action": {"quiz"}}, false); status != 400 {
		t.Errorf("empty quiz status = %d, want 400", status)
	}
	check("1 completion · 1 active day")
	other := &browser{t: t, app: b.app}
	if _, body := other.do("GET", "/", nil, false); !strings.Contains(body, "0 completions · 0 active days") {
		t.Error("activity leaked between browsers")
	}
	b.do("POST", "/progress/reset", url.Values{}, false)
	check("0 completions · 0 active days")
}

func TestBrowserTimezone(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error { return c.SendString(browserNow(c).Location().String()) })
	for _, tc := range []struct{ cookie, want string }{
		{"", "UTC"}, {"America%2FChicago", "America/Chicago"}, {"Mars%2FInvalid", "UTC"}, {"Local", "UTC"}, {"%zz", "UTC"},
	} {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "goland_timezone", Value: tc.cookie})
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		if string(body) != tc.want {
			t.Errorf("timezone %q = %q, want %q", tc.cookie, body, tc.want)
		}
	}
}
