package main

import (
	"fmt"
	"net/http"
	"todo-api/internal/handlers"
	"todo-api/internal/middleware"
	"todo-api/pkg/db"
)

var (
	jwtSecret = []byte("your-secret-key") // Секретный ключ для JWT (в продакшене храните в переменных окружения)
)

func main() {
	if err := db.InitDB(); err != nil {
		panic(err)
	}

	// Открытые эндпоинты
	http.HandleFunc("/register", handlers.Register)
	http.HandleFunc("/login", handlers.Login)

	// Защищённые эндпоинты
	http.HandleFunc("/tasks", middleware.AuthMiddleware(handlers.TasksHandler))
	http.HandleFunc("/tasks/", middleware.AuthMiddleware(handlers.TaskHandler))

	fmt.Println("Сервер запущен на :8080")
	http.ListenAndServe(":8080", nil)
}
