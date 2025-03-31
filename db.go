package main

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	// Строка подключения (DSN)
	dsn := "host=localhost user=user password=secret dbname=todo port=5432 sslmode=disable TimeZone=UTC"

	// Открываем соединение с базой данных
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}

	// Автомиграция: создаём таблицу tasks, если её нет
	err = db.AutoMigrate(&Task{})
	if err != nil {
		return fmt.Errorf("ошибка миграции: %v", err)
	}

	DB = db
	fmt.Println("База данных успешно подключена")
	return nil
}
