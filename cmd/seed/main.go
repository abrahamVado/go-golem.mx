package main

import (
	"github.com/example/gin-multitenant-backend/internal/config"
	"github.com/example/gin-multitenant-backend/internal/database"
	"github.com/example/gin-multitenant-backend/seeders"
	"log"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := seeders.Run(db, cfg); err != nil {
		log.Fatal(err)
	}
}
