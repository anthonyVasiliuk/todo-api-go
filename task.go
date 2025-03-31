package main

// Task — структура для задачи
type Task struct {
	ID    int    `json:"id" gorm:"primaryKey"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}
