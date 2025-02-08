package license

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"tedalogger-licensemanager/internal/database"
	_ "time"
)

func GenerateAndStoreLicense(input LicenseInput, privateKey *rsa.PrivateKey) (License, error) {
	modulesBytes, err := json.Marshal(input.Modules)
	if err != nil {
		return License{}, fmt.Errorf("modules JSON error: %w", err)
	}

	licModel := License{
		LicenseID:   input.LicenseID,
		CustomerID:  input.CustomerID,
		ModulesJSON: string(modulesBytes),
		ValidFrom:   input.ValidFrom,
		ValidUntil:  input.ValidUntil,
		MachineID:   input.MachineID,
		GracePeriod: input.GracePeriod,
	}

	if err := database.DB.Create(&licModel).Error; err != nil {
		return License{}, fmt.Errorf("db create error: %w", err)
	}

	outForSign := LicenseOutput{
		LicenseID:   licModel.LicenseID,
		CustomerID:  licModel.CustomerID,
		ValidFrom:   licModel.ValidFrom,
		ValidUntil:  licModel.ValidUntil,
		MachineID:   licModel.MachineID,
		GracePeriod: licModel.GracePeriod,
	}
	var m LicenseModules
	json.Unmarshal([]byte(licModel.ModulesJSON), &m)
	outForSign.Modules = m

	signBytes, err := json.Marshal(outForSign)
	if err != nil {
		return licModel, fmt.Errorf("json marshal for sign: %w", err)
	}
	hash := sha256.Sum256(signBytes)
	signatureBytes, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return licModel, fmt.Errorf("rsa sign error: %w", err)
	}
	signatureB64 := base64.StdEncoding.EncodeToString(signatureBytes)

	if err := database.DB.Model(&licModel).Updates(map[string]interface{}{
		"signature": signatureB64,
	}).Error; err != nil {
		return licModel, fmt.Errorf("db update signature error: %w", err)
	}

	licModel.Signature = signatureB64
	return licModel, nil
}

func GetLicenseByID(id uint) (License, error) {
	var lic License
	err := database.DB.First(&lic, id).Error
	if err != nil {
		return lic, err
	}
	return lic, nil
}

func ListLicenses() ([]License, error) {
	var list []License
	if err := database.DB.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func DeleteLicense(id uint) error {
	return database.DB.Delete(&License{}, id).Error
}

func CheckCustomerExists(customerID uint) error {
	var count int64
	if err := database.DB.Table("customers").Where("id = ?", customerID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
