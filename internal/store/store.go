// Package store keeps course material and learner progress in SQLite.
//
// The embedded content is the source of truth. On startup Seed replaces
// the content tables so the database always matches the binary. Progress
// lives in its own table and is kept.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"

	"goland-engine/internal/course"
)

// ErrNotFound is returned when a course or lesson doesn't exist.
var ErrNotFound = errors.New("not found")

type Store struct{ db *sql.DB }

// CourseSummary is a course as listed on the home page.
type CourseSummary struct {
	Slug        string
	Title       string
	Description string
	Beyond      bool // listed after the core roadmap
	Chapters    int
	Lessons     int
}

const schema = `
DROP TABLE IF EXISTS hints;
DROP TABLE IF EXISTS problems;
DROP TABLE IF EXISTS exercises;
DROP TABLE IF EXISTS options;
DROP TABLE IF EXISTS questions;
DROP TABLE IF EXISTS lessons;
DROP TABLE IF EXISTS chapters;
DROP TABLE IF EXISTS courses;

CREATE TABLE courses (
	id          INTEGER PRIMARY KEY,
	slug        TEXT NOT NULL UNIQUE,
	title       TEXT NOT NULL,
	description TEXT NOT NULL,
	beyond      INTEGER NOT NULL DEFAULT 0,
	position    INTEGER NOT NULL
);
CREATE TABLE chapters (
	id          INTEGER PRIMARY KEY,
	course_id   INTEGER NOT NULL REFERENCES courses(id),
	slug        TEXT NOT NULL,
	title       TEXT NOT NULL,
	description TEXT NOT NULL,
	position    INTEGER NOT NULL,
	UNIQUE (course_id, slug)
);
CREATE TABLE lessons (
	id         INTEGER PRIMARY KEY,
	chapter_id INTEGER NOT NULL REFERENCES chapters(id),
	slug       TEXT NOT NULL,
	title      TEXT NOT NULL,
	html       TEXT NOT NULL,
	position   INTEGER NOT NULL,
	UNIQUE (chapter_id, slug)
);
CREATE TABLE questions (
	id          INTEGER PRIMARY KEY,
	lesson_id   INTEGER NOT NULL REFERENCES lessons(id),
	prompt      TEXT NOT NULL,
	explanation TEXT NOT NULL,
	position    INTEGER NOT NULL
);
CREATE TABLE exercises (
	lesson_id       INTEGER PRIMARY KEY REFERENCES lessons(id),
	starter         TEXT NOT NULL,
	tests           TEXT NOT NULL,
	expected_output TEXT NOT NULL
);
CREATE TABLE problems (
	id              INTEGER PRIMARY KEY,
	course_id       INTEGER NOT NULL REFERENCES courses(id),
	slug            TEXT NOT NULL,
	title           TEXT NOT NULL,
	html            TEXT NOT NULL,
	difficulty      TEXT NOT NULL,
	after_chapter   TEXT NOT NULL,
	position        INTEGER NOT NULL,
	starter         TEXT NOT NULL,
	tests           TEXT NOT NULL,
	expected_output TEXT NOT NULL,
	UNIQUE (course_id, slug)
);
CREATE TABLE hints (
	problem_id INTEGER NOT NULL REFERENCES problems(id),
	html       TEXT NOT NULL,
	position   INTEGER NOT NULL
);
CREATE TABLE options (
	id          INTEGER PRIMARY KEY,
	question_id INTEGER NOT NULL REFERENCES questions(id),
	text        TEXT NOT NULL,
	correct     INTEGER NOT NULL,
	position    INTEGER NOT NULL
);`

// Open opens (creating if needed) the SQLite database at path.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_txlock=immediate")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(progressSchema + activitySchema + practiceSchema); err != nil {
		db.Close()
		return nil, err
	}
	if err := migrateActivity(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrating activity: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// Seed replaces all course material with courses.
func (s *Store) Seed(ctx context.Context, courses []course.Course) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("schema: %w", err)
	}
	for i, c := range courses {
		courseID, err := insert(ctx, tx,
			`INSERT INTO courses (slug, title, description, beyond, position) VALUES (?, ?, ?, ?, ?)`,
			c.Slug, c.Title, c.Description, c.Beyond, i)
		if err != nil {
			return fmt.Errorf("course %s: %w", c.Slug, err)
		}
		for m, p := range c.Problems {
			problemID, err := insert(ctx, tx,
				`INSERT INTO problems (course_id, slug, title, html, difficulty, after_chapter, position, starter, tests, expected_output)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				courseID, p.Slug, p.Title, p.HTML, p.Difficulty, p.After, m, p.Exercise.Starter, p.Exercise.Tests, p.Exercise.ExpectedOutput)
			if err != nil {
				return fmt.Errorf("problem %s/%s: %w", c.Slug, p.Slug, err)
			}
			for n, h := range p.Hints {
				if _, err := insert(ctx, tx, `INSERT INTO hints (problem_id, html, position) VALUES (?, ?, ?)`, problemID, h, n); err != nil {
					return err
				}
			}
		}
		for j, ch := range c.Chapters {
			chapterID, err := insert(ctx, tx,
				`INSERT INTO chapters (course_id, slug, title, description, position) VALUES (?, ?, ?, ?, ?)`,
				courseID, ch.Slug, ch.Title, ch.Description, j)
			if err != nil {
				return fmt.Errorf("chapter %s/%s: %w", c.Slug, ch.Slug, err)
			}
			for k, l := range ch.Lessons {
				lessonID, err := insert(ctx, tx,
					`INSERT INTO lessons (chapter_id, slug, title, html, position) VALUES (?, ?, ?, ?, ?)`,
					chapterID, l.Slug, l.Title, l.HTML, k)
				if err != nil {
					return fmt.Errorf("lesson %s/%s/%s: %w", c.Slug, ch.Slug, l.Slug, err)
				}
				if e := l.Exercise; e != nil {
					if _, err := insert(ctx, tx,
						`INSERT INTO exercises (lesson_id, starter, tests, expected_output) VALUES (?, ?, ?, ?)`,
						lessonID, e.Starter, e.Tests, e.ExpectedOutput); err != nil {
						return err
					}
				}
				for m, q := range l.Quiz {
					questionID, err := insert(ctx, tx,
						`INSERT INTO questions (lesson_id, prompt, explanation, position) VALUES (?, ?, ?, ?)`,
						lessonID, q.Prompt, q.Explanation, m)
					if err != nil {
						return err
					}
					for n, o := range q.Options {
						if _, err := insert(ctx, tx,
							`INSERT INTO options (question_id, text, correct, position) VALUES (?, ?, ?, ?)`,
							questionID, o.Text, o.Correct, n); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return tx.Commit()
}

func insert(ctx context.Context, tx *sql.Tx, query string, args ...any) (int64, error) {
	res, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Courses lists every course in roadmap order.
func (s *Store) Courses(ctx context.Context) ([]CourseSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.slug, c.title, c.description, c.beyond,
		       (SELECT count(*) FROM chapters ch WHERE ch.course_id = c.id),
		       (SELECT count(*) FROM lessons l JOIN chapters ch ON ch.id = l.chapter_id WHERE ch.course_id = c.id)
		FROM courses c ORDER BY c.position`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CourseSummary
	for rows.Next() {
		var c CourseSummary
		if err := rows.Scan(&c.Slug, &c.Title, &c.Description, &c.Beyond, &c.Chapters, &c.Lessons); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Outline returns a course with its chapters and lesson titles, without
// lesson bodies or quizzes.
func (s *Store) Outline(ctx context.Context, courseSlug string) (course.Course, error) {
	var c course.Course
	err := s.db.QueryRowContext(ctx,
		`SELECT slug, title, description FROM courses WHERE slug = ?`, courseSlug,
	).Scan(&c.Slug, &c.Title, &c.Description)
	if errors.Is(err, sql.ErrNoRows) {
		return c, ErrNotFound
	}
	if err != nil {
		return c, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT ch.slug, ch.title, ch.description, l.slug, l.title
		FROM chapters ch
		JOIN courses c ON c.id = ch.course_id
		JOIN lessons l ON l.chapter_id = ch.id
		WHERE c.slug = ?
		ORDER BY ch.position, l.position`, courseSlug)
	if err != nil {
		return c, err
	}
	defer rows.Close()
	for rows.Next() {
		var ch course.Chapter
		var l course.Lesson
		if err := rows.Scan(&ch.Slug, &ch.Title, &ch.Description, &l.Slug, &l.Title); err != nil {
			return c, err
		}
		if n := len(c.Chapters); n == 0 || c.Chapters[n-1].Slug != ch.Slug {
			c.Chapters = append(c.Chapters, ch)
		}
		last := &c.Chapters[len(c.Chapters)-1]
		last.Lessons = append(last.Lessons, l)
	}
	return c, rows.Err()
}

// Lesson returns one lesson with its body, quiz and exercise. The
// exercise's reference solution isn't stored, so it's never included.
func (s *Store) Lesson(ctx context.Context, courseSlug, chapterSlug, lessonSlug string) (course.Lesson, error) {
	var (
		l  course.Lesson
		id int64
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT l.id, l.slug, l.title, l.html
		FROM lessons l
		JOIN chapters ch ON ch.id = l.chapter_id
		JOIN courses c ON c.id = ch.course_id
		WHERE c.slug = ? AND ch.slug = ? AND l.slug = ?`,
		courseSlug, chapterSlug, lessonSlug,
	).Scan(&id, &l.Slug, &l.Title, &l.HTML)
	if errors.Is(err, sql.ErrNoRows) {
		return l, ErrNotFound
	}
	if err != nil {
		return l, err
	}

	var e course.Exercise
	err = s.db.QueryRowContext(ctx,
		`SELECT starter, tests, expected_output FROM exercises WHERE lesson_id = ?`, id,
	).Scan(&e.Starter, &e.Tests, &e.ExpectedOutput)
	switch {
	case err == nil:
		l.Exercise = &e
	case !errors.Is(err, sql.ErrNoRows):
		return l, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT q.id, q.prompt, q.explanation, o.text, o.correct
		FROM questions q
		JOIN options o ON o.question_id = q.id
		WHERE q.lesson_id = ?
		ORDER BY q.position, o.position`, id)
	if err != nil {
		return l, err
	}
	defer rows.Close()
	lastID := int64(-1)
	for rows.Next() {
		var (
			qid int64
			q   course.Question
			o   course.Option
		)
		if err := rows.Scan(&qid, &q.Prompt, &q.Explanation, &o.Text, &o.Correct); err != nil {
			return l, err
		}
		if qid != lastID {
			l.Quiz = append(l.Quiz, q)
			lastID = qid
		}
		last := &l.Quiz[len(l.Quiz)-1]
		last.Options = append(last.Options, o)
	}
	return l, rows.Err()
}
