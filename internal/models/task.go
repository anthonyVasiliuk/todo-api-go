package models

import "time"

// Task — структура для задачи
type Task struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" validate:"required,min=3,max=255"`
	Done      bool      `json:"done" gorm:"default:false" validate:"required"`
	UserID    int       `json:"user_id" gorm:"index"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime,default:now()"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime,default:now()"`
}
