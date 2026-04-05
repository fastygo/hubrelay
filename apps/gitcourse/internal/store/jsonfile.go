package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type JSONFileStore struct {
	path string
	mu   sync.Mutex
}

func NewJSONFileStore(dir string) (*JSONFileStore, error) {
	if dir == "" {
		return nil, errors.New("data dir must not be empty")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	store := &JSONFileStore{path: filepath.Join(dir, "courses.json")}
	if _, err := os.Stat(store.path); errors.Is(err, os.ErrNotExist) {
		if err := store.writeRegistry(Registry{}); err != nil {
			return nil, err
		}
	}

	return store, nil
}

func (s *JSONFileStore) List(_ context.Context) ([]RegisteredCourse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	registry, err := s.readRegistry()
	if err != nil {
		return nil, err
	}
	return append([]RegisteredCourse(nil), registry.Courses...), nil
}

func (s *JSONFileStore) Get(_ context.Context, id string) (RegisteredCourse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	registry, err := s.readRegistry()
	if err != nil {
		return RegisteredCourse{}, err
	}
	for _, course := range registry.Courses {
		if course.ID == id {
			return course, nil
		}
	}
	return RegisteredCourse{}, fmt.Errorf("course %q not found", id)
}

func (s *JSONFileStore) Add(_ context.Context, course RegisteredCourse) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	registry, err := s.readRegistry()
	if err != nil {
		return err
	}
	for idx, existing := range registry.Courses {
		if existing.ID == course.ID {
			registry.Courses[idx] = course
			return s.writeRegistry(registry)
		}
	}
	registry.Courses = append(registry.Courses, course)
	return s.writeRegistry(registry)
}

func (s *JSONFileStore) Remove(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	registry, err := s.readRegistry()
	if err != nil {
		return err
	}

	filtered := registry.Courses[:0]
	for _, course := range registry.Courses {
		if course.ID != id {
			filtered = append(filtered, course)
		}
	}
	registry.Courses = filtered

	enrollments := registry.Enrollments[:0]
	for _, enrollment := range registry.Enrollments {
		if enrollment.CourseID != id {
			enrollments = append(enrollments, enrollment)
		}
	}
	registry.Enrollments = enrollments
	return s.writeRegistry(registry)
}

func (s *JSONFileStore) Enroll(_ context.Context, courseID, studentRepo string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	registry, err := s.readRegistry()
	if err != nil {
		return err
	}
	for idx, enrollment := range registry.Enrollments {
		if enrollment.CourseID == courseID {
			registry.Enrollments[idx].StudentRepo = studentRepo
			return s.writeRegistry(registry)
		}
	}
	registry.Enrollments = append(registry.Enrollments, Enrollment{
		CourseID:    courseID,
		StudentRepo: studentRepo,
	})
	return s.writeRegistry(registry)
}

func (s *JSONFileStore) Enrollment(_ context.Context, courseID string) (Enrollment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	registry, err := s.readRegistry()
	if err != nil {
		return Enrollment{}, err
	}
	for _, enrollment := range registry.Enrollments {
		if enrollment.CourseID == courseID {
			return enrollment, nil
		}
	}
	return Enrollment{}, fmt.Errorf("enrollment for course %q not found", courseID)
}

func (s *JSONFileStore) readRegistry() (Registry, error) {
	body, err := os.ReadFile(s.path)
	if err != nil {
		return Registry{}, err
	}
	var registry Registry
	if len(body) == 0 {
		return Registry{}, nil
	}
	if err := json.Unmarshal(body, &registry); err != nil {
		return Registry{}, err
	}
	if registry.Courses == nil {
		registry.Courses = []RegisteredCourse{}
	}
	if registry.Enrollments == nil {
		registry.Enrollments = []Enrollment{}
	}
	return registry, nil
}

func (s *JSONFileStore) writeRegistry(registry Registry) error {
	if registry.Courses == nil {
		registry.Courses = []RegisteredCourse{}
	}
	if registry.Enrollments == nil {
		registry.Enrollments = []Enrollment{}
	}

	body, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}

	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, append(body, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tempPath, s.path)
}
