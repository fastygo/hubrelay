package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	bolt "go.etcd.io/bbolt"

	"sshbot/internal/buildprofile"
	"sshbot/internal/core"
)

type BboltStore struct {
	db *bolt.DB
}

func Open(path string) (*BboltStore, error) {
	if path == "" {
		return nil, errors.New("database path cannot be empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return nil, err
	}

	return &BboltStore{db: db}, nil
}

func (s *BboltStore) EnsureSchema(_ buildprofile.Profile) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range []string{
			bucketMeta,
			bucketPrincipals,
			bucketSessions,
			bucketAudit,
			bucketPluginState,
		} {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return err
			}
		}

		meta := tx.Bucket([]byte(bucketMeta))
		if err := meta.Put([]byte(keySchemaVersion), []byte(currentSchema)); err != nil {
			return err
		}

		return nil
	})
}

func (s *BboltStore) UpsertPrincipal(principal core.Principal) error {
	return s.putJSON(bucketPrincipals, principal.ID, principal)
}

func (s *BboltStore) SaveSession(session core.SessionState) error {
	return s.putJSON(bucketSessions, session.ID, session)
}

func (s *BboltStore) RecordAudit(entry core.AuditEntry) error {
	return s.putJSON(bucketAudit, entry.ID, entry)
}

func (s *BboltStore) ListRecentAudit(limit int) ([]core.AuditEntry, error) {
	if limit <= 0 {
		limit = 20
	}

	entries := make([]core.AuditEntry, 0, limit)
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketAudit))
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(_, value []byte) error {
			var entry core.AuditEntry
			if err := json.Unmarshal(value, &entry); err != nil {
				return err
			}
			entries = append(entries, entry)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].RecordedAt.After(entries[j].RecordedAt)
	})
	if len(entries) > limit {
		entries = entries[:limit]
	}
	return entries, nil
}

func (s *BboltStore) Close() error {
	return s.db.Close()
}

func (s *BboltStore) putJSON(bucketName, key string, value any) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty for bucket %s", bucketName)
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket %s does not exist", bucketName)
		}
		return bucket.Put([]byte(key), payload)
	})
}
