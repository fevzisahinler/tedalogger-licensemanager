package customer

import (
	"tedalogger-licensemanager/internal/database"
)

func CreateCustomer(cust Customer) error {
	return database.DB.Create(&cust).Error
}

func ListCustomers() ([]Customer, error) {
	var list []Customer
	if err := database.DB.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func GetCustomerByID(id uint) (Customer, error) {
	var c Customer
	err := database.DB.First(&c, id).Error
	return c, err
}

func DeleteCustomer(id uint) error {
	return database.DB.Delete(&Customer{}, id).Error
}
