package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" // Подключаем pprof
	"todo-api/internal/handlers"
	"todo-api/internal/middleware"
	"todo-api/pkg/db"
	"todo-api/pkg/logger"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

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

	// Защищённые эндпоинты
	http.HandleFunc("/tasks", middleware.AuthMiddleware(handlers.TasksHandler))
	http.HandleFunc("/tasks/", middleware.AuthMiddleware(handlers.TaskHandler))

	http.HandleFunc("/users", middleware.AuthMiddleware(handlers.UsersHandler))

	ctx := context.Background()
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		logger.Log.Fatalf("Ошибка подключения к Redis: %v", err)
	}
	logger.Log.Infof("Redis подключён: %s", pong)

	// Запускаем сервер для pprof на отдельном порту
	go func() {
		fmt.Println("pprof доступен на :6060")
		http.ListenAndServe(":6060", nil)
	}()

	fmt.Println("Сервер запущен на :8080")
	http.ListenAndServe(":8080", nil)
}
