module github.com/mmtaee/ocserv-dashboard/user_expiry

go 1.25.0

require (
	github.com/mmtaee/ocserv-dashboard/core v0.0.0-00010101000000-000000000000
	github.com/robfig/cron/v3 v3.0.1
	gorm.io/gorm v1.30.1
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	gorm.io/driver/postgres v1.6.0 // indirect
)

replace github.com/mmtaee/ocserv-dashboard/core => ./../core
