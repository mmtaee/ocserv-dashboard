package connectionexpiry

import (
	"context"
	"fmt"
	"regexp"
	"time"
)

var newSessionPattern = regexp.MustCompile(`main\[([^\]]+)\]:\S+\s+new user session`)

type Repository interface {
	RecordFirstConnection(ctx context.Context, username string, connectedAt time.Time) (bool, error)
}

type Observer struct {
	repository Repository
}

func New(repository Repository) *Observer {
	return &Observer{repository: repository}
}

// Observe records the first successful OCServ session using the timestamp
// supplied by the live OCServ log source.
func (o *Observer) Observe(ctx context.Context, line string, observedAt time.Time) error {
	match := newSessionPattern.FindStringSubmatch(line)
	if len(match) != 2 {
		return nil
	}
	if _, err := o.repository.RecordFirstConnection(ctx, match[1], observedAt.UTC()); err != nil {
		return fmt.Errorf("record first connection for %s: %w", match[1], err)
	}
	return nil
}
