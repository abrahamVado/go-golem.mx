package seeders

import (
	"github.com/abrahamVado/go-golem.mx/internal/platform/config"
	"gorm.io/gorm"
)

func Run(db *gorm.DB, cfg config.Config) error {
	// TODO: add real seed data later.
	_ = cfg

	return nil
}