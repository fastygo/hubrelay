package source

import (
	"context"
	"fmt"
	"path"
	"strings"

	"gitcourse/fixtures"
	"gitcourse/internal/course"
)

type Fixture struct {
	courses []CourseView
	ask     askFixtureData
}

type fixtureCoursesData struct {
	Courses []struct {
		ID          string `json:"id"`
		RepoURL     string `json:"repo_url"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Language    string `json:"language"`
	} `json:"courses"`
}

type askFixtureData struct {
	Result CommandResult `json:"result"`
	Stream struct {
		Chunks []StreamChunk `json:"chunks"`
		Result CommandResult `json:"result"`
	} `json:"stream"`
}

func NewFixture(locale string) (*Fixture, error) {
	locale = strings.TrimSpace(locale)
	if locale == "" {
		return nil, fmt.Errorf("locale must not be empty")
	}

	var coursesData fixtureCoursesData
	if err := fixtures.Decode(path.Join(locale, "mocks", "courses.json"), &coursesData); err != nil {
		return nil, err
	}

	var ask askFixtureData
	if err := fixtures.Decode(path.Join(locale, "mocks", "ask.json"), &ask); err != nil {
		return nil, err
	}

	out := &Fixture{ask: ask}
	for _, item := range coursesData.Courses {
		out.courses = append(out.courses, CourseView{
			ID:          item.ID,
			RepoURL:     item.RepoURL,
			Title:       item.Title,
			Description: item.Description,
			Language:    item.Language,
			ProgressKnown: true,
			Progress: course.Progress{
				CourseID: item.ID,
				Lessons: []course.LessonStatus{
					{ID: "001", Status: "done"},
					{ID: "002", Status: "in_progress"},
					{ID: "003", Status: "pending"},
				},
			},
		})
	}
	return out, nil
}

func (s *Fixture) Courses(_ context.Context) ([]CourseView, error) {
	return append([]CourseView(nil), s.courses...), nil
}

func (s *Fixture) Course(_ context.Context, id string) (CourseDetailView, error) {
	for _, item := range s.courses {
		if item.ID == id {
			return CourseDetailView{
				CourseView: item,
				Course: course.Course{
					ID:          item.ID,
					Version:     "1.0.0",
					Title:       item.Title,
					Description: item.Description,
					Language:    item.Language,
					Sections: []course.Section{
						{
							ID:    "basics",
							Title: "Basics",
							Lessons: []course.Lesson{
								{
									ID:        "001",
									Title:     "Run the project",
									Objective: "Install dependencies and boot the dev server.",
									AskContext: "The student starts the Vite project and validates the toolchain.",
									Hints: []string{
										"Run npm install before starting Vite.",
										"Make sure the build command passes in CI too.",
									},
									Checklist: []course.CheckItem{
										{ID: "node_modules", Label: "Dependencies installed", Verify: "dir_exists:node_modules"},
										{ID: "build_ok", Label: "Build command passes", Verify: "build_succeeds:npm run build"},
									},
								},
								{
									ID:        "002",
									Title:     "Create the header",
									Objective: "Add navigation to the app shell.",
									AskContext: "The student is creating the first reusable React component.",
									Hints: []string{
										"Keep the component presentational.",
										"Use semantic nav markup.",
									},
									Checklist: []course.CheckItem{
										{ID: "header_file", Label: "Header component exists", Verify: "file_exists:src/components/Header.tsx"},
									},
								},
							},
						},
					},
				},
			}, nil
		}
	}
	return CourseDetailView{}, fmt.Errorf("course %q not found", id)
}

func (s *Fixture) AddCourse(_ context.Context, _ string) error { return nil }
func (s *Fixture) RemoveCourse(_ context.Context, _ string) error { return nil }
func (s *Fixture) Enroll(_ context.Context, _, _ string) error { return nil }

func (s *Fixture) Progress(ctx context.Context, courseID string) (course.Progress, error) {
	item, err := s.Course(ctx, courseID)
	if err != nil {
		return course.Progress{}, err
	}
	return item.Progress, nil
}

func (s *Fixture) Ask(_ context.Context, _, _, _ string) (CommandResult, error) {
	return s.ask.Result, nil
}

func (s *Fixture) AskStream(_ context.Context, _, _, _ string) (AskStream, error) {
	return &fixtureAskStream{
		chunks: append([]StreamChunk(nil), s.ask.Stream.Chunks...),
		result: s.ask.Stream.Result,
	}, nil
}

func (s *Fixture) RagAvailable() bool {
	return false
}

type fixtureAskStream struct {
	chunks  []StreamChunk
	result  CommandResult
	index   int
	current StreamChunk
}

func (s *fixtureAskStream) Next() bool {
	if s.index >= len(s.chunks) {
		return false
	}
	s.current = s.chunks[s.index]
	s.index++
	return true
}

func (s *fixtureAskStream) Chunk() StreamChunk { return s.current }
func (s *fixtureAskStream) Result() (CommandResult, error) { return s.result, nil }
func (s *fixtureAskStream) Close() error { return nil }
