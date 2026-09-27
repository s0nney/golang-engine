// Command validate checks course material and reports every problem found.
//
//	go run ./cmd/validate                                 # all courses
//	go run ./cmd/validate content/courses/01-learn-go     # one course
//	go run ./cmd/validate -exec content/courses/01-learn-go
//
// With -exec it also runs every exercise: the solution must pass and the
// starter must not.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"goland-engine/internal/course"
	"goland-engine/internal/runner"
)

func main() {
	execute := flag.Bool("exec", false, "check starter compilation and grade every exercise's solution and starter")
	flag.Parse()

	var (
		courses []course.Course
		err     error
	)
	if dir := flag.Arg(0); dir != "" {
		dir = filepath.Clean(dir)
		var c course.Course
		c, err = course.LoadCourse(os.DirFS(filepath.Dir(dir)), filepath.Base(dir))
		courses = append(courses, c)
	} else {
		courses, err = course.Load(os.DirFS("content"), "courses")
	}
	for _, c := range courses {
		lessons, questions, exercises := 0, 0, 0
		for _, ch := range c.Chapters {
			lessons += len(ch.Lessons)
			for _, l := range ch.Lessons {
				questions += len(l.Quiz)
				if l.Exercise != nil {
					exercises++
				}
			}
		}
		levels := make(map[string]int)
		for _, p := range c.Problems {
			levels[p.Difficulty]++
		}
		fmt.Printf("%-40s %2d chapters %3d lessons %3d questions %3d exercises %3d problems (%d/%d/%d)\n",
			c.Title, len(c.Chapters), lessons, questions, exercises,
			len(c.Problems), levels["easy"], levels["medium"], levels["hard"])
	}
	if *execute {
		err = errors.Join(err, checkExercises(courses))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func checkExercises(courses []course.Course) error {
	run, err := runner.New()
	if err != nil {
		return err
	}
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)
	fail := func(format string, args ...any) {
		mu.Lock()
		defer mu.Unlock()
		errs = append(errs, fmt.Errorf(format, args...))
	}
	ctx := context.Background()
	type named struct {
		name string
		e    *course.Exercise
	}
	var all []named
	for _, c := range courses {
		for _, ch := range c.Chapters {
			for _, l := range ch.Lessons {
				if l.Exercise != nil {
					all = append(all, named{c.Slug + "/" + ch.Slug + "/" + l.Slug, l.Exercise})
				}
			}
		}
		for _, p := range c.Problems {
			all = append(all, named{c.Slug + "/exercises/" + p.Slug, p.Exercise})
		}
	}
	for _, x := range all {
		name, e := x.name, x.e
		wg.Go(func() {
			if res, ok := run.Submit(ctx, e.Solution, e.Tests, e.ExpectedOutput); !ok {
				fail("%s: solution doesn't pass:\n%s", name, res.Output)
			}
			if res := run.Check(ctx, e.Starter); !res.OK {
				fail("%s: starter doesn't compile:\n%s", name, res.Output)
				return
			}
			if _, ok := run.Submit(ctx, e.Starter, e.Tests, e.ExpectedOutput); ok {
				fail("%s: starter code already passes", name)
			}
		})
	}
	wg.Wait()
	return errors.Join(errs...)
}
