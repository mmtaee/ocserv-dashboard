package worker

import (
	"context"
	"errors"
	"fmt"

	occtldocker "github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/docker"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/occtl"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/user"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/database"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/services/worker/connectionexpiry"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/services/worker/readers"
	userexpiryservice "github.com/mmtaee/ocserv-dashboard/backend/internal/services/worker/userexpiry"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/worker/logprocessor"
	userexpiryusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/worker/userexpiry"
	"golang.org/x/sync/errgroup"
)

const ocservServiceName = "ocserv"

type accessController struct {
	disconnect func(string) (string, error)
	lock       func(string) (string, error)
	unlock     func(string) (string, error)
}

func (c accessController) Disconnect(username string) error {
	_, err := c.disconnect(username)
	return err
}

func (c accessController) Lock(username string) error {
	_, err := c.lock(username)
	return err
}

func (c accessController) Unlock(username string) error {
	_, err := c.unlock(username)
	return err
}

type logJob struct {
	dockerMode  bool
	access      accessController
	repository  logprocessor.Repository
	connections logprocessor.ConnectionObserver
}

type Component interface {
	Run(ctx context.Context) error
}

type Service struct {
	components []Component
}

func (j *logJob) Run(ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)
	lines := make(chan logger.StreamEntry, 1000)
	group.Go(func() error {
		defer close(lines)
		var err error
		if j.dockerMode {
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
	processor := logprocessor.New(groupCtx, lines, j.repository, j.access, j.connections)
	group.Go(processor.CalculateUserStats)
	return group.Wait()
}

func New(dockerMode bool) *Service {
	var access accessController
	if dockerMode {
		client := occtldocker.NewOcservOcctlDocker()
		access = accessController{disconnect: client.DisconnectUser, lock: client.Lock, unlock: client.Unlock}
	} else {
		control := occtl.NewOcservOcctl()
		users := user.NewOcservUser()
		access = accessController{disconnect: control.DisconnectUser, lock: users.Lock, unlock: users.UnLock}
	}
	db := database.GetConnection()
	statsRepository := repository.NewWorkerStatsRepository(db)
	expiryRepository := repository.NewWorkerUserExpiryRepository(db)
	expiryUsecase := userexpiryusecase.New(expiryRepository, access)
	return &Service{components: []Component{
		&logJob{dockerMode: dockerMode, access: access, repository: statsRepository, connections: connectionexpiry.New(expiryRepository)},
		userexpiryservice.NewCronService(expiryUsecase),
	}}
}

func (s *Service) Run(ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)
	for _, component := range s.components {
		component := component
		group.Go(func() error { return component.Run(groupCtx) })
	}
	return group.Wait()
}
