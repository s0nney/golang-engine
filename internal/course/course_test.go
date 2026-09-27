package course

import (
	"strings"
	"testing"
	"testing/fstest"
)

const questionYAML = `  - question: Which value?
    options:
      - text: one
        correct: true
      - text: two
    explanation: One is correct.
`

func TestQuizAuthoringLimits(t *testing.T) {
	for _, tc := range []struct{ name, quiz, want string }{
		{"valid", questionYAML, ""},
		{"three questions", strings.Repeat(questionYAML, 3), ""},
		{"too many questions", strings.Repeat(questionYAML, 4), "at most 3"},
		{"too few options", strings.Replace(questionYAML, "      - text: two\n", "", 1), "2 to 5"},
		{"too many options", strings.Replace(questionYAML, "    explanation:", strings.Repeat("      - text: extra\n", 4)+"    explanation:", 1), "2 to 5"},
		{"blank option", strings.Replace(questionYAML, "text: two", "text: ' '", 1), "missing text"},
		{"two correct", strings.Replace(questionYAML, "      - text: two", "      - text: two\n        correct: true", 1), "exactly 1 correct"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fsys := fstest.MapFS{"01-example.md": {Data: []byte("---\ntitle: Example\nquiz:\n" + tc.quiz + "---\nLesson body.\n")}}
			_, err := loadLesson(fsys, "01-example.md")
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestLoadReportsBrokenLessonAndKeepsValidContent(t *testing.T) {
	fsys := fstest.MapFS{
		"courses/01-demo/course.yaml":           {Data: []byte("title: Demo\n")},
		"courses/01-demo/01-intro/chapter.yaml": {Data: []byte("title: Intro\n")},
		"courses/01-demo/01-intro/01-valid.md":  {Data: []byte("---\ntitle: Valid\nquiz:\n" + questionYAML + "---\nValid lesson.\n")},
		"courses/01-demo/01-intro/02-broken.md": {Data: []byte("No frontmatter")},
	}
	courses, err := Load(fsys, "courses")
	if err == nil || !strings.Contains(err.Error(), "02-broken.md") {
		t.Fatalf("error = %v", err)
	}
	if len(courses) != 1 || len(courses[0].Chapters) != 1 || len(courses[0].Chapters[0].Lessons) != 1 {
		t.Fatalf("valid material was lost: %+v", courses)
	}
}

func TestCourseTrack(t *testing.T) {
	for _, tc := range []struct{ name, yaml, want string }{
		{"default core", "title: Demo\n", ""},
		{"core", "title: Demo\ntrack: core\n", ""},
		{"beyond", "title: Demo\ntrack: beyond\n", ""},
		{"unknown", "title: Demo\ntrack: advanced\n", "track must be core or beyond"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fsys := fstest.MapFS{
				"01-demo/course.yaml":           {Data: []byte(tc.yaml)},
				"01-demo/01-intro/chapter.yaml": {Data: []byte("title: Intro\n")},
				"01-demo/01-intro/01-valid.md":  {Data: []byte("---\ntitle: Valid\nquiz:\n" + questionYAML + "---\nValid lesson.\n")},
			}
			c, err := LoadCourse(fsys, "01-demo")
			if tc.want != "" {
				if err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("error = %v, want %q", err, tc.want)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got, want := c.Beyond, tc.name == "beyond"; got != want {
				t.Errorf("Beyond = %v, want %v", got, want)
			}
		})
	}
}

const problemMD = `---
title: Two Sum
difficulty: medium
after: intro
hints:
  - A map remembers what you've **seen**.
  - Look for target - n.
exercise:
  starter: |
    package main

    func main() {}
  solution: |
    package main

    import "fmt"

    func main() { fmt.Println("ok") }
  expected_output: ok
---
Find two numbers.
`

func problemFS(problem string) fstest.MapFS {
	return fstest.MapFS{
		"01-demo/course.yaml":             {Data: []byte("title: Demo\n")},
		"01-demo/01-intro/chapter.yaml":   {Data: []byte("title: Intro\n")},
		"01-demo/01-intro/01-valid.md":    {Data: []byte("---\ntitle: Valid\nquiz:\n" + questionYAML + "---\nValid lesson.\n")},
		"01-demo/exercises/01-two-sum.md": {Data: []byte(problem)},
	}
}

func TestLoadProblems(t *testing.T) {
	c, err := LoadCourse(problemFS(problemMD), "01-demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Chapters) != 1 {
		t.Fatalf("exercises/ must not load as a chapter: %d chapters", len(c.Chapters))
	}
	if len(c.Problems) != 1 {
		t.Fatalf("got %d problems, want 1", len(c.Problems))
	}
	p := c.Problems[0]
	if p.Slug != "two-sum" || p.Title != "Two Sum" || p.Difficulty != "medium" || p.After != "intro" {
		t.Errorf("problem = %+v", p)
	}
	if len(p.Hints) != 2 || !strings.Contains(p.Hints[0], "<strong>seen</strong>") {
		t.Errorf("hints = %q, want 2 rendered hints", p.Hints)
	}
	if p.Exercise == nil || !strings.Contains(p.HTML, "Find two numbers.") {
		t.Errorf("exercise or body missing: %+v", p)
	}
}

func TestProblemValidation(t *testing.T) {
	for _, tc := range []struct{ name, old, new, want string }{
		{"bad difficulty", "difficulty: medium", "difficulty: brutal", "difficulty must be easy, medium or hard"},
		{"missing difficulty", "difficulty: medium\n", "", "difficulty must be easy, medium or hard"},
		{"unknown chapter", "after: intro", "after: nowhere", `after: no chapter "nowhere"`},
		{"no exercise", "exercise:", "notes:", "needs an exercise"},
		{"blank hint", "  - Look for target - n.", "  - ' '", "hints[1]: empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadCourse(problemFS(strings.Replace(problemMD, tc.old, tc.new, 1)), "01-demo")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}
