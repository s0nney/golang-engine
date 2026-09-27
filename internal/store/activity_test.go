package store

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
	_ "time/tzdata"
)

func TestActivityHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "activity.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	ctx := t.Context()
	day := time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC)
	key := LessonKey{"go", "intro", "hello"}
	for range 2 {
		if err := s.PassQuiz(ctx, "one", key, day); err != nil {
			t.Fatal(err)
		}
		if err := s.SaveCode(ctx, "one", key, "solution", true, day); err != nil {
			t.Fatal(err)
		}
	}
	// Runs, failed submissions, and code resets don't earn activity.
	if err := s.SaveCode(ctx, "one", key, "", false, day.AddDate(0, 0, 1)); err != nil {
		t.Fatal(err)
	}
	a, err := s.Activity(ctx, "one", day)
	if err != nil || a.Today != 2 || a.CurrentStreak != 1 {
		t.Fatalf("activity = %+v, %v", a, err)
	}
	if err := s.PassQuiz(ctx, "one", key, day.AddDate(0, 0, 1)); err != nil {
		t.Fatal(err)
	}
	if err := s.PassQuiz(ctx, "two", key, day); err != nil {
		t.Fatal(err)
	}
	// Content reseeding and reopening must preserve history.
	if err := s.Seed(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	a, err = s.Activity(ctx, "one", day.AddDate(0, 0, 1))
	if err != nil || a.Today != 1 || a.Total != 3 || a.CurrentStreak != 2 {
		t.Fatalf("persisted activity = %+v, %v", a, err)
	}
	if err := s.ResetProgress(ctx, "one"); err != nil {
		t.Fatal(err)
	}
	a, err = s.Activity(ctx, "one", day)
	if err != nil || a.Total != 0 || a.BestStreak != 0 {
		t.Fatalf("reset activity = %+v, %v", a, err)
	}
	p, err := s.Progress(ctx, "one", key)
	if err != nil || p.QuizPassed || p.ExercisePassed {
		t.Fatalf("reset progress = %+v, %v", p, err)
	}
	a, err = s.Activity(ctx, "two", day)
	if err != nil || a.Today != 1 {
		t.Fatalf("other session = %+v, %v", a, err)
	}
}

func TestCalendarStreaks(t *testing.T) {
	for _, tc := range []struct {
		name, today   string
		days          []string
		current, best int
	}{
		{"empty", "2026-09-27", nil, 0, 0},
		{"today", "2026-09-27", []string{"2026-09-27"}, 1, 1},
		{"yesterday grace", "2026-09-27", []string{"2026-09-25", "2026-09-26"}, 2, 2},
		{"broken", "2026-09-27", []string{"2026-09-24", "2026-09-25"}, 0, 2},
		{"best preserved", "2026-09-27", []string{"2025-01-01", "2025-01-02", "2025-01-03", "2026-09-27"}, 1, 3},
		{"year boundary", "2027-01-01", []string{"2026-12-30", "2026-12-31", "2027-01-01"}, 3, 3},
		{"leap day", "2028-03-01", []string{"2028-02-28", "2028-02-29", "2028-03-01"}, 3, 3},
		{"DST spring", "2026-03-09", []string{"2026-03-07", "2026-03-08", "2026-03-09"}, 3, 3},
		{"DST fall", "2026-11-02", []string{"2026-10-31", "2026-11-01", "2026-11-02"}, 3, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loc, err := time.LoadLocation("America/Chicago")
			if err != nil {
				t.Fatal(err)
			}
			today, err := time.ParseInLocation(time.DateOnly, tc.today, loc)
			if err != nil {
				t.Fatal(err)
			}
			counts := map[string]int{}
			for _, d := range tc.days {
				counts[d] = 3
			}
			a := summarizeActivity(counts, today)
			if a.CurrentStreak != tc.current || a.BestStreak != tc.best {
				t.Fatalf("streak = %d/%d, want %d/%d", a.CurrentStreak, a.BestStreak, tc.current, tc.best)
			}
			if len(a.Days) != 182 {
				t.Fatalf("heatmap has %d cells", len(a.Days))
			}
			first, _ := time.Parse(time.DateOnly, a.Days[0].Date)
			if first.Weekday() != time.Sunday {
				t.Error("first column doesn't start on Sunday")
			}
			todayCells := 0
			for _, d := range a.Days {
				if d.Date == tc.today {
					todayCells++
					if d.Future {
						t.Error("today marked future")
					}
				}
				if d.Date > tc.today && (!d.Future || d.Count != 0) {
					t.Error("future cell contains activity")
				}
			}
			if todayCells != 1 {
				t.Errorf("today cells = %d", todayCells)
			}
		})
	}
}

func TestActivityUsesLocalDate(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "local.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		t.Fatal(err)
	}
	utc := time.Date(2026, 9, 28, 2, 0, 0, 0, time.UTC)
	local := utc.In(loc) // September 27 in Chicago.
	if err := s.PassQuiz(t.Context(), "one", LessonKey{"go", "intro", "hello"}, local); err != nil {
		t.Fatal(err)
	}
	a, err := s.Activity(t.Context(), "one", local)
	if err != nil || a.Today != 1 {
		t.Fatalf("local activity: %+v, %v", a, err)
	}
	for _, d := range a.Days {
		if d.Count > 0 && d.Date != "2026-09-27" {
			t.Errorf("recorded date %s", d.Date)
		}
	}
}

// Two tabs or a double-clicked submit write at once; SQLite must wait for
// the lock instead of failing with SQLITE_BUSY.
func TestConcurrentWrites(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "busy.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	ctx := t.Context()
	day := time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC)
	errs := make(chan error, 40)
	for i := range 40 {
		key := LessonKey{"go", "intro", fmt.Sprint("lesson-", i%5)}
		go func() {
			if i%2 == 0 {
				errs <- s.PassQuiz(ctx, "one", key, day)
			} else {
				errs <- s.SaveCode(ctx, "one", key, "code", true, day)
			}
		}()
	}
	for range 40 {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
	// Concurrent retries must neither lose the other completion kind nor
	// inflate the calendar: five lessons each have one quiz and one exercise.
	a, err := s.Activity(ctx, "one", day)
	if err != nil {
		t.Fatal(err)
	}
	if a.Today != 10 || a.Total != 10 || a.ActiveDays != 1 || a.CurrentStreak != 1 {
		t.Fatalf("concurrent activity = %+v", a)
	}
	for i := range 5 {
		p, err := s.Progress(ctx, "one", LessonKey{"go", "intro", fmt.Sprint("lesson-", i)})
		if err != nil {
			t.Fatal(err)
		}
		if !p.QuizPassed || !p.ExercisePassed || p.Code != "code" {
			t.Errorf("lesson %d lost concurrent progress: %+v", i, p)
		}
	}
}
