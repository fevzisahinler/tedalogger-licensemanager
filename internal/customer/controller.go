package customer

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"tedalogger-licensemanager/internal/models"
)

func CreateCustomerHandler(c *fiber.Ctx) error {
	var cust models.Customer
	if err := c.BodyParser(&cust); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid JSON",
		})
	}
	if err := CreateCustomer(cust); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}
	return c.JSON(cust)
}

func ListCustomersHandler(c *fiber.Ctx) error {
	custs, err := ListCustomers()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}
	return c.JSON(custs)
}

func GetCustomerHandler(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id < 1 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid customer id",
		})
	}
	cust, err := GetCustomerByID(uint(id))
	if err != nil {
		if err.Error() == "record not found" {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "customer not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}
	return c.JSON(cust)
}

func DeleteCustomerHandler(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id < 1 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid customer id",
		})
	}
	if err := DeleteCustomer(uint(id)); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "customer deleted",
	})
}
