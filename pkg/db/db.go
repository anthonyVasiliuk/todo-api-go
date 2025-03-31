package db

import (
	"fmt"
	"time"
	"todo-api/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	// Строка подключения (DSN)
	dsn := "host=localhost user=user password=secret dbname=todo port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Task{}); err != nil {
		return fmt.Errorf("ошибка миграции: %v", err)
	}
	DB = db
	return nil
}

// InitTestDB — для тестов, создаёт временную схему
func InitTestDB() string {
	dsn := "host=localhost user=user password=secret dbname=testing port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
			panic("Ошибка подключения к PostgreSQL: " + err.Error())
	}

	schemaName := "test_schema_" + fmt.Sprintf("%d", time.Now().UnixNano())
	db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	db.Exec(fmt.Sprintf("CREATE SCHEMA %s", schemaName))
	db.Exec(fmt.Sprintf("SET search_path TO %s", schemaName))

	if err := db.AutoMigrate(&models.User{}, &models.Task{}); err != nil {
			panic("Ошибка миграции базы: " + err.Error())
	}

	DB = db
	return schemaName
}