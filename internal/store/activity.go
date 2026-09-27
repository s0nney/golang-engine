package store

import (
	"context"
	"database/sql"
	"time"
)

const activitySchema = `
CREATE TABLE IF NOT EXISTS activity (
 session TEXT NOT NULL,
 day TEXT NOT NULL,
 kind TEXT NOT NULL CHECK (kind IN ('lesson', 'exercise')),
 course TEXT NOT NULL,
 chapter TEXT NOT NULL,
 lesson TEXT NOT NULL,
 PRIMARY KEY (session, day, kind, course, chapter, lesson)
);`

// ActivityDay is a calendar day, including empty days in the heatmap.
type ActivityDay struct {
	Date   string
	Count  int
	Future bool
}

// Activity summarizes passed lesson quizzes and exercises for one browser.
type Activity struct {
	Days          []ActivityDay
	CurrentStreak int
	BestStreak    int
	Today         int
	Total         int // quiz and exercise completions in the displayed period
	ActiveDays    int // active days in the displayed period
	Timezone      string
}

// recordActivity is part of the progress transaction. Each passed quiz or
// exercise counts once per calendar day, including reviews on later days.
func recordActivity(ctx context.Context, tx *sql.Tx, session string, k LessonKey, kind string, at time.Time) error {
	_, err := tx.ExecContext(ctx, `
 INSERT INTO activity (session, day, kind, course, chapter, lesson)
 VALUES (?, ?, ?, ?, ?, ?)
 ON CONFLICT DO NOTHING`, session, at.Format(time.DateOnly), kind, k.Course, k.Chapter, k.Lesson)
	return err
}

// Activity uses calendar dates, not elapsed 24-hour periods, so DST doesn't
// shorten or lengthen a streak. Historical dates stay as recorded when studying.
func (s *Store) Activity(ctx context.Context, session string, now time.Time) (Activity, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT day, count(*) FROM activity
 WHERE session = ? AND day <= ? GROUP BY day ORDER BY day`, session, now.Format(time.DateOnly))
	if err != nil {
		return Activity{}, err
	}
	defer rows.Close()
	counts := make(map[string]int)
	for rows.Next() {
		var day string
		var count int
		if err := rows.Scan(&day, &count); err != nil {
			return Activity{}, err
		}
		counts[day] = count
	}
	if err := rows.Err(); err != nil {
		return Activity{}, err
	}
	return summarizeActivity(counts, now), nil
}

func summarizeActivity(counts map[string]int, now time.Time) Activity {
	// Work in UTC with local date components: these values represent dates only.
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	a := Activity{Timezone: now.Location().String(), Today: counts[today.Format(time.DateOnly)]}
	last := today
	if a.Today == 0 {
		last = last.AddDate(0, 0, -1)
	}
	for counts[last.Format(time.DateOnly)] > 0 {
		a.CurrentStreak++
		last = last.AddDate(0, 0, -1)
	}
	// Count a run only at its first day, keeping total work linear in history.
	for day, count := range counts {
		date, err := time.Parse(time.DateOnly, day)
		if err != nil || count <= 0 || date.After(today) || counts[date.AddDate(0, 0, -1).Format(time.DateOnly)] > 0 {
			continue
		}
		run := 0
		for !date.After(today) && counts[date.Format(time.DateOnly)] > 0 {
			run++
			date = date.AddDate(0, 0, 1)
		}
		a.BestStreak = max(a.BestStreak, run)
	}
	start := today.AddDate(0, 0, -int(today.Weekday())-25*7)
	for i := range 26 * 7 {
		date := start.AddDate(0, 0, i)
		day := ActivityDay{Date: date.Format(time.DateOnly), Future: date.After(today)}
		if !day.Future {
			day.Count = counts[day.Date]
			a.Total += day.Count
			if day.Count > 0 {
				a.ActiveDays++
			}
		}
		a.Days = append(a.Days, day)
	}
	return a
}
