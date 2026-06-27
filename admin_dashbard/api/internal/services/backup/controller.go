package backup

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/core/models"
)

type Controller struct {
	request request.CustomRequestInterface
	usecase usecase.BackupUsecaseInterface
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

func New(uc usecase.BackupUsecaseInterface) *Controller {
	return &Controller{
		request: request.NewCustomRequest(),
		usecase: uc,
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
	c.Response().Header().Set(echo.HeaderContentType, "application/json")
	c.Response().Header().Set(
		echo.HeaderContentDisposition,
		"attachment; filename=ocserv_groups_backup.json.gz",
	)
	c.Response().Header().Set(echo.HeaderContentEncoding, "gzip")

	c.Response().WriteHeader(http.StatusOK)

	gz := gzip.NewWriter(c.Response())
	defer gz.Close()

	if err := ctl.usecase.OcservGroupBackup(gz); err != nil {
		return ctl.request.BadRequest(c, err, "1000")
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
// @Success      200 {object} usecase.RestoreResponse
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Router       /backup/ocserv_groups [post]
func (ctl *Controller) OcservGroupRestore(c echo.Context) error {
	owner := c.Get("username").(string)
	if owner == "" {
		return ctl.request.BadRequest(c, errors.New("admin or staff username not found"), "1005")
	}

	reader, err := ctl.fileUploadValidator(c)
	if err != nil {
		if err.Error() == "file is required" {
			return ctl.request.BadRequest(c, err, "1020")
		}
		if err.Error() == "invalid gzip file" {
			return ctl.request.BadRequest(c, err, "1021")
		}
		if err.Error() == "file must be .json or .json.gz" {
			return ctl.request.BadRequest(c, err, "1022")
		}
		return ctl.request.BadRequest(c, err, "1000")
	}

	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)

	var groupData usecase.BackupGroupFile

	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	if err = decoder.Decode(&groupData); err != nil {
		return ctl.request.BadRequest(c, errors.New("invalid json file"), "1023")
	}

	if err = decoder.Decode(&struct{}{}); err != io.EOF {
		return ctl.request.BadRequest(c, errors.New("invalid json EOF file"), "1024")
	}

	inserted, existing, err := ctl.usecase.OcservGroupRestore(owner, &groupData)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}

	return c.JSON(http.StatusOK, usecase.RestoreResponse{
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
	defer gz.Close()

	if err := ctl.usecase.OcservUserBackup(gz); err != nil {
		return ctl.request.BadRequest(c, err, "1000")
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
// @Success      200 {object} usecase.RestoreResponse
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Failure      403 {object} middlewares.PermissionDenied
// @Router       /backup/ocserv_users [post]
func (ctl *Controller) OcservUserRestore(c echo.Context) error {
	owner := c.Get("username").(string)
	if owner == "" {
		return ctl.request.BadRequest(c, errors.New("admin or staff username not found"), "1005")
	}

	reader, err := ctl.fileUploadValidator(c)
	if err != nil {
		if err.Error() == "file is required" {
			return ctl.request.BadRequest(c, err, "1020")
		}
		if err.Error() == "invalid gzip file" {
			return ctl.request.BadRequest(c, err, "1021")
		}
		if err.Error() == "file must be .json or .json.gz" {
			return ctl.request.BadRequest(c, err, "1022")
		}
		return ctl.request.BadRequest(c, err, "1000")
	}

	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)

	var users []models.OcservUser
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	if err = decoder.Decode(&users); err != nil {
		return ctl.request.BadRequest(c, errors.New("invalid json file"), "1023")
	}

	if err = decoder.Decode(&struct{}{}); err != io.EOF {
		return ctl.request.BadRequest(c, errors.New("invalid json EOF file"), "1024")
	}

	inserted, existing, err := ctl.usecase.OcservUserRestore(owner, users)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}

	return c.JSON(http.StatusOK, usecase.RestoreResponse{
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
