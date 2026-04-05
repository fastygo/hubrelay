package course

type Course struct {
	ID          string    `json:"id"`
	Version     string    `json:"version"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Language    string    `json:"language"`
	Sections    []Section `json:"sections"`
}

type Section struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Lessons     []Lesson `json:"lessons"`
}

type Lesson struct {
	ID         string      `json:"id"`
	Title      string      `json:"title"`
	Objective  string      `json:"objective"`
	AskContext string      `json:"ask_context,omitempty"`
	Hints      []string    `json:"hints,omitempty"`
	Checklist  []CheckItem `json:"checklist"`
}

type CheckItem struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Verify string `json:"verify"`
}

type Progress struct {
	CourseID string         `json:"course_id"`
	Lessons  []LessonStatus `json:"lessons"`
}

type LessonStatus struct {
	ID        string            `json:"id"`
	Status    string            `json:"status"`
	Checks    map[string]bool   `json:"checks,omitempty"`
	Messages  map[string]string `json:"messages,omitempty"`
	Completed string            `json:"completed_at,omitempty"`
}
