package telegram

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	telegramusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/telegram"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request  request.CustomRequestInterface
	telegram *telegramusecase.Usecase
}

func New(usecase *telegramusecase.Usecase) *Controller {
	return &Controller{request: request.NewCustomRequest(), telegram: usecase}
}

func (ctl *Controller) GetSettings(c *echo.Context) error {
	result, err := ctl.telegram.GetSettings(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (ctl *Controller) UpdateSettings(c *echo.Context) error {
	var input PatchSettingsData
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.telegram.PatchSettings(c.Request().Context(), input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (ctl *Controller) Test(c *echo.Context) error {
	var input TestData
	_ = ctl.request.DoValidate(c, &input)
	if err := ctl.telegram.Test(c.Request().Context(), input.Message); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (ctl *Controller) ListPackages(c *echo.Context) error {
	result, err := ctl.telegram.ListPackages(c.Request().Context(), c.QueryParam("include_inactive") == "true")
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (ctl *Controller) CreatePackage(c *echo.Context) error {
	var input CreatePackageData
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.telegram.CreatePackage(c.Request().Context(), input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusCreated, result)
}

func (ctl *Controller) UpdatePackage(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	var input PatchPackageData
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.telegram.PatchPackage(c.Request().Context(), id, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (ctl *Controller) DeletePackage(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.telegram.DeletePackage(c.Request().Context(), id); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusNoContent, nil)
}

func (ctl *Controller) ListRequests(c *echo.Context) error {
	pagination := ctl.request.Pagination(c)
	query := c.Request().URL.Query()
	if query.Get("order") == "" {
		pagination.Order = "created_at"
	}
	if query.Get("sort") == "" {
		pagination.Sort = "DESC"
	}
	result, total, err := ctl.telegram.ListRequests(c.Request().Context(), pagination, c.QueryParam("status"), c.QueryParam("type"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, RequestsResponse{Meta: request.Meta{Page: pagination.Page, PageSize: pagination.PageSize, TotalRecords: total}, Result: result})
}

func (ctl *Controller) GetRequest(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.telegram.RequestByID(c.Request().Context(), id)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (ctl *Controller) GetReceipt(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	path, err := ctl.telegram.ReceiptPath(c.Request().Context(), id)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.File(path)
}

func (ctl *Controller) DeleteRequest(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.telegram.DeletePaymentRequest(c.Request().Context(), id); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (ctl *Controller) Approve(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	var input ApproveData
	_ = ctl.request.DoValidate(c, &input)
	result, err := ctl.telegram.Approve(c.Request().Context(), id, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (ctl *Controller) Reject(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	var input RejectData
	_ = ctl.request.DoValidate(c, &input)
	result, err := ctl.telegram.Reject(c.Request().Context(), id, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (ctl *Controller) ConfirmPayment(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	var input ConfirmPaymentData
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.telegram.ConfirmPayment(c.Request().Context(), id, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (ctl *Controller) AccountsForOcservUser(c *echo.Context) error {
	id, err := parseID(c.QueryParam("ocserv_user_id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.telegram.Accounts(c.Request().Context(), id)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

func (ctl *Controller) DeleteAccount(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	if err := ctl.telegram.DeleteAccount(c.Request().Context(), id); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusNoContent, nil)
}

func parseID(value string) (uint, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, strconv.ErrSyntax
	}
	return uint(id), nil
}
