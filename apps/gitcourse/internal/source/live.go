package source

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	hubrelay "github.com/fastygo/hubrelay-sdk"
	"gitcourse/internal/config"
	"gitcourse/internal/course"
	gitreader "gitcourse/internal/git"
	"gitcourse/internal/progress"
	"gitcourse/internal/relay"
	"gitcourse/internal/store"
)

type Live struct {
	store      store.CourseStore
	reader     gitreader.Reader
	progress   *progress.Cache
	relay      *relay.Client
	qdrantURL  string
}

func NewLive(cfg config.Config, courseStore store.CourseStore, reader gitreader.Reader, progressCache *progress.Cache, client *relay.Client) *Live {
	return &Live{
		store:     courseStore,
		reader:    reader,
		progress:  progressCache,
		relay:     client,
		qdrantURL: strings.TrimSpace(cfg.QdrantURL),
	}
}

func (s *Live) Courses(ctx context.Context) ([]CourseView, error) {
	registered, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]CourseView, 0, len(registered))
	for _, item := range registered {
		view := CourseView{
			ID:          item.ID,
			RepoURL:     item.RepoURL,
			Title:       item.Title,
			Description: item.Description,
			Language:    item.Language,
		}

		if enrollment, err := s.store.Enrollment(ctx, item.ID); err == nil {
			view.StudentRepoURL = enrollment.StudentRepo
			if progressValue, err := s.progress.Get(ctx, item.ID, enrollment.StudentRepo); err == nil {
				view.Progress = progressValue
				view.ProgressKnown = true
			}
		}

		out = append(out, view)
	}
	return out, nil
}

func (s *Live) Course(ctx context.Context, id string) (CourseDetailView, error) {
	registered, err := s.store.Get(ctx, id)
	if err != nil {
		return CourseDetailView{}, err
	}

	courseData, err := s.readCourse(ctx, registered.RepoURL)
	if err != nil {
		return CourseDetailView{}, err
	}

	view := CourseDetailView{
		CourseView: CourseView{
			ID:          registered.ID,
			RepoURL:     registered.RepoURL,
			Title:       registered.Title,
			Description: registered.Description,
			Language:    registered.Language,
		},
		Course: courseData,
	}
	if enrollment, err := s.store.Enrollment(ctx, id); err == nil {
		view.StudentRepoURL = enrollment.StudentRepo
		if progressValue, err := s.progress.Get(ctx, id, enrollment.StudentRepo); err == nil {
			view.Progress = progressValue
			view.ProgressKnown = true
		}
	}
	return view, nil
}

func (s *Live) AddCourse(ctx context.Context, repoURL string) error {
	courseBody, err := s.reader.ReadFile(ctx, repoURL, ".course/course.json")
	if err != nil {
		return err
	}

	var parsed course.Course
	if err := json.Unmarshal(courseBody, &parsed); err != nil {
		return fmt.Errorf("decode course.json: %w", err)
	}
	if strings.TrimSpace(parsed.ID) == "" || strings.TrimSpace(parsed.Title) == "" {
		return fmt.Errorf("course.json must define id and title")
	}

	verifyBody, err := s.reader.ReadFile(ctx, repoURL, ".course/ci/verify.sh")
	if err != nil {
		return err
	}
	workflowBody, err := s.reader.ReadFile(ctx, repoURL, ".github/workflows/course-check.yml")
	if err != nil {
		return err
	}

	return s.store.Add(ctx, store.RegisteredCourse{
		ID:          parsed.ID,
		RepoURL:     strings.TrimSpace(repoURL),
		Title:       parsed.Title,
		Language:    parsed.Language,
		Description: parsed.Description,
		AddedAt:     time.Now().UTC(),
		FileHashes: map[string]string{
			".course/course.json":                sha256Hex(courseBody),
			".course/ci/verify.sh":               sha256Hex(verifyBody),
			".github/workflows/course-check.yml": sha256Hex(workflowBody),
		},
		QdrantIndexed: false,
	})
}

func (s *Live) RemoveCourse(ctx context.Context, id string) error {
	return s.store.Remove(ctx, id)
}

func (s *Live) Enroll(ctx context.Context, courseID, studentRepo string) error {
	return s.store.Enroll(ctx, courseID, studentRepo)
}

func (s *Live) Progress(ctx context.Context, courseID string) (course.Progress, error) {
	enrollment, err := s.store.Enrollment(ctx, courseID)
	if err != nil {
		return course.Progress{}, err
	}
	return s.progress.Get(ctx, courseID, enrollment.StudentRepo)
}

func (s *Live) Ask(ctx context.Context, prompt, model, courseContext string) (CommandResult, error) {
	if s.relay == nil {
		return CommandResult{}, fmt.Errorf("relay client is not configured")
	}
	result, err := s.relay.Ask(ctx, mergePrompt(prompt, courseContext), model)
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{
		Status:          result.Status,
		Message:         result.Message,
		Data:            cloneMap(result.Data),
		RequiresConfirm: result.RequiresConfirm,
	}, nil
}

func (s *Live) AskStream(ctx context.Context, prompt, model, courseContext string) (AskStream, error) {
	if s.relay == nil {
		return nil, fmt.Errorf("relay client is not configured")
	}
	stream, err := s.relay.AskStream(ctx, mergePrompt(prompt, courseContext), model)
	if err != nil {
		return nil, err
	}
	return &liveAskStream{stream: stream}, nil
}

func (s *Live) RagAvailable() bool {
	return s.qdrantURL != ""
}

func (s *Live) readCourse(ctx context.Context, repoURL string) (course.Course, error) {
	body, err := s.reader.ReadFile(ctx, repoURL, ".course/course.json")
	if err != nil {
		return course.Course{}, err
	}
	var parsed course.Course
	if err := json.Unmarshal(body, &parsed); err != nil {
		return course.Course{}, err
	}
	return parsed, nil
}

type liveAskStream struct {
	stream hubrelay.ResultStream
}

func (s *liveAskStream) Next() bool {
	return s.stream.Next()
}

func (s *liveAskStream) Chunk() StreamChunk {
	chunk := s.stream.Chunk()
	return StreamChunk{Delta: chunk.Delta, Metadata: cloneMap(chunk.Metadata)}
}

func (s *liveAskStream) Result() (CommandResult, error) {
	result, err := s.stream.Result()
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{
		Status:          result.Status,
		Message:         result.Message,
		Data:            cloneMap(result.Data),
		RequiresConfirm: result.RequiresConfirm,
	}, nil
}

func (s *liveAskStream) Close() error {
	return s.stream.Close()
}

func sha256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func mergePrompt(prompt, courseContext string) string {
	courseContext = strings.TrimSpace(courseContext)
	if courseContext == "" {
		return prompt
	}
	return prompt + "\n\nCourse context:\n" + courseContext
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
