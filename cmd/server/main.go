package main

import (
	"fmt"
	"net/http"
	"todo-api/internal/handlers"
	"todo-api/internal/middleware"
	"todo-api/pkg/db"
	"todo-api/pkg/logger"

	"github.com/joho/godotenv"
)

var (
	jwtSecret = []byte("your-secret-key") // Секретный ключ для JWT (в продакшене храните в переменных окружения)
)

func main() {
	if err := logger.InitLogger(); err != nil {
		panic(err)
	}

	err := godotenv.Load()
	if err != nil {
		logger.Log.Infof("Error loading .env file")
	}

	if err := db.InitDB(); err != nil {
		panic(err)
	}

	// Открытые эндпоинты
	http.HandleFunc("/register", handlers.Register)
	http.HandleFunc("/login", handlers.Login)

	// Защищённые эндпоинты
	http.HandleFunc("/tasks", middleware.AuthMiddleware(handlers.TasksHandler))
	http.HandleFunc("/tasks/", middleware.AuthMiddleware(handlers.TaskHandler))

	http.HandleFunc("/users", middleware.AuthMiddleware(handlers.UsersHandler))

	fmt.Println("Сервер запущен на :8080")
	http.ListenAndServe(":8080", nil)
}
