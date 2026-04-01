package source

import (
	"context"
	"fmt"
	"path"
	"strings"

	"hubrelay-dashboard/fixtures"
)

type Fixture struct {
	health       HealthData
	capabilities CapabilitiesData
	ask          askFixtureData
	egress       EgressData
	audit        auditFixtureData
}

type askFixtureData struct {
	Result CommandResult    `json:"result"`
	Stream askStreamFixture `json:"stream"`
}

type askStreamFixture struct {
	Chunks []StreamChunk `json:"chunks"`
	Result CommandResult `json:"result"`
}

type auditFixtureData struct {
	Entries []AuditEntry `json:"entries"`
}

func NewFixture(locale string) (*Fixture, error) {
	locale = strings.TrimSpace(locale)
	if locale == "" {
		return nil, fmt.Errorf("locale must not be empty")
	}

	var out Fixture
	loaders := []struct {
		name   string
		target any
	}{
		{name: "health.json", target: &out.health},
		{name: "capabilities.json", target: &out.capabilities},
		{name: "ask.json", target: &out.ask},
		{name: "egress.json", target: &out.egress},
		{name: "audit.json", target: &out.audit},
	}

	for _, loader := range loaders {
		if err := fixtures.Decode(path.Join(locale, "mocks", loader.name), loader.target); err != nil {
			return nil, err
		}
	}

	return &out, nil
}

func (s *Fixture) Health(_ context.Context) (HealthData, error) {
	return s.health, nil
}

func (s *Fixture) Capabilities(_ context.Context) (CapabilitiesData, error) {
	return s.capabilities, nil
}

func (s *Fixture) Ask(_ context.Context, _, _ string) (CommandResult, error) {
	return cloneCommandResult(s.ask.Result), nil
}

func (s *Fixture) AskStream(_ context.Context, _, _ string) (AskStream, error) {
	return &fixtureAskStream{
		chunks: append([]StreamChunk(nil), s.ask.Stream.Chunks...),
		result: cloneCommandResult(s.ask.Stream.Result),
	}, nil
}

func (s *Fixture) Egress(_ context.Context) (EgressData, error) {
	return EgressData{
		Gateways: append([]GatewayData(nil), s.egress.Gateways...),
	}, nil
}

func (s *Fixture) Audit(_ context.Context, limit int) ([]AuditEntry, error) {
	entries := append([]AuditEntry(nil), s.audit.Entries...)
	if limit > 0 && limit < len(entries) {
		return entries[:limit], nil
	}
	return entries, nil
}

type fixtureAskStream struct {
	chunks  []StreamChunk
	result  CommandResult
	index   int
	current StreamChunk
}

func (s *fixtureAskStream) Next() bool {
	if s.index >= len(s.chunks) {
		return false
	}

	s.current = s.chunks[s.index]
	s.index++
	return true
}

func (s *fixtureAskStream) Chunk() StreamChunk {
	return s.current
}

func (s *fixtureAskStream) Result() (CommandResult, error) {
	return cloneCommandResult(s.result), nil
}

func (s *fixtureAskStream) Close() error {
	return nil
}

func cloneCommandResult(input CommandResult) CommandResult {
	input.Data = cloneMap(input.Data)
	return input
}
