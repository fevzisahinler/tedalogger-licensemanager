package models

import "time"

type License struct {
	ID          uint      `gorm:"primaryKey"`
	LicenseID   string    `gorm:"uniqueIndex;not null"`
	CustomerID  uint      `gorm:"not null"`
	ModulesJSON string    `gorm:"type:text;not null"`
	ValidFrom   time.Time `gorm:"not null"`
	ValidUntil  time.Time `gorm:"not null"`
	MachineID   string    `gorm:"not null"`
	GracePeriod int       `gorm:"default:0"`
	Signature   string    `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
