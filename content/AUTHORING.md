# Authoring courses for goland-engine

goland-engine teaches programming through Go, modelled on boot.dev's course
structure: short text lessons, each paired with a multiple-choice quiz and/or a coding
exercise shown in a side panel. Learners type Go into the exercise editor, press
**Run** to execute it (never graded) and **Submit** to have it graded.

## Target Go version

**Go 1.27** (released 2026-08-19). Write modern, idiomatic Go:

- `any` not `interface{}`; `for i := range 10` (range over int, 1.22+);
  per-iteration loop variables (1.22+), so don't write `v := v` copies.
- `min`, `max` and `clear` built-ins (1.21+).
- The `slices`, `maps`, `cmp` and `iter` packages (`slices.Sort`, `slices.Contains`,
  `slices.SortFunc(s, func(a, b T) int { return cmp.Compare(a.X, b.X) })`,
  `maps.Keys` returning an `iter.Seq`, `slices.Collect(maps.Keys(m))`).
- Range-over-func iterators: `iter.Seq[V]`, `iter.Seq2[K, V]`, `for v := range seq` (1.23+).
- Generic type aliases (1.24), `strings.Lines`, `strings.SplitSeq`, `b.Loop()` in benchmarks,
  `t.Context()` in tests (1.24), `testing/synctest` (1.25), `sync.WaitGroup.Go` (1.25).
- `new(expr)`: `p := new(42)` gives an `*int` pointing at 42 (1.26).
- `errors.AsType[T](err) (T, bool)`, the generic replacement for `errors.As` (1.26).
- Self-referential generic constraints such as `type Adder[A Adder[A]] interface{ Add(A) A }` (1.26).
- **Generic methods** (1.27): methods may declare their own type parameters,
  e.g. `func (r *Rand) N[Int intType](n Int) Int`. Interface methods still can't have type
  parameters, and a generic method can't satisfy an interface method.
- Struct literals may set promoted fields of embedded structs directly (1.27):
  `Hero{HP: 100}` when `Hero` embeds `Stats`. It can't be mixed with the `Stats:` key and
  doesn't work through an embedded pointer.
- `strings.CutLast` / `bytes.CutLast` (1.27), the `uuid` package (1.27),
  `encoding/json/v2` (1.27).
- `go fix` runs "modernizers" that upgrade code to current idioms (1.26+).
- `go mod init` writes the previous Go minor version into `go.mod` (1.26+).

### Reference material

Use these free resources for topic coverage, ordering and example ideas:

- A Tour of Go: https://go.dev/tour/
- Go by Example: https://gobyexample.com/
- Learn Go with Tests: https://quii.gitbook.io/learn-go-with-tests
- Effective Go and the Go blog (https://go.dev/doc/effective_go, https://go.dev/blog/)

Write the lessons in your own words and with your own examples in the course's theme.
Don't paste their prose or code wholesale. Where a lesson closely follows one of them,
end it with a short "Further reading" link to that page.

Don't invent APIs. If you aren't sure something exists in Go 1.27, check it with
`go doc <pkg>.<Name>` (Go 1.27.1 is installed) or leave it out.

**Verify your code.** Every snippet that is a complete program (`package main` + `func main`)
must compile and print what the lesson says it prints. Test them in your own scratch
directory (`go run file.go`) and never inside this repository.

## File layout

```
content/courses/NN-course-slug/course.yaml
content/courses/NN-course-slug/NN-chapter-slug/chapter.yaml
content/courses/NN-course-slug/NN-chapter-slug/NN-lesson-slug.md
```

- The `NN-` two-digit prefix sets the order. The slug after it (lowercase, digits, hyphens)
  becomes part of the URL.
- `course.yaml` and `chapter.yaml` hold `title:` and `description:` (one or two sentences).

## Lesson file

```markdown
---
title: Zero Values
quiz:
  - question: What is the zero value of a `string`?
    options:
      - text: '`nil`'
      - text: '`""` (the empty string)'
        correct: true
      - text: '`" "`'
    explanation: |
      Every type has a zero value. For strings it's the empty string,
      never `nil`.
---

Markdown body. Don't repeat the title as an `# H1`, because the page already shows it.
Use `##` and `###` for sections and fenced ```go blocks for code.
```

Quiz rules (enforced by the validator):

- Each lesson has 1 to 3 questions. Each question has 2 to 5 options, with **exactly one**
  `correct: true`, and a non-empty `explanation` that teaches *why*.
- Questions, options and explanations are Markdown, so backticks work. **Quote any YAML
  scalar** that starts with a backtick, `*`, `[`, `{`, `&`, `!`, `%`, `@`, or contains `: `.
  Single quotes are safest (write `''` for a literal `'`). Use `|` blocks for multi-line text.
- Mix up where the correct option sits. Don't always put it first.
- The best questions make the learner *read code and predict output* ("What does this print?")
  or spot a bug. Put a fenced code block inside a `|` question when it helps.

## Coding exercises

A lesson can have an `exercise:` instead of (or as well as) a `quiz:`. boot.dev is
coding-heavy, so aim for **at least two exercises per chapter** (in course 01, about
half of all lessons), and end every course with an **"assemble it" exercise** that
combines pieces built earlier into the course's running project. Use exercises wherever
typing code teaches more than picking an answer: implementing a function, fixing a bug,
or finishing a data structure. Cheap `expected_output` exercises are fine early on.

Each course's last lesson should wrap up the course and link to the next one on the
roadmap. A course's URL slug is its directory name without the `NN-` prefix, e.g.
`/courses/learn-oop`. When a topic is taught in depth elsewhere on the roadmap, recap it
briefly and link there instead of re-teaching it.

Every exercise is a single `main.go` in `package main`. It's graded one of two ways:

- **`expected_output`**: Submit runs the program and compares stdout+stderr to this text
  (trailing whitespace ignored). Best for the early "print this" lessons.
- **`tests`**: a hidden `main_test.go` (also `package main`). Submit runs `go test -v`.
  Best once functions exist. Write clear failure messages
  (`t.Errorf("getMessageCost(%q) = %v, want %v", ...)`), since they're all the learner sees.

```markdown
---
title: Message Costs
exercise:
  starter: |
    package main

    import "fmt"

    // getMessageCost returns the cost in cents of sending msg:
    // 2 cents per character.
    func getMessageCost(msg string) int {
    	// ?
    	return 0
    }

    func main() {
    	fmt.Println(getMessageCost("hi")) // should print 4
    }
  solution: |
    package main

    import "fmt"

    func getMessageCost(msg string) int {
    	return len(msg) * 2
    }

    func main() {
    	fmt.Println(getMessageCost("hi"))
    }
  tests: |
    package main

    import "testing"

    func TestGetMessageCost(t *testing.T) {
    	for _, tt := range []struct {
    		msg  string
    		want int
    	}{{"hi", 4}, {"", 0}, {"hello there", 22}} {
    		if got := getMessageCost(tt.msg); got != tt.want {
    			t.Errorf("getMessageCost(%q) = %d, want %d", tt.msg, got, tt.want)
    		}
    	}
    }
---
```

Rules (checked by `go run ./cmd/validate -exec <course>`):

- `starter` and `solution` are required, plus **exactly one** of `tests` or `expected_output`.
- The solution must pass. The starter must **not** pass, but it should compile, and its
  `main` should print something useful so **Run** means something.
- Indent code inside the YAML `|` blocks with **tabs**, as gofmt does. YAML allows tabs
  inside block scalars as long as each line starts with the block's space indentation.
- Standard library only (no network, no module downloads), and it must finish in well
  under 5 seconds.
- Say in the lesson body what the learner must do ("Complete `getMessageCost` so that…").

## Style (boot.dev tone)

- Short, direct, a bit playful. Speak to the learner as "you". Assume no prior programming
  knowledge in course 01, and only the previous courses' knowledge in later ones.
- Each lesson covers one idea, in roughly 250 to 700 words, with at least one code example.
- Ground examples in a running theme per course, as boot.dev does (a game, a messaging
  app, a bank, and so on). Keep it consistent within the course.
- Point out Go-specific gotchas (nil maps, slice aliasing, loop semantics, shadowing).

## Validate

```
go run ./cmd/validate content/courses/NN-course-slug
go run ./cmd/validate -exec content/courses/NN-course-slug
```

The validator must exit 0 before you're done. If the course has exercises, also run
it with `-exec`.
