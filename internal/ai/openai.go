package ai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"

	"sshbot/internal/buildprofile"
	"sshbot/internal/outbound"
	proxymgr "sshbot/internal/proxy"
)

type OpenAICompatibleProvider struct {
	name                 string
	model                string
	apiKey               string
	baseURL              string
	apiMode              string
	timeout              time.Duration
	policy               outbound.Policy
	privateEgress        buildprofile.PrivateEgressConfig
	privateEgressChecker outbound.PrivateEgressChecker
	proxyManager         *proxymgr.Manager
}

func NewOpenAICompatibleProvider(name, apiKey, baseURL, model, apiMode string, policy outbound.Policy, privateEgress buildprofile.PrivateEgressConfig, privateEgressChecker outbound.PrivateEgressChecker, proxyManager *proxymgr.Manager) (*OpenAICompatibleProvider, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("ai api key is required")
	}
	if strings.TrimSpace(model) == "" {
		return nil, errors.New("ai model is required")
	}

	normalizedBaseURL := strings.TrimSpace(baseURL)
	if normalizedBaseURL != "" && !strings.HasSuffix(normalizedBaseURL, "/") {
		normalizedBaseURL += "/"
	}

	return &OpenAICompatibleProvider{
		name:                 strings.TrimSpace(strings.ToLower(name)),
		model:                strings.TrimSpace(model),
		apiKey:               apiKey,
		baseURL:              normalizedBaseURL,
		apiMode:              normalizeAPIMode(apiMode),
		timeout:              60 * time.Second,
		policy:               policy,
		privateEgress:        privateEgress,
		privateEgressChecker: privateEgressChecker,
		proxyManager:         proxyManager,
	}, nil
}

func (p *OpenAICompatibleProvider) Name() string {
	return p.name
}

func (p *OpenAICompatibleProvider) Ask(ctx context.Context, request AskRequest) (AskResponse, error) {
	model := strings.TrimSpace(request.Model)
	if model == "" {
		model = p.model
	}
	if p.privateEgress.Required {
		if err := p.requirePrivateEgress(ctx); err != nil {
			return AskResponse{}, err
		}
	}

	proxySessionID := strings.TrimSpace(request.ProxySessionID)
	proxyAddress, err := p.policy.ResolveProxyAddress(proxySessionID, outbound.ProxyLeaseResolver{Manager: p.proxyManager})
	if err != nil {
		return AskResponse{}, err
	}

	response, err := p.askOnce(ctx, request, model, proxyAddress)
	if err != nil && proxyAddress != "" && proxySessionID != "" && p.proxyManager != nil && isTransportError(err) {
		_, _ = p.proxyManager.ReportFailure(proxySessionID, proxyAddress, err.Error())
		lease, leaseErr := p.proxyManager.AcquireLease(proxySessionID)
		if leaseErr == nil && lease.Address != "" && lease.Address != proxyAddress {
			return p.askOnce(ctx, request, model, lease.Address)
		}
	}
	return response, err
}

func (p *OpenAICompatibleProvider) AskStream(ctx context.Context, request AskRequest, callback StreamCallback) (AskResponse, error) {
	model := strings.TrimSpace(request.Model)
	if model == "" {
		model = p.model
	}
	if p.privateEgress.Required {
		if err := p.requirePrivateEgress(ctx); err != nil {
			return AskResponse{}, err
		}
	}

	proxySessionID := strings.TrimSpace(request.ProxySessionID)
	proxyAddress, err := p.policy.ResolveProxyAddress(proxySessionID, outbound.ProxyLeaseResolver{Manager: p.proxyManager})
	if err != nil {
		return AskResponse{}, err
	}

	emitted := false
	wrappedCallback := func(chunk AskStreamChunk) error {
		if chunk.Delta != "" {
			emitted = true
		}
		return callback(chunk)
	}

	response, err := p.askStreamOnce(ctx, request, model, proxyAddress, wrappedCallback)
	if err != nil && proxyAddress != "" && proxySessionID != "" && p.proxyManager != nil && isTransportError(err) && !emitted {
		_, _ = p.proxyManager.ReportFailure(proxySessionID, proxyAddress, err.Error())
		lease, leaseErr := p.proxyManager.AcquireLease(proxySessionID)
		if leaseErr == nil && lease.Address != "" && lease.Address != proxyAddress {
			return p.askStreamOnce(ctx, request, model, lease.Address, wrappedCallback)
		}
	}
	return response, err
}

func (p *OpenAICompatibleProvider) requirePrivateEgress(ctx context.Context) error {
	if p.privateEgressChecker == nil {
		if p.privateEgress.FailClosed {
			return errors.New("private egress checker is not configured")
		}
		return nil
	}
	if err := p.privateEgressChecker.Check(ctx, p.privateEgress.Interface, p.privateEgress.TestHost); err != nil {
		if p.privateEgress.FailClosed {
			return err
		}
		log.Printf("[ai] private egress validation failed but continuing: %v", err)
	}
	return nil
}

func (p *OpenAICompatibleProvider) askOnce(ctx context.Context, request AskRequest, model, proxyAddress string) (AskResponse, error) {
	log.Printf("[ai] askOnce name=%s model=%s base_url=%s api_mode=%s proxy=%q",
		p.name, model, p.baseURL, p.apiMode, proxyAddress)

	client, err := p.client(proxyAddress)
	if err != nil {
		log.Printf("[ai] client creation failed: %v", err)
		return AskResponse{}, err
	}

	if p.apiMode == "responses" {
		return p.askViaResponses(ctx, client, request, model)
	}
	return p.askViaChatCompletions(ctx, client, request, model)
}

func (p *OpenAICompatibleProvider) askStreamOnce(ctx context.Context, request AskRequest, model, proxyAddress string, callback StreamCallback) (AskResponse, error) {
	log.Printf("[ai] askStreamOnce name=%s model=%s base_url=%s api_mode=%s proxy=%q",
		p.name, model, p.baseURL, p.apiMode, proxyAddress)

	client, err := p.client(proxyAddress)
	if err != nil {
		log.Printf("[ai] streaming client creation failed: %v", err)
		return AskResponse{}, err
	}

	if p.apiMode == "responses" {
		return p.askViaResponsesStream(ctx, client, request, model, callback)
	}
	return p.askViaChatCompletionsStream(ctx, client, request, model, callback)
}

func (p *OpenAICompatibleProvider) askViaResponses(ctx context.Context, client openai.Client, request AskRequest, model string) (AskResponse, error) {
	params := responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(request.Prompt),
		},
		Model: shared.ResponsesModel(model),
		Store: openai.Bool(false),
	}
	if instruction := strings.TrimSpace(request.System); instruction != "" {
		params.Instructions = openai.String(instruction)
	}
	if userID := strings.TrimSpace(request.UserID); userID != "" {
		params.SafetyIdentifier = openai.String(userID)
	}

	response, err := client.Responses.New(ctx, params)
	if err != nil {
		return AskResponse{}, err
	}

	return AskResponse{
		Provider:   p.name,
		Model:      model,
		Answer:     response.OutputText(),
		ResponseID: response.ID,
	}, nil
}

func (p *OpenAICompatibleProvider) askViaChatCompletions(ctx context.Context, client openai.Client, request AskRequest, model string) (AskResponse, error) {
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, 2)
	if instruction := strings.TrimSpace(request.System); instruction != "" {
		messages = append(messages, openai.SystemMessage(instruction))
	}
	messages = append(messages, openai.UserMessage(request.Prompt))

	params := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    openai.ChatModel(model),
	}
	if userID := strings.TrimSpace(request.UserID); userID != "" {
		params.User = openai.String(userID)
	}

	completion, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return AskResponse{}, err
	}

	answer := ""
	if len(completion.Choices) > 0 {
		answer = strings.TrimSpace(completion.Choices[0].Message.Content)
	}

	return AskResponse{
		Provider:   p.name,
		Model:      model,
		Answer:     answer,
		ResponseID: completion.ID,
	}, nil
}

func (p *OpenAICompatibleProvider) askViaResponsesStream(ctx context.Context, client openai.Client, request AskRequest, model string, callback StreamCallback) (AskResponse, error) {
	params := responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(request.Prompt),
		},
		Model: shared.ResponsesModel(model),
		Store: openai.Bool(false),
	}
	if instruction := strings.TrimSpace(request.System); instruction != "" {
		params.Instructions = openai.String(instruction)
	}
	if userID := strings.TrimSpace(request.UserID); userID != "" {
		params.SafetyIdentifier = openai.String(userID)
	}

	stream := client.Responses.NewStreaming(ctx, params)
	defer stream.Close()

	final := AskResponse{
		Provider: p.name,
		Model:    model,
	}
	for stream.Next() {
		event := stream.Current()
		switch event.Type {
		case "response.output_text.delta":
			delta := event.AsResponseOutputTextDelta()
			if delta.Delta == "" {
				continue
			}
			if err := callback(AskStreamChunk{Delta: delta.Delta}); err != nil {
				return AskResponse{}, err
			}
		case "response.completed":
			completed := event.AsResponseCompleted()
			final.Model = completed.Response.Model
			final.Answer = completed.Response.OutputText()
			final.ResponseID = completed.Response.ID
		case "error":
			failure := event.AsError()
			if failure.Message != "" {
				return AskResponse{}, errors.New(failure.Message)
			}
		}
	}
	if err := stream.Err(); err != nil {
		return AskResponse{}, err
	}
	return final, nil
}

func (p *OpenAICompatibleProvider) askViaChatCompletionsStream(ctx context.Context, client openai.Client, request AskRequest, model string, callback StreamCallback) (AskResponse, error) {
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, 2)
	if instruction := strings.TrimSpace(request.System); instruction != "" {
		messages = append(messages, openai.SystemMessage(instruction))
	}
	messages = append(messages, openai.UserMessage(request.Prompt))

	params := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    openai.ChatModel(model),
	}
	if userID := strings.TrimSpace(request.UserID); userID != "" {
		params.User = openai.String(userID)
	}

	stream := client.Chat.Completions.NewStreaming(ctx, params)
	defer stream.Close()

	accumulator := openai.ChatCompletionAccumulator{}
	for stream.Next() {
		chunk := stream.Current()
		_ = accumulator.AddChunk(chunk)
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				if err := callback(AskStreamChunk{
					Delta:        choice.Delta.Content,
					FinishReason: string(choice.FinishReason),
				}); err != nil {
					return AskResponse{}, err
				}
			}
		}
	}
	if err := stream.Err(); err != nil {
		return AskResponse{}, err
	}

	answer := ""
	responseID := accumulator.ID
	if len(accumulator.Choices) > 0 {
		answer = strings.TrimSpace(accumulator.Choices[0].Message.Content)
	}

	return AskResponse{
		Provider:   p.name,
		Model:      model,
		Answer:     answer,
		ResponseID: responseID,
	}, nil
}

func (p *OpenAICompatibleProvider) client(proxyAddress string) (openai.Client, error) {
	if p.policy.RequireProxy && strings.TrimSpace(proxyAddress) == "" {
		return openai.Client{}, errors.New("proxy session is required by outbound policy")
	}
	options := []option.RequestOption{
		option.WithAPIKey(p.apiKey),
	}
	if p.baseURL != "" {
		options = append(options, option.WithBaseURL(p.baseURL))
	}
	if strings.TrimSpace(proxyAddress) != "" {
		httpClient, err := proxymgr.NewHTTPClient(proxyAddress, p.timeout)
		if err != nil {
			return openai.Client{}, err
		}
		options = append(options, option.WithHTTPClient(httpClient))
	}
	return openai.NewClient(options...), nil
}

func normalizeAPIMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "responses":
		return "responses"
	default:
		return "chat_completions"
	}
}

func NewOpenAIProxyProber(apiKey, baseURL string, timeout time.Duration) proxymgr.ProbeFunc {
	endpoint := modelsEndpoint(baseURL)
	return func(ctx context.Context, proxyAddress string) (time.Duration, error) {
		client, err := proxymgr.NewHTTPClient(proxyAddress, timeout)
		if err != nil {
			return 0, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return 0, err
		}
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))

		started := time.Now()
		response, err := client.Do(req)
		if err != nil {
			return 0, err
		}
		defer response.Body.Close()
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1024))
		if response.StatusCode >= http.StatusInternalServerError {
			return 0, fmt.Errorf("provider probe failed with status %d", response.StatusCode)
		}
		return time.Since(started), nil
	}
}

func modelsEndpoint(baseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmed == "" {
		trimmed = "https://api.openai.com/v1"
	}
	if strings.HasSuffix(trimmed, "/models") {
		return trimmed
	}
	return trimmed + "/models"
}

func isTransportError(err error) bool {
	var urlErr *url.Error
	var netErr net.Error
	return errors.Is(err, context.DeadlineExceeded) ||
		errors.As(err, &urlErr) ||
		errors.As(err, &netErr)
}
