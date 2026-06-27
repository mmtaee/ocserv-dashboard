module github.com/mmtaee/ocserv-dashboard/webhook

go 1.25.0

require github.com/mmtaee/ocserv-dashboard/core v0.0.0-00010101000000-000000000000

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	golang.org/x/text v0.28.0 // indirect
	gorm.io/gorm v1.30.1 // indirect
)

replace github.com/mmtaee/ocserv-dashboard/core => ./../core
