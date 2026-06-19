package backup

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/dashboard/api/internal/usecase"
)

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

type BackupController struct {
	backupUseCase      usecase.BackupUseCase
	ocservGroupUseCase usecase.OcservGroupUseCase
	ocservGroupRepo    repository.OcservGroupRepository
	req                *request.Request
	validator          *request.Validator
}

func NewBackupController(backupUseCase usecase.BackupUseCase, ocservGroupUseCase usecase.OcservGroupUseCase, ocservGroupRepo repository.OcservGroupRepository) *BackupController {
	return &BackupController{
		backupUseCase:      backupUseCase,
		ocservGroupUseCase: ocservGroupUseCase,
		ocservGroupRepo:    ocservGroupRepo,
		req:                &request.Request{},
		validator:          request.NewValidator(),
	}
}

func (ctrl *BackupController) OcservGroupBackup(c echo.Context) error {
	defaultGroup, err := ctrl.ocservGroupRepo.DefaultGroup()
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/json")
	c.Response().Header().Set(
		echo.HeaderContentDisposition,
		"attachment; filename=ocserv_groups_backup.json.gz",
	)
	c.Response().Header().Set(echo.HeaderContentEncoding, "gzip")

	c.Response().WriteHeader(http.StatusOK)

	gz := gzip.NewWriter(c.Response())

	if err = ctrl.backupUseCase.OcservGroupBackup(c.Request().Context(), gz, defaultGroup); err != nil {
		_ = gz.Close()
		return ctrl.req.BadRequest(c, err)
	}

	if err = gz.Close(); err != nil {
		return err
	}

	return nil
}

func (ctrl *BackupController) OcservGroupRestore(c echo.Context) error {
	adminID, ok := c.Get("admin_id").(uint)
	if !ok {
		return ctrl.req.BadRequest(c, errors.New("admin_id not found in context"))
	}

	reader, err := ctrl.fileUploadValidator(c)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
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
		return ctrl.req.BadRequest(c, errors.New("invalid json file"))
	}

	if err = decoder.Decode(&struct{}{}); err != io.EOF {
		return ctrl.req.BadRequest(c, errors.New("invalid json EOF file"))
	}

	if err = ctrl.ocservGroupRepo.UpdateDefaultGroup(groupData.DefaultGroup); err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	var inserted, existing *[]string

	if len(groupData.Groups) > 0 {
		inserted, existing, err = ctrl.backupUseCase.OcservGroupRestore(c.Request().Context(), adminID, &groupData.Groups)
		if err != nil {
			return ctrl.req.BadRequest(c, err)
		}
	}

	return c.JSON(http.StatusOK, RestoreResponse{
		Inserted: inserted,
		Existing: existing,
	})
}

func (ctrl *BackupController) OcservUserBackup(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, "application/json")
	c.Response().Header().Set(
		echo.HeaderContentDisposition,
		"attachment; filename=ocserv_users_backup.json.gz",
	)
	c.Response().Header().Set(echo.HeaderContentEncoding, "gzip")

	c.Response().WriteHeader(http.StatusOK)

	gz := gzip.NewWriter(c.Response())

	if err := ctrl.backupUseCase.OcservUserBackup(c.Request().Context(), gz); err != nil {
		_ = gz.Close()
		return ctrl.req.BadRequest(c, err)
	}

	if err := gz.Close(); err != nil {
		return err
	}

	return nil
}

func (ctrl *BackupController) OcservUserRestore(c echo.Context) error {
	adminID, ok := c.Get("admin_id").(uint)
	if !ok {
		return ctrl.req.BadRequest(c, errors.New("admin_id not found in context"))
	}

	reader, err := ctrl.fileUploadValidator(c)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)

	var users []models.OcservUser
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	if err = decoder.Decode(&users); err != nil {
		return ctrl.req.BadRequest(c, errors.New("invalid json file"))
	}

	if err = decoder.Decode(&struct{}{}); err != io.EOF {
		return ctrl.req.BadRequest(c, errors.New("invalid json EOF file"))
	}

	if len(users) == 0 {
		return c.JSON(http.StatusOK, RestoreResponse{
			Inserted: nil,
			Existing: nil,
		})
	}

	inserted, existing, err := ctrl.backupUseCase.OcservUserRestore(c.Request().Context(), adminID, &users)
	if err != nil {
		return ctrl.req.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, RestoreResponse{
		Inserted: inserted,
		Existing: existing,
	})
}

func (ctrl *BackupController) fileUploadValidator(c echo.Context) (io.ReadCloser, error) {
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
			return nil, errors.New("invalid gzip file")
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
