package models

import "gorm.io/gorm"

type License struct {
	gorm.Model
	CustomerID    uint   `json:"customer_id"`
	LicenseKey    string `json:"license_key"`
	EncryptedData string `json:"encrypted_data"`
}
