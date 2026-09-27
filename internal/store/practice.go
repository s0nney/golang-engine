package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"goland-engine/internal/course"
)

// Practice progress is keyed by slugs, like lesson progress, so it survives
// the content tables being rebuilt.
const practiceSchema = `
CREATE TABLE IF NOT EXISTS practice (
	session    TEXT NOT NULL,
	course     TEXT NOT NULL,
	problem    TEXT NOT NULL,
	passed     INTEGER NOT NULL DEFAULT 0,
	code       TEXT NOT NULL DEFAULT '',
	attempts   INTEGER NOT NULL DEFAULT 0,
	updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (session, course, problem)
);`

// PracticeKey identifies a practice problem.
type PracticeKey struct{ Course, Problem string }

// PracticeProgress is one learner's state on one problem.
type PracticeProgress struct {
	Passed   bool
	Code     string // last code run or submitted, "" if none
	Attempts int    // submissions, passed or not
}

// ProblemSummary is a problem as listed on the exercises page.
type ProblemSummary struct {
	Course      string
	CourseTitle string
	Slug        string
	Title       string
	Difficulty  string
	After       string // chapter slug, "" if none
	AfterTitle  string
}

// Problem is a problem with its statement, hints and exercise. The
// reference solution isn't stored, so it's never included.
type Problem struct {
	course.Problem
	Course      string // course slug
	CourseTitle string
	AfterTitle  string
}

const problemColumns = `
	SELECT c.slug, c.title, p.slug, p.title, p.difficulty, p.after_chapter, coalesce(ch.title, '')
	FROM problems p
	JOIN courses c ON c.id = p.course_id
	LEFT JOIN chapters ch ON ch.course_id = c.id AND ch.slug = p.after_chapter`

// difficultyOrder sorts easy before medium before hard.
const difficultyOrder = `CASE p.difficulty WHEN 'easy' THEN 0 WHEN 'medium' THEN 1 ELSE 2 END`

// Problems lists every problem in roadmap order, easiest first within a course.
func (s *Store) Problems(ctx context.Context) ([]ProblemSummary, error) {
	rows, err := s.db.QueryContext(ctx, problemColumns+`
		ORDER BY c.position, `+difficultyOrder+`, p.position`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProblemSummary
	for rows.Next() {
		var p ProblemSummary
		if err := rows.Scan(&p.Course, &p.CourseTitle, &p.Slug, &p.Title, &p.Difficulty, &p.After, &p.AfterTitle); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Problem returns one problem with its statement, hints and exercise.
func (s *Store) Problem(ctx context.Context, courseSlug, slug string) (Problem, error) {
	var (
		p  Problem
		e  course.Exercise
		id int64
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT p.id, p.slug, p.title, p.html, p.difficulty, p.after_chapter, c.title, coalesce(ch.title, ''),
		       p.starter, p.tests, p.expected_output
		FROM problems p
		JOIN courses c ON c.id = p.course_id
		LEFT JOIN chapters ch ON ch.course_id = c.id AND ch.slug = p.after_chapter
		WHERE c.slug = ? AND p.slug = ?`, courseSlug, slug,
	).Scan(&id, &p.Slug, &p.Title, &p.HTML, &p.Difficulty, &p.After, &p.CourseTitle, &p.AfterTitle,
		&e.Starter, &e.Tests, &e.ExpectedOutput)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrNotFound
	}
	if err != nil {
		return p, err
	}
	p.Course = courseSlug
	p.Exercise = &e
	rows, err := s.db.QueryContext(ctx, `SELECT html FROM hints WHERE problem_id = ? ORDER BY position`, id)
	if err != nil {
		return p, err
	}
	defer rows.Close()
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return p, err
		}
		p.Hints = append(p.Hints, h)
	}
	return p, rows.Err()
}

// PracticeProgress returns session's state on one problem.
func (s *Store) PracticeProgress(ctx context.Context, session string, k PracticeKey) (PracticeProgress, error) {
	var p PracticeProgress
	err := s.db.QueryRowContext(ctx, `
		SELECT passed, code, attempts FROM practice WHERE session = ? AND course = ? AND problem = ?`,
		session, k.Course, k.Problem,
	).Scan(&p.Passed, &p.Code, &p.Attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return p, nil
	}
	return p, err
}

// SolvedProblems returns the set of problems session has passed.
func (s *Store) SolvedProblems(ctx context.Context, session string) (map[PracticeKey]bool, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT course, problem FROM practice WHERE session = ? AND passed`, session)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	solved := make(map[PracticeKey]bool)
	for rows.Next() {
		var k PracticeKey
		if err := rows.Scan(&k.Course, &k.Problem); err != nil {
			return nil, err
		}
		solved[k] = true
	}
	return solved, rows.Err()
}

// SavePractice stores session's latest code for a problem. submitted counts
// an attempt; once passed, a problem stays passed. A pass also counts toward
// the activity streak, once per problem per day.
func (s *Store) SavePractice(ctx context.Context, session string, k PracticeKey, code string, submitted, passed bool, at time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	attempt := 0
	if submitted {
		attempt = 1
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO practice (session, course, problem, code, passed, attempts) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT DO UPDATE SET
			code = excluded.code,
			passed = max(passed, excluded.passed),
			attempts = attempts + excluded.attempts,
			updated_at = CURRENT_TIMESTAMP`,
		session, k.Course, k.Problem, code, passed, attempt)
	if err != nil {
		return err
	}
	if passed {
		if err := recordActivity(ctx, tx, session, LessonKey{Course: k.Course, Lesson: k.Problem}, "practice", at); err != nil {
			return err
		}
	}
	return tx.Commit()
}
