package main

import (
	"log"

	"github.com/yourusername/gophkeeper/internal/api"
	"github.com/yourusername/gophkeeper/internal/config"
	"github.com/yourusername/gophkeeper/internal/database"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	api := api.NewAPI(db, cfg.JWTSecret)

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := api.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
