package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"todo-api/internal/handlers"
	"todo-api/internal/models"
	"todo-api/pkg/db"
)

func TestUsersHandler(t *testing.T) {
	schemaName := db.InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		db.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	}()

	users := []models.User{{Username: "testuser", Password: "password", Role: models.RoleAdmin}}
	if err := db.DB.Create(&users).Error; err != nil {
		t.Fatalf("Ошибка создания тестовой записей: %v", err)
	}

	req, err := http.NewRequest("GET", "/users", nil)
	if err != nil {
		t.Fatalf("Ошибка создания запроса: %v", err)
	}
	req.Header.Set("Role", models.RoleAdmin)

	rr := httptest.NewRecorder()

	handlers.UsersHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("UsersHandler() error = %v, wantErr %v", rr.Code, http.StatusOK)
	}
}
