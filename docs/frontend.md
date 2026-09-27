# Frontend

The UI is server-rendered HTML from [templ](https://templ.guide) components, styled by one
hand-written stylesheet. Design inspiration is
[motherfuckingwebsite.com](https://motherfuckingwebsite.com/): readable text, a narrow
column, no chrome. Core learning and progress flows work with JavaScript disabled;
the enhanced editor and automatic timezone detection require JavaScript.

## Pages and components

All in `internal/web/views/views.templ`. After editing it, run `templ generate` to update
`views_templ.go` (which is committed).

| Component | Used for |
|---|---|
| `layout` | `<html>` shell: header, footer, stylesheet, `htmx.min.js`, `activity.js`. |
| `Home` | Roadmap page: intro, `Activity`, the ordered course list with progress bars, "Reset my progress". |
| `Activity` | Streak stats (current, best, today) and the 26-week heatmap with legend. |
| `Course` | Course outline: chapters, lessons with ✓ marks, "Start"/"Continue" link. |
| `LessonPage` | Two-column lesson: article on the left; exercise and quiz panels on the right (stacked on narrow screens). Prev/next pager. |
| `exercise` / `Exercise` | Code form with Run, Submit and Reset code. `Exercise` is the htmx fragment for Reset. |
| `Output` | Result of Run/Submit: verdict line, output (or Expected vs Your output), "Next lesson →" on a pass, and an out-of-band `#lesson-status` update. |
| `quiz` | Radio-button quiz; after grading shows right/wrong per question and explanations. |
| `NotFound` | 404 page. |

View models (`Lesson`, `Outline`, `CourseCard`, `QuizResult`, `RunResult`, `Link`) are
plain structs at the top of `views.templ`; handlers fill them in `internal/web/web.go`.

## Progressive enhancement

| Feature | Without JavaScript | With JavaScript |
|---|---|---|
| Quiz | Form POST, full page re-render | Same (quizzes don't use htmx) |
| Run / Submit | Form POST to `…#exercise`, full page re-render | htmx `hx-post`, swaps only `#output`; buttons disabled and "Running…" shown while waiting |
| Reset code | Form POST | htmx swaps the `#exercise` panel, after `hx-confirm` |
| Editor | Plain `<textarea>` with normal browser keyboard behavior | CodeMirror 6 with Go highlighting, optional vim mode; the htmx textarea fallback supports Tab indentation |
| Activity days | UTC calendar days | The browser's IANA timezone (see below) |

### htmx

htmx 2 is vendored as `static/htmx.min.js`. The server checks the `HX-Request: true`
header to decide between a fragment and a full page.

### Code editor (`static/editor.js`)

Built from `internal/web/editor/editor.js` with esbuild (see [development.md](development.md)).
It's loaded only on lessons that have an exercise. It replaces each
`textarea[data-editor]` with a CodeMirror view, keeping the hidden textarea in sync so
the form (and htmx) submits exactly as before. Features:

- Go syntax highlighting, line numbers, bracket matching, tab indentation (tab size 4).
- **Vim mode** toggle, remembered in `localStorage` (`goland-vim-mode`); `:w` submits.
- `Ctrl/⌘+Enter` runs, `Shift+Ctrl/⌘+Enter` submits. `Esc`, then `Tab`, leaves the editor
  (keyboard trap avoidance).
- Re-enhances the new textarea after htmx swaps in a reset panel, and destroys stale views.

Third-party licences for the bundle are written to `static/editor-licenses.txt`.

Typing alone does not save work to the server. Run and Submit save the submitted code;
Reset code saves an empty value and restores the starter while preserving prior passes.
There is no draft autosave or cross-tab editing coordination. Vim preference is local
to the browser; lesson progress is in SQLite behind the session cookie.

### Timezone (`static/activity.js`)

A tiny script loaded on every page. It reads
`Intl.DateTimeFormat().resolvedOptions().timeZone`, stores it in the `goland_timezone`
cookie for a year, and, if the value changed and the page shows the activity calendar,
reloads once so the calendar is drawn in the right timezone. The browser only ever
supplies a timezone *name*; dates and times always come from the server clock.

## Styling (`static/style.css`)

- System font stack, readable measure, generous line height.
- Light and dark themes via `prefers-color-scheme`.
- The lesson page switches to two columns at `min-width: 960px`.
- Heatmap cells use five shade levels (`level-0` … `level-4`); each cell also has a text
  label (`aria-label`/tooltip), so colour isn't the only signal.

## Accessibility notes

- Quiz and code forms are real forms with labels; results are announced with
  `aria-live="polite"` on the output region.
- The heatmap is a focusable, scrollable region with a list of labelled days.
- Visually hidden labels use the `sr-only` class.
