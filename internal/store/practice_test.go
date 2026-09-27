package store

import (
	"database/sql"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"goland-engine/internal/course"
)

func practiceStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "practice.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	ex := &course.Exercise{Starter: "package main", Tests: "package main"}
	courses := []course.Course{
		{Slug: "algos", Title: "Algorithms", Chapters: []course.Chapter{{Slug: "sorting", Title: "Sorting", Lessons: []course.Lesson{{Slug: "bubble", Title: "Bubble", Quiz: []course.Question{{Options: []course.Option{{Correct: true}}}}}}}},
			Problems: []course.Problem{
				{Slug: "hard-one", Title: "Hard One", Difficulty: "hard", After: "sorting", HTML: "<p>Hard.</p>", Hints: []string{"<p>h1</p>", "<p>h2</p>"}, Exercise: ex},
				{Slug: "easy-one", Title: "Easy One", Difficulty: "easy", HTML: "<p>Easy.</p>", Exercise: ex},
				{Slug: "easy-two", Title: "Easy Two", Difficulty: "easy", After: "sorting", HTML: "<p>Easy 2.</p>", Exercise: ex},
			}},
	}
	if err := s.Seed(t.Context(), courses); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestProblems(t *testing.T) {
	s := practiceStore(t)
	ctx := t.Context()
	list, err := s.Problems(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var slugs []string
	for _, p := range list {
		slugs = append(slugs, p.Slug)
	}
	// Easiest first, then file order within a level.
	if want := []string{"easy-one", "easy-two", "hard-one"}; !slices.Equal(slugs, want) {
		t.Errorf("order = %v, want %v", slugs, want)
	}
	if list[1].AfterTitle != "Sorting" || list[1].CourseTitle != "Algorithms" {
		t.Errorf("summary = %+v", list[1])
	}

	p, err := s.Problem(ctx, "algos", "hard-one")
	if err != nil {
		t.Fatal(err)
	}
	if p.Title != "Hard One" || p.Difficulty != "hard" || len(p.Hints) != 2 || p.Hints[1] != "<p>h2</p>" ||
		p.Exercise == nil || p.Exercise.Starter != "package main" || p.Exercise.Solution != "" {
		t.Errorf("problem = %+v (solutions must never be stored)", p)
	}
	if _, err := s.Problem(ctx, "algos", "nope"); err != ErrNotFound {
		t.Errorf("missing problem: err = %v, want ErrNotFound", err)
	}
}

func TestPracticeProgress(t *testing.T) {
	s := practiceStore(t)
	ctx := t.Context()
	day := time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC)
	k := PracticeKey{Course: "algos", Problem: "easy-one"}

	// A run saves code but isn't an attempt; failed submits count as attempts.
	if err := s.SavePractice(ctx, "one", k, "draft", false, false, day); err != nil {
		t.Fatal(err)
	}
	if err := s.SavePractice(ctx, "one", k, "wrong", true, false, day); err != nil {
		t.Fatal(err)
	}
	if err := s.SavePractice(ctx, "one", k, "right", true, true, day); err != nil {
		t.Fatal(err)
	}
	// Passing stays passed even if later code fails.
	if err := s.SavePractice(ctx, "one", k, "broken", true, false, day); err != nil {
		t.Fatal(err)
	}
	p, err := s.PracticeProgress(ctx, "one", k)
	if err != nil || !p.Passed || p.Code != "broken" || p.Attempts != 3 {
		t.Fatalf("progress = %+v, %v; want passed, code broken, 3 attempts", p, err)
	}
	solved, err := s.SolvedProblems(ctx, "one")
	if err != nil || !solved[k] || len(solved) != 1 {
		t.Fatalf("solved = %v, %v", solved, err)
	}

	// A pass counts toward the streak once per day.
	a, err := s.Activity(ctx, "one", day)
	if err != nil || a.Today != 1 {
		t.Fatalf("activity today = %d, %v; want 1", a.Today, err)
	}

	if err := s.ResetProgress(ctx, "one"); err != nil {
		t.Fatal(err)
	}
	if p, _ := s.PracticeProgress(ctx, "one", k); p != (PracticeProgress{}) {
		t.Errorf("after reset: %+v", p)
	}
}

// Databases created before practice existed have an activity table whose
// CHECK constraint rejects kind 'practice'. Open must upgrade it, keeping rows.
func TestActivityMigration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE activity (
 session TEXT NOT NULL, day TEXT NOT NULL,
 kind TEXT NOT NULL CHECK (kind IN ('lesson', 'exercise')),
 course TEXT NOT NULL, chapter TEXT NOT NULL, lesson TEXT NOT NULL,
 PRIMARY KEY (session, day, kind, course, chapter, lesson));
INSERT INTO activity VALUES ('one', '2026-09-26', 'lesson', 'algos', 'sorting', 'bubble');`)
	db.Close()
	if err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := t.Context()
	if err := s.Seed(ctx, nil); err != nil {
		t.Fatal(err)
	}
	day := time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC)
	if err := s.SavePractice(ctx, "one", PracticeKey{"algos", "easy-one"}, "x", true, true, day); err != nil {
		t.Fatalf("practice activity after migration: %v", err)
	}
	a, err := s.Activity(ctx, "one", day)
	if err != nil || a.CurrentStreak != 2 {
		t.Fatalf("streak = %d, %v; want 2 (old row kept plus the new one)", a.CurrentStreak, err)
	}
}
