package core_test

import (
	"context"
	"testing"

	"sshbot/internal/core"
)

type streamingOnlyPlugin struct{}

func (streamingOnlyPlugin) Descriptor() core.PluginDescriptor {
	return core.PluginDescriptor{Name: "streaming"}
}

func (streamingOnlyPlugin) Execute(context.Context, core.CommandContext, core.CommandEnvelope) (core.CommandResult, error) {
	return core.CommandResult{}, nil
}

func (streamingOnlyPlugin) ExecuteStream(context.Context, core.CommandContext, core.CommandEnvelope, core.StreamWriter) error {
	return nil
}

func TestBufferedStreamWriterFlushAddsCombinedAnswer(t *testing.T) {
	writer := &core.BufferedStreamWriter{}

	if err := writer.WriteChunk(core.StreamChunk{Delta: "Hello"}); err != nil {
		t.Fatalf("WriteChunk() error = %v", err)
	}
	if err := writer.WriteChunk(core.StreamChunk{Delta: " world"}); err != nil {
		t.Fatalf("WriteChunk() error = %v", err)
	}
	writer.SetResult(core.CommandResult{
		Message: "ai answer generated",
		Data: map[string]any{
			"provider": "openai",
		},
	})

	if err := writer.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	result := writer.Result()
	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %q", result.Status)
	}
	if got := result.Data["answer"]; got != "Hello world" {
		t.Fatalf("expected combined answer, got %v", got)
	}
	if got := result.Data["provider"]; got != "openai" {
		t.Fatalf("expected provider metadata to survive, got %v", got)
	}
}

func TestBufferedStreamWriterFlushDoesNotOverwriteExplicitAnswer(t *testing.T) {
	writer := &core.BufferedStreamWriter{}

	_ = writer.WriteChunk(core.StreamChunk{Delta: "partial"})
	writer.SetResult(core.CommandResult{
		Status:  "ok",
		Message: "done",
		Data: map[string]any{
			"answer": "final answer",
		},
	})

	if err := writer.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	result := writer.Result()
	if got := result.Data["answer"]; got != "final answer" {
		t.Fatalf("expected explicit answer to win, got %v", got)
	}
}

func TestBufferedStreamWriterFlushIsIdempotent(t *testing.T) {
	writer := &core.BufferedStreamWriter{}
	_ = writer.WriteChunk(core.StreamChunk{Delta: "abc"})
	writer.SetResult(core.CommandResult{Message: "done"})

	if err := writer.Flush(); err != nil {
		t.Fatalf("first Flush() error = %v", err)
	}
	first := writer.Result()

	if err := writer.Flush(); err != nil {
		t.Fatalf("second Flush() error = %v", err)
	}
	second := writer.Result()

	if first.Status != second.Status || first.Message != second.Message {
		t.Fatalf("expected identical result after second flush")
	}
	if first.Data["answer"] != second.Data["answer"] {
		t.Fatalf("expected identical answer after second flush")
	}
}

func TestStreamingPluginSatisfiesPlugin(t *testing.T) {
	var plugin any = streamingOnlyPlugin{}
	if _, ok := plugin.(core.StreamingPlugin); !ok {
		t.Fatal("expected streamingOnlyPlugin to satisfy core.StreamingPlugin")
	}
}
