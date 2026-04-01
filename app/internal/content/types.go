package content

type Catalog struct {
	Locale       string
	Common       CommonContent
	Shell        ShellContent
	Health       HealthContent
	Capabilities CapabilitiesContent
	Ask          AskContent
	Egress       EgressContent
	Audit        AuditContent
}

type CommonContent struct {
	LocaleCode         string `json:"localeCode"`
	LocaleName         string `json:"localeName"`
	SwitchLanguage     string `json:"switchLanguage"`
	ToggleTheme        string `json:"toggleTheme"`
	SwitchToDarkTheme  string `json:"switchToDarkTheme"`
	SwitchToLightTheme string `json:"switchToLightTheme"`
	BoolTrue           string `json:"boolTrue"`
	BoolFalse          string `json:"boolFalse"`
	EmptyValue         string `json:"emptyValue"`
}

type NavItemContent struct {
	Path  string `json:"path"`
	Label string `json:"label"`
	Icon  string `json:"icon"`
}

type ShellContent struct {
	BrandName string           `json:"brandName"`
	Nav       []NavItemContent `json:"nav"`
}

type PageHeaderContent struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type PageErrors struct {
	Default string `json:"default"`
}

type HealthContent struct {
	Title        string             `json:"title"`
	ActivePath   string             `json:"activePath"`
	Header       PageHeaderContent  `json:"header"`
	KPILabels    HealthKPILabels    `json:"kpiLabels"`
	DetailsTitle string             `json:"detailsTitle"`
	DetailLabels HealthDetailLabels `json:"detailLabels"`
	Errors       PageErrors         `json:"errors"`
}

type HealthKPILabels struct {
	Service         string `json:"service"`
	Profile         string `json:"profile"`
	DiscoveryStatus string `json:"discoveryStatus"`
	AdapterStatus   string `json:"adapterStatus"`
}

type HealthDetailLabels struct {
	Adapter string `json:"adapter"`
	Profile string `json:"profile"`
}

type CapabilitiesContent struct {
	Title               string                 `json:"title"`
	ActivePath          string                 `json:"activePath"`
	Header              PageHeaderContent      `json:"header"`
	KPILabels           CapabilitiesKPILabels  `json:"kpiLabels"`
	FlagsTitle          string                 `json:"flagsTitle"`
	FlagLabels          CapabilitiesFlagLabels `json:"flagLabels"`
	CapabilityListTitle string                 `json:"capabilityListTitle"`
	Errors              PageErrors             `json:"errors"`
}

type CapabilitiesKPILabels struct {
	ProfileID   string `json:"profileID"`
	DisplayName string `json:"displayName"`
	AIModel     string `json:"aiModel"`
}

type CapabilitiesFlagLabels struct {
	HTTPBind     string `json:"httpBind"`
	AIEnabled    string `json:"aiEnabled"`
	AIProvider   string `json:"aiProvider"`
	AIAPIMode    string `json:"aiApiMode"`
	AIBaseURL    string `json:"aiBaseURL"`
	AIHasAPIKey  string `json:"aiHasApiKey"`
	ChatHistory  string `json:"chatHistory"`
	EmailEnabled string `json:"emailEnabled"`
	ProxySession string `json:"proxySession"`
	ProxyForce   string `json:"proxyForce"`
}

type AskContent struct {
	Title      string            `json:"title"`
	ActivePath string            `json:"activePath"`
	Header     PageHeaderContent `json:"header"`
	Fields     AskFieldsContent  `json:"fields"`
	Actions    AskActionsContent `json:"actions"`
	Stream     AskStreamContent  `json:"stream"`
	Result     AskResultContent  `json:"result"`
	Errors     AskErrorsContent  `json:"errors"`
}

type AskFieldsContent struct {
	Prompt AskFieldContent `json:"prompt"`
	Model  AskFieldContent `json:"model"`
}

type AskFieldContent struct {
	Label       string `json:"label"`
	Placeholder string `json:"placeholder"`
}

type AskActionsContent struct {
	StreamSubmit string `json:"streamSubmit"`
	SyncSubmit   string `json:"syncSubmit"`
}

type AskStreamContent struct {
	Title                string `json:"title"`
	InitialStatus        string `json:"initialStatus"`
	StatusSync           string `json:"statusSync"`
	StatusPromptRequired string `json:"statusPromptRequired"`
	StatusConnecting     string `json:"statusConnecting"`
	StatusStreaming      string `json:"statusStreaming"`
	StatusDone           string `json:"statusDone"`
	StatusError          string `json:"statusError"`
	DefaultError         string `json:"defaultError"`
}

type AskResultContent struct {
	Title        string `json:"title"`
	StatusLabel  string `json:"statusLabel"`
	MessageLabel string `json:"messageLabel"`
}

type AskErrorsContent struct {
	InvalidForm    string `json:"invalidForm"`
	PromptRequired string `json:"promptRequired"`
}

type EgressContent struct {
	Title        string             `json:"title"`
	ActivePath   string             `json:"activePath"`
	Header       PageHeaderContent  `json:"header"`
	TableHeaders EgressTableHeaders `json:"tableHeaders"`
	Errors       PageErrors         `json:"errors"`
}

type EgressTableHeaders struct {
	Gateway        string `json:"gateway"`
	Interface      string `json:"interface"`
	Priority       string `json:"priority"`
	Healthy        string `json:"healthy"`
	Level          string `json:"level"`
	Checks         string `json:"checks"`
	Active         string `json:"active"`
	LastCheck      string `json:"lastCheck"`
	LastTransition string `json:"lastTransition"`
	LastError      string `json:"lastError"`
}

type AuditContent struct {
	Title        string            `json:"title"`
	ActivePath   string            `json:"activePath"`
	Header       PageHeaderContent `json:"header"`
	TableHeaders AuditTableHeaders `json:"tableHeaders"`
	Errors       PageErrors        `json:"errors"`
}

type AuditTableHeaders struct {
	Time      string `json:"time"`
	Command   string `json:"command"`
	Principal string `json:"principal"`
	Transport string `json:"transport"`
	Outcome   string `json:"outcome"`
	Message   string `json:"message"`
}
