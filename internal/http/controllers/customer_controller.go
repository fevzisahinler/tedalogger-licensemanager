package controllers

import (
	"tedalogger-licensemanager/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CustomerController struct {
	DB *gorm.DB
}

func NewCustomerController(db *gorm.DB) *CustomerController {
	return &CustomerController{DB: db}
}

func (cc *CustomerController) CreateCustomer(c *fiber.Ctx) error {
	customer := new(models.Customer)
	if err := c.BodyParser(customer); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := cc.DB.Create(&customer).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(customer)
}

func (cc *CustomerController) GetAllCustomers(c *fiber.Ctx) error {
	var customers []models.Customer
	if err := cc.DB.Preload("Licenses").Find(&customers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(customers)
}

func (cc *CustomerController) GetCustomerByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var customer models.Customer
	if err := cc.DB.Preload("Licenses").First(&customer, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Customer not found"})
	}
	return c.JSON(customer)
}
