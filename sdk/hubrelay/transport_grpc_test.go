package hubrelay

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
)

func TestGRPCClientExecuteHealthAndStream(t *testing.T) {
	target := startTestGRPCServer(t)

	client, err := NewClient("grpc",
		WithGRPCTarget(target),
		WithPrincipal(Principal{
			ID:    "operator-local",
			Roles: []string{"operator"},
		}),
	)
	if err != nil {
		t.Fatalf("NewClient(grpc) error = %v", err)
	}
	defer client.Close()

	discovery, err := client.Discover(context.Background())
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if discovery.Service != "hubrelay" || discovery.Profile != "grpc-test-profile" {
		t.Fatalf("unexpected discovery response %+v", discovery)
	}

	health, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if health.Status != "ok" || health.Adapter != "grpc" {
		t.Fatalf("unexpected health response %+v", health)
	}

	capabilities, err := client.Capabilities(context.Background(), Principal{})
	if err != nil {
		t.Fatalf("Capabilities() error = %v", err)
	}
	if capabilities.ProfileID != "grpc-test-profile" || len(capabilities.Capabilities) != 2 {
		t.Fatalf("unexpected capabilities response %+v", capabilities)
	}

	result, err := client.Execute(context.Background(), CommandRequest{
		Command: "system-info",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Status != "ok" || result.Data["host_name"] != "grpc-devbox" {
		t.Fatalf("unexpected execute result %+v", result)
	}

	stream, err := client.ExecuteStream(context.Background(), CommandRequest{
		Command: "ask",
		Args: map[string]string{
			"prompt": "hello",
		},
	})
	if err != nil {
		t.Fatalf("ExecuteStream() error = %v", err)
	}
	defer stream.Close()

	var deltas []string
	for stream.Next() {
		deltas = append(deltas, stream.Chunk().Delta)
	}

	streamResult, err := stream.Result()
	if err != nil {
		t.Fatalf("Result() error = %v", err)
	}
	if len(deltas) != 2 || deltas[0]+deltas[1] != "hello grpc" {
		t.Fatalf("unexpected deltas %v", deltas)
	}
	if streamResult.Status != "ok" || streamResult.Data["provider"] != "grpc-test" {
		t.Fatalf("unexpected stream result %+v", streamResult)
	}
}

type testCommandServiceServer interface {
	Discover(context.Context, *grpcEmpty) (*DiscoveryResponse, error)
	Health(context.Context, *grpcEmpty) (*HealthResponse, error)
	Execute(context.Context, *grpcCommandEnvelope) (*CommandResult, error)
	ExecuteStream(*grpcCommandEnvelope, testCommandServiceExecuteStreamServer) error
}

type testCommandServiceExecuteStreamServer interface {
	Send(*grpcStreamEvent) error
	grpc.ServerStream
}

type testCommandServiceExecuteStreamServerImpl struct {
	grpc.ServerStream
}

func (s *testCommandServiceExecuteStreamServerImpl) Send(event *grpcStreamEvent) error {
	return s.ServerStream.SendMsg(event)
}

type grpcTestServer struct{}

func (grpcTestServer) Discover(context.Context, *grpcEmpty) (*DiscoveryResponse, error) {
	return &DiscoveryResponse{
		Service: "hubrelay",
		Profile: "grpc-test-profile",
		Status:  "ok",
	}, nil
}

func (grpcTestServer) Health(context.Context, *grpcEmpty) (*HealthResponse, error) {
	return &HealthResponse{
		Status:  "ok",
		Adapter: "grpc",
		Profile: "grpc-test-profile",
	}, nil
}

func (grpcTestServer) Execute(_ context.Context, request *grpcCommandEnvelope) (*CommandResult, error) {
	switch request.Name {
	case "capabilities":
		return &CommandResult{
			Status:  "ok",
			Message: "capabilities returned",
			Data: map[string]any{
				"profile_id":   "grpc-test-profile",
				"display_name": "gRPC Test Profile",
				"capabilities": []string{"adapter.grpc", "plugin.system.info"},
			},
		}, nil
	case "system-info":
		return &CommandResult{
			Status:  "ok",
			Message: "system info returned",
			Data: map[string]any{
				"host_name":       "grpc-devbox",
				"go_os":           "linux",
				"memory_total_mb": float64(1024),
			},
		}, nil
	default:
		return &CommandResult{
			Status:  "ok",
			Message: "command executed",
			Data: map[string]any{
				"command": request.Name,
			},
		}, nil
	}
}

func (grpcTestServer) ExecuteStream(request *grpcCommandEnvelope, stream testCommandServiceExecuteStreamServer) error {
	if request.Name != "ask" {
		return stream.Send(&grpcStreamEvent{
			Event: "error",
			Result: &CommandResult{
				Status:  "error",
				Message: "unsupported stream command",
			},
		})
	}

	if err := stream.Send(&grpcStreamEvent{
		Event: "chunk",
		Chunk: &StreamChunk{Delta: "hello"},
	}); err != nil {
		return err
	}
	if err := stream.Send(&grpcStreamEvent{
		Event: "chunk",
		Chunk: &StreamChunk{Delta: " grpc"},
	}); err != nil {
		return err
	}
	return stream.Send(&grpcStreamEvent{
		Event: "done",
		Result: &CommandResult{
			Status:  "ok",
			Message: "done",
			Data: map[string]any{
				"provider": "grpc-test",
			},
		},
	})
}

func startTestGRPCServer(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}

	server := grpc.NewServer(grpc.ForceServerCodec(grpcJSONCodec{}))
	registerTestCommandService(server, grpcTestServer{})

	go func() {
		_ = server.Serve(listener)
	}()

	t.Cleanup(func() {
		server.GracefulStop()
		_ = listener.Close()
	})

	return listener.Addr().String()
}

func registerTestCommandService(server *grpc.Server, srv testCommandServiceServer) {
	server.RegisterService(&grpc.ServiceDesc{
		ServiceName: "hubrelay.CommandService",
		HandlerType: (*testCommandServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{MethodName: "Discover", Handler: testDiscoverHandler},
			{MethodName: "Health", Handler: testHealthHandler},
			{MethodName: "Execute", Handler: testExecuteHandler},
		},
		Streams: []grpc.StreamDesc{
			{
				StreamName:    "ExecuteStream",
				Handler:       testExecuteStreamHandler,
				ServerStreams: true,
			},
		},
	}, srv)
}

func testDiscoverHandler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	request := new(grpcEmpty)
	if err := dec(request); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(testCommandServiceServer).Discover(ctx, request)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: grpcMethodDiscover}
	handler := func(callCtx context.Context, req any) (any, error) {
		return srv.(testCommandServiceServer).Discover(callCtx, req.(*grpcEmpty))
	}
	return interceptor(ctx, request, info, handler)
}

func testHealthHandler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	request := new(grpcEmpty)
	if err := dec(request); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(testCommandServiceServer).Health(ctx, request)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: grpcMethodHealth}
	handler := func(callCtx context.Context, req any) (any, error) {
		return srv.(testCommandServiceServer).Health(callCtx, req.(*grpcEmpty))
	}
	return interceptor(ctx, request, info, handler)
}

func testExecuteHandler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	request := new(grpcCommandEnvelope)
	if err := dec(request); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(testCommandServiceServer).Execute(ctx, request)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: grpcMethodExecute}
	handler := func(callCtx context.Context, req any) (any, error) {
		return srv.(testCommandServiceServer).Execute(callCtx, req.(*grpcCommandEnvelope))
	}
	return interceptor(ctx, request, info, handler)
}

func testExecuteStreamHandler(srv any, stream grpc.ServerStream) error {
	request := new(grpcCommandEnvelope)
	if err := stream.RecvMsg(request); err != nil {
		return err
	}
	return srv.(testCommandServiceServer).ExecuteStream(request, &testCommandServiceExecuteStreamServerImpl{
		ServerStream: stream,
	})
}
