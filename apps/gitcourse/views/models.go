package views

import (
	ui8layout "github.com/fastygo/ui8kit/layout"
	"gitcourse/views/partials"
)

type LayoutData struct {
	Title          string
	Locale         string
	Active         string
	BrandName      string
	NavItems       []ui8layout.NavItem
	LanguageToggle partials.LanguageToggleData
	LogoutAction   partials.LogoutActionData
	ThemeToggle    ThemeToggleData
}

type ThemeToggleData struct {
	Label              string
	SwitchToDarkLabel  string
	SwitchToLightLabel string
}

type PageHeader struct {
	Title       string
	Description string
}

type LoginPageData struct {
	Layout        LayoutData
	Header        PageHeader
	Error         string
	FormAction    string
	UsernameLabel string
	UsernameValue string
	PasswordLabel string
	SubmitLabel   string
}

type AskFieldView struct {
	ID          string
	Name        string
	Label       string
	Placeholder string
	Value       string
}

type AskActionsView struct {
	StreamSubmitLabel string
	SyncSubmitLabel   string
}

type AskResultView struct {
	Title string
	Rows  []KV
}

type AskPageData struct {
	Layout         LayoutData
	Header         PageHeader
	Error          string
	FormAction     string
	StreamEndpoint string
	PromptField    AskFieldView
	ModelField     AskFieldView
	ContextField   AskFieldView
	Actions        AskActionsView
	Stream         partials.StreamOutputData
	Result         *AskResultView
}

type KV struct {
	Label string
	Value string
}

type CatalogCourseCard struct {
	ID             string
	Title          string
	Description    string
	Language       string
	RepoURL        string
	StudentRepoURL string
	ProgressLabel  string
	ProgressValue  int
}

type AddCourseFormData struct {
	Title          string
	Description    string
	Action         string
	RepoURLLabel   string
	RepoURLValue   string
	RepoURLHint    string
	SubmitLabel    string
}

type CatalogPageData struct {
	Layout   LayoutData
	Header   PageHeader
	Error    string
	Success  string
	Form     AddCourseFormData
	Courses  []CatalogCourseCard
	Empty    string
}

type CourseLessonView struct {
	ID          string
	Title       string
	Objective   string
	StatusLabel string
	Path        string
}

type CourseSectionView struct {
	ID      string
	Title   string
	Lessons []CourseLessonView
}

type CoursePageData struct {
	Layout          LayoutData
	Header          PageHeader
	Error           string
	CourseID        string
	RepoURL         string
	StudentRepoURL  string
	ProgressLabel   string
	ProgressValue   int
	ProgressKnown   bool
	EnrollLabel     string
	EnrollValue     string
	EnrollHint      string
	EnrollAction    string
	EnrollSubmit    string
	RemoveAction    string
	RemoveLabel     string
	SectionsTitle   string
	Sections        []CourseSectionView
}

type LessonCheckView struct {
	Label     string
	Verify    string
	Done      bool
	Message   string
}

type LessonPageData struct {
	Layout         LayoutData
	Header         PageHeader
	Error          string
	CourseTitle    string
	CoursePath     string
	ProgressLabel  string
	ProgressValue  int
	Checks         []LessonCheckView
	HintsTitle     string
	Hints          []string
	AskLabel       string
	AskPath        string
}
