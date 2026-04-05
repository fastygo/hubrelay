package storage_test

import (
	"path/filepath"
	"testing"
	"time"

	"sshbot/internal/buildprofile"
	"sshbot/internal/core"
	"sshbot/internal/storage"
)

func TestBboltStorePersistsAuditEntries(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "bot.db")
	store, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	if err := store.EnsureSchema(buildprofile.Current().RuntimeProfile()); err != nil {
		t.Fatalf("EnsureSchema() error = %v", err)
	}

	if err := store.RecordAudit(core.AuditEntry{
		ID:          "audit-1",
		CommandID:   "cmd-1",
		PrincipalID: "operator",
		Transport:   "http_chat",
		CommandName: "capabilities",
		Outcome:     "ok",
		Message:     "worked",
		RecordedAt:  time.Now().UTC(),
	}); err != nil {
		t.Fatalf("RecordAudit() error = %v", err)
	}

	entries, err := store.ListRecentAudit(5)
	if err != nil {
		t.Fatalf("ListRecentAudit() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(entries))
	}
	if entries[0].CommandName != "capabilities" {
		t.Fatalf("unexpected command name %q", entries[0].CommandName)
	}
}
