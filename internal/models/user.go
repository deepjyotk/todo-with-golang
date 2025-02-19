package models

import "time"

// User represents a user in the system.
//
// swagger:model
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"unique;not null" json:"username" example:"johndoe"`
	Email     string    `gorm:"unique;not null" json:"email" example:"john@example.com"`
	Password  string    `gorm:"not null" json:"-"` // Hashed password (never returned)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
