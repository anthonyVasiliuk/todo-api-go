package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"todo-api/internal/models"
	"todo-api/pkg/db"
)

// Обработчик для списка задач
func TasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var tasks []models.Task
		userIDStr := r.Header.Get("UserID")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "UserID не найден в заголовке", http.StatusBadRequest)
			return
		}

		var user models.User
		if err := db.DB.First(&user, userID).Error; err != nil {
			http.Error(w, "Пользователь не найден", http.StatusInternalServerError)
			return
		}

		query := db.DB
		if user.Role != "admin" {
			query = query.Where("user_id = ?", userID)
		}

		if err := query.Find(&tasks).Error; err != nil {
			http.Error(w, "Ошибка получения задач", http.StatusInternalServerError)
			return
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
