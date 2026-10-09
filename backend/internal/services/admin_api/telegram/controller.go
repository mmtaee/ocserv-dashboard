package telegram

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/authz"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	telegramusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/telegram"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/middlewares"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
)

type Controller struct {
	request  request.CustomRequestInterface
	telegram *telegramusecase.Usecase
}

func New(usecase *telegramusecase.Usecase) *Controller {
	return &Controller{request: request.NewCustomRequest(), telegram: usecase}
}

// GetSettings returns the Telegram bot settings used by the admin UI.
// @Summary Get Telegram settings
// @Tags Telegram
// @Param Authorization header string true "Bearer TOKEN"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Success 200 {object} SettingsResponse
// @Router /telegram/settings [get]
func (ctl *Controller) GetSettings(c *echo.Context) error {
	result, err := ctl.telegram.GetSettings(c.Request().Context())
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// UpdateSettings partially updates the Telegram bot settings.
// @Summary Update Telegram settings
// @Tags Telegram
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body PatchSettingsData true "Telegram settings changes"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} SettingsResponse
// @Router /telegram/settings [patch]
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

// Test sends a test message using the configured Telegram bot.
// @Summary Test Telegram bot delivery
// @Tags Telegram
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body TestData false "Optional test message"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} map[string]string
// @Router /telegram/test [post]
func (ctl *Controller) Test(c *echo.Context) error {
	var input TestData
	_ = ctl.request.DoValidate(c, &input)
	if err := ctl.telegram.Test(c.Request().Context(), input.Message); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// ListPackages returns Telegram packages, optionally including inactive ones.
// @Summary List Telegram packages
// @Tags Telegram
// @Param Authorization header string true "Bearer TOKEN"
// @Param include_inactive query bool false "Include inactive packages"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Success 200 {array} models.TelegramPackage
// @Router /telegram/packages [get]
func (ctl *Controller) ListPackages(c *echo.Context) error {
	result, err := ctl.telegram.ListPackages(c.Request().Context(), c.QueryParam("include_inactive") == "true")
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// CreatePackage creates a Telegram package.
// @Summary Create Telegram package
// @Tags Telegram
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param request body CreatePackageData true "Telegram package"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 201 {object} models.TelegramPackage
// @Router /telegram/packages [post]
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

// UpdatePackage partially updates a Telegram package.
// @Summary Update Telegram package
// @Tags Telegram
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Package ID" minimum(1)
// @Param request body PatchPackageData true "Telegram package changes"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} models.TelegramPackage
// @Router /telegram/packages/{id} [patch]
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

// DeletePackage deletes a Telegram package.
// @Summary Delete Telegram package
// @Tags Telegram
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Package ID" minimum(1)
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 204
// @Router /telegram/packages/{id} [delete]
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

// ListRequests returns filtered Telegram payment requests.
// @Summary List Telegram requests
// @Tags Telegram
// @Param Authorization header string true "Bearer TOKEN"
// @Param page query int false "Page number, starting from 1" minimum(1)
// @Param size query int false "Number of items per page" minimum(1) maximum(100)
// @Param order query string false "Field to order by" default(created_at)
// @Param sort query string false "Sort order" Enums(ASC,DESC) default(DESC)
// @Param status query string false "Request status" Enums(pending,awaiting_payment,payment_uploaded,approved,rejected,delivered)
// @Param type query string false "Request type" Enums(new,renew)
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} RequestsResponse
// @Router /telegram/requests [get]
func (ctl *Controller) ListRequests(c *echo.Context) error {
	pagination := ctl.request.Pagination(c)
	query := c.Request().URL.Query()
	if query.Get("order") == "" {
		pagination.Order = "created_at"
	}
	if query.Get("sort") == "" {
		pagination.Sort = "DESC"
	}
	result, total, err := ctl.telegram.ListRequests(
		c.Request().Context(),
		pagination,
		models.TelegramRequestStatus(c.QueryParam("status")),
		models.TelegramRequestType(c.QueryParam("type")),
	)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, RequestsResponse{Meta: request.Meta{Page: pagination.Page, PageSize: pagination.PageSize, TotalRecords: total}, Result: result})
}

// GetRequest returns one Telegram payment request.
// @Summary Get Telegram request
// @Tags Telegram
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Request ID" minimum(1)
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} models.TelegramRequest
// @Router /telegram/requests/{id} [get]
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

// GetReceipt downloads the receipt attached to a Telegram payment request.
// @Summary Download Telegram request receipt
// @Tags Telegram
// @Produce application/octet-stream
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Request ID" minimum(1)
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {file} file "Receipt file"
// @Router /telegram/requests/{id}/receipt [get]
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

// DeleteRequest deletes a completed or rejected Telegram payment request.
// @Summary Delete Telegram request
// @Tags Telegram
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Request ID" minimum(1)
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 204
// @Router /telegram/requests/{id} [delete]
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

// Approve moves a pending Telegram request to awaiting payment.
// @Summary Approve Telegram request
// @Tags Telegram
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Request ID" minimum(1)
// @Param request body ApproveData false "Approval details"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} models.TelegramRequest
// @Router /telegram/requests/{id}/approve [post]
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

// Reject rejects a Telegram request.
// @Summary Reject Telegram request
// @Tags Telegram
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Request ID" minimum(1)
// @Param request body RejectData false "Rejection details"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} models.TelegramRequest
// @Router /telegram/requests/{id}/reject [post]
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

// ConfirmPayment confirms payment and delivers or renews the requested account.
// @Summary Confirm Telegram request payment
// @Tags Telegram
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Request ID" minimum(1)
// @Param request body ConfirmPaymentData true "Delivery overrides"
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {object} map[string]interface{}
// @Router /telegram/requests/{id}/confirm-payment [post]
func (ctl *Controller) ConfirmPayment(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	var input ConfirmPaymentData
	if err := ctl.request.DoValidate(c, &input); err != nil {
		return ctl.request.BadRequest(c, err)
	}
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.telegram.ConfirmPayment(c.Request().Context(), principal, id, input)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// AccountsForOcservUser returns Telegram accounts linked to an accessible Ocserv user.
// @Summary List Telegram accounts for an Ocserv user
// @Tags Telegram
// @Param Authorization header string true "Bearer TOKEN"
// @Param ocserv_user_id query int true "Ocserv user ID" minimum(1)
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 200 {array} models.TelegramAccount
// @Router /telegram/accounts [get]
func (ctl *Controller) AccountsForOcservUser(c *echo.Context) error {
	id, err := parseID(c.QueryParam("ocserv_user_id"))
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	principal, err := middlewares.Principal(c)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	result, err := ctl.telegram.Accounts(c.Request().Context(), principal, id)
	if err != nil {
		if errors.Is(err, authz.ErrForbidden) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return ctl.request.BadRequest(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// DeleteAccount removes a Telegram account link.
// @Summary Delete Telegram account link
// @Tags Telegram
// @Param Authorization header string true "Bearer TOKEN"
// @Param id path int true "Telegram account ID" minimum(1)
// @Failure 400 {object} request.ErrorResponse
// @Failure 401 {object} request.ErrorResponse
// @Failure 403 {object} request.ErrorResponse
// @Success 204
// @Router /telegram/accounts/{id} [delete]
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
