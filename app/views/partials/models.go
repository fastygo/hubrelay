package partials

type LanguageToggleData struct {
	Enabled          bool
	CurrentLocale    string
	CurrentLabel     string
	AvailableLocales string
	DefaultLocale    string
	NextLocale       string
	Title            string
}

type StreamOutputData struct {
	Title                string
	InitialStatus        string
	StatusSync           string
	StatusPromptRequired string
	StatusConnecting     string
	StatusStreaming      string
	StatusDone           string
	StatusError          string
	DefaultError         string
}
