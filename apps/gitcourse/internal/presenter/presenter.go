package presenter

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	ui8layout "github.com/fastygo/ui8kit/layout"
	"gitcourse/internal/content"
	"gitcourse/internal/course"
	appmodule "gitcourse/internal/module"
	"gitcourse/internal/source"
	"gitcourse/views"
	"gitcourse/views/partials"
)

type Config struct {
	AuthEnabled      bool
	DefaultLocale    string
	AvailableLocales []string
	ToggleLocales    []string
	Catalogs         map[string]content.Catalog
	Locale           string
}

type Presenter struct {
	catalog  content.Catalog
	cfg      Config
	navItems []appmodule.NavItem
}

type LoginPageState struct {
	Username string
	Error    string
}

type AskPageState struct {
	Prompt  string
	Model   string
	Context string
	Error   string
	Result  *source.CommandResult
}

type CatalogPageState struct {
	RepoURL string
	Error   string
	Success string
	Courses []source.CourseView
}

type CoursePageState struct {
	View  source.CourseDetailView
	Error string
}

type LessonPageState struct {
	Course source.CourseDetailView
	LessonID string
	Error string
}

func New(catalog content.Catalog, cfg Config) *Presenter {
	return &Presenter{catalog: catalog, cfg: cfg}
}

func (p *Presenter) SetNavItems(items []appmodule.NavItem) {
	p.navItems = append([]appmodule.NavItem(nil), items...)
}

func (p *Presenter) HomePath() string { return "/" }
func (p *Presenter) LoginPath() string { return "/login" }

func (p *Presenter) LoginInvalidFormError() string {
	return p.catalog.Login.InvalidFormError
}

func (p *Presenter) LoginInvalidCredentialsError() string {
	return p.catalog.Login.InvalidCredsError
}

func (p *Presenter) AskPromptRequiredError() string {
	return p.catalog.Ask.PromptRequiredError
}

func (p *Presenter) LoginPage(state LoginPageState) views.LoginPageData {
	return views.LoginPageData{
		Layout: p.layout(p.catalog.Login.Title, "/login"),
		Header: views.PageHeader{
			Title:       p.catalog.Login.Title,
			Description: p.catalog.Login.Description,
		},
		Error:         state.Error,
		FormAction:    "/login",
		UsernameLabel: p.catalog.Login.UsernameLabel,
		UsernameValue: state.Username,
		PasswordLabel: p.catalog.Login.PasswordLabel,
		SubmitLabel:   p.catalog.Login.SubmitLabel,
	}
}

func (p *Presenter) AskPage(state AskPageState) views.AskPageData {
	var result *views.AskResultView
	if state.Result != nil {
		rows := []views.KV{
			{Label: "Status", Value: state.Result.Status},
			{Label: "Message", Value: state.Result.Message},
		}
		keys := make([]string, 0, len(state.Result.Data))
		for key := range state.Result.Data {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			rows = append(rows, views.KV{
				Label: key,
				Value: fmt.Sprint(state.Result.Data[key]),
			})
		}
		result = &views.AskResultView{Title: p.catalog.Ask.ResultTitle, Rows: rows}
	}

	return views.AskPageData{
		Layout: p.layout(p.catalog.Ask.Title, "/ask"),
		Header: views.PageHeader{
			Title:       p.catalog.Ask.Title,
			Description: p.catalog.Ask.Description,
		},
		Error:          state.Error,
		FormAction:     "/ask",
		StreamEndpoint: "/ask/stream",
		PromptField: views.AskFieldView{
			ID:          "prompt-input",
			Name:        "prompt",
			Label:       p.catalog.Ask.PromptLabel,
			Placeholder: p.catalog.Ask.PromptPlaceholder,
			Value:       state.Prompt,
		},
		ModelField: views.AskFieldView{
			ID:          "model-input",
			Name:        "model",
			Label:       p.catalog.Ask.ModelLabel,
			Placeholder: p.catalog.Ask.ModelPlaceholder,
			Value:       state.Model,
		},
		ContextField: views.AskFieldView{
			ID:          "course-context-input",
			Name:        "context",
			Label:       p.catalog.Ask.ContextLabel,
			Placeholder: p.catalog.Ask.ContextPlaceholder,
			Value:       state.Context,
		},
		Actions: views.AskActionsView{
			StreamSubmitLabel: p.catalog.Ask.StreamSubmitLabel,
			SyncSubmitLabel:   p.catalog.Ask.SyncSubmitLabel,
		},
		Stream: partials.StreamOutputData{
			Title:                p.catalog.Ask.StreamTitle,
			InitialStatus:        p.catalog.Ask.StatusIdle,
			StatusSync:           p.catalog.Ask.StatusSync,
			StatusPromptRequired: p.catalog.Ask.StatusPromptRequired,
			StatusConnecting:     p.catalog.Ask.StatusConnecting,
			StatusStreaming:      p.catalog.Ask.StatusStreaming,
			StatusDone:           p.catalog.Ask.StatusDone,
			StatusError:          p.catalog.Ask.StatusError,
			DefaultError:         p.catalog.Ask.DefaultError,
		},
		Result: result,
	}
}

func (p *Presenter) CatalogPage(state CatalogPageState) views.CatalogPageData {
	cards := make([]views.CatalogCourseCard, 0, len(state.Courses))
	for _, item := range state.Courses {
		cards = append(cards, views.CatalogCourseCard{
			ID:             item.ID,
			Title:          item.Title,
			Description:    item.Description,
			Language:       item.Language,
			RepoURL:        item.RepoURL,
			StudentRepoURL: item.StudentRepoURL,
			ProgressLabel:  progressLabel(item.Progress, item.ProgressKnown),
			ProgressValue:  progressPercent(item.Progress),
		})
	}

	return views.CatalogPageData{
		Layout: p.layout(p.catalog.Course.Title, "/"),
		Header: views.PageHeader{
			Title:       p.catalog.Course.Title,
			Description: p.catalog.Course.Description,
		},
		Error:   state.Error,
		Success: state.Success,
		Form: views.AddCourseFormData{
			Title:        p.catalog.Course.AddCourseTitle,
			Description:  p.catalog.Course.AddCourseDescription,
			Action:       "/courses/add",
			RepoURLLabel: p.catalog.Course.RepoURLLabel,
			RepoURLValue: state.RepoURL,
			RepoURLHint:  p.catalog.Course.RepoURLPlaceholder,
			SubmitLabel:  p.catalog.Course.AddCourseSubmit,
		},
		Courses: cards,
		Empty:   p.catalog.Course.NoCourses,
	}
}

func (p *Presenter) CoursePage(state CoursePageState) views.CoursePageData {
	sections := make([]views.CourseSectionView, 0, len(state.View.Course.Sections))
	for _, section := range state.View.Course.Sections {
		sectionView := views.CourseSectionView{ID: section.ID, Title: section.Title}
		for _, lesson := range section.Lessons {
			sectionView.Lessons = append(sectionView.Lessons, views.CourseLessonView{
				ID:          lesson.ID,
				Title:       lesson.Title,
				Objective:   lesson.Objective,
				StatusLabel: lessonStatus(state.View.Progress, lesson.ID, state.View.ProgressKnown),
				Path:        "/course/" + state.View.ID + "/lesson/" + lesson.ID,
			})
		}
		sections = append(sections, sectionView)
	}

	return views.CoursePageData{
		Layout: p.layout(state.View.Title, "/"),
		Header: views.PageHeader{
			Title:       state.View.Title,
			Description: state.View.Description,
		},
		Error:          state.Error,
		CourseID:       state.View.ID,
		RepoURL:        state.View.RepoURL,
		StudentRepoURL: state.View.StudentRepoURL,
		ProgressLabel:  progressLabel(state.View.Progress, state.View.ProgressKnown),
		ProgressValue:  progressPercent(state.View.Progress),
		ProgressKnown:  state.View.ProgressKnown,
		EnrollLabel:    p.catalog.Course.EnrollLabel,
		EnrollValue:    state.View.StudentRepoURL,
		EnrollHint:     p.catalog.Course.EnrollPlaceholder,
		EnrollAction:   "/courses/" + state.View.ID + "/enroll",
		EnrollSubmit:   p.catalog.Course.EnrollSubmit,
		RemoveAction:   "/courses/" + state.View.ID + "/remove",
		RemoveLabel:    p.catalog.Course.RemoveCourse,
		SectionsTitle:  p.catalog.Course.LessonsTitle,
		Sections:       sections,
	}
}

func (p *Presenter) LessonPage(state LessonPageState) views.LessonPageData {
	lesson := findLesson(state.Course.Course, state.LessonID)
	checks := make([]views.LessonCheckView, 0, len(lesson.Checklist))
	statusMap, msgMap := lessonState(state.Course.Progress, state.LessonID)
	for _, item := range lesson.Checklist {
		checks = append(checks, views.LessonCheckView{
			Label:   item.Label,
			Verify:  item.Verify,
			Done:    statusMap[item.ID],
			Message: msgMap[item.ID],
		})
	}

	askPath := "/ask"
	query := url.Values{}
	query.Set("context", lesson.AskContext)
	query.Set("prompt", "")
	askPath += "?" + query.Encode()

	return views.LessonPageData{
		Layout: p.layout(lesson.Title, "/"),
		Header: views.PageHeader{
			Title:       lesson.Title,
			Description: lesson.Objective,
		},
		Error:         state.Error,
		CourseTitle:   state.Course.Title,
		CoursePath:    "/course/" + state.Course.ID,
		ProgressLabel: progressLabel(state.Course.Progress, state.Course.ProgressKnown),
		ProgressValue: progressPercent(state.Course.Progress),
		Checks:        checks,
		HintsTitle:    p.catalog.Course.HintsTitle,
		Hints:         lesson.Hints,
		AskLabel:      p.catalog.Course.AskLesson,
		AskPath:       askPath,
	}
}

func (p *Presenter) layout(title, active string) views.LayoutData {
	return views.LayoutData{
		Title:          title + " | " + p.catalog.Common.BrandName,
		Locale:         p.locale(),
		Active:         active,
		BrandName:      p.catalog.Common.BrandName,
		NavItems:       p.layoutNav(),
		LanguageToggle: p.languageToggle(),
		LogoutAction: partials.LogoutActionData{
			Enabled: p.cfg.AuthEnabled,
			Action:  "/logout",
			Title:   p.catalog.Common.Actions.Logout,
		},
		ThemeToggle: views.ThemeToggleData{
			Label:              p.catalog.Common.Theme.Label,
			SwitchToDarkLabel:  p.catalog.Common.Theme.SwitchToDarkLabel,
			SwitchToLightLabel: p.catalog.Common.Theme.SwitchToLightLabel,
		},
	}
}

func (p *Presenter) layoutNav() []ui8layout.NavItem {
	items := make([]ui8layout.NavItem, 0, len(p.navItems))
	for _, item := range p.navItems {
		label := item.Label
		switch item.Path {
		case "/":
			if value := strings.TrimSpace(p.catalog.Common.Nav["courses"]); value != "" {
				label = value
			}
		case "/ask":
			if value := strings.TrimSpace(p.catalog.Common.Nav["ask"]); value != "" {
				label = value
			}
		}
		items = append(items, ui8layout.NavItem{
			Path:  item.Path,
			Label: label,
			Icon:  item.Icon,
		})
	}
	return items
}

func (p *Presenter) locale() string {
	if strings.TrimSpace(p.cfg.Locale) != "" {
		return p.cfg.Locale
	}
	return p.cfg.DefaultLocale
}

func (p *Presenter) languageToggle() partials.LanguageToggleData {
	current := p.locale()
	available := append([]string(nil), p.cfg.ToggleLocales...)
	if len(available) == 0 {
		available = append([]string(nil), p.cfg.AvailableLocales...)
	}
	next := current
	for _, locale := range available {
		if locale != current {
			next = locale
			break
		}
	}
	return partials.LanguageToggleData{
		Enabled:          len(available) > 1,
		CurrentLocale:    current,
		CurrentLabel:     p.catalog.Common.LocaleLabels[current],
		AvailableLocales: strings.Join(available, ","),
		DefaultLocale:    p.cfg.DefaultLocale,
		NextLocale:       next,
		Title:            "Toggle language",
	}
}

func progressPercent(progress course.Progress) int {
	done, total := progressCount(progress)
	if total == 0 {
		return 0
	}
	return int(float64(done) / float64(total) * 100)
}

func progressLabel(progress course.Progress, known bool) string {
	if !known {
		return "0%"
	}
	done, total := progressCount(progress)
	if total == 0 {
		return "0%"
	}
	return fmt.Sprintf("%d/%d complete", done, total)
}

func progressCount(progress course.Progress) (int, int) {
	done := 0
	total := len(progress.Lessons)
	for _, lesson := range progress.Lessons {
		if strings.EqualFold(lesson.Status, "done") {
			done++
		}
	}
	return done, total
}

func lessonStatus(progress course.Progress, lessonID string, known bool) string {
	if !known {
		return "pending"
	}
	for _, lesson := range progress.Lessons {
		if lesson.ID == lessonID && lesson.Status != "" {
			return lesson.Status
		}
	}
	return "pending"
}

func lessonState(progress course.Progress, lessonID string) (map[string]bool, map[string]string) {
	for _, lesson := range progress.Lessons {
		if lesson.ID == lessonID {
			return lesson.Checks, lesson.Messages
		}
	}
	return map[string]bool{}, map[string]string{}
}

func findLesson(data course.Course, lessonID string) course.Lesson {
	for _, section := range data.Sections {
		for _, lesson := range section.Lessons {
			if lesson.ID == lessonID {
				return lesson
			}
		}
	}
	return course.Lesson{ID: lessonID, Title: lessonID, Objective: lessonID}
}
