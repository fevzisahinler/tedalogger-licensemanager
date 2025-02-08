package license

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"

	"gorm.io/gorm"
	"tedalogger-licensemanager/internal/database"
	"tedalogger-licensemanager/internal/models"
)

const licenseKeyCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateLicenseKey() (string, error) {
	keyLength := 25
	b := make([]byte, keyLength)
	for i := 0; i < keyLength; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(licenseKeyCharset))))
		if err != nil {
			return "", err
		}
		b[i] = licenseKeyCharset[num.Int64()]
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		string(b[0:5]),
		string(b[5:10]),
		string(b[10:15]),
		string(b[15:20]),
		string(b[20:25]),
	), nil
}

func GenerateAndStoreLicense(input LicenseInput, privateKey *rsa.PrivateKey) (models.License, error) {
	modulesBytes, err := json.Marshal(input.Modules)
	if err != nil {
		return models.License{}, fmt.Errorf("modules JSON error: %w", err)
	}

	licenseKey, err := generateLicenseKey()
	if err != nil {
		return models.License{}, fmt.Errorf("failed to generate license key: %w", err)
	}

	licModel := models.License{
		LicenseID:   licenseKey,
		CustomerID:  input.CustomerID,
		ModulesJSON: string(modulesBytes),
		ValidFrom:   input.ValidFrom,
		ValidUntil:  input.ValidUntil,
		MachineID:   input.MachineID,
		GracePeriod: input.GracePeriod,
	}

	if err := database.DB.Create(&licModel).Error; err != nil {
		return models.License{}, fmt.Errorf("db create error: %w", err)
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

func GetLicenseByID(id uint) (models.License, error) {
	var lic models.License
	err := database.DB.First(&lic, id).Error
	if err != nil {
		return lic, err
	}
	return lic, nil
}

func ListLicenses() ([]models.License, error) {
	var list []models.License
	if err := database.DB.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func DeleteLicense(id uint) error {
	return database.DB.Delete(&models.License{}, id).Error
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
