package license

import (
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

func CreateLicenseHandler(privateKey *rsa.PrivateKey) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var inp LicenseInput
		if err := c.BodyParser(&inp); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "invalid JSON",
			})
		}

		if inp.ValidFrom.IsZero() {
			inp.ValidFrom = time.Now()
		}
		if inp.ValidUntil.IsZero() {
			inp.ValidUntil = inp.ValidFrom.AddDate(1, 0, 0) // 1 yıl varsayılan
		}

		if err := CheckCustomerExists(inp.CustomerID); err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "customer not found",
			})
		}

		licModel, err := GenerateAndStoreLicense(inp, privateKey)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		// Artık yalnızca oluşturulan lisans anahtarını (license_id) döndürüyoruz.
		return c.JSON(fiber.Map{
			"license": licModel.LicenseID,
		})
	}
}

func ListLicensesHandler(c *fiber.Ctx) error {
	licenses, err := ListLicenses()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}
	var output []LicenseOutput
	for _, l := range licenses {
		var m LicenseModules
		json.Unmarshal([]byte(l.ModulesJSON), &m)
		output = append(output, LicenseOutput{
			LicenseID:   l.LicenseID,
			CustomerID:  l.CustomerID,
			Modules:     m,
			ValidFrom:   l.ValidFrom,
			ValidUntil:  l.ValidUntil,
			MachineID:   l.MachineID,
			GracePeriod: l.GracePeriod,
			Signature:   l.Signature,
		})
	}
	return c.JSON(output)
}

func GetLicenseHandler(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id < 1 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid license id",
		})
	}
	lic, err := GetLicenseByID(uint(id))
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "license not found",
		})
	}
	var m LicenseModules
	json.Unmarshal([]byte(lic.ModulesJSON), &m)
	out := LicenseOutput{
		LicenseID:   lic.LicenseID,
		CustomerID:  lic.CustomerID,
		Modules:     m,
		ValidFrom:   lic.ValidFrom,
		ValidUntil:  lic.ValidUntil,
		MachineID:   lic.MachineID,
		GracePeriod: lic.GracePeriod,
		Signature:   lic.Signature,
	}
	return c.JSON(out)
}

func DeleteLicenseHandler(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id < 1 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid license id",
		})
	}
	if err := DeleteLicense(uint(id)); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "license deleted",
	})
}
