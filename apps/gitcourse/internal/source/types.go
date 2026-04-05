package source

import (
	"context"

	"gitcourse/internal/course"
)

type Source interface {
	Courses(ctx context.Context) ([]CourseView, error)
	Course(ctx context.Context, id string) (CourseDetailView, error)
	AddCourse(ctx context.Context, repoURL string) error
	RemoveCourse(ctx context.Context, id string) error
	Enroll(ctx context.Context, courseID, studentRepo string) error
	Progress(ctx context.Context, courseID string) (course.Progress, error)
	Ask(ctx context.Context, prompt, model, courseContext string) (CommandResult, error)
	AskStream(ctx context.Context, prompt, model, courseContext string) (AskStream, error)
	RagAvailable() bool
}

type CourseView struct {
	ID             string
	RepoURL        string
	Title          string
	Description    string
	Language       string
	Progress       course.Progress
	ProgressKnown  bool
	StudentRepoURL string
}

type CourseDetailView struct {
	CourseView
	Course course.Course
}

type AskStream interface {
	Next() bool
	Chunk() StreamChunk
	Result() (CommandResult, error)
	Close() error
}

type CommandResult struct {
	Status          string         `json:"status"`
	Message         string         `json:"message"`
	Data            map[string]any `json:"data,omitempty"`
	RequiresConfirm bool           `json:"requires_confirm,omitempty"`
}

type StreamChunk struct {
	Delta    string         `json:"delta"`
	Metadata map[string]any `json:"metadata,omitempty"`
}
