package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"tedalogger-licensemanager/internal/models"
)

var DB *gorm.DB

func Connect(dsn string) error {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&models.Customer{},
		&models.License{},
	); err != nil {
		return err
	}

	DB = db
	log.Println("[DB] PostgreSQL connected & migrated successfully!")
	return nil
}
