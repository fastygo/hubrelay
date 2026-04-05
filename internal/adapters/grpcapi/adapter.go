package grpcapi

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"

	"sshbot/internal/core"
	proxymgr "sshbot/internal/proxy"
	"sshbot/pkg/contract"
)

const (
	serviceName         = "hubrelay.CommandService"
	methodDiscover      = "/hubrelay.CommandService/Discover"
	methodHealth        = "/hubrelay.CommandService/Health"
	methodExecute       = "/hubrelay.CommandService/Execute"
	methodExecuteStream = "/hubrelay.CommandService/ExecuteStream"
)

type Adapter struct {
	bindAddress string
	service     *core.Service
	server      *grpc.Server
	listener    net.Listener
}

type emptyRequest struct{}

type discoveryResponse struct {
	Service string `json:"service"`
	Profile string `json:"profile"`
	Status  string `json:"status"`
}

type healthResponse struct {
	Status  string `json:"status"`
	Adapter string `json:"adapter,omitempty"`
	Profile string `json:"profile,omitempty"`
}

type streamEvent struct {
	Event  string                  `json:"event"`
	Chunk  *contract.StreamChunk   `json:"chunk,omitempty"`
	Result *contract.CommandResult `json:"result,omitempty"`
}

type commandServiceServer interface {
	Discover(context.Context, *emptyRequest) (*discoveryResponse, error)
	Health(context.Context, *emptyRequest) (*healthResponse, error)
	Execute(context.Context, *contract.CommandEnvelope) (*contract.CommandResult, error)
	ExecuteStream(*contract.CommandEnvelope, commandServiceExecuteStreamServer) error
}

type commandServiceExecuteStreamServer interface {
	Send(*streamEvent) error
	grpc.ServerStream
}

type commandServiceExecuteStreamServerImpl struct {
	grpc.ServerStream
}

func (s *commandServiceExecuteStreamServerImpl) Send(event *streamEvent) error {
	return s.ServerStream.SendMsg(event)
}

type commandServer struct {
	service       *core.Service
	transportName string
}

type grpcStreamWriter struct {
	stream commandServiceExecuteStreamServer
	result contract.CommandResult
}

func New(bindAddress string, service *core.Service, _ *proxymgr.Manager) *Adapter {
	return &Adapter{
		bindAddress: bindAddress,
		service:     service,
	}
}

func (a *Adapter) Name() string {
	return "grpc"
}

func (a *Adapter) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", a.bindAddress)
	if err != nil {
		return err
	}
	a.listener = listener

	server := grpc.NewServer(grpc.ForceServerCodec(jsonCodec{}))
	a.server = server
	registerCommandService(server, &commandServer{
		service:       a.service,
		transportName: a.Name(),
	})

	go func() {
		<-ctx.Done()
		stopped := make(chan struct{})
		go func() {
			server.GracefulStop()
			close(stopped)
		}()

		select {
		case <-stopped:
		case <-time.After(5 * time.Second):
			server.Stop()
		}

		if a.listener != nil {
			_ = a.listener.Close()
		}
	}()

	err = server.Serve(listener)
	if err != nil && ctx.Err() == nil {
		return err
	}
	return nil
}

func (s *commandServer) Discover(_ context.Context, _ *emptyRequest) (*discoveryResponse, error) {
	return &discoveryResponse{
		Service: "hubrelay",
		Profile: s.service.Profile().ID,
		Status:  "ok",
	}, nil
}

func (s *commandServer) Health(_ context.Context, _ *emptyRequest) (*healthResponse, error) {
	return &healthResponse{
		Status:  "ok",
		Adapter: s.transportName,
		Profile: s.service.Profile().ID,
	}, nil
}

func (s *commandServer) Execute(ctx context.Context, envelope *contract.CommandEnvelope) (*contract.CommandResult, error) {
	normalized := s.normalizeEnvelope(envelope)
	result, err := s.service.Execute(ctx, normalized)
	if err != nil && result.Message == "" {
		result.Message = err.Error()
	}
	return &result, nil
}

func (s *commandServer) ExecuteStream(envelope *contract.CommandEnvelope, stream commandServiceExecuteStreamServer) error {
	writer := &grpcStreamWriter{stream: stream}
	return s.service.ExecuteStream(stream.Context(), s.normalizeEnvelope(envelope), writer)
}

func (s *commandServer) normalizeEnvelope(envelope *contract.CommandEnvelope) contract.CommandEnvelope {
	now := time.Now().UTC()
	if envelope == nil {
		envelope = &contract.CommandEnvelope{}
	}

	normalized := *envelope
	if normalized.ID == "" {
		normalized.ID = fmt.Sprintf("%s-%d", s.transportName, now.UnixNano())
	}
	normalized.Transport = s.transportName
	if normalized.RequestedAt.IsZero() {
		normalized.RequestedAt = now
	}
	if normalized.Principal.Display == "" {
		normalized.Principal.Display = normalized.Principal.ID
	}
	normalized.Principal.Transport = s.transportName
	return normalized
}

func (w *grpcStreamWriter) WriteChunk(chunk contract.StreamChunk) error {
	chunkCopy := chunk
	return w.stream.Send(&streamEvent{
		Event: "chunk",
		Chunk: &chunkCopy,
	})
}

func (w *grpcStreamWriter) SetResult(result contract.CommandResult) {
	w.result = contract.CloneCommandResult(result)
}

func (w *grpcStreamWriter) Flush() error {
	if w.result.Status == "" {
		w.result.Status = "ok"
	}

	eventName := "done"
	if w.result.Status == "error" {
		eventName = "error"
	}

	result := contract.CloneCommandResult(w.result)
	return w.stream.Send(&streamEvent{
		Event:  eventName,
		Result: &result,
	})
}

func registerCommandService(server *grpc.Server, srv commandServiceServer) {
	server.RegisterService(&grpc.ServiceDesc{
		ServiceName: serviceName,
		HandlerType: (*commandServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "Discover",
				Handler:    discoverHandler,
			},
			{
				MethodName: "Health",
				Handler:    healthHandler,
			},
			{
				MethodName: "Execute",
				Handler:    executeHandler,
			},
		},
		Streams: []grpc.StreamDesc{
			{
				StreamName:    "ExecuteStream",
				Handler:       executeStreamHandler,
				ServerStreams: true,
			},
		},
	}, srv)
}

func discoverHandler(
	srv any,
	ctx context.Context,
	dec func(any) error,
	interceptor grpc.UnaryServerInterceptor,
) (any, error) {
	request := new(emptyRequest)
	if err := dec(request); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(commandServiceServer).Discover(ctx, request)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: methodDiscover,
	}
	handler := func(callCtx context.Context, req any) (any, error) {
		return srv.(commandServiceServer).Discover(callCtx, req.(*emptyRequest))
	}
	return interceptor(ctx, request, info, handler)
}

func healthHandler(
	srv any,
	ctx context.Context,
	dec func(any) error,
	interceptor grpc.UnaryServerInterceptor,
) (any, error) {
	request := new(emptyRequest)
	if err := dec(request); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(commandServiceServer).Health(ctx, request)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: methodHealth,
	}
	handler := func(callCtx context.Context, req any) (any, error) {
		return srv.(commandServiceServer).Health(callCtx, req.(*emptyRequest))
	}
	return interceptor(ctx, request, info, handler)
}

func executeHandler(
	srv any,
	ctx context.Context,
	dec func(any) error,
	interceptor grpc.UnaryServerInterceptor,
) (any, error) {
	request := new(contract.CommandEnvelope)
	if err := dec(request); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(commandServiceServer).Execute(ctx, request)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: methodExecute,
	}
	handler := func(callCtx context.Context, req any) (any, error) {
		return srv.(commandServiceServer).Execute(callCtx, req.(*contract.CommandEnvelope))
	}
	return interceptor(ctx, request, info, handler)
}

func executeStreamHandler(srv any, stream grpc.ServerStream) error {
	request := new(contract.CommandEnvelope)
	if err := stream.RecvMsg(request); err != nil {
		return err
	}
	return srv.(commandServiceServer).ExecuteStream(request, &commandServiceExecuteStreamServerImpl{
		ServerStream: stream,
	})
}
