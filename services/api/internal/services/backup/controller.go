package backup

import (
	"compress/gzip"
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"net/http"
)

type Controller struct {
	request         request.CustomRequestInterface
	ocservUserRepo  repository.OcservUserRepositoryInterface
	ocservGroupRepo repository.OcservGroupRepositoryInterface
	backupRepo      repository.BackupRepositoryInterface
}

func New() *Controller {
	return &Controller{
		request:         request.NewCustomRequest(),
		ocservUserRepo:  repository.NewtOcservUserRepository(),
		ocservGroupRepo: repository.NewOcservGroupRepository(),
		backupRepo:      repository.NewBackupRepository(),
	}
}

// OcservGroupBackup
// @Summary      Backup ocserv groups
// @Description  Download gzip compressed JSON backup of all ocserv groups including default group configuration
// @Tags         System(Backup)
// @Produce      application/json
// @Produce      application/gzip
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Success      200 {file} file "ocserv_groups_backup.json.gz"
// @Router       /backup/ocserv_groups [get]
func (ctl *Controller) OcservGroupBackup(c echo.Context) error {
	defaultGroup, err := ctl.ocservGroupRepo.DefaultGroup()
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/json")
	c.Response().Header().Set(
		echo.HeaderContentDisposition,
		"attachment; filename=ocserv_groups_backup.json.gz",
	)
	c.Response().Header().Set(echo.HeaderContentEncoding, "gzip")

	c.Response().WriteHeader(http.StatusOK)

	gz := gzip.NewWriter(c.Response())

	if err = ctl.backupRepo.OcservGroupBackup(
		c.Request().Context(),
		gz,
		defaultGroup,
	); err != nil {
		gz.Close()
		return ctl.request.BadRequest(c, err)
	}

	if err = gz.Close(); err != nil {
		return err
	}

	return nil
}

// OcservUserBackup
// @Summary      Backup ocserv users
// @Description  Download gzip compressed JSON backup of all ocserv users (including default group mapping if exists)
// @Tags         System(Backup)
// @Produce      application/json
// @Produce      application/gzip
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Success      200 {file} file "ocserv_users_backup.json.gz"
// @Router       /backup/ocserv_users [get]
func (ctl *Controller) OcservUserBackup(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, "application/json")
	c.Response().Header().Set(
		echo.HeaderContentDisposition,
		"attachment; filename=ocserv_users_backup.json.gz",
	)
	c.Response().Header().Set(echo.HeaderContentEncoding, "gzip")

	c.Response().WriteHeader(http.StatusOK)

	gz := gzip.NewWriter(c.Response())

	if err := ctl.backupRepo.OcservUserBackup(
		c.Request().Context(),
		gz,
	); err != nil {
		gz.Close()
		return ctl.request.BadRequest(c, err)
	}

	if err := gz.Close(); err != nil {
		return err
	}

	return nil
}
