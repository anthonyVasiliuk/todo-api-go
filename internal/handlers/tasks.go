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
		if err := db.DB.Find(&tasks).Error; err != nil {
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
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Некорректный запрос", http.StatusBadRequest)
			return
		}
		t.ID = id
		if err := db.DB.Save(&t).Error; err != nil {
			http.Error(w, "Ошибка обновления задачи", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(t)
	case "DELETE":
		if err := db.DB.Delete(&models.Task{}, id).Error; err != nil {
			http.Error(w, "Ошибка удаления задачи", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}
