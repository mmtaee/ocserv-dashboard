package main

import (
	"github.com/mmtaee/ocserv-users-management/api/cmd"
	_ "github.com/mmtaee/ocserv-users-management/api/docs"
)

// @title Ocserv User management Example Api
// @version 1.0
// @description This is a sample Ocserv User management Api server.
// @BasePath /api
func main() {
	cmd.Execute()
}
