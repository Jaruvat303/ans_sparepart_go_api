package domain

import (
	"time"

	"gorm.io/gorm"
)

// User stuct
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"unique;not null"`
	Username  string         `json:"username" gorm:"notnull"`
	Password  string         `json:"-" gorm:"not null"`
	Role      string         `json:"role" gorm:"default:cashier"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
