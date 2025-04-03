package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"todo-api/internal/models"
	"todo-api/pkg/db"
	"todo-api/pkg/logger"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
)

var validate = validator.New()
var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

// Обработчик для списка задач
func TasksHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("UserID")
	role := r.Header.Get("Role")
	userID, _ := strconv.Atoi(userIDStr)
	switch r.Method {
	case "GET":
		pageStr := r.URL.Query().Get("page")
		limitStr := r.URL.Query().Get("limit")
		doneStr := r.URL.Query().Get("done")

		// Устанавливаем значения по умолчанию
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1 // По умолчанию первая страница
		}
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 10 // По умолчанию 10 записей
		}

		// Преобразуем done в bool (если параметр передан)
		var doneFilter *bool
		if doneStr != "" {
			done, err := strconv.ParseBool(doneStr)
			if err != nil {
				logger.Log.Warnf("Неверный параметр done: %v", err)
				http.Error(w, "Неверный параметр done", http.StatusBadRequest)
				return
			}
			doneFilter = &done
		}

		// Вычисляем смещение
		offset := (page - 1) * limit

		// Формируем ключ с учётом nil и разыменования
		var doneValue string
		if doneFilter != nil {
			doneValue = strconv.FormatBool(*doneFilter)
		} else {
			doneValue = "nil"
		}
		// Ключ для кэша
		cacheKey := fmt.Sprintf("tasks:user:%d:page:%d:limit:%d:done:%s", userID, page, limit, doneValue)
		ctx := context.Background()

		// Проверяем кэш
		cached, err := redisClient.Get(ctx, cacheKey).Bytes()
		if err == nil {
			logger.Log.Info("Данные взяты из кэша")
			w.Header().Set("Content-Type", "application/json")
			w.Write(cached)
			return
		} else {
			logger.Log.Info("Данные не взяты из кэша", err)
		}

		// Формируем запрос
		query := db.DB.Model(&models.Task{})
		if role != models.RoleAdmin {
			query = query.Where("user_id = ?", userID)
		}
		if doneFilter != nil {
			query = query.Where("done = ?", *doneFilter)
		}

		// Применяем пагинацию и получаем задачи
		var tasks []models.Task
		if err := query.Offset(offset).Limit(limit).Find(&tasks).Error; err != nil {
			logger.Log.Errorf("Ошибка получения задач: %v", err)
			http.Error(w, "Ошибка получения задач", http.StatusInternalServerError)
			return
		}

		// Сериализуем и кэшируем
		jsonData, _ := json.Marshal(tasks)
		if err := redisClient.Set(ctx, cacheKey, jsonData, 10*time.Minute).Err(); err != nil {
			logger.Log.Errorf("Ошибка записи в Redis: %v", err)
			// Не прерываем выполнение, так как это не критично
		} else {
			logger.Log.Info("Данные сохранены в кэш")
		}

		json.NewEncoder(w).Encode(tasks)
	case "POST":
		var t models.Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Некорректный запрос", http.StatusBadRequest)
			return
		}

		// Добавляем ID пользователя, создавшего задачу
		userIDStr := r.Header.Get("UserID")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "UserID не найден в заголовке", http.StatusBadRequest)
			return
		}
		t.UserID = userID

		if err := validate.Struct(t); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Сохраняем задачу в базе данных
		if err := db.DB.Create(&t).Error; err != nil {
			http.Error(w, "Ошибка создания задачи", http.StatusInternalServerError)
			return
		}

		ch := make(chan string, 1) // Буферизированный канал
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		go notifyUser(ctx, userID, t.Title, ch)

		// Можно не ждать результата в реальном коде, но для примера:
		notificationResult := <-ch
		logger.Log.Infof("Результат уведомления: %s", notificationResult)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(t)

	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// Обработчик для конкретной задачи
func TaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		var t models.Task
		if err := db.DB.First(&t, id).Error; err != nil {
			http.Error(w, "Задача не найдена", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(t)
	case "PUT":
		var t models.Task
		if !taskExists(t) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Задача не найдена"})
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Некорректный запрос", http.StatusBadRequest)
			return
		}
		if err := validate.Struct(t); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		t.ID = id
		if err := db.DB.Save(&t).Error; err != nil {
			http.Error(w, "Ошибка обновления задачи", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(t)
	case "DELETE":
		var t models.Task
		if !taskExists(t) {
			http.Error(w, "Задача не найдена", http.StatusBadRequest)
			return
		}
		if err := db.DB.Delete(&models.Task{}, id).Error; err != nil {
			http.Error(w, "Ошибка удаления задачи", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func taskExists(task models.Task) bool {
	if err := db.DB.First(&task).Error; err != nil {
		return false
	}
	return true
}
