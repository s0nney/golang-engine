package main

import (
	"strings"
	"testing"

	"goland-engine/internal/course"
)

func TestExerciseCheckRejectsBrokenStarter(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles Go programs")
	}
	courses := []course.Course{{Slug: "demo", Chapters: []course.Chapter{{Slug: "intro", Lessons: []course.Lesson{{
		Slug: "print", Exercise: &course.Exercise{
			Starter:        "package main\nfunc main() { unused := 1 }",
			Solution:       "package main\nimport \"fmt\"\nfunc main() { fmt.Println(42) }",
			ExpectedOutput: "42",
		},
	}}}}}}
	err := checkExercises(courses)
	if err == nil || !strings.Contains(err.Error(), "starter doesn't compile") {
		t.Fatalf("error = %v, want starter compilation failure", err)
	}
}
