package usecase

import (
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
)

type OcctlUsecaseInterface interface {
	GetServerInfo() (models.OcservInfo, error)
	ExecuteCommand(action int, value string) (interface{}, error)
}

type OcctlUsecase struct {
	occtlRepo repository.OcctlRepositoryInterface
}

func NewOcctlUsecase(occtlRepo repository.OcctlRepositoryInterface) *OcctlUsecase {
	return &OcctlUsecase{
		occtlRepo: occtlRepo,
	}
}

func (uc *OcctlUsecase) GetServerInfo() (models.OcservInfo, error) {
	serverVersion := uc.occtlRepo.Version()
	info := models.OcservInfo{
		Version: serverVersion,
		Status:  "error",
	}

	serverStatus, err := uc.occtlRepo.Status()
	if err != nil {
		logger.Error("Get server status error: %v", err)
		return info, nil
	}

	serverStatusMap, ok := serverStatus.(map[string]interface{})
	if !ok {
		logger.Error("Invalid server status format")
		return info, nil
	}

	status := models.ParseOcservServerStatus(serverStatusMap)
	if status.GeneralInfo.Status != "" {
		info.Status = status.GeneralInfo.Status
	}

	return info, nil
}

func (uc *OcctlUsecase) ExecuteCommand(action int, value string) (interface{}, error) {
	actions := map[int]func(string) (interface{}, error){
		1:  func(_ string) (interface{}, error) { return uc.occtlRepo.OnlineSessions() },
		2:  func(val string) (interface{}, error) { return uc.occtlRepo.ShowUserByUsername(val) },
		3:  func(val string) (interface{}, error) { return uc.occtlRepo.ShowUserByID(val) },
		4:  func(val string) (interface{}, error) { return uc.occtlRepo.Disconnect(val) },
		5:  func(_ string) (interface{}, error) { return uc.occtlRepo.ShowSessionsAll() },
		6:  func(_ string) (interface{}, error) { return uc.occtlRepo.ShowSessionsValid() },
		7:  func(val string) (interface{}, error) { return uc.occtlRepo.ShowSessionBySID(val) },
		8:  func(_ string) (interface{}, error) { return uc.occtlRepo.IPBans() },
		9:  func(val string) (interface{}, error) { return uc.occtlRepo.UnbanIP(val) },
		10: func(_ string) (interface{}, error) { return uc.occtlRepo.Status() },
		11: func(_ string) (interface{}, error) { return uc.occtlRepo.ShowEvent(), nil },
		12: func(_ string) (interface{}, error) { return uc.occtlRepo.IRoutes() },
		13: func(_ string) (interface{}, error) { return uc.occtlRepo.Reload() },
		14: func(val string) (interface{}, error) { return uc.occtlRepo.DisconnectSession(val) },
		15: func(val string) (interface{}, error) { return uc.occtlRepo.Terminate(val) },
		16: func(val string) (interface{}, error) { return uc.occtlRepo.TerminateSession(val) },
	}

	handler, exists := actions[action]
	if !exists {
		return nil, nil
	}

	return handler(value)
}
