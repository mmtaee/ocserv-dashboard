module github.com/mmtaee/ocserv-dashboard/ocserv_telegram_bot

go 1.26.4

require (
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/mmtaee/ocserv-dashboard/core v0.0.0
	github.com/spf13/cobra v1.8.1
)

replace github.com/mmtaee/ocserv-dashboard/core => ../core
