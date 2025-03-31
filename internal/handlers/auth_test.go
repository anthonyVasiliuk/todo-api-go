package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"todo-api/internal/handlers"
	"todo-api/internal/models"
	"todo-api/pkg/db"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type TokenResponse struct {
	Token string `json:"token"`
}

func TestRegistration(t *testing.T) {
	// Инициализируем тестовую базу и сохраняем имя схемы
	schemaName := db.InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		db.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	}()

	user := models.User{
		Username: "test",
		Password: "password",
	}

	body, err := json.Marshal(user)

	if err != nil {
		t.Fatalf("Ошибка сериализации задачи: %v", err)
	}

	// Создаём POST-запрос
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Создаём ResponseRecorder
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handlers.Register(rr, req)

	// Проверяем статус-код
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Ожидался статус %v, получен %v", http.StatusCreated, status)
	}

	// Проверяем тело ответа
	var createdUser models.User
	if err := json.NewDecoder(rr.Body).Decode(&createdUser); err != nil {
		t.Fatalf("Ошибка десериализации ответа: %v", err)
	}
}

func TestLogin(t *testing.T) {
	// Инициализируем тестовую базу и сохраняем имя схемы
	schemaName := db.InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		db.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	}()

	var jwtSecret = []byte(os.Getenv("JWT_SECRET_TESTING"))

	// Создаём тестового пользователя с хешированным паролем
	user := models.User{
		Username: "testuser",
		Password: "password123", // Оригинальный пароль
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Ошибка хеширования пароля: %v", err)
	}
	user.Password = string(hashedPassword) // Сохраняем хешированный пароль

	if err := db.DB.Create(&user).Error; err != nil {
		t.Fatalf("Ошибка создания тестовой записи: %v", err)
	}

	creds := models.Creds{
		Username: "testuser",
		Password: "password123",
	}

	body, err := json.Marshal(creds)

	if err != nil {
		t.Fatalf("Ошибка сериализации задачи: %v", err)
	}

	// Создаём POST-запрос
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Создаём ResponseRecorder
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handlers.Login(rr, req)

	// Проверяем статус-код
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Ожидался статус %v, получен %v", http.StatusOK, status)
	}

	// Проверяем тело ответа
	var tokenResp TokenResponse
	if err := json.NewDecoder(rr.Body).Decode(&tokenResp); err != nil {
		t.Fatalf("Ошибка десериализации ответа: %v", err)
	}

	// Проверяем, что токен не пустой
	if tokenResp.Token == "" {
		t.Errorf("Ожидался непустой токен, получен пустой")
	}

	// Проверяем валидность токена
	token, err := jwt.Parse(tokenResp.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неверный метод подписи: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		t.Errorf("Ошибка валидации токена: %v", err)
		return
	}

	// Проверяем, что токен валиден
	if !token.Valid {
		t.Errorf("Токен невалиден")
	}
}
