package customer

import "time"

type Customer struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null;uniqueIndex"`
	Email     string `gorm:"not null;uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
