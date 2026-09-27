// Package web serves the courses over HTTP.
package web

import (
	"cmp"
	"embed"
	"errors"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
	"uuid"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"

	"goland-engine/internal/course"
	"goland-engine/internal/runner"
	"goland-engine/internal/store"
	"goland-engine/internal/web/views"
)

//go:embed static
var staticFS embed.FS

const (
	sessionCookie = "goland_session"
	maxCodeBytes  = 64 << 10
)

// New returns the app with every route registered. run may be nil, in
// which case exercises can't be run.
func New(s *store.Store, run *runner.Runner) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "goland-engine",
		ErrorHandler: errorHandler,
		BodyLimit:    256 << 10,
	})
	h := handler{store: s, runner: run}
	app.Use("/static", static.New("static", static.Config{FS: staticFS, MaxAge: 3600, Compress: true}))
	app.Use(session)
	app.Get("/", h.home)
	app.Post("/progress/reset", h.resetProgress)
	app.Get("/courses/:course", h.course)
	app.Get("/courses/:course/:chapter/:lesson", h.lesson)
	app.Post("/courses/:course/:chapter/:lesson", h.lesson)
	app.Use(func(fiber.Ctx) error { return fiber.ErrNotFound })
	return app
}

// session gives every browser a random anonymous ID in a long-lived cookie.
// Progress is stored against that ID; there are no accounts.
func session(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Cookies(sessionCookie))
	if err != nil {
		id = uuid.NewV4()
	}
	// Refresh on every visit so active learners never expire.
	c.Cookie(&fiber.Cookie{
		Name:     sessionCookie,
		Value:    id.String(),
		Path:     "/",
		Expires:  time.Now().AddDate(1, 0, 0),
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
	c.Locals(sessionCookie, id.String())
	return c.Next()
}

func sessionID(c fiber.Ctx) string { return c.Locals(sessionCookie).(string) }

type handler struct {
	store  *store.Store
	runner *runner.Runner
}

func (h handler) home(c fiber.Ctx) error {
	courses, err := h.store.Courses(c.Context())
	if err != nil {
		return err
	}
	done, err := h.store.Completed(c.Context(), sessionID(c))
	if err != nil {
		return err
	}
	cards := make([]views.CourseCard, len(courses))
	for i, cs := range courses {
		cards[i].CourseSummary = cs
	}
	for k := range done {
		for i := range cards {
			if cards[i].Slug == k.Course {
				cards[i].Done++
			}
		}
	}
	activity, err := h.store.Activity(c.Context(), sessionID(c), browserNow(c))
	if err != nil {
		return err
	}
	return render(c, views.Home(cards, activity))
}

func (h handler) resetProgress(c fiber.Ctx) error {
	if err := h.store.ResetProgress(c.Context(), sessionID(c)); err != nil {
		return err
	}
	return c.Redirect().Status(fiber.StatusSeeOther).To("/")
}

func (h handler) course(c fiber.Ctx) error {
	outline, err := h.store.Outline(c.Context(), c.Params("course"))
	if err != nil {
		return err
	}
	done, err := h.store.Completed(c.Context(), sessionID(c))
	if err != nil {
		return err
	}
	v := views.Outline{Course: outline, Done: make(map[store.LessonKey]bool)}
	for _, ch := range outline.Chapters {
		for _, l := range ch.Lessons {
			k := store.LessonKey{Course: outline.Slug, Chapter: ch.Slug, Lesson: l.Slug}
			switch {
			case done[k]:
				v.Done[k] = true
			case v.Continue == nil:
				v.Continue = &views.Link{URL: views.LessonURL(k.Course, k.Chapter, k.Lesson), Title: l.Title}
			}
		}
	}
	return render(c, views.Course(v))
}

// lesson renders a lesson. POSTs grade the quiz or run the exercise,
// depending on the form's action field.
func (h handler) lesson(c fiber.Ctx) error {
	ctx, sess := c.Context(), sessionID(c)
	key := store.LessonKey{Course: c.Params("course"), Chapter: c.Params("chapter"), Lesson: c.Params("lesson")}
	outline, err := h.store.Outline(ctx, key.Course)
	if err != nil {
		return err
	}
	lesson, err := h.store.Lesson(ctx, key.Course, key.Chapter, key.Lesson)
	if err != nil {
		return err
	}
	progress, err := h.store.Progress(ctx, sess, key)
	if err != nil {
		return err
	}

	v := views.Lesson{Course: outline, Lesson: lesson, URL: views.LessonURL(key.Course, key.Chapter, key.Lesson)}
	var flat []views.Link
	for _, ch := range outline.Chapters {
		for _, l := range ch.Lessons {
			if ch.Slug == key.Chapter && l.Slug == key.Lesson {
				v.ChapterTitle = ch.Title
				v.Number = len(flat) + 1
			}
			flat = append(flat, views.Link{URL: views.LessonURL(key.Course, ch.Slug, l.Slug), Title: l.Title})
		}
	}
	v.Total = len(flat)
	if i := v.Number - 1; i > 0 {
		v.Prev = &flat[i-1]
	}
	if i := v.Number; i < len(flat) {
		v.Next = &flat[i]
	}
	if e := lesson.Exercise; e != nil {
		v.Code = cmp.Or(progress.Code, e.Starter)
	}
	complete := func() bool {
		return (len(lesson.Quiz) == 0 || progress.QuizPassed) &&
			(lesson.Exercise == nil || progress.ExercisePassed)
	}
	v.Complete = complete()

	htmx := c.Get("HX-Request") == "true"
	switch action := c.FormValue("action"); {
	case c.Method() != fiber.MethodPost:
	case action == "quiz" && len(lesson.Quiz) > 0:
		v.Quiz = grade(c, lesson.Quiz)
		if v.Quiz.Passed {
			if err := h.store.PassQuiz(ctx, sess, key, browserNow(c)); err != nil {
				return err
			}
			progress.QuizPassed = true
		}
	case (action == "run" || action == "submit") && lesson.Exercise != nil:
		v.Code = normalizeCode(c.FormValue("code"))
		out := h.runExercise(c, lesson.Exercise, v.Code, action == "submit")
		if err := h.store.SaveCode(ctx, sess, key, v.Code, out.Passed, browserNow(c)); err != nil {
			return err
		}
		progress.ExercisePassed = progress.ExercisePassed || out.Passed
		out.Complete = !v.Complete && complete()
		out.Next = v.Next
		if htmx {
			return render(c, views.Output(out))
		}
		v.Output = &out
	case action == "reset" && lesson.Exercise != nil:
		v.Code = lesson.Exercise.Starter
		if err := h.store.SaveCode(ctx, sess, key, "", false, browserNow(c)); err != nil {
			return err
		}
		if htmx {
			return render(c, views.Exercise(v))
		}
	default:
		return fiber.ErrBadRequest
	}
	v.Complete = complete()
	return render(c, views.LessonPage(v))
}

func (h handler) runExercise(c fiber.Ctx, e *course.Exercise, code string, submit bool) views.RunResult {
	r := views.RunResult{Submit: submit}
	switch {
	case h.runner == nil:
		r.Output = "Running code isn't available: the server couldn't find the go toolchain."
	case len(code) > maxCodeBytes:
		r.Output = "That's too much code for one exercise."
	case submit:
		res, passed := h.runner.Submit(c.Context(), code, e.Tests, e.ExpectedOutput)
		r.Output, r.Passed = res.Output, passed
		if e.Tests == "" && res.OK && !passed {
			r.Expected = e.ExpectedOutput
		}
	default:
		r.Output = h.runner.Run(c.Context(), code).Output
	}
	return r
}

// normalizeCode undoes the CRLF line endings browsers submit textareas with.
func normalizeCode(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }

func grade(c fiber.Ctx, quiz []course.Question) *views.QuizResult {
	r := &views.QuizResult{Passed: true}
	for i, q := range quiz {
		chosen, err := strconv.Atoi(c.FormValue("q" + strconv.Itoa(i)))
		if err != nil || chosen < 0 || chosen >= len(q.Options) {
			chosen = -1
		}
		ok := chosen >= 0 && q.Options[chosen].Correct
		r.Chosen = append(r.Chosen, chosen)
		r.Correct = append(r.Correct, ok)
		r.Passed = r.Passed && ok
	}
	return r
}

func render(c fiber.Ctx, t templ.Component) error {
	c.Type("html", "utf-8")
	return t.Render(c.Context(), c.Response().BodyWriter())
}

func errorHandler(c fiber.Ctx, err error) error {
	code, msg := fiber.StatusInternalServerError, fiber.ErrInternalServerError.Message
	if e, ok := errors.AsType[*fiber.Error](err); ok {
		code, msg = e.Code, e.Message
	}
	if errors.Is(err, store.ErrNotFound) {
		code = fiber.StatusNotFound
	}
	c.Status(code)
	switch code {
	case fiber.StatusNotFound:
		return render(c, views.NotFound())
	case fiber.StatusInternalServerError:
		slog.Error("request failed", "path", c.Path(), "err", err)
	}
	return c.SendString(msg)
}

// The browser supplies an IANA timezone, never an activity date or timestamp.
func browserNow(c fiber.Ctx) time.Time {
	name, err := url.QueryUnescape(c.Cookies("goland_timezone"))
	if err == nil && name != "" && name != "Local" && len(name) <= 100 {
		if loc, err := time.LoadLocation(name); err == nil {
			return time.Now().In(loc)
		}
	}
	return time.Now().UTC()
}
