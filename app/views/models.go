package views

import (
	"time"

	"sshbot/sdk/hubrelay"
)

type KV struct {
	Label string
	Value string
}

type HealthPageData struct {
	Discovery hubrelay.DiscoveryResponse
	Health    hubrelay.HealthResponse
	Error     string
}

type CapabilitiesPageData struct {
	Capabilities hubrelay.CapabilitiesResponse
	Error        string
}

type AskPageData struct {
	Prompt string
	Model  string
	Result *hubrelay.CommandResult
	Error  string
}

type EgressPageData struct {
	Gateways []hubrelay.GatewayStatus
	Error    string
}

type AuditEntryView struct {
	Command   string
	Principal string
	Transport string
	Outcome   string
	Message   string
	At        time.Time
}

type AuditPageData struct {
	Entries []AuditEntryView
	Error   string
}
