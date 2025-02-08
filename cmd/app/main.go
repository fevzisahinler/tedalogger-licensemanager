package main

import (
	"crypto/rsa"
	"fmt"
	"log"
	_ "os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"tedalogger-licensemanager/internal/config"
	"tedalogger-licensemanager/internal/customer"
	"tedalogger-licensemanager/internal/database"
	"tedalogger-licensemanager/internal/license"

	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
)

func ParsePrivateKey(path string) (*rsa.PrivateKey, error) {
	pemBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func main() {
	cfg := config.LoadConfig()

	if err := database.Connect(cfg.GetDBConnectionString()); err != nil {
		log.Fatalf("DB connect error: %v", err)
	}

	privateKey, err := ParsePrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		log.Fatalf("Private key parse error: %v", err)
	}

	app := fiber.New()
	app.Use(logger.New()) // Request log middleware

	app.Post("/customers", customer.CreateCustomerHandler)
	app.Get("/customers", customer.ListCustomersHandler)
	app.Get("/customers/:id", customer.GetCustomerHandler)
	app.Delete("/customers/:id", customer.DeleteCustomerHandler)

	app.Post("/licenses", license.CreateLicenseHandler(privateKey))
	app.Get("/licenses", license.ListLicensesHandler)
	app.Get("/licenses/:id", license.GetLicenseHandler)
	app.Delete("/licenses/:id", license.DeleteLicenseHandler)

	addr := fmt.Sprintf(":%d", cfg.AppPort)
	log.Printf("License Manager running on port %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
