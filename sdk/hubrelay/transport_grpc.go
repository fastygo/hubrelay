package hubrelay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
)

const (
	grpcCodecName           = "json"
	grpcMethodDiscover      = "/hubrelay.CommandService/Discover"
	grpcMethodHealth        = "/hubrelay.CommandService/Health"
	grpcMethodExecute       = "/hubrelay.CommandService/Execute"
	grpcMethodExecuteStream = "/hubrelay.CommandService/ExecuteStream"
)

func init() {
	encoding.RegisterCodec(grpcJSONCodec{})
}

type grpcJSONCodec struct{}

type grpcTransport struct {
	conn             *grpc.ClientConn
	defaultPrincipal Principal
}

type grpcEmpty struct{}

type grpcPrincipal struct {
	Version   string            `json:"version,omitempty"`
	ID        string            `json:"id"`
	Display   string            `json:"display"`
	Transport string            `json:"transport"`
	Scope     string            `json:"scope,omitempty"`
	Roles     []string          `json:"roles"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type grpcCommandEnvelope struct {
	Version     string            `json:"version,omitempty"`
	ID          string            `json:"id"`
	Transport   string            `json:"transport"`
	Name        string            `json:"name"`
	Args        map[string]string `json:"args,omitempty"`
	RawText     string            `json:"raw_text,omitempty"`
	Principal   grpcPrincipal     `json:"principal"`
	RequestedAt time.Time         `json:"requested_at"`
}

type grpcStreamEvent struct {
	Event  string         `json:"event"`
	Chunk  *StreamChunk   `json:"chunk,omitempty"`
	Result *CommandResult `json:"result,omitempty"`
}

type grpcResultStream struct {
	stream grpc.ClientStream
	chunk  StreamChunk
	result CommandResult
	err    error
	done   bool
}

func newGRPCTransport(target string, cfg clientConfig) (*grpcTransport, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return nil, errors.New("grpc target is required")
	}

	dialCtx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()

	conn, err := grpc.DialContext(
		dialCtx,
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(
			grpc.ForceCodec(grpcJSONCodec{}),
			grpc.CallContentSubtype(grpcCodecName),
		),
	)
	if err != nil {
		return nil, err
	}

	return &grpcTransport{
		conn:             conn,
		defaultPrincipal: cfg.principal,
	}, nil
}

func (grpcJSONCodec) Marshal(value any) ([]byte, error) {
	return json.Marshal(value)
}

func (grpcJSONCodec) Unmarshal(data []byte, value any) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, value)
}

func (grpcJSONCodec) Name() string {
	return grpcCodecName
}

func (t *grpcTransport) Discover(ctx context.Context) (DiscoveryResponse, error) {
	var response DiscoveryResponse
	if err := t.conn.Invoke(ctx, grpcMethodDiscover, &grpcEmpty{}, &response); err != nil {
		return DiscoveryResponse{}, err
	}
	return response, nil
}

func (t *grpcTransport) Health(ctx context.Context) (HealthResponse, error) {
	var response HealthResponse
	if err := t.conn.Invoke(ctx, grpcMethodHealth, &grpcEmpty{}, &response); err != nil {
		return HealthResponse{}, err
	}
	return response, nil
}

func (t *grpcTransport) Capabilities(ctx context.Context, principal Principal) (CapabilitiesResponse, error) {
	result, err := t.Execute(ctx, CommandRequest{
		Principal: principal,
		Command:   "capabilities",
	})
	if err != nil {
		return CapabilitiesResponse{}, err
	}
	if result.Status == "error" {
		return CapabilitiesResponse{}, errors.New(result.Message)
	}

	var response CapabilitiesResponse
	if err := decodeResultData(result.Data, &response); err != nil {
		return CapabilitiesResponse{}, err
	}
	return response, nil
}

func (t *grpcTransport) Execute(ctx context.Context, req CommandRequest) (CommandResult, error) {
	var result CommandResult
	if err := t.conn.Invoke(ctx, grpcMethodExecute, t.commandEnvelope(req), &result); err != nil {
		return CommandResult{}, err
	}
	return result, nil
}

func (t *grpcTransport) ExecuteStream(ctx context.Context, req CommandRequest) (ResultStream, error) {
	stream, err := t.conn.NewStream(ctx, &grpc.StreamDesc{
		ServerStreams: true,
	}, grpcMethodExecuteStream)
	if err != nil {
		return nil, err
	}

	if err := stream.SendMsg(t.commandEnvelope(req)); err != nil {
		_ = stream.CloseSend()
		return nil, err
	}
	if err := stream.CloseSend(); err != nil {
		return nil, err
	}

	return &grpcResultStream{stream: stream}, nil
}

func (t *grpcTransport) EgressStatus(ctx context.Context, principal Principal) (EgressStatusResponse, error) {
	result, err := t.Execute(ctx, CommandRequest{
		Principal: principal,
		Command:   "egress-status",
	})
	if err != nil {
		return EgressStatusResponse{}, err
	}
	if result.Status == "error" {
		return EgressStatusResponse{}, errors.New(result.Message)
	}

	var response EgressStatusResponse
	if err := decodeResultData(result.Data, &response); err != nil {
		return EgressStatusResponse{}, err
	}
	return response, nil
}

func (t *grpcTransport) Close() error {
	if t == nil || t.conn == nil {
		return nil
	}
	return t.conn.Close()
}

func (t *grpcTransport) commandEnvelope(req CommandRequest) *grpcCommandEnvelope {
	principal := t.principalOrDefault(req.Principal)
	now := time.Now().UTC()
	return &grpcCommandEnvelope{
		ID:        fmt.Sprintf("grpc-%d", now.UnixNano()),
		Transport: "grpc",
		Name:      strings.ToLower(strings.TrimSpace(req.Command)),
		Args:      cloneStringMap(req.Args),
		RawText:   strings.TrimSpace(req.Command),
		Principal: grpcPrincipal{
			ID:        principal.ID,
			Display:   firstString(principal.Display, principal.ID),
			Transport: "grpc",
			Roles:     append([]string(nil), principal.Roles...),
		},
		RequestedAt: now,
	}
}

func (t *grpcTransport) principalOrDefault(principal Principal) Principal {
	if strings.TrimSpace(principal.ID) == "" {
		principal = t.defaultPrincipal
	}
	if strings.TrimSpace(principal.Display) == "" {
		principal.Display = principal.ID
	}
	return principal
}

func (s *grpcResultStream) Next() bool {
	if s == nil || s.done {
		return false
	}

	for {
		var event grpcStreamEvent
		err := s.stream.RecvMsg(&event)
		if errors.Is(err, io.EOF) {
			s.done = true
			return false
		}
		if err != nil {
			s.err = err
			s.done = true
			return false
		}

		switch event.Event {
		case "chunk":
			if event.Chunk == nil {
				continue
			}
			s.chunk = *event.Chunk
			return true
		case "done":
			if event.Result != nil {
				s.result = *event.Result
			}
			s.done = true
			return false
		case "error":
			if event.Result != nil {
				s.result = *event.Result
			}
			if s.result.Message != "" {
				s.err = errors.New(s.result.Message)
			} else {
				s.err = errors.New("hubrelay stream returned error")
			}
			s.done = true
			return false
		default:
			continue
		}
	}
}

func (s *grpcResultStream) Chunk() StreamChunk {
	return s.chunk
}

func (s *grpcResultStream) Result() (CommandResult, error) {
	return s.result, s.err
}

func (s *grpcResultStream) Close() error {
	if s == nil || s.stream == nil {
		return nil
	}
	return s.stream.CloseSend()
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func firstString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
