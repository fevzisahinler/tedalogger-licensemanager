package models

import "gorm.io/gorm"

type Customer struct {
	gorm.Model
	CompanyName string    `json:"company_name"`
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
	Licenses    []License `json:"licenses"`
}
