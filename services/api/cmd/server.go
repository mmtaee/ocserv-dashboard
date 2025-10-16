package cmd

import (
	"github.com/joho/godotenv"
	"github.com/mmtaee/ocserv-users-management/api/pkg/bootstrap"
	"github.com/spf13/cobra"
	"log"
)

var (
	debug bool
	host  string
	port  int
)

var serverCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the api server",
	Run: func(cmd *cobra.Command, args []string) {
		bootstrap.Serve(debug, host, port)
	},
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	serverCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	serverCmd.Flags().StringVar(&host, "host", "0.0.0.0", "Server Host")
	serverCmd.Flags().IntVarP(&port, "port", "p", 8080, "Server Port")
	rootCmd.AddCommand(serverCmd)
}
