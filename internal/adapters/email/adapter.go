package email

import (
	"context"
	"log"
	"time"
)

type Adapter struct {
	provider string
	mode     string
}

func New(provider, mode string) *Adapter {
	return &Adapter{
		provider: provider,
		mode:     mode,
	}
}

func (a *Adapter) Name() string {
	return "email"
}

func (a *Adapter) Start(ctx context.Context) error {
	log.Printf("email adapter started in %s mode for provider=%s", a.mode, a.provider)

	// The first slice keeps email transport boundaries visible in the runtime
	// without introducing a full IMAP/SMTP workflow before the core contract stabilizes.
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("email adapter stopped")
			return nil
		case <-ticker.C:
			log.Printf("email adapter heartbeat provider=%s mode=%s", a.provider, a.mode)
		}
	}
}
