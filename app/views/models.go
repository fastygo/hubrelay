package views

import (
	ui8layout "github.com/fastygo/ui8kit/layout"
	"hubrelay-dashboard/views/partials"
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

type KV struct {
	Label string
	Value string
}

type KPIView struct {
	Label string
	Value string
}

type HealthPageData struct {
	Layout       LayoutData
	Header       PageHeader
	Error        string
	KPIs         []KPIView
	DetailsTitle string
	Details      []KV
}

type CapabilitiesPageData struct {
	Layout              LayoutData
	Header              PageHeader
	Error               string
	KPIs                []KPIView
	FlagsTitle          string
	Flags               []KV
	CapabilityListTitle string
	CapabilityList      []string
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
	Actions        AskActionsView
	Stream         partials.StreamOutputData
	Result         *AskResultView
}

type EgressHeaders struct {
	Gateway        string
	Interface      string
	Priority       string
	Healthy        string
	Level          string
	Checks         string
	Active         string
	LastCheck      string
	LastTransition string
	LastError      string
}

type EgressGatewayView struct {
	Gateway        string
	Interface      string
	Priority       string
	Healthy        string
	Level          string
	Checks         string
	Active         string
	LastCheck      string
	LastTransition string
	LastError      string
}

type EgressPageData struct {
	Layout  LayoutData
	Header  PageHeader
	Error   string
	Headers EgressHeaders
	Rows    []EgressGatewayView
}

type AuditHeaders struct {
	Time      string
	Command   string
	Principal string
	Transport string
	Outcome   string
	Message   string
}

type AuditEntryView struct {
	Time      string
	Command   string
	Principal string
	Transport string
	Outcome   string
	Message   string
}

type AuditPageData struct {
	Layout  LayoutData
	Header  PageHeader
	Error   string
	Headers AuditHeaders
	Rows    []AuditEntryView
}
