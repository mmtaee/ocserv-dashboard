module github.com/mmtaee/ocserv-dashboard/ocserv_webhook

go 1.26.4

require github.com/mmtaee/ocserv-dashboard/core v0.0.0

require (
	github.com/labstack/echo/v5 v5.0.3
	github.com/spf13/cobra v1.8.1
)

replace github.com/mmtaee/ocserv-dashboard/core => ../core
