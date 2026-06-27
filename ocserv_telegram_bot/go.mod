module github.com/mmtaee/ocserv-dashboard/ocserv_telegram_bot

go 1.26.4

require (
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/mmtaee/ocserv-dashboard/core v0.0.0
	github.com/spf13/cobra v1.8.1
	gorm.io/gorm v1.31.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/crypto v0.52.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	gorm.io/driver/postgres v1.6.0 // indirect
	gorm.io/driver/sqlite v1.6.0 // indirect
)

replace github.com/mmtaee/ocserv-dashboard/core => ../core
