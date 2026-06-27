package backup

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/core/models"
)

type Controller struct {
	request         request.CustomRequestInterface
	ocservUserRepo  repository.OcservUserRepositoryInterface
	ocservGroupRepo repository.OcservGroupRepositoryInterface
	backupRepo      repository.BackupRepositoryInterface
}
type multiReadCloser struct {
	io.Reader
	closers []io.Closer
}

func (m *multiReadCloser) Close() error {
	for _, c := range m.closers {
		_ = c.Close()
	}
	return nil
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
		_ = gz.Close()
		return ctl.request.BadRequest(c, err)
	}

	if err = gz.Close(); err != nil {
		return err
	}

	return nil
}

// OcservGroupRestore
// @Summary      Restore ocserv groups
// @Description  Upload JSON or gzip-compressed (.json.gz) backup of ocserv groups
// @Tags         System(Restore)
// @Produce      application/json
// @Accept       multipart/form-data
// @Param        Authorization header string true "Bearer TOKEN"
// @Param        file formData file true "JSON or JSON.GZ file"
// @Success      200 {object} RestoreResponse
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Router       /backup/ocserv_groups [post]
func (ctl *Controller) OcservGroupRestore(c echo.Context) error {
	owner := c.Get("username").(string)
	if owner == "" {
		return ctl.request.BadRequest(c, errors.New("admin or staff username not found"))
	}

	reader, err := ctl.fileUploadValidator(c)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)

	type groupFile struct {
		DefaultGroup *models.OcservGroupConfig `json:"default_group" validate:"required"`
		Groups       []models.OcservGroup      `json:"groups"`
	}

	var groupData groupFile

	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	if err = decoder.Decode(&groupData); err != nil {
		return ctl.request.BadRequest(c, errors.New("invalid json file"))
	}

	if err = decoder.Decode(&struct{}{}); err != io.EOF {
		return ctl.request.BadRequest(c, errors.New("invalid json EOF file"))
	}

	if err = ctl.ocservGroupRepo.UpdateDefaultGroup(groupData.DefaultGroup); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	var inserted, existing *[]string

	if len(groupData.Groups) > 0 {
		inserted, existing, err = ctl.backupRepo.OcservGroupRestore(c.Request().Context(), owner, &groupData.Groups)
		if err != nil {
			return ctl.request.BadRequest(c, err)
		}
	}

	return c.JSON(http.StatusOK, RestoreResponse{
		Inserted: inserted,
		Existing: existing,
	})
}

// OcservUserBackup
// @Summary      Backup ocserv users
// @Description  Download gzip compressed JSON backup of all ocserv users
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
		_ = gz.Close()
		return ctl.request.BadRequest(c, err)
	}

	if err := gz.Close(); err != nil {
		return err
	}

	return nil
}

// OcservUserRestore
// @Summary      Restore ocserv users
// @Description  Upload JSON or gzip-compressed (.json.gz) backup of ocserv users
// @Tags         System(Restore)
// @Produce      application/json
// @Accept       multipart/form-data
// @Param        Authorization header string true "Bearer TOKEN"
// @Param        file formData file true "JSON or JSON.GZ file"
// @Success      200 {object} RestoreResponse
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Router       /backup/ocserv_users [post]
func (ctl *Controller) OcservUserRestore(c echo.Context) error {
	owner := c.Get("username").(string)
	if owner == "" {
		return ctl.request.BadRequest(c, errors.New("admin or staff username not found"))
	}

	reader, err := ctl.fileUploadValidator(c)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)

	var users []models.OcservUser
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	if err = decoder.Decode(&users); err != nil {
		return ctl.request.BadRequest(c, errors.New("invalid json file"))
	}

	if err = decoder.Decode(&struct{}{}); err != io.EOF {
		return ctl.request.BadRequest(c, errors.New("invalid json EOF file"))
	}

	if len(users) == 0 {
		return c.JSON(http.StatusOK, RestoreResponse{
			Inserted: nil,
			Existing: nil,
		})
	}

	inserted, existing, err := ctl.backupRepo.OcservUserRestore(c.Request().Context(), owner, &users)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, RestoreResponse{
		Inserted: inserted,
		Existing: existing,
	})
}

// fileUploadValidator validate request file format and extension
func (ctl *Controller) fileUploadValidator(c echo.Context) (io.ReadCloser, error) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, errors.New("file is required")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}

	filename := strings.ToLower(fileHeader.Filename)

	switch {
	case strings.HasSuffix(filename, ".json.gz"):
		gz, err := gzip.NewReader(file)
		if err != nil {
			_ = file.Close()
			return nil, fmt.Errorf("invalid gzip file")
		}

		return &multiReadCloser{
			Reader:  gz,
			closers: []io.Closer{gz, file},
		}, nil

	case strings.HasSuffix(filename, ".json"):
		return file, nil

	default:
		_ = file.Close()
		return nil, errors.New("file must be .json or .json.gz")
	}
}
