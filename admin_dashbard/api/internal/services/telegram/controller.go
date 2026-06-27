package telegram

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/usecase"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
)

type Controller struct {
	request request.CustomRequestInterface
	usecase usecase.TelegramUsecaseInterface
}

func New(telegramUsecase usecase.TelegramUsecaseInterface) *Controller {
	return &Controller{
		request: request.NewCustomRequest(),
		usecase: telegramUsecase,
	}
}

func (ctl *Controller) GetSettings(c echo.Context) error {
	s, err := ctl.usecase.GetSettings()
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, s)
}

func (ctl *Controller) UpdateSettings(c echo.Context) error {
	var data usecase.PatchSettingsData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err, "1003")
	}

	s, err := ctl.usecase.UpdateSettings(data)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}

	return c.JSON(http.StatusOK, s)
}

func (ctl *Controller) Test(c echo.Context) error {
	var data usecase.TestData
	_ = ctl.request.DoValidate(c, &data)

	if err := ctl.usecase.Test(data); err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (ctl *Controller) ListPackages(c echo.Context) error {
	includeInactive := c.QueryParam("include_inactive") == "true"
	packages, err := ctl.usecase.ListPackages(includeInactive)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, packages)
}

func (ctl *Controller) CreatePackage(c echo.Context) error {
	var data usecase.CreatePackageData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err, "1003")
	}
	created, err := ctl.usecase.CreatePackage(data)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusCreated, created)
}

func (ctl *Controller) UpdatePackage(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1029")
	}

	var data usecase.PatchPackageData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err, "1003")
	}

	pkg, err := ctl.usecase.UpdatePackage(uint(id), data)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, pkg)
}

func (ctl *Controller) DeletePackage(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1029")
	}
	if err := ctl.usecase.DeletePackage(uint(id)); err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusNoContent, nil)
}

func (ctl *Controller) ListRequests(c echo.Context) error {
	pagination := ctl.request.Pagination(c)
	// Default listing matches historical behavior (newest first) when client omits order/sort.
	q := c.Request().URL.Query()
	if q.Get("order") == "" {
		pagination.Order = "created_at"
	}
	if q.Get("sort") == "" {
		pagination.Sort = "DESC"
	}
	status := c.QueryParam("status")
	reqType := c.QueryParam("type")

	requests, total, err := ctl.usecase.ListRequests(pagination, status, reqType)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, usecase.RequestsResponse{
		Meta: request.Meta{
			Page:         pagination.Page,
			TotalRecords: total,
			PageSize:     pagination.PageSize,
		},
		Result: requests,
	})
}

func (ctl *Controller) GetRequest(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1029")
	}
	req, err := ctl.usecase.GetRequest(uint(id))
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, req)
}

func (ctl *Controller) GetReceipt(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1029")
	}
	filePath, err := ctl.usecase.GetReceipt(uint(id))
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.File(filePath)
}

func (ctl *Controller) DeleteRequest(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1029")
	}
	if err := ctl.usecase.DeleteRequest(uint(id)); err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.NoContent(http.StatusNoContent)
}

func (ctl *Controller) Approve(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1029")
	}

	var data usecase.ApproveData
	_ = ctl.request.DoValidate(c, &data)

	updated, err := ctl.usecase.Approve(uint(id), data)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, updated)
}

func (ctl *Controller) Reject(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1029")
	}

	var data usecase.RejectData
	_ = ctl.request.DoValidate(c, &data)

	updated, err := ctl.usecase.Reject(uint(id), data)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, updated)
}

func (ctl *Controller) ConfirmPayment(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1029")
	}

	var data usecase.ConfirmPaymentData
	if err := ctl.request.DoValidate(c, &data); err != nil {
		return ctl.request.BadRequest(c, err, "1003")
	}

	res, err := ctl.usecase.ConfirmPayment(uint(id), data)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, res)
}

func (ctl *Controller) AccountsForOcservUser(c echo.Context) error {
	uid := c.QueryParam("ocserv_user_uid")
	if uid == "" {
		return ctl.request.BadRequest(c, ctl.request.BadRequest(c, "ocserv_user_uid query parameter is required"), "1032")
	}

	accounts, err := ctl.usecase.AccountsForOcservUser(uid)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusOK, accounts)
}

func (ctl *Controller) DeleteAccount(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return ctl.request.BadRequest(c, err, "1029")
	}
	if err := ctl.usecase.DeleteAccount(uint(id)); err != nil {
		return ctl.request.BadRequest(c, err, "1000")
	}
	return c.JSON(http.StatusNoContent, nil)
}
