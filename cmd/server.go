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
	"github.com/unshade/citadelle/internal/handlers"
	"github.com/unshade/citadelle/internal/helpers"
	"github.com/unshade/citadelle/internal/middleware"
	"github.com/unshade/citadelle/internal/models"
	"github.com/unshade/citadelle/internal/repositories"
	"github.com/unshade/citadelle/internal/services"
	"github.com/unshade/citadelle/internal/storage"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP server",
	Long: `Start the citadelle HTTP server with REST API endpoints.

This command initializes the PostgreSQL database and starts the Fuego web framework server.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		db, err := helpers.InitServerDb(cfg.Database)
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}

		if err := db.AutoMigrate(&models.ServerNode{}, &models.User{}); err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}

		// Repositories
		database := repositories.NewDatabase(db)

		// Storage
		fileStorage := storage.NewDiskStorage("./data")

		// Services
		authService := services.NewAuthService(database.Users, cfg.JWTSecret)
		userService := services.NewUserService(database.Users)
		nodeService := services.NewNodeService(database.ServerNodes, fileStorage)

		// Auth middleware
		authMiddleware := middleware.NewJWTAuthMiddleware(cfg.JWTSecret)

		// HTTP server
		s := fuego.NewServer(
			fuego.WithAddr("localhost:"+cfg.Port),
			fuego.WithGlobalMiddlewares(cors.New(cors.Options{
				AllowedOrigins:   cfg.AllowedOrigins,
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Authorization", "Content-Type"},
				AllowCredentials: true,
			}).Handler),
		)

		apiGroup := fuego.Group(s, "/api")

		handlers.NewNodeHandler(nodeService).Register(apiGroup, authMiddleware)
		handlers.NewUserHandler(userService).Register(apiGroup)
		handlers.NewAuthHandler(authService).Register(apiGroup)

		log.Printf("Starting server on http://localhost:%s", cfg.Port)

		if err := s.Run(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
