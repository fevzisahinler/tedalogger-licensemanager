package controllers

import (
	"encoding/base64"
	"encoding/json"
	"tedalogger-licensemanager/internal/config"
	"tedalogger-licensemanager/internal/license"
	"tedalogger-licensemanager/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ValidateController struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func NewValidateController(db *gorm.DB, cfg *config.Config) *ValidateController {
	return &ValidateController{DB: db, Cfg: cfg}
}

// /validate?license_key=XXXX-XXXX-XXXX-XXXX şeklinde kullanılacak.
func (vc *ValidateController) ValidateLicense(c *fiber.Ctx) error {
	licenseKey := c.Query("license_key")
	if licenseKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "license_key query parameter required"})
	}

	// Lisansı veritabanından getir
	var lic models.License
	if err := vc.DB.Where("license_key = ?", licenseKey).First(&lic).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "License not found"})
	}

	// Lisans verisini AES çözme ve RSA imza kontrolü ile doğrula
	decrypted, err := license.ValidateLicense(lic.EncryptedData, vc.Cfg)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Deşifre edilen veriyi JSON parse edip döndür
	var licenseData map[string]interface{}
	if err := json.Unmarshal([]byte(decrypted), &licenseData); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse license data"})
	}

	return c.JSON(fiber.Map{
		"valid":        true,
		"license_data": licenseData,
	})
}
