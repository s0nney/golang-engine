package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Progress is keyed by slugs rather than row IDs, because the content tables
// are rebuilt on every start and it must survive that.
const progressSchema = `
CREATE TABLE IF NOT EXISTS progress (
	session         TEXT NOT NULL,
	course          TEXT NOT NULL,
	chapter         TEXT NOT NULL,
	lesson          TEXT NOT NULL,
	quiz_passed     INTEGER NOT NULL DEFAULT 0,
	exercise_passed INTEGER NOT NULL DEFAULT 0,
	code            TEXT NOT NULL DEFAULT '',
	updated_at      TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (session, course, chapter, lesson)
);`

// LessonKey identifies a lesson.
type LessonKey struct{ Course, Chapter, Lesson string }

// LessonProgress is one learner's state on one lesson.
type LessonProgress struct {
	QuizPassed     bool
	ExercisePassed bool
	Code           string // last code run or submitted, "" if none
}

// A lesson is complete once every part it has (quiz, exercise) is passed.
const completeLessons = `
	SELECT c.slug AS course, ch.slug AS chapter, l.slug AS lesson
	FROM progress p
	JOIN courses c   ON c.slug = p.course
	JOIN chapters ch ON ch.course_id = c.id AND ch.slug = p.chapter
	JOIN lessons l   ON l.chapter_id = ch.id AND l.slug = p.lesson
	WHERE p.session = ?
	  AND (p.quiz_passed OR NOT EXISTS (SELECT 1 FROM questions q WHERE q.lesson_id = l.id))
	  AND (p.exercise_passed OR NOT EXISTS (SELECT 1 FROM exercises e WHERE e.lesson_id = l.id))`

// Completed returns the set of lessons session has completed.
func (s *Store) Completed(ctx context.Context, session string) (map[LessonKey]bool, error) {
	rows, err := s.db.QueryContext(ctx, completeLessons, session)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	done := make(map[LessonKey]bool)
	for rows.Next() {
		var k LessonKey
		if err := rows.Scan(&k.Course, &k.Chapter, &k.Lesson); err != nil {
			return nil, err
		}
		done[k] = true
	}
	return done, rows.Err()
}

// Progress returns session's state on one lesson.
func (s *Store) Progress(ctx context.Context, session string, k LessonKey) (LessonProgress, error) {
	var p LessonProgress
	err := s.db.QueryRowContext(ctx, `
		SELECT quiz_passed, exercise_passed, code FROM progress
		WHERE session = ? AND course = ? AND chapter = ? AND lesson = ?`,
		session, k.Course, k.Chapter, k.Lesson,
	).Scan(&p.QuizPassed, &p.ExercisePassed, &p.Code)
	if errors.Is(err, sql.ErrNoRows) {
		return p, nil
	}
	return p, err
}

// PassQuiz records that session answered every question correctly.
func (s *Store) PassQuiz(ctx context.Context, session string, k LessonKey, at time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO progress (session, course, chapter, lesson, quiz_passed) VALUES (?, ?, ?, ?, 1)
		ON CONFLICT DO UPDATE SET quiz_passed = 1, updated_at = CURRENT_TIMESTAMP`,
		session, k.Course, k.Chapter, k.Lesson)
	if err != nil {
		return err
	}
	if err := recordActivity(ctx, tx, session, k, "lesson", at); err != nil {
		return err
	}
	return tx.Commit()
}

// SaveCode stores session's latest code for an exercise. Once an exercise
// is passed it stays passed, even if later code fails.
func (s *Store) SaveCode(ctx context.Context, session string, k LessonKey, code string, passed bool, at time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO progress (session, course, chapter, lesson, code, exercise_passed) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT DO UPDATE SET
			code = excluded.code,
			exercise_passed = max(exercise_passed, excluded.exercise_passed),
			updated_at = CURRENT_TIMESTAMP`,
		session, k.Course, k.Chapter, k.Lesson, code, passed)
	if err != nil {
		return err
	}
	if passed {
		if err := recordActivity(ctx, tx, session, k, "exercise", at); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ResetProgress forgets everything session has done.
func (s *Store) ResetProgress(ctx context.Context, session string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM progress WHERE session = ?`, session); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM activity WHERE session = ?`, session); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM practice WHERE session = ?`, session); err != nil {
		return err
	}
	return tx.Commit()
}
