package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var (
	// jwtSecret      = []byte("your-secret-key")
	authServiceURL = "http://localhost:8081/verify" // URL для проверки токена
)

// func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		authHeader := r.Header.Get("Authorization")
// 		if authHeader == "" {
// 			http.Error(w, "Требуется токен", http.StatusUnauthorized)
// 			return
// 		}

// 		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

// 		// Локальная проверка токена (для простоты, можно заменить на вызов auth-service)
// 		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
// 			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 				return nil, fmt.Errorf("неверный метод подписи")
// 			}
// 			return jwtSecret, nil
// 		})

// 		if err != nil || !token.Valid {
// 			http.Error(w, "Неверный токен", http.StatusUnauthorized)
// 			return
// 		}

// 		claims, ok := token.Claims.(jwt.MapClaims)
// 		if !ok || !token.Valid {
// 			http.Error(w, "Неверные данные токена", http.StatusUnauthorized)
// 			return
// 		}

// 		userID, ok := claims["user_id"].(float64)
// 		if !ok {
// 			http.Error(w, "ID пользователя не найден в токене", http.StatusUnauthorized)
// 			return
// 		}

// 		role, ok := claims["role"].(string)
// 		if !ok {
// 			http.Error(w, "Роль не найдена в токене", http.StatusUnauthorized)
// 			return
// 		}

// 		r.Header.Set("UserID", fmt.Sprintf("%d", int(userID)))
// 		r.Header.Set("Role", role)
// 		next(w, r)
// 	}
// }

// Вариант с вызовом auth-service (раскомментировать для использования)
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Требуется токен", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// Вызов auth-service для проверки токена
		req, _ := http.NewRequest("POST", authServiceURL, strings.NewReader(tokenStr))
		req.Header.Set("Content-Type", "text/plain")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			http.Error(w, "Ошибка проверки токена", http.StatusUnauthorized)
			return
		}
		defer resp.Body.Close()

		var claims struct {
			UserID int    `json:"user_id"`
			Role   string `json:"role"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
			http.Error(w, "Неверный токен", http.StatusUnauthorized)
			return
		}

		r.Header.Set("UserID", fmt.Sprintf("%d", claims.UserID))
		r.Header.Set("Role", claims.Role)
		next(w, r)
	}
}
