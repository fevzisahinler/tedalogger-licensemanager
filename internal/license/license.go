package license

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"

	"tedalogger-licensemanager/internal/config"
)

// LicensePackage, imzalanmış veriyi ve imzayı içeren yapıdır.
type LicensePackage struct {
	Data      map[string]interface{} `json:"data"`
	Signature string                 `json:"signature"`
}

// GenerateLicense imzalama, paketleme ve AES şifreleme işlemlerini yapar.
// Geriye üretilen kısa lisans anahtarı ve şifrelenmiş verinin (base64) string’ini döner.
func GenerateLicense(licenseData map[string]interface{}, cfg *config.Config) (string, string, error) {
	// Lisans verisini JSON’a çevir
	dataBytes, err := json.Marshal(licenseData)
	if err != nil {
		return "", "", err
	}

	// RSA özel anahtarını yükle
	privKeyBytes, err := ioutil.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read private key: %v", err)
	}
	block, _ := pem.Decode(privKeyBytes)
	if block == nil {
		return "", "", errors.New("failed to parse PEM block containing the private key")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse RSA private key: %v", err)
	}

	// SHA256 ile imzala (RSA PKCS1v15)
	hash := sha256.Sum256(dataBytes)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", "", fmt.Errorf("failed to sign data: %v", err)
	}

	// Paket oluştur: orijinal veriyi ve base64 kodlanmış imzayı ekle
	pkg := LicensePackage{
		Data:      licenseData,
		Signature: base64.StdEncoding.EncodeToString(signature),
	}
	pkgBytes, err := json.Marshal(pkg)
	if err != nil {
		return "", "", err
	}

	// AES anahtarını dosyadan yükle (AES-256 için 32 bayt)
	aesKey, err := ioutil.ReadFile(cfg.AESKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read AES key: %v", err)
	}
	if len(aesKey) != 32 {
		return "", "", errors.New("AES key must be 32 bytes for AES-256")
	}

	// pkgBytes verisini AES CBC ile PKCS7 padding uygulayarak şifrele
	encrypted, err := encryptAES(aesKey, pkgBytes)
	if err != nil {
		return "", "", fmt.Errorf("AES encryption error: %v", err)
	}

	// Rastgele lisans anahtarı üret (XXXX-XXXX-XXXX-XXXX formatında)
	licenseKey, err := generateLicenseKey()
	if err != nil {
		return "", "", err
	}

	// Encrypted veriyi base64 ile kodlayıp, lisans anahtarı ile birlikte geri döndür
	return licenseKey, base64.StdEncoding.EncodeToString(encrypted), nil
}

// ValidateLicense, şifrelenmiş lisans verisini çözer ve RSA imzasını doğrular.
// İmzalama doğru ise deşifre edilen lisans verisini (pretty JSON string) geri döner.
func ValidateLicense(encryptedData string, cfg *config.Config) (string, error) {
	// Base64 çöz
	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("base64 decode error: %v", err)
	}

	// AES anahtarını yükle
	aesKey, err := ioutil.ReadFile(cfg.AESKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read AES key: %v", err)
	}
	if len(aesKey) != 32 {
		return "", errors.New("AES key must be 32 bytes for AES-256")
	}

	// AES ile şifreyi çöz
	decrypted, err := decryptAES(aesKey, encryptedBytes)
	if err != nil {
		return "", fmt.Errorf("AES decryption error: %v", err)
	}

	// Deşifre edilen paketi parse et
	var pkg LicensePackage
	if err := json.Unmarshal(decrypted, &pkg); err != nil {
		return "", fmt.Errorf("failed to parse license package: %v", err)
	}

	// İmza doğrulaması için "data" alanını yeniden JSON’a çevir
	dataBytes, err := json.Marshal(pkg.Data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal license data: %v", err)
	}

	// RSA public anahtarını yükle
	pubKeyBytes, err := ioutil.ReadFile(cfg.PublicKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read public key: %v", err)
	}
	block, _ = pem.Decode(pubKeyBytes)
	if block == nil {
		return "", errors.New("failed to parse PEM block containing the public key")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse RSA public key: %v", err)
	}
	publicKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("not RSA public key")
	}

	// İmzayı decode et
	signature, err := base64.StdEncoding.DecodeString(pkg.Signature)
	if err != nil {
		return "", fmt.Errorf("failed to decode signature: %v", err)
	}

	// RSA ile imzayı doğrula
	hash := sha256.Sum256(dataBytes)
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return "", fmt.Errorf("license signature verification failed: %v", err)
	}

	// Deşifre edilen lisans verisini pretty JSON olarak döndür
	pretty, err := json.MarshalIndent(pkg.Data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(pretty), nil
}

// encryptAES, verilen plaintext verisini AES-CBC ve PKCS7 padding ile şifreler.
func encryptAES(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	plaintext = pkcs7Pad(plaintext, blockSize)
	ciphertext := make([]byte, blockSize+len(plaintext))
	iv := ciphertext[:blockSize]
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[blockSize:], plaintext)
	return ciphertext, nil
}

// decryptAES, verilen ciphertext verisini AES-CBC ve PKCS7 padding kaldırarak çözer.
func decryptAES(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	if len(ciphertext) < blockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ciphertext[:blockSize]
	ciphertext = ciphertext[blockSize:]
	if len(ciphertext)%blockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)
	plaintext, err = pkcs7Unpad(plaintext, blockSize)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// pkcs7Pad, verilen veriyi blockSize katı olacak şekilde pad eder.
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7Unpad, PKCS7 padding’i kaldırır.
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("invalid padding size")
	}
	padding := data[len(data)-1]
	if int(padding) > len(data) || padding == 0 {
		return nil, errors.New("invalid padding")
	}
	for i := 0; i < int(padding); i++ {
		if data[len(data)-1-i] != padding {
			return nil, errors.New("invalid padding")
		}
	}
	return data[:len(data)-int(padding)], nil
}

// generateLicenseKey, "XXXX-XXXX-XXXX-XXXX" formatında rastgele lisans anahtarı üretir.
func generateLicenseKey() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	key := make([]byte, 16)
	for i := range key {
		num, err := randInt(len(charset))
		if err != nil {
			return "", err
		}
		key[i] = charset[num]
	}
	// Formatla: XXXX-XXXX-XXXX-XXXX
	return fmt.Sprintf("%s-%s-%s-%s", string(key[0:4]), string(key[4:8]), string(key[8:12]), string(key[12:16])), nil
}

// randInt, 0 ile max arasında rastgele sayı üretir.
func randInt(max int) (int, error) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(nBig.Int64()), nil
}
