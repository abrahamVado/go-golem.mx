package main

import (
	"github.com/example/gin-multitenant-backend/internal/config"
	"github.com/example/gin-multitenant-backend/internal/database"
	"github.com/example/gin-multitenant-backend/internal/router"
	"log"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	r := router.New(db, cfg)
	log.Fatal(r.Run(":" + cfg.AppPort))
}
