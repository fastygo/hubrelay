package content

type Catalog struct {
	Common CommonContent `json:"common"`
	Login  LoginContent  `json:"login"`
	Ask    AskContent    `json:"ask"`
	Course CourseContent `json:"course"`
}

type CommonContent struct {
	BrandName    string            `json:"brand_name"`
	DefaultTitle string            `json:"default_title"`
	Nav          map[string]string `json:"nav"`
	LocaleLabels map[string]string `json:"locale_labels"`
	Actions      CommonActions     `json:"actions"`
	Theme        ThemeContent      `json:"theme"`
}

type CommonActions struct {
	Logout string `json:"logout"`
}

type ThemeContent struct {
	Label              string `json:"label"`
	SwitchToDarkLabel  string `json:"switch_to_dark_label"`
	SwitchToLightLabel string `json:"switch_to_light_label"`
}

type LoginContent struct {
	Title            string `json:"title"`
	Description      string `json:"description"`
	UsernameLabel    string `json:"username_label"`
	PasswordLabel    string `json:"password_label"`
	SubmitLabel      string `json:"submit_label"`
	InvalidFormError string `json:"invalid_form_error"`
	InvalidCredsError string `json:"invalid_creds_error"`
}

type AskContent struct {
	Title                string `json:"title"`
	Description          string `json:"description"`
	PromptLabel          string `json:"prompt_label"`
	PromptPlaceholder    string `json:"prompt_placeholder"`
	ModelLabel           string `json:"model_label"`
	ModelPlaceholder     string `json:"model_placeholder"`
	ContextLabel         string `json:"context_label"`
	ContextPlaceholder   string `json:"context_placeholder"`
	StreamSubmitLabel    string `json:"stream_submit_label"`
	SyncSubmitLabel      string `json:"sync_submit_label"`
	PromptRequiredError  string `json:"prompt_required_error"`
	StreamTitle          string `json:"stream_title"`
	StatusIdle           string `json:"status_idle"`
	StatusSync           string `json:"status_sync"`
	StatusPromptRequired string `json:"status_prompt_required"`
	StatusConnecting     string `json:"status_connecting"`
	StatusStreaming      string `json:"status_streaming"`
	StatusDone           string `json:"status_done"`
	StatusError          string `json:"status_error"`
	DefaultError         string `json:"default_error"`
	ResultTitle          string `json:"result_title"`
}

type CourseContent struct {
	Title               string `json:"title"`
	Description         string `json:"description"`
	AddCourseTitle      string `json:"add_course_title"`
	AddCourseDescription string `json:"add_course_description"`
	RepoURLLabel        string `json:"repo_url_label"`
	RepoURLPlaceholder  string `json:"repo_url_placeholder"`
	AddCourseSubmit     string `json:"add_course_submit"`
	OpenCourse          string `json:"open_course"`
	RemoveCourse        string `json:"remove_course"`
	EnrollLabel         string `json:"enroll_label"`
	EnrollPlaceholder   string `json:"enroll_placeholder"`
	EnrollSubmit        string `json:"enroll_submit"`
	ProgressTitle       string `json:"progress_title"`
	LessonsTitle        string `json:"lessons_title"`
	HintsTitle          string `json:"hints_title"`
	AskLesson           string `json:"ask_lesson"`
	NoCourses           string `json:"no_courses"`
	AddCourseSuccess    string `json:"add_course_success"`
	ProgressUnavailable string `json:"progress_unavailable"`
}
