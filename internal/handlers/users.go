package handlers

import (
	"encoding/json"
	"net/http"
	"todo-api/internal/models"
	"todo-api/pkg/db"
)

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Role") != models.RoleAdmin {
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
		return
	}
	var users []models.User
	db.DB.Find(&users)
	json.NewEncoder(w).Encode(users)
}
