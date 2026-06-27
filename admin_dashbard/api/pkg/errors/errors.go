package errors

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

type ErrorInfo struct {
	Status int               `json:"status"`
	Fa     string            `json:"fa"`
	En     string            `json:"en"`
	It     string            `json:"it"`
	Ru     string            `json:"ru"`
	ZhCn   string            `json:"zh-cn"`
	ZhTw   string            `json:"zh-tw"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var errorCodes map[string]ErrorInfo

func init() {
	// Load errors from errors.json
	// Try to find the errors.json file
	paths := []string{
		"./config/errors.json",
		"../config/errors.json",
		"../../config/errors.json",
		"admin_dashbard/api/config/errors.json",
		"config/errors.json",
	}

	var data []byte
	var err error
	for _, p := range paths {
		if abs, absErr := filepath.Abs(p); absErr == nil {
			if data, err = os.ReadFile(abs); err == nil {
				break
			}
		}
		if data, err = os.ReadFile(p); err == nil {
			break
		}
	}

	if err != nil {
		// Fallback to default errors
		errorCodes = map[string]ErrorInfo{
			"1000": {Status: 400, Fa: "درخواست نامعتبر است", En: "Bad Request"},
		}
		return
	}

	if err = json.Unmarshal(data, &errorCodes); err != nil {
		errorCodes = map[string]ErrorInfo{
			"1000": {Status: 400, Fa: "درخواست نامعتبر است", En: "Bad Request"},
		}
	}
}

func GetError(code string, lang string) ErrorInfo {
	info, ok := errorCodes[code]
	if !ok {
		info = errorCodes["1000"]
	}
	return info
}

func GetMessage(info ErrorInfo, lang string) string {
	lang = strings.ToLower(lang)
	switch lang {
	case "fa":
		return info.Fa
	case "it":
		return info.It
	case "ru":
		return info.Ru
	case "zh-cn", "zh_cn":
		return info.ZhCn
	case "zh-tw", "zh_tw":
		return info.ZhTw
	default:
		return info.En
	}
}

func NewErrorResponse(code string, lang string) ErrorResponse {
	info := GetError(code, lang)
	return ErrorResponse{
		Code:    code,
		Status:  info.Status,
		Message: GetMessage(info, lang),
	}
}

func RespondError(c echo.Context, code string, lang ...string) error {
	l := "en"
	if len(lang) > 0 && lang[0] != "" {
		l = lang[0]
	} else {
		// Try to get language from header
		acceptLang := c.Request().Header.Get("Accept-Language")
		if acceptLang != "" {
			// Parse simple language header
			parts := strings.Split(acceptLang, ",")
			if len(parts) > 0 {
				l = strings.TrimSpace(strings.Split(parts[0], ";")[0])
			}
		}
	}
	resp := NewErrorResponse(code, l)
	return c.JSON(resp.Status, resp)
}
