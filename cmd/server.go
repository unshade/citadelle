/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/go-fuego/fuego"
	"github.com/spf13/cobra"
	"github.com/unshade/citadelle/internal/controllers"
	"github.com/unshade/citadelle/internal/helpers"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/repositories"
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

		db, err := helpers.InitServerDb(dbPath)
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}

		// Auto-migrate the database
		if err := db.AutoMigrate(&models.ServerNode{}); err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}

		// Initialize repositories
		database := repositories.NewDatabase(db)

		// Create Fuego server
		s := fuego.NewServer(
			fuego.WithAddr("localhost:" + port),
		)

		// Create /api group
		apiGroup := fuego.Group(s, "/api")

		// Initialize and register controllers under /api
		fileCtrl := controllers.NewNodeController(*database)
		fileCtrl.Register(apiGroup)

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
