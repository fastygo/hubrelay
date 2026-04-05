package progress

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"gitcourse/internal/course"
	gitreader "gitcourse/internal/git"
)

type Cache struct {
	reader  gitreader.Reader
	ttl     time.Duration
	entries sync.Map
}

type entry struct {
	value     course.Progress
	expiresAt time.Time
}

func NewCache(reader gitreader.Reader, ttl time.Duration) *Cache {
	return &Cache{
		reader: reader,
		ttl:    ttl,
	}
}

func (c *Cache) Get(ctx context.Context, courseID, repoURL string) (course.Progress, error) {
	if cached, ok := c.entries.Load(courseID); ok {
		item := cached.(entry)
		if time.Now().Before(item.expiresAt) {
			return item.value, nil
		}
	}

	body, err := c.reader.ReadFile(ctx, repoURL, ".course/progress.json")
	if err != nil {
		return course.Progress{}, err
	}

	var progress course.Progress
	if err := json.Unmarshal(body, &progress); err != nil {
		return course.Progress{}, err
	}
	c.entries.Store(courseID, entry{
		value:     progress,
		expiresAt: time.Now().Add(c.ttl),
	})
	return progress, nil
}

func (c *Cache) Put(courseID string, progress course.Progress) {
	c.entries.Store(courseID, entry{
		value:     progress,
		expiresAt: time.Now().Add(c.ttl),
	})
}
