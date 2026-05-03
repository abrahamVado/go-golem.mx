package main

import (
	"github.com/golem-mx/core-api/internal/config"
	"github.com/golem-mx/core-api/internal/database"
	"github.com/golem-mx/core-api/internal/router"
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
