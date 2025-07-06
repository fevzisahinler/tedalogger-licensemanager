package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type License struct {
	CustomerID    string    `json:"customer_id"`
	ExpiryDate    time.Time `json:"expiry_date"`
	Features      []string  `json:"features"`
	MaxNASDevices int       `json:"max_nas_devices"`
	MaxUsers      int       `json:"max_users"`
	Signature     string    `json:"signature"`
}

func main() {
	keyPath := flag.String("key", "", "Path to existing private key file (optional)")
	customerID := flag.String("customer", "ACME-12345", "Customer ID")
	featuresStr := flag.String("features", "captive_portal,5651_logging", "Comma-separated list of features")
	maxUsers := flag.Int("maxUsers", 1000, "Maximum number of users")
	maxNAS := flag.Int("maxNAS", 10, "Maximum number of NAS devices")
	expiry := flag.String("expiry", "2026-04-20T00:00:00Z", "Expiry date in RFC3339 format (e.g., 2026-01-01T00:00:00Z)")
	outFile := flag.String("out", "license.json", "Output license file")
	flag.Parse()

	// Parse expiry date
	expiryTime, err := parseExpiryDate(*expiry)
	if err != nil {
		fmt.Printf("Error: invalid expiry date => %v\n", err)
		fmt.Println("Supported formats:")
		fmt.Println("  - RFC3339: 2026-01-01T00:00:00Z")
		fmt.Println("  - RFC3339 with timezone: 2026-01-01T00:00:00+03:00")
		fmt.Println("  - Date only: 2026-01-01 (will be converted to 00:00:00Z)")
		os.Exit(1)
	}

	fList := splitAndTrim(*featuresStr, ",")

	var privKey *ecdsa.PrivateKey

	if *keyPath == "" {
		fmt.Println("No key provided. Generating new ECDSA key pair...")

		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			fmt.Printf("Error: cannot generate key => %v\n", err)
			os.Exit(1)
		}

		privBytes, err := x509.MarshalECPrivateKey(priv)
		if err != nil {
			fmt.Printf("Error: cannot marshal private key => %v\n", err)
			os.Exit(1)
		}

		privPem := &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}
		if err := os.WriteFile("private_key.pem", pem.EncodeToMemory(privPem), 0600); err != nil {
			fmt.Printf("Error: cannot write private key => %v\n", err)
			os.Exit(1)
		}

		pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
		if err != nil {
			fmt.Printf("Error: cannot marshal public key => %v\n", err)
			os.Exit(1)
		}
		pubPem := &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}
		if err := os.WriteFile("public_key.pem", pem.EncodeToMemory(pubPem), 0644); err != nil {
			fmt.Printf("Error: cannot write public key => %v\n", err)
			os.Exit(1)
		}

		fmt.Println("ECDSA key pair generated: private_key.pem & public_key.pem")
		privKey = priv

	} else {
		privKeyBytes, err := os.ReadFile(*keyPath)
		if err != nil {
			fmt.Printf("Error: cannot read key => %v\n", err)
			os.Exit(1)
		}
		block, _ := pem.Decode(privKeyBytes)
		if block == nil {
			fmt.Println("Error: failed to decode PEM block")
			os.Exit(1)
		}

		// Try different key parsing methods
		k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			privKey, err = x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				fmt.Printf("Error: cannot parse EC key => %v\n", err)
				os.Exit(1)
			}
		} else {
			switch kt := k.(type) {
			case *ecdsa.PrivateKey:
				privKey = kt
			case *rsa.PrivateKey:
				fmt.Println("Error: this is RSA, not ECDSA")
				os.Exit(1)
			default:
				fmt.Println("Error: unknown key type")
				os.Exit(1)
			}
		}
	}

	// Create license
	lic := License{
		CustomerID:    *customerID,
		Features:      fList,
		MaxUsers:      *maxUsers,
		MaxNASDevices: *maxNAS,
		ExpiryDate:    expiryTime,
	}

	// Deterministik JSON oluştur ve imzala
	licenseData := createDeterministicJSON(&lic)
	fmt.Printf("License data to sign: %s\n", licenseData)

	// SHA256 hash
	hash := sha256.Sum256([]byte(licenseData))

	// Sign the hash
	sig, err := ecdsa.SignASN1(rand.Reader, privKey, hash[:])
	if err != nil {
		fmt.Printf("Error: sign failed => %v\n", err)
		os.Exit(1)
	}

	// Add signature to license
	lic.Signature = base64.StdEncoding.EncodeToString(sig)

	// Marshal final license with signature (field sırası önemli)
	finalBytes, err := json.MarshalIndent(lic, "", "  ")
	if err != nil {
		fmt.Printf("Error: cannot marshal final license => %v\n", err)
		os.Exit(1)
	}

	// Write to file
	if err := os.WriteFile(*outFile, finalBytes, 0644); err != nil {
		fmt.Printf("Error: cannot write license => %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nLicense generated successfully -> %s\n", *outFile)
	fmt.Printf("Customer: %s\n", lic.CustomerID)
	fmt.Printf("Features: %s\n", strings.Join(lic.Features, ", "))
	fmt.Printf("Max Users: %d\n", lic.MaxUsers)
	fmt.Printf("Max NAS Devices: %d\n", lic.MaxNASDevices)
	fmt.Printf("Expiry Date: %s\n", lic.ExpiryDate.Format(time.RFC3339))
	fmt.Printf("\nSignature: %s\n", lic.Signature)
}

// Deterministik JSON string oluştur (signature hariç)
func createDeterministicJSON(l *License) string {
	// Features array'ini sırala
	sortedFeatures := make([]string, len(l.Features))
	copy(sortedFeatures, l.Features)
	sort.Strings(sortedFeatures)

	// Sabit format ile JSON oluştur
	return fmt.Sprintf(`{"customer_id":"%s","expiry_date":"%s","features":%s,"max_nas_devices":%d,"max_users":%d}`,
		l.CustomerID,
		l.ExpiryDate.Format(time.RFC3339),
		toJSONArray(sortedFeatures),
		l.MaxNASDevices,
		l.MaxUsers,
	)
}

// String array'i JSON array string'e çevir
func toJSONArray(arr []string) string {
	if len(arr) == 0 {
		return "[]"
	}
	result := "["
	for i, v := range arr {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`"%s"`, v)
	}
	result += "]"
	return result
}

// parseExpiryDate parses expiry date with multiple format support
func parseExpiryDate(expiry string) (time.Time, error) {
	// List of supported time formats
	formats := []string{
		time.RFC3339,           // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04:05Z", // 2006-01-02T15:04:05Z
		"2006-01-02T15:04:05",  // 2006-01-02T15:04:05
		"2006-01-02 15:04:05",  // 2006-01-02 15:04:05
		"2006-01-02",           // 2006-01-02 (date only)
	}

	for _, format := range formats {
		if t, err := time.Parse(format, expiry); err == nil {
			// If date only format, ensure it's in UTC
			if format == "2006-01-02" {
				return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported time format: %s", expiry)
}

func splitAndTrim(s, sep string) []string {
	var out []string
	for _, p := range splitNonEmpty(s, sep) {
		trimmed := trimSpaces(p)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func splitNonEmpty(s, sep string) []string {
	var result []string
	parts := []rune(s)
	start := 0
	for i := 0; i < len(parts); i++ {
		if string(parts[i]) == sep {
			if i > start {
				result = append(result, string(parts[start:i]))
			}
			start = i + 1
		}
	}
	if start < len(parts) {
		result = append(result, string(parts[start:]))
	}
	return result
}

func trimSpaces(str string) string {
	return strings.TrimSpace(str)
}
