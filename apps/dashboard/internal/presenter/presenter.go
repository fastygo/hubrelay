package presenter

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	coreshell "github.com/fastygo/hubcore/shell"
	ui8layout "github.com/fastygo/ui8kit/layout"
	"hubrelay-dashboard/internal/content"
	appmodule "hubrelay-dashboard/internal/module"
	"hubrelay-dashboard/internal/source"
	"hubrelay-dashboard/views"
	"hubrelay-dashboard/views/partials"
)

type Presenter struct {
	content          content.Catalog
	authEnabled      bool
	defaultLocale    string
	availableLocales []string
	toggleLocales    []string
	localeMeta       map[string]localeMeta
	navItems         []appmodule.NavItem
}

type Config struct {
	AuthEnabled      bool
	DefaultLocale    string
	AvailableLocales []string
	ToggleLocales    []string
	Catalogs         map[string]content.Catalog
	NavItems         []appmodule.NavItem
}

type localeMeta struct {
	Code string
	Name string
}

type AskPageState struct {
	Prompt string
	Model  string
	Error  string
	Result *source.CommandResult
}

type LoginPageState struct {
	Username string
	Error    string
}

func New(catalog content.Catalog, cfg Config) *Presenter {
	meta := make(map[string]localeMeta, len(cfg.Catalogs))
	for locale, item := range cfg.Catalogs {
		meta[locale] = localeMeta{
			Code: item.Common.LocaleCode,
			Name: item.Common.LocaleName,
		}
	}

	return &Presenter{
		content:          catalog,
		authEnabled:      cfg.AuthEnabled,
		defaultLocale:    cfg.DefaultLocale,
		availableLocales: append([]string(nil), cfg.AvailableLocales...),
		toggleLocales:    append([]string(nil), cfg.ToggleLocales...),
		localeMeta:       meta,
		navItems:         append([]appmodule.NavItem(nil), cfg.NavItems...),
	}
}

func (p *Presenter) SetNavItems(items []appmodule.NavItem) {
	p.navItems = append([]appmodule.NavItem(nil), items...)
}

func (p *Presenter) AskPromptRequiredError() string {
	return p.content.Ask.Errors.PromptRequired
}

func (p *Presenter) AskInvalidFormError() string {
	return p.content.Ask.Errors.InvalidForm
}

func (p *Presenter) LoginInvalidFormError() string {
	return p.content.Login.Errors.InvalidForm
}

func (p *Presenter) LoginInvalidCredentialsError() string {
	return p.content.Login.Errors.InvalidCredentials
}

func (p *Presenter) HomePath() string {
	return p.localizePath("/")
}

func (p *Presenter) LoginPath() string {
	return p.localizePath("/login")
}

func (p *Presenter) LoginPage(state LoginPageState) views.LoginPageData {
	copy := p.content.Login

	return views.LoginPageData{
		Layout: p.layout(copy.Title, ""),
		Header: views.PageHeader{
			Title:       copy.Header.Title,
			Description: copy.Header.Description,
		},
		Error:         strings.TrimSpace(state.Error),
		FormAction:    p.LoginPath(),
		UsernameLabel: copy.UsernameLabel,
		UsernameValue: strings.TrimSpace(state.Username),
		PasswordLabel: copy.PasswordLabel,
		SubmitLabel:   copy.SubmitLabel,
	}
}

func (p *Presenter) HealthPage(data source.HealthData, err error) views.HealthPageData {
	copy := p.content.Health

	return views.HealthPageData{
		Layout: p.layout(copy.Title, copy.ActivePath),
		Header: views.PageHeader{Title: copy.Header.Title, Description: copy.Header.Description},
		Error:  presentError(err, copy.Errors.Default),
		KPIs: []views.KPIView{
			{Label: copy.KPILabels.Service, Value: p.valueOrEmpty(data.Discovery.Service)},
			{Label: copy.KPILabels.Profile, Value: p.valueOrEmpty(data.Discovery.Profile)},
			{Label: copy.KPILabels.DiscoveryStatus, Value: p.valueOrEmpty(data.Discovery.Status)},
			{Label: copy.KPILabels.AdapterStatus, Value: p.valueOrEmpty(data.Health.Status)},
		},
		DetailsTitle: copy.DetailsTitle,
		Details: []views.KV{
			{Label: copy.DetailLabels.Adapter, Value: p.valueOrEmpty(data.Health.Adapter)},
			{Label: copy.DetailLabels.Profile, Value: p.valueOrEmpty(data.Health.Profile)},
		},
	}
}

func (p *Presenter) CapabilitiesPage(data source.CapabilitiesData, err error) views.CapabilitiesPageData {
	copy := p.content.Capabilities

	return views.CapabilitiesPageData{
		Layout: p.layout(copy.Title, copy.ActivePath),
		Header: views.PageHeader{Title: copy.Header.Title, Description: copy.Header.Description},
		Error:  presentError(err, copy.Errors.Default),
		KPIs: []views.KPIView{
			{Label: copy.KPILabels.ProfileID, Value: p.valueOrEmpty(data.ProfileID)},
			{Label: copy.KPILabels.DisplayName, Value: p.valueOrEmpty(data.DisplayName)},
			{Label: copy.KPILabels.AIModel, Value: p.valueOrEmpty(data.AIModel)},
		},
		FlagsTitle: copy.FlagsTitle,
		Flags: []views.KV{
			{Label: copy.FlagLabels.HTTPBind, Value: p.valueOrEmpty(data.HTTPBind)},
			{Label: copy.FlagLabels.AIEnabled, Value: p.boolLabel(data.AIEnabled)},
			{Label: copy.FlagLabels.AIProvider, Value: p.valueOrEmpty(data.AIProvider)},
			{Label: copy.FlagLabels.AIAPIMode, Value: p.valueOrEmpty(data.AIAPIMode)},
			{Label: copy.FlagLabels.AIBaseURL, Value: p.valueOrEmpty(data.AIBaseURL)},
			{Label: copy.FlagLabels.AIHasAPIKey, Value: p.boolLabel(data.AIHasAPIKey)},
			{Label: copy.FlagLabels.ChatHistory, Value: p.boolLabel(data.ChatHistory)},
			{Label: copy.FlagLabels.EmailEnabled, Value: p.boolLabel(data.EmailEnabled)},
			{Label: copy.FlagLabels.ProxySession, Value: p.boolLabel(data.ProxySession)},
			{Label: copy.FlagLabels.ProxyForce, Value: p.boolLabel(data.ProxyForce)},
		},
		CapabilityListTitle: copy.CapabilityListTitle,
		CapabilityList:      append([]string(nil), data.Capabilities...),
	}
}

func (p *Presenter) AskPage(state AskPageState) views.AskPageData {
	copy := p.content.Ask

	page := views.AskPageData{
		Layout: p.layout(copy.Title, copy.ActivePath),
		Header: views.PageHeader{
			Title:       copy.Header.Title,
			Description: copy.Header.Description,
		},
		Error:          state.Error,
		FormAction:     p.localizePath("/ask"),
		StreamEndpoint: p.localizePath("/ask/stream"),
		PromptField: views.AskFieldView{
			ID:          "prompt-input",
			Name:        "prompt",
			Label:       copy.Fields.Prompt.Label,
			Placeholder: copy.Fields.Prompt.Placeholder,
			Value:       state.Prompt,
		},
		ModelField: views.AskFieldView{
			ID:          "model-input",
			Name:        "model",
			Label:       copy.Fields.Model.Label,
			Placeholder: copy.Fields.Model.Placeholder,
			Value:       state.Model,
		},
		Actions: views.AskActionsView{
			StreamSubmitLabel: copy.Actions.StreamSubmit,
			SyncSubmitLabel:   copy.Actions.SyncSubmit,
		},
		Stream: partials.StreamOutputData{
			Title:                copy.Stream.Title,
			InitialStatus:        copy.Stream.InitialStatus,
			StatusSync:           copy.Stream.StatusSync,
			StatusPromptRequired: copy.Stream.StatusPromptRequired,
			StatusConnecting:     copy.Stream.StatusConnecting,
			StatusStreaming:      copy.Stream.StatusStreaming,
			StatusDone:           copy.Stream.StatusDone,
			StatusError:          copy.Stream.StatusError,
			DefaultError:         copy.Stream.DefaultError,
		},
	}

	if state.Result == nil {
		return page
	}

	rows := []views.KV{
		{Label: copy.Result.StatusLabel, Value: p.valueOrEmpty(state.Result.Status)},
		{Label: copy.Result.MessageLabel, Value: p.valueOrEmpty(state.Result.Message)},
	}

	keys := make([]string, 0, len(state.Result.Data))
	for key := range state.Result.Data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		rows = append(rows, views.KV{
			Label: key,
			Value: p.stringify(state.Result.Data[key]),
		})
	}

	page.Result = &views.AskResultView{
		Title: copy.Result.Title,
		Rows:  rows,
	}

	return page
}

func (p *Presenter) EgressPage(data source.EgressData, err error) views.EgressPageData {
	copy := p.content.Egress
	rows := make([]views.EgressGatewayView, 0, len(data.Gateways))
	for _, gateway := range data.Gateways {
		rows = append(rows, views.EgressGatewayView{
			Gateway:        p.valueOrEmpty(gateway.Name),
			Interface:      p.valueOrEmpty(gateway.Interface),
			Priority:       strconv.Itoa(gateway.Priority),
			Healthy:        p.boolLabel(gateway.Healthy),
			Level:          p.valueOrEmpty(gateway.HealthLevel),
			Checks:         p.gatewaySummary(gateway),
			Active:         p.boolLabel(gateway.Active),
			LastCheck:      p.formatTime(gateway.LastCheckAt),
			LastTransition: p.formatTime(gateway.LastTransitionAt),
			LastError:      p.valueOrEmpty(gateway.LastError),
		})
	}

	return views.EgressPageData{
		Layout: p.layout(copy.Title, copy.ActivePath),
		Header: views.PageHeader{
			Title:       copy.Header.Title,
			Description: copy.Header.Description,
		},
		Error: presentError(err, copy.Errors.Default),
		Headers: views.EgressHeaders{
			Gateway:        copy.TableHeaders.Gateway,
			Interface:      copy.TableHeaders.Interface,
			Priority:       copy.TableHeaders.Priority,
			Healthy:        copy.TableHeaders.Healthy,
			Level:          copy.TableHeaders.Level,
			Checks:         copy.TableHeaders.Checks,
			Active:         copy.TableHeaders.Active,
			LastCheck:      copy.TableHeaders.LastCheck,
			LastTransition: copy.TableHeaders.LastTransition,
			LastError:      copy.TableHeaders.LastError,
		},
		Rows: rows,
	}
}

func (p *Presenter) AuditPage(entries []source.AuditEntry, err error) views.AuditPageData {
	copy := p.content.Audit
	rows := make([]views.AuditEntryView, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, views.AuditEntryView{
			Time:      p.formatTime(entry.At),
			Command:   p.valueOrEmpty(entry.Command),
			Principal: p.valueOrEmpty(entry.Principal),
			Transport: p.valueOrEmpty(entry.Transport),
			Outcome:   p.valueOrEmpty(entry.Outcome),
			Message:   p.valueOrEmpty(entry.Message),
		})
	}

	return views.AuditPageData{
		Layout: p.layout(copy.Title, copy.ActivePath),
		Header: views.PageHeader{
			Title:       copy.Header.Title,
			Description: copy.Header.Description,
		},
		Error: presentError(err, copy.Errors.Default),
		Headers: views.AuditHeaders{
			Time:      copy.TableHeaders.Time,
			Command:   copy.TableHeaders.Command,
			Principal: copy.TableHeaders.Principal,
			Transport: copy.TableHeaders.Transport,
			Outcome:   copy.TableHeaders.Outcome,
			Message:   copy.TableHeaders.Message,
		},
		Rows: rows,
	}
}

func (p *Presenter) layout(title, active string) views.LayoutData {
	currentLocale := p.locale()

	return views.LayoutData{
		Title:          title,
		Locale:         currentLocale,
		Active:         p.localizePath(active),
		BrandName:      p.content.Shell.BrandName,
		NavItems:       p.layoutNavItems(),
		LanguageToggle: p.languageToggle(currentLocale),
		LogoutAction: partials.LogoutActionData{
			Enabled: p.authEnabled,
			Action:  p.localizePath("/logout"),
			Title:   p.content.Common.SignOut,
		},
		ThemeToggle: views.ThemeToggleData{
			Label:              p.content.Common.ToggleTheme,
			SwitchToDarkLabel:  p.content.Common.SwitchToDarkTheme,
			SwitchToLightLabel: p.content.Common.SwitchToLightTheme,
		},
	}
}

func (p *Presenter) layoutNavItems() []ui8layout.NavItem {
	contentItems := make([]coreshell.ContentNavItem, 0, len(p.content.Shell.Nav))
	for _, item := range p.content.Shell.Nav {
		contentItems = append(contentItems, coreshell.ContentNavItem{
			Path:  item.Path,
			Label: item.Label,
			Icon:  item.Icon,
		})
	}

	navItems := coreshell.ResolveNavItems(p.defaultLocale, p.locale(), contentItems, p.navItems)
	for i := range navItems {
		navItems[i].Label = p.valueOrEmpty(navItems[i].Label)
	}
	return navItems
}

func (p *Presenter) locale() string {
	if p.content.Locale == "" {
		return p.defaultLocale
	}
	return p.content.Locale
}

func (p *Presenter) localizePath(raw string) string {
	return coreshell.LocalizePath(p.defaultLocale, p.locale(), raw)
}

func (p *Presenter) languageToggle(currentLocale string) partials.LanguageToggleData {
	if len(p.toggleLocales) < 2 {
		return partials.LanguageToggleData{}
	}

	nextLocale := p.toggleLocales[0]
	if currentLocale == nextLocale {
		nextLocale = p.toggleLocales[1]
	}

	meta := p.localeMeta[currentLocale]
	currentLabel := meta.Code
	if strings.TrimSpace(currentLabel) == "" {
		currentLabel = strings.ToUpper(currentLocale)
	}

	title := p.content.Common.SwitchLanguage
	if nextMeta := p.localeMeta[nextLocale]; strings.TrimSpace(nextMeta.Name) != "" {
		title = strings.TrimSpace(title + ": " + nextMeta.Name)
	}

	return partials.LanguageToggleData{
		Enabled:          true,
		CurrentLocale:    currentLocale,
		CurrentLabel:     currentLabel,
		AvailableLocales: strings.Join(p.availableLocales, ","),
		DefaultLocale:    p.defaultLocale,
		NextLocale:       nextLocale,
		Title:            title,
	}
}

func (p *Presenter) boolLabel(value bool) string {
	if value {
		return p.content.Common.BoolTrue
	}
	return p.content.Common.BoolFalse
}

func (p *Presenter) valueOrEmpty(value string) string {
	if strings.TrimSpace(value) == "" {
		return p.content.Common.EmptyValue
	}
	return value
}

func (p *Presenter) formatTime(value time.Time) string {
	if value.IsZero() {
		return p.content.Common.EmptyValue
	}
	return value.UTC().Format(time.RFC3339)
}

func (p *Presenter) gatewaySummary(gateway source.GatewayData) string {
	parts := []string{
		"WG=" + p.boolLabel(gateway.Levels.WG.OK),
		"transport=" + p.boolLabel(gateway.Levels.Transport.OK),
		"business=" + p.boolLabel(gateway.Levels.Business.OK),
	}
	return strings.Join(parts, " | ")
}

func (p *Presenter) stringify(value any) string {
	if value == nil {
		return p.content.Common.EmptyValue
	}

	switch typed := value.(type) {
	case string:
		return p.valueOrEmpty(typed)
	case fmt.Stringer:
		return p.valueOrEmpty(typed.String())
	}

	body, err := json.Marshal(value)
	if err == nil && string(body) != "null" {
		return string(body)
	}

	return fmt.Sprint(value)
}

func presentError(err error, fallback string) string {
	if err == nil {
		return ""
	}
	if strings.TrimSpace(err.Error()) != "" {
		return err.Error()
	}
	return fallback
}
