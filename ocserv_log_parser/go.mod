module github.com/mmtaee/ocserv-dashboard/ocserv_log_parser

go 1.26.4

require (
	github.com/docker/docker v28.3.3+incompatible
	github.com/mmtaee/ocserv-dashboard/core v0.0.0
	github.com/spf13/cobra v1.8.1
)

replace github.com/mmtaee/ocserv-dashboard/core => ../core
