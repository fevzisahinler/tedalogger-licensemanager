package main

import (
	"log"
	"tedalogger-licensemanager/internal/config"
	"tedalogger-licensemanager/internal/database"
	"tedalogger-licensemanager/internal/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}
	database.Migrate(db)

	// Fiber uygulamasını oluştur
	app := fiber.New()

	// Rotaları tanımla
	routes.Setup(app, db, cfg)

	// Sunucuyu başlat
	log.Fatal(app.Listen(":" + cfg.Port))
}
