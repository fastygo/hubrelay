package source

import (
	"context"
	"strings"
	"testing"
)

func TestFixtureSource(t *testing.T) {
	src, err := NewFixture("en")
	if err != nil {
		t.Fatalf("NewFixture(en) error = %v", err)
	}

	health, err := src.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if health.Discovery.Service != "hubrelay" {
		t.Fatalf("expected hubrelay service, got %q", health.Discovery.Service)
	}

	result, err := src.Ask(context.Background(), "hello", "")
	if err != nil {
		t.Fatalf("Ask() error = %v", err)
	}
	if result.Status != "ok" {
		t.Fatalf("expected ok ask status, got %q", result.Status)
	}

	stream, err := src.AskStream(context.Background(), "hello", "")
	if err != nil {
		t.Fatalf("AskStream() error = %v", err)
	}
	defer stream.Close()

	var chunks []string
	for stream.Next() {
		chunks = append(chunks, stream.Chunk().Delta)
	}

	if got := strings.Join(chunks, ""); got != "This is a fixture stream." {
		t.Fatalf("unexpected stream output %q", got)
	}

	done, err := stream.Result()
	if err != nil {
		t.Fatalf("Result() error = %v", err)
	}
	if done.Message != "stream completed" {
		t.Fatalf("expected stream completed message, got %q", done.Message)
	}
}

func TestRussianFixtureSource(t *testing.T) {
	src, err := NewFixture("ru")
	if err != nil {
		t.Fatalf("NewFixture(ru) error = %v", err)
	}

	result, err := src.Ask(context.Background(), "privet", "")
	if err != nil {
		t.Fatalf("Ask() error = %v", err)
	}
	if result.Message != "команда ask завершена" {
		t.Fatalf("unexpected russian ask message %q", result.Message)
	}
}

func TestSpanishFixtureSource(t *testing.T) {
	src, err := NewFixture("es")
	if err != nil {
		t.Fatalf("NewFixture(es) error = %v", err)
	}

	result, err := src.Ask(context.Background(), "hola", "")
	if err != nil {
		t.Fatalf("Ask() error = %v", err)
	}
	if result.Message != "comando ask completado" {
		t.Fatalf("unexpected spanish ask message %q", result.Message)
	}
}
