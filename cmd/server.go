/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/go-fuego/fuego"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/unshade/citadelle/internal/config"
	"github.com/unshade/citadelle/internal/controllers"
	"github.com/unshade/citadelle/internal/helpers"
	"github.com/unshade/citadelle/internal/middleware"
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
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		db, err := helpers.InitServerDb(cfg.DBPath)
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}

		if err := db.AutoMigrate(&models.ServerNode{}, &models.User{}); err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}

		database := repositories.NewDatabase(db)

		// Initialize auth middleware
		authMiddleware := middleware.NewJWTAuthMiddleware(cfg.JWTSecret)

		s := fuego.NewServer(
			fuego.WithAddr("localhost:"+cfg.Port),
			fuego.WithGlobalMiddlewares(cors.New(cors.Options{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"*"},
				AllowCredentials: true,
			}).Handler),
		)

		apiGroup := fuego.Group(s, "/api")

		nodeCtrl := controllers.NewNodeController(*database)
		nodeCtrl.Register(apiGroup, authMiddleware)

		userCtrl := controllers.NewUserController(*database)
		userCtrl.Register(apiGroup)

		authCtrl := controllers.NewAuthController(*database, cfg.JWTSecret)
		authCtrl.Register(apiGroup)

		log.Printf("Starting server on http://localhost:%s", cfg.Port)

		if err := s.Run(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
