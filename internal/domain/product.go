package domain

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	Name        string  `json:"name" gorm:"not null"`
	Description string  `json:"description"`
	Price       float64 `json:"price" gorm:"not null"`
	SKU         string  `json:"sku" gorm:"uniqueIndex;not null"`

	// Foreign Key ไป Category (Meny-to-One Relationship)
	CategoryID uint     `json:"category_id" gorm:"not null" validate:"required"`
	Category   Category `json:"category"`

	// ความสัมพันธ์แบบ one-to-one ไป Inventory (GORM จะใช้ ProductID ใน Inventory)
	Inventory Inventory      `json:"inventory"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"create_at"`
	UpdatedAt time.Time      `json:"update_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	SKU         string  `json:"sku" validate:"reuired"`
	CategoryID  uint    `json:"catefory_id" validate:"required"`
	Quantity    int64   `json:"quantity" validate:"required,gte=0"`
}
