package controllers

import (
	"encoding/json"
	"tedalogger-licensemanager/internal/config"
	"tedalogger-licensemanager/internal/license"
	"tedalogger-licensemanager/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type LicenseController struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func NewLicenseController(db *gorm.DB, cfg *config.Config) *LicenseController {
	return &LicenseController{DB: db, Cfg: cfg}
}

type LicenseRequest struct {
	CustomerID  uint            `json:"customer_id"`
	Modules     map[string]bool `json:"modules"`
	ValidFrom   string          `json:"valid_from"`
	ValidUntil  string          `json:"valid_until"`
	MachineID   string          `json:"machine_id"`
	GracePeriod int             `json:"grace_period"`
}

func (lc *LicenseController) CreateLicense(c *fiber.Ctx) error {
	req := new(LicenseRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Müşterinin varlığını kontrol et
	var customer models.Customer
	if err := lc.DB.First(&customer, req.CustomerID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Customer not found"})
	}

	// Lisans verisini oluştur
	licenseData := map[string]interface{}{
		"customer_id":  req.CustomerID,
		"modules":      req.Modules,
		"valid_from":   req.ValidFrom,
		"valid_until":  req.ValidUntil,
		"machine_id":   req.MachineID,
		"grace_period": req.GracePeriod,
	}

	// Lisansı üret (imzalama, paketleme, AES ile şifreleme)
	licenseKey, encryptedData, err := license.GenerateLicense(licenseData, lc.Cfg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Lisansı veritabanına kaydet
	lic := models.License{
		CustomerID:    req.CustomerID,
		LicenseKey:    licenseKey,
		EncryptedData: encryptedData,
	}
	if err := lc.DB.Create(&lic).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Lisans anahtarı ve şifrelenmiş veriyi döndür (isterseniz sadece anahtarı da dönebilirsiniz)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"license_key":    licenseKey,
		"encrypted_data": encryptedData,
	})
}
