package customer

import (
	"tedalogger-licensemanager/internal/database"
	"tedalogger-licensemanager/internal/models"
)

func CreateCustomer(cust models.Customer) error {
	return database.DB.Create(&cust).Error
}

func ListCustomers() ([]models.Customer, error) {
	var list []models.Customer
	if err := database.DB.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func GetCustomerByID(id uint) (models.Customer, error) {
	var c models.Customer
	err := database.DB.First(&c, id).Error
	return c, err
}

func DeleteCustomer(id uint) error {
	return database.DB.Delete(&models.Customer{}, id).Error
}
