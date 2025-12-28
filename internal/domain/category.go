package domain

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID       uint           `json:"id" gorm:"primaryKey"`
	Name     string         `json:"name" gorm:"unique;not null" validate:"required,min-2,max-50"`
	CreateAt time.Time      `json:"create_at"`
	UpdateAt time.Time      `json:"update_at"`
	DeleteAt gorm.DeletedAt `json:"-" gorm:"index"`
}
