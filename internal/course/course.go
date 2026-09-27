// Package course loads course material from a file system, validates it
// and renders its Markdown to HTML.
package course

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type Course struct {
	Slug        string
	Title       string
	Description string
	Chapters    []Chapter
}

type Chapter struct {
	Slug        string
	Title       string
	Description string
	Lessons     []Lesson
}

type Lesson struct {
	Slug     string
	Title    string
	HTML     string // rendered lesson body
	Quiz     []Question
	Exercise *Exercise
}

// Exercise is code the learner writes in the browser. Run executes it
// without grading; Submit grades it, either by running Tests with go test
// or, when Tests is empty, by comparing the program's output to
// ExpectedOutput.
type Exercise struct {
	Starter        string // main.go shown in the editor
	Solution       string // reference main.go, used by the validator
	Tests          string // hidden main_test.go
	ExpectedOutput string
}

type Question struct {
	Prompt      string // rendered HTML
	Explanation string // rendered HTML
	Options     []Option
}

type Option struct {
	Text    string // rendered inline HTML
	Correct bool
}

type courseMeta struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

type lessonMeta struct {
	Title string `yaml:"title"`
	Quiz  []struct {
		Question    string `yaml:"question"`
		Explanation string `yaml:"explanation"`
		Options     []struct {
			Text    string `yaml:"text"`
			Correct bool   `yaml:"correct"`
		} `yaml:"options"`
	} `yaml:"quiz"`
	Exercise *struct {
		Starter        string `yaml:"starter"`
		Solution       string `yaml:"solution"`
		Tests          string `yaml:"tests"`
		ExpectedOutput string `yaml:"expected_output"`
	} `yaml:"exercise"`
}

var (
	orderedName = regexp.MustCompile(`^(\d{2})-([a-z0-9]+(?:-[a-z0-9]+)*)$`)
	md          = goldmark.New(goldmark.WithExtensions(extension.GFM))
)

// Load reads every course under root in fsys. It returns all problems
// found joined together, so a single run reports every broken lesson,
// along with everything that did load: broken lessons and empty chapters
// are left out, and a course with no valid chapters is dropped.
func Load(fsys fs.FS, root string) ([]Course, error) {
	var (
		courses []Course
		errs    []error
	)
	for name := range orderedDirs(fsys, root, &errs) {
		c, err := LoadCourse(fsys, path.Join(root, name))
		if err != nil {
			errs = append(errs, err)
		}
		if len(c.Chapters) > 0 {
			courses = append(courses, c)
		}
	}
	return courses, errors.Join(errs...)
}

// LoadCourse reads and validates the single course rooted at dir.
func LoadCourse(fsys fs.FS, dir string) (Course, error) {
	if !orderedName.MatchString(path.Base(dir)) {
		return Course{}, fmt.Errorf("%s: directory must be named NN-slug", dir)
	}
	var meta courseMeta
	if err := readYAML(fsys, path.Join(dir, "course.yaml"), &meta); err != nil {
		return Course{}, err
	}
	c := Course{Slug: slug(path.Base(dir)), Title: meta.Title, Description: meta.Description}
	var errs []error
	for name := range orderedDirs(fsys, dir, &errs) {
		ch, err := loadChapter(fsys, path.Join(dir, name))
		if err != nil {
			errs = append(errs, err)
		}
		if len(ch.Lessons) > 0 {
			c.Chapters = append(c.Chapters, ch)
		}
	}
	if len(c.Chapters) == 0 {
		errs = append(errs, fmt.Errorf("%s: course has no chapters", dir))
	}
	return c, errors.Join(errs...)
}

func loadChapter(fsys fs.FS, dir string) (Chapter, error) {
	var meta courseMeta
	if err := readYAML(fsys, path.Join(dir, "chapter.yaml"), &meta); err != nil {
		return Chapter{}, err
	}
	ch := Chapter{Slug: slug(path.Base(dir)), Title: meta.Title, Description: meta.Description}
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return Chapter{}, err
	}
	var errs []error
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), ".md")
		if e.IsDir() || !ok {
			continue
		}
		if !orderedName.MatchString(name) {
			errs = append(errs, fmt.Errorf("%s/%s: lesson file must be named NN-slug.md", dir, e.Name()))
			continue
		}
		l, err := loadLesson(fsys, path.Join(dir, e.Name()))
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ch.Lessons = append(ch.Lessons, l)
	}
	if len(ch.Lessons) == 0 {
		errs = append(errs, fmt.Errorf("%s: chapter has no lessons", dir))
	}
	return ch, errors.Join(errs...)
}

func loadLesson(fsys fs.FS, file string) (Lesson, error) {
	raw, err := fs.ReadFile(fsys, file)
	if err != nil {
		return Lesson{}, err
	}
	front, body, err := splitFrontmatter(raw)
	if err != nil {
		return Lesson{}, fmt.Errorf("%s: %w", file, err)
	}
	var meta lessonMeta
	if err := yaml.Unmarshal(front, &meta); err != nil {
		return Lesson{}, fmt.Errorf("%s: frontmatter: %w", file, err)
	}

	var errs []error
	if strings.TrimSpace(meta.Title) == "" {
		errs = append(errs, errors.New("missing title"))
	}
	if len(bytes.TrimSpace(body)) == 0 {
		errs = append(errs, errors.New("empty lesson body"))
	}
	if len(meta.Quiz) == 0 && meta.Exercise == nil {
		errs = append(errs, errors.New("lesson needs a quiz or an exercise"))
	}
	if len(meta.Quiz) > 3 {
		errs = append(errs, errors.New("lesson must have at most 3 quiz questions"))
	}

	l := Lesson{Slug: slug(strings.TrimSuffix(path.Base(file), ".md")), Title: meta.Title, HTML: render(body)}
	if e := meta.Exercise; e != nil {
		hasTests, hasOutput := strings.TrimSpace(e.Tests) != "", strings.TrimSpace(e.ExpectedOutput) != ""
		switch {
		case strings.TrimSpace(e.Starter) == "" || strings.TrimSpace(e.Solution) == "":
			errs = append(errs, errors.New("exercise needs a starter and a solution"))
		case hasTests == hasOutput:
			errs = append(errs, errors.New("exercise needs exactly one of tests or expected_output"))
		}
		l.Exercise = &Exercise{Starter: e.Starter, Solution: e.Solution, Tests: e.Tests, ExpectedOutput: e.ExpectedOutput}
	}
	for i, q := range meta.Quiz {
		correct := 0
		opts := make([]Option, 0, len(q.Options))
		for j, o := range q.Options {
			if strings.TrimSpace(o.Text) == "" {
				errs = append(errs, fmt.Errorf("quiz[%d].options[%d]: missing text", i, j))
			}
			if o.Correct {
				correct++
			}
			opts = append(opts, Option{Text: renderInline(o.Text), Correct: o.Correct})
		}
		switch {
		case strings.TrimSpace(q.Question) == "":
			errs = append(errs, fmt.Errorf("quiz[%d]: missing question", i))
		case len(opts) < 2 || len(opts) > 5:
			errs = append(errs, fmt.Errorf("quiz[%d]: needs 2 to 5 options", i))
		case correct != 1:
			errs = append(errs, fmt.Errorf("quiz[%d]: needs exactly 1 correct option, has %d", i, correct))
		case strings.TrimSpace(q.Explanation) == "":
			errs = append(errs, fmt.Errorf("quiz[%d]: missing explanation", i))
		}
		l.Quiz = append(l.Quiz, Question{
			Prompt:      render([]byte(q.Question)),
			Explanation: render([]byte(q.Explanation)),
			Options:     opts,
		})
	}
	if err := errors.Join(errs...); err != nil {
		return Lesson{}, fmt.Errorf("%s: %w", file, err)
	}
	return l, nil
}

// splitFrontmatter separates a leading "---" delimited YAML block from the body.
func splitFrontmatter(raw []byte) (front, body []byte, err error) {
	raw = bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
	rest, ok := bytes.CutPrefix(raw, []byte("---\n"))
	if !ok {
		return nil, nil, errors.New("missing frontmatter: file must start with ---")
	}
	front, body, ok = bytes.Cut(rest, []byte("\n---\n"))
	if !ok {
		return nil, nil, errors.New("unterminated frontmatter")
	}
	return front, body, nil
}

func readYAML(fsys fs.FS, file string, v *courseMeta) error {
	raw, err := fs.ReadFile(fsys, file)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	if strings.TrimSpace(v.Title) == "" {
		return fmt.Errorf("%s: missing title", file)
	}
	return nil
}

// orderedDirs yields the NN-slug subdirectories of dir in order.
func orderedDirs(fsys fs.FS, dir string, errs *[]error) func(func(string) bool) {
	return func(yield func(string) bool) {
		entries, err := fs.ReadDir(fsys, dir)
		if err != nil {
			*errs = append(*errs, err)
			return
		}
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			if !orderedName.MatchString(e.Name()) {
				*errs = append(*errs, fmt.Errorf("%s/%s: directory must be named NN-slug", dir, e.Name()))
				continue
			}
			names = append(names, e.Name())
		}
		slices.Sort(names)
		for _, n := range names {
			if !yield(n) {
				return
			}
		}
	}
}

func slug(name string) string { return orderedName.FindStringSubmatch(name)[2] }

func render(src []byte) string {
	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		return ""
	}
	return buf.String()
}

func renderInline(s string) string {
	h := strings.TrimSpace(render([]byte(s)))
	h = strings.TrimPrefix(h, "<p>")
	return strings.TrimSuffix(h, "</p>")
}
