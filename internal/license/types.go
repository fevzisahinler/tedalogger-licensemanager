package license

import "time"

type LicenseModules struct {
	LogModule  bool `json:"log_module"`
	WifiModule bool `json:"wifi_module"`
}

type LicenseInput struct {
	LicenseID   string         `json:"license_id" validate:"required"`
	CustomerID  uint           `json:"customer_id" validate:"required"`
	Modules     LicenseModules `json:"modules" validate:"required"`
	ValidFrom   time.Time      `json:"valid_from"`
	ValidUntil  time.Time      `json:"valid_until"`
	MachineID   string         `json:"machine_id" validate:"required"`
	GracePeriod int            `json:"grace_period"`
}

type LicenseOutput struct {
	LicenseID   string         `json:"license_id"`
	CustomerID  uint           `json:"customer_id"`
	Modules     LicenseModules `json:"modules"`
	ValidFrom   time.Time      `json:"valid_from"`
	ValidUntil  time.Time      `json:"valid_until"`
	MachineID   string         `json:"machine_id"`
	GracePeriod int            `json:"grace_period"`
	Signature   string         `json:"signature"`
}
