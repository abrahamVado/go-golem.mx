package main

import (
	"github.com/golem-mx/core-api/internal/config"
	"github.com/golem-mx/core-api/internal/database"
	"github.com/golem-mx/core-api/seeders"
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
