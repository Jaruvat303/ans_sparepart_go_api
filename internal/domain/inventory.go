package domain

import (
	"time"

	"gorm.io/gorm"
)

// ข้อมูล Inventory จะถูกสร้างหลังจาก Product หนึ่งขิ้นถูกสร้างขึ้น

type Inventory struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ProductID uint           `json:"product_id" gorm:"unique;not null"`
	Quantity  int            `json:"quantity" gorm:"not null;default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
