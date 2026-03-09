/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/go-fuego/fuego"
	"github.com/spf13/cobra"
	"github.com/unshade/citadelle/internal/controllers"
	"github.com/unshade/citadelle/internal/db"
)

// serverCmd represents the server subcommand
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP server",
	Long: `Start the citadelle HTTP server with REST API endpoints.
	
This command initializes the SQLite database and starts the Fuego web framework server.`,
	Run: func(cmd *cobra.Command, args []string) {
		dbPath, _ := cmd.Flags().GetString("db")
		port, _ := cmd.Flags().GetString("port")

		if err := db.InitDB(dbPath); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}

		// Create Fuego server
		s := fuego.NewServer(
			fuego.WithAddr("localhost:" + port),
		)

		// Create /api group
		apiGroup := fuego.Group(s, "/api")

		// Initialize and register controllers under /api
		userCtrl := controllers.NewUserController()
		userCtrl.Register(apiGroup)

		log.Printf("Starting server on http://localhost:%s", port)

		if err := s.Run(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Add flags
	serverCmd.Flags().StringP("port", "p", "8080", "Port to run the server on")
	serverCmd.Flags().StringP("db", "d", "citadelle.db", "Path to SQLite database file")
}
