package entity

import "time"

// User represents an admin user (single-user system for now).
type User struct {
	ID           string    `gorm:"column:id;primaryKey;type:uuid"`
	Email        string    `gorm:"column:email;uniqueIndex;size:255"`
	PasswordHash string    `gorm:"column:password_hash;size:255"`
	Role         string    `gorm:"column:role;size:32;default:admin"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (u *User) TableName() string { return "users" }
