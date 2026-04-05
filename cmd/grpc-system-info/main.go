package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"

	"sshbot/pkg/contract"
)

const executeMethod = "/hubrelay.CommandService/Execute"

func main() {
	target := flag.String("target", "127.0.0.1:5501", "gRPC target in host:port format")
	timeout := flag.Duration("timeout", 10*time.Second, "request timeout")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		*target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(
			grpc.ForceCodec(jsonCodec{}),
			grpc.CallContentSubtype(codecName),
		),
	)
	if err != nil {
		fatalf("dial grpc target: %v", err)
	}
	defer conn.Close()

	request := &contract.CommandEnvelope{
		ID:        fmt.Sprintf("grpc-system-info-%d", time.Now().UTC().UnixNano()),
		Transport: "grpc",
		Name:      "system-info",
		Principal: contract.Principal{
			ID:        "operator-local",
			Display:   "operator-local",
			Transport: "grpc",
			Roles:     []string{"operator"},
		},
		RequestedAt: time.Now().UTC(),
	}

	var result contract.CommandResult
	if err := conn.Invoke(ctx, executeMethod, request, &result); err != nil {
		fatalf("execute system-info: %v", err)
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fatalf("marshal result: %v", err)
	}
	fmt.Println(string(output))
	if result.Status == "error" {
		os.Exit(1)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

const codecName = "json"

func init() {
	encoding.RegisterCodec(jsonCodec{})
}

type jsonCodec struct{}

func (jsonCodec) Marshal(value any) ([]byte, error) {
	return json.Marshal(value)
}

func (jsonCodec) Unmarshal(data []byte, value any) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, value)
}

func (jsonCodec) Name() string {
	return codecName
}
