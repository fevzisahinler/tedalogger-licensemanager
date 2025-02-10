package routes

import (
	"tedalogger-licensemanager/internal/config"
	"tedalogger-licensemanager/internal/controllers"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB, cfg *config.Config) {
	customerController := controllers.NewCustomerController(db)
	licenseController := controllers.NewLicenseController(db, cfg)
	validateController := controllers.NewValidateController(db, cfg)

	// Customer rotaları
	app.Post("/customers", customerController.CreateCustomer)
	app.Get("/customers", customerController.GetAllCustomers)
	app.Get("/customers/:id", customerController.GetCustomerByID)

	// License rotası
	app.Post("/licenses", licenseController.CreateLicense)

	// Validate rotası: Örneğin /validate?license_key=XXXX-XXXX-XXXX-XXXX şeklinde kullanılabilir
	app.Get("/validate", validateController.ValidateLicense)
}
