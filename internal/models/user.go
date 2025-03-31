package models

import "time"

type User struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique" validate:"required,min=3,max=255"`
	Password  string    `json:"password" validate:"required,min=3,max=32"` // Хешированный пароль
	Role      string    `json:"role" gorm:"default:user"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime,default:now()"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime,default:now()"`
}

// Определяем константы для ролей
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)