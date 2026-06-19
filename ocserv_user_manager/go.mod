module github.com/mmtaee/ocserv-dashboard/ocserv_user_manager

go 1.26.4

require (
	github.com/mmtaee/ocserv-dashboard/core v0.0.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/spf13/cobra v1.8.1
)

replace github.com/mmtaee/ocserv-dashboard/core => ../core
