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

const printSixProblem = `---
title: Print Six
difficulty: easy
after: intro
hints:
  - Use **fmt.Println**.
  - 'Print 6.'
exercise:
  starter: |
    package main

    func main() {}
  solution: |
    package main

    import "fmt"

    func main() { fmt.Println(6) }
  expected_output: "6"
---
Print the number six.
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
		"courses/01-demo/exercises/01-print-six.md":  {Data: []byte(printSixProblem)},
		"courses/01-demo/exercises/02-hard-thing.md": {Data: []byte(strings.NewReplacer("title: Print Six", "title: Hard Thing", "difficulty: easy", "difficulty: hard").Replace(printSixProblem))},
		"courses/02-next/course.yaml":                {Data: []byte("title: Next Up\ntrack: beyond\n")},
		"courses/02-next/01-start/chapter.yaml":      {Data: []byte("title: Start\n")},
		"courses/02-next/01-start/01-first.md":       {Data: []byte(strings.Replace(lessonMD, "title: Hello", "title: First Steps", 1))},
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
		{"/courses/demo/intro/hello", 200, "Hello Again"},
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

// Every lesson links to its neighbours, crossing into the next or previous
// course at the edges, so learners can page through the whole roadmap.
func TestLessonNavigation(t *testing.T) {
	b := &browser{t: t, app: newTestApp(t, nil)}
	tests := []struct {
		path       string
		want, deny []string
	}{
		{"/courses/demo/intro/hello",
			[]string{`href="/courses/demo/intro/hello-again" rel="next"`, "Next lesson", "Hello Again"},
			[]string{`rel="prev"`}},
		{"/courses/demo/intro/print-it",
			[]string{`href="/courses/demo/intro/hello-again" rel="prev"`, `href="/courses/demo/intro/add" rel="next"`},
			nil},
		{"/courses/demo/intro/add",
			[]string{`href="/courses/next/start/first" rel="next"`, "Next course: Next Up"},
			[]string{"Next lesson"}},
		{"/courses/next/start/first",
			[]string{`href="/courses/demo/intro/add" rel="prev"`, "Previous course: Demo", `href="/"`, "Back to the roadmap"},
			[]string{`rel="next"`}},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			_, body := b.do("GET", tt.path, nil, false)
			for _, w := range tt.want {
				if !strings.Contains(body, w) {
					t.Errorf("GET %s: missing %q", tt.path, w)
				}
			}
			for _, d := range tt.deny {
				if strings.Contains(body, d) {
					t.Errorf("GET %s: unexpected %q", tt.path, d)
				}
			}
		})
	}
}

// Courses marked "track: beyond" are listed after the core roadmap under their
// own heading, with numbering that carries on from the core list.
func TestRoadmapTracks(t *testing.T) {
	b := &browser{t: t, app: newTestApp(t, nil)}
	_, body := b.do("GET", "/", nil, false)
	core, beyond := strings.Index(body, ">Demo<"), strings.Index(body, ">Next Up<")
	heading := strings.Index(body, "Beyond the core</h2>")
	if core < 0 || beyond < 0 || heading < 0 || !(core < heading && heading < beyond) {
		t.Fatalf("want Demo, then the Beyond the core heading, then Next Up; indexes %d, %d, %d", core, heading, beyond)
	}
	if !strings.Contains(body, `<ol class="roadmap" start="2">`) {
		t.Error(`beyond list should continue numbering with start="2"`)
	}
}

func TestExercisesPage(t *testing.T) {
	b := &browser{t: t, app: newTestApp(t, nil)}
	tests := []struct {
		path       string
		code       int
		want, deny []string
	}{
		{"/", 200, []string{`href="/exercises"`}, nil},
		{"/exercises", 200,
			[]string{"Print Six", "Hard Thing", `href="/exercises/demo/print-six"`, "badge easy", "badge hard", "Intro", "Easy 0/1"},
			nil},
		{"/exercises?level=hard", 200, []string{"Hard Thing"}, []string{"Print Six"}},
		{"/exercises?course=next", 200, []string{"No exercises match"}, []string{"Print Six"}},
		{"/exercises/demo/print-six", 200,
			[]string{"Print the number six.", "Hint 1", "<strong>fmt.Println</strong>", "Hint 2", `<textarea id="code"`, `href="/courses/demo#intro"`, `href="/exercises/demo/hard-thing" rel="next"`},
			[]string{"fmt.Println(6)"}}, // never leak the solution
		{"/exercises/demo/nope", 404, []string{"doesn't exist"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			code, body := b.do("GET", tt.path, nil, false)
			if code != tt.code {
				t.Fatalf("GET %s = %d, want %d", tt.path, code, tt.code)
			}
			for _, w := range tt.want {
				if !strings.Contains(body, w) {
					t.Errorf("GET %s: missing %q", tt.path, w)
				}
			}
			for _, d := range tt.deny {
				if strings.Contains(body, d) {
					t.Errorf("GET %s: unexpected %q", tt.path, d)
				}
			}
		})
	}
}

func TestSolvingProblems(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles Go programs")
	}
	run, err := runner.New()
	if err != nil {
		t.Fatal(err)
	}
	b := &browser{t: t, app: newTestApp(t, run)}
	const path = "/exercises/demo/print-six"
	send := func(action, code string) string {
		_, body := b.do("POST", path, url.Values{"action": {action}, "code": {code}}, true)
		return body
	}
	if body := send("submit", "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(5) }\n"); !strings.Contains(body, "Not quite") {
		t.Fatalf("wrong answer: %s", body)
	}
	body := send("submit", "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(6) }\n")
	if !strings.Contains(body, "Passed.") || !strings.Contains(body, "Next exercise →") || !strings.Contains(body, `hx-swap-oob="true"><span class="done">✓ Solved`) {
		t.Fatalf("right answer: %s", body)
	}
	if _, body := b.do("GET", "/exercises", nil, false); !strings.Contains(body, "Easy 1/1") {
		t.Errorf("list should count the solved problem:\n%s", body)
	}
	if _, body := b.do("GET", path, nil, false); !strings.Contains(body, "✓ Solved") || !strings.Contains(body, "fmt.Println(6)") {
		t.Errorf("problem page should show solved status and the saved code")
	}
	if _, body := b.do("GET", "/", nil, false); !strings.Contains(body, "<dd>1<span> completed") {
		t.Errorf("a solved problem should count toward today's activity")
	}
	if body := send("reset", ""); !strings.Contains(body, "func main() {}") || strings.Contains(body, "<html") {
		t.Errorf("reset should return the starter panel fragment:\n%s", body)
	}
}
