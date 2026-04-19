package cmd

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	LabstackLog "github.com/labstack/gommon/log"
	"github.com/spf13/cobra"
	echoSwagger "github.com/swaggo/echo-swagger"
	"net/http"
	"os"
)

var swaggerCmd = &cobra.Command{
	Use:   "swagger",
	Short: "Run the swagger api server",
	Run: func(cmd *cobra.Command, args []string) {
		e := echo.New()
		e.Debug = true
		e.Logger.SetLevel(LabstackLog.DEBUG)
		e.GET("/swagger/*", echoSwagger.WrapHandler)
		server := fmt.Sprintf("%s:%d", "0.0.0.0", 8888)
		err := e.Start(server)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("shutting down the server: %v\n", err) // use fmt, not Fatal
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(swaggerCmd)
}
