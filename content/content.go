// Package content embeds the course material shipped with the binary.
//
// Layout:
//
//	courses/NN-course-slug/course.yaml
//	courses/NN-course-slug/NN-chapter-slug/chapter.yaml
//	courses/NN-course-slug/NN-chapter-slug/NN-lesson-slug.md
package content

import "embed"

//go:embed all:courses
var FS embed.FS
