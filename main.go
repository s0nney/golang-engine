// Command goland-engine serves the goland-engine courses.
package main

import (
	"context"
	"errors"
	"flag"
	"io/fs"
	"log/slog"
	"os"

	"goland-engine/content"
	"goland-engine/internal/course"
	"goland-engine/internal/runner"
	"goland-engine/internal/store"
	"goland-engine/internal/web"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:3000", "listen address (the server runs learners' code, so keep it off public interfaces)")
	dbPath := flag.String("db", "goland.db", "SQLite database path")
	contentDir := flag.String("content", "", "load courses from this directory instead of the embedded ones (for authoring)")
	flag.Parse()

	var fsys fs.FS = content.FS
	if *contentDir != "" {
		fsys = os.DirFS(*contentDir)
	}
	if err := run(fsys, *addr, *dbPath); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(fsys fs.FS, addr, dbPath string) error {
	courses, err := course.Load(fsys, "courses")
	if err != nil {
		// Serve what's valid; `go run ./cmd/validate` lists every problem.
		slog.Warn("skipping invalid course content", "err", err)
	}
	if len(courses) == 0 {
		return errors.New("no valid courses to serve")
	}
	s, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	defer s.Close()
	if err := s.Seed(context.Background(), courses); err != nil {
		return err
	}
	slog.Info("seeded courses", "count", len(courses))
	run, err := runner.New()
	if err != nil {
		slog.Warn("exercises can't be run", "err", err)
	}
	return web.New(s, run).Listen(addr)
}
