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
