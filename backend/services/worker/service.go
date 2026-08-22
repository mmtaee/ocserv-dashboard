package worker

import (
	"context"
	"errors"
	"fmt"

	"github.com/mmtaee/ocserv-dashboard/backend/services/worker/internal/readers"
	"github.com/mmtaee/ocserv-dashboard/backend/services/worker/internal/stats"
	userexpiry "github.com/mmtaee/ocserv-dashboard/backend/services/worker/internal/userexpiry"
	"golang.org/x/sync/errgroup"
)

const ocservServiceName = "ocserv"

// Service owns the log-processing and user-expiry background jobs.
type Service struct {
	dockerMode bool
}

func New(dockerMode bool) *Service {
	return &Service{dockerMode: dockerMode}
}

func (s *Service) Run(ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)
	lines := make(chan string, 1000)

	group.Go(func() error {
		defer close(lines)
		var err error
		if s.dockerMode {
			err = readers.DockerStreamLogs(groupCtx, ocservServiceName, lines)
		} else {
			err = readers.SystemdStreamLogs(groupCtx, ocservServiceName, lines)
		}
		if groupCtx.Err() != nil {
			return nil
		}
		if err == nil {
			return errors.New("ocserv log reader stopped unexpectedly")
		}
		return fmt.Errorf("ocserv log reader: %w", err)
	})

	processor := stats.NewStatService(groupCtx, lines, s.dockerMode)
	group.Go(processor.CalculateUserStats)

	expiry := userexpiry.NewCronService(s.dockerMode)
	group.Go(func() error {
		if err := expiry.MissedCron(groupCtx); err != nil {
			return err
		}
		return expiry.UserExpiryCron(groupCtx)
	})

	return group.Wait()
}
