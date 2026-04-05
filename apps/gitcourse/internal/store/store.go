package store

import (
	"context"
	"time"
)

type RegisteredCourse struct {
	ID            string            `json:"id"`
	RepoURL       string            `json:"repo_url"`
	Title         string            `json:"title"`
	Language      string            `json:"language"`
	Description   string            `json:"description,omitempty"`
	AddedAt       time.Time         `json:"added_at"`
	FileHashes    map[string]string `json:"file_hashes"`
	QdrantIndexed bool              `json:"qdrant_indexed"`
}

type Enrollment struct {
	CourseID    string `json:"course_id"`
	StudentRepo string `json:"student_repo"`
}

type Registry struct {
	Courses     []RegisteredCourse `json:"courses"`
	Enrollments []Enrollment       `json:"enrollments"`
}

type CourseStore interface {
	List(ctx context.Context) ([]RegisteredCourse, error)
	Get(ctx context.Context, id string) (RegisteredCourse, error)
	Add(ctx context.Context, course RegisteredCourse) error
	Remove(ctx context.Context, id string) error
	Enroll(ctx context.Context, courseID, studentRepo string) error
	Enrollment(ctx context.Context, courseID string) (Enrollment, error)
}
