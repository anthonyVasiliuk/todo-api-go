package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"todo-api/internal/handlers"
	"todo-api/internal/models"
	"todo-api/pkg/db"
	"todo-api/pkg/logger"
)

func TestMain(m *testing.M) {
	// Инициализируем зависимости
	if err := logger.InitLogger(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка инициализации логгера: %v\n", err)
		os.Exit(1)
	}

	// Запускаем тесты
	code := m.Run()

	// Очистка (опционально)
	// Здесь можно закрыть файлы логов или соединение с базой, если нужно

	os.Exit(code)
}

func SeedTasks(count int) []models.Task {
	tasks := make([]models.Task, count)
	for i := 0; i < count; i++ {
		tasks[i] = models.Task{
			Title:  "Task " + string(rune('A'+i)), // Task A, Task B, Task C...
			Done:   i%2 == 0,                      // Чередуем true/false
			UserID: 1,
		}
		if err := db.DB.Create(&tasks[i]).Error; err != nil {
			panic("Ошибка создания тестовой записи: " + err.Error())
		}
	}
	return tasks
}

func TestGetTasksForUser(t *testing.T) {
	// Инициализируем тестовую базу и сохраняем имя схемы
	schemaName := db.InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		db.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	}()

	user1 := models.User{
		Username: "user1",
		Password: "password1",
		Role:     "user",
	}
	if err := db.DB.Create(&user1).Error; err != nil {
		t.Fatalf("Ошибка создания тестового пользователя 1: %v", err)
	}
	user2 := models.User{
		Username: "user2",
		Password: "password2",
		Role:     "user",
	}
	if err := db.DB.Create(&user2).Error; err != nil {
		t.Fatalf("Ошибка создания тестового пользователя 2: %v", err)
	}
	// Генерируем тестовые записи для нескольких пользователей
	user1Tasks := SeedTasks(2)
	user2Tasks := make([]models.Task, 1)
	user2Tasks[0] = models.Task{
		Title:  "Task for User 2",
		Done:   false,
		UserID: user2.ID,
	}
	if err := db.DB.Create(&user2Tasks[0]).Error; err != nil {
		t.Fatalf("Ошибка создания тестовой записи для User 2: %v", err)
	}

	// Создаём HTTP-запрос от имени User 1
	req, err := http.NewRequest("GET", "/tasks", nil)
	if err != nil {
		t.Fatalf("Ошибка создания запроса: %v", err)
	}
	req.Header.Set("UserID", fmt.Sprintf("%d", user1.ID)) // Устанавливаем заголовок с ID пользователя

	// Создаём ResponseRecorder
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handlers.TasksHandler(rr, req)

	// Проверяем статус-код
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Ожидался статус %v, получен %v", http.StatusOK, status)
	}

	// Проверяем тело ответа
	var gotTasks []models.Task
	if err := json.NewDecoder(rr.Body).Decode(&gotTasks); err != nil {
		t.Fatalf("Ошибка десериализации ответа: %v", err)
	}

	// Проверяем, что возвращаются только задачи User 1
	if len(gotTasks) != len(user1Tasks) {
		t.Errorf("Ожидалось %d задач, получено %d", len(user1Tasks), len(gotTasks))
	}

	// Проверяем содержимое задач
	for i, task := range gotTasks {
		if task.ID != user1Tasks[i].ID || task.Title != user1Tasks[i].Title || task.Done != user1Tasks[i].Done {
			t.Errorf("Ожидалась задача %+v, получена %+v", user1Tasks[i], task)
		}
	}
}

func TestGetTask(t *testing.T) {
	schemaName := db.InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		db.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	}()

	// Генерируем тестовую запись
	expectedTask := SeedTasks(1)

	// Формируем URL с ID
	id := 1
	url := fmt.Sprintf("/tasks/%d", id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Ошибка создания запроса: %v", err)
	}

	// Создаём ResponseRecorder
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handlers.TaskHandler(rr, req) // Предполагается, что у вас есть handlers.taskHandler для /tasks/{id}

	// Проверяем статус-код
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Ожидался статус %v, получен %v", http.StatusOK, status)
	}

	// Проверяем тело ответа
	var gotTasks models.Task
	if err := json.NewDecoder(rr.Body).Decode(&gotTasks); err != nil {
		t.Fatalf("Ошибка десериализации ответа: %v", err)
	}

	if gotTasks.ID != expectedTask[0].ID || gotTasks.Title != expectedTask[0].Title || gotTasks.Done != expectedTask[0].Done {
		t.Errorf("Ожидалась задача %+v, получена %+v", expectedTask[0], gotTasks)
	}
}

func TestGetNotExistingTask(t *testing.T) {
	schemaName := db.InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		db.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	}()

	// Генерируем тестовую запись
	SeedTasks(1)

	// Формируем URL с ID не существующей записи
	id := 2
	url := fmt.Sprintf("/tasks/%d", id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Ошибка создания запроса: %v", err)
	}

	// Создаём ResponseRecorder
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handlers.TaskHandler(rr, req)

	// Проверяем статус-код
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Ожидался статус %v, получен %v", http.StatusNotFound, status)
	}

}

func TestCreateTask(t *testing.T) {
	schemaName := db.InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		db.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	}()

	// Создаём задачу для отправки
	task := models.Task{
		Title: "task title",
		Done:  true,
	}

	// Сериализуем задачу в JSON
	body, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Ошибка сериализации задачи: %v", err)
	}

	// Создаём POST-запрос
	req, err := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Ошибка создания запроса: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Добавляем заголовок с ID пользователя
	req.Header.Set("UserID", "1")

	// Создаём ResponseRecorder
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handlers.TasksHandler(rr, req)

	// Проверяем статус-код
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Ожидался статус %v, получен %v", http.StatusCreated, status)
	}

	// Проверяем тело ответа
	var createdTask models.Task
	if err := json.NewDecoder(rr.Body).Decode(&createdTask); err != nil {
		t.Fatalf("Ошибка десериализации ответа: %v", err)
	}

	// Проверяем, что задача создана корректно
	if createdTask.ID != 1 {
		t.Errorf("Ожидался ID %d, получен %d", 1, createdTask.ID)
	}
	if createdTask.Title != task.Title {
		t.Errorf("Ожидался Title %v, получен %v", task.Title, createdTask.Title)
	}
	if createdTask.Done != task.Done {
		t.Errorf("Ожидался Done %v, получен %v", task.Done, createdTask.Done)
	}
	if createdTask.UserID != 1 {
		t.Errorf("Ожидался UserID %d, получен %d", 1, createdTask.UserID)
	}

	var savedTask models.Task
	if err := db.DB.First(&savedTask, 1).Error; err != nil {
		t.Errorf("Задача не найдена в базе: %v", err)
	}
}

func TestTasksHandler_Post_AsyncNotification(t *testing.T) {
	schemaName := db.InitTestDB()
	defer db.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))

	user := models.User{Username: "testuser", Password: "hashed", Role: models.RoleUser}
	db.DB.Create(&user)

	task := models.Task{Title: "Test Task", UserID: user.ID}
	body, _ := json.Marshal(task)

	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
	req.Header.Set("UserID", fmt.Sprintf("%d", user.ID))
	req.Header.Set("Role", models.RoleUser)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handlers.TasksHandler(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Ожидался статус %v, получен %v", http.StatusCreated, status)
	}

	var createdTask models.Task
	json.NewDecoder(rr.Body).Decode(&createdTask)
	if createdTask.Title != task.Title {
		t.Errorf("Ожидалось название %v, получено %v", task.Title, createdTask.Title)
	}
}

func TestUpdateTask(t *testing.T) {
	// Подготовка: добавляем тестовую задачу
	schemaName := db.InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		db.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	}()

	task1 := SeedTasks(1)
	// Формируем URL с ID
	id := task1[0].ID
	url := fmt.Sprintf("/tasks/%d", id)
	task := models.Task{
		ID:        id,
		Title:     "task updated",
		Done:      true,
		UserID:    1,
		CreatedAt: task1[0].CreatedAt,
		UpdatedAt: task1[0].UpdatedAt,
	}

	// 2. Сериализуем структуру в JSON
	body, err := json.Marshal(task)
	if err != nil {
		panic(err) // Обработка ошибки в реальном коде должна быть лучше
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Создаём ResponseRecorder
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handlers.TaskHandler(rr, req) // Предполагается, что у вас есть handlers.taskHandler для /tasks

	// Проверяем статус-код
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Ожидался статус %v, получен %v", http.StatusOK, status)
	}

	// Проверяем тело ответа
	var updatedTask models.Task
	if err := json.NewDecoder(rr.Body).Decode(&updatedTask); err != nil {
		t.Fatalf("Ошибка десериализации ответа: %v", err)
	}

}

func TestUpdateNonExistingTask(t *testing.T) {
	// Подготовка: добавляем тестовую задачу
	schemaName := db.InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		db.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	}()

	task := models.Task{
		ID:    1,
		Title: "task updated",
		Done:  true,
	}

	// 2. Сериализуем структуру в JSON
	body, err := json.Marshal(task)
	if err != nil {
		panic(err) // Обработка ошибки в реальном коде должна быть лучше
	}

	req, err := http.NewRequest("PUT", "/tasks/2", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Создаём ResponseRecorder
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handlers.TaskHandler(rr, req)

	// Проверяем статус-код
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Ожидался статус %v, получен %v", http.StatusNotFound, status)
	}

	// Проверяем тело ответа
	expected := `{"message":"Задача не найдена"}`
	got := strings.TrimSpace(rr.Body.String())
	if got != expected {
		t.Errorf("Ожидался ответ %v, получен %v", expected, got)
	}
}

func TestDeleteTask(t *testing.T) {
	schemaName := db.InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		db.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	}()

	task1 := SeedTasks(1)

	// Формируем URL с ID
	id := task1[0].ID
	url := fmt.Sprintf("/tasks/%d", id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Fatalf("Ошибка создания запроса: %v", err)
	}

	// Создаём ResponseRecorder
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handlers.TaskHandler(rr, req) // Предполагается, что у вас есть handlers.taskHandler для /tasks/{id}

	// Проверяем статус-код
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Ожидался статус %v, получен %v", http.StatusNoContent, status)
	}

	// Проверяем тело ответа
	expected := ``
	got := strings.TrimSpace(rr.Body.String())
	if got != expected {
		t.Errorf("Ожидался ответ %v, получен %v", expected, got)
	}
}

func TestTasksHandler_Get_Pagination(t *testing.T) {
	schemaName := db.InitTestDB()
	defer db.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))

	user := models.User{Username: "testuser", Password: "hashed", Role: models.RoleUser}
	db.DB.Create(&user)

	// Создаём 10 задач
	for i := 1; i <= 10; i++ {
			task := models.Task{
					Title:  fmt.Sprintf("Task %d", i),
					Done:   i%2 == 0, // Чётные завершены
					UserID: user.ID,
			}
			db.DB.Create(&task)
	}

	tests := []struct {
			name      string
			query     string
			wantCount int
	}{
			{"Page 1, limit 5", "?page=1&limit=5", 5},
			{"Page 2, limit 3", "?page=2&limit=3", 3},
			{"Done true, limit 5", "?done=true&limit=5", 5}, // 5 чётных задач
	}

	for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
					req, _ := http.NewRequest("GET", "/tasks"+tt.query, nil)
					req.Header.Set("UserID", fmt.Sprintf("%d", user.ID))
					req.Header.Set("Role", models.RoleUser)
					rr := httptest.NewRecorder()

					handlers.TasksHandler(rr, req)

					if status := rr.Code; status != http.StatusOK {
							t.Errorf("Ожидался статус %v, получен %v", http.StatusOK, status)
					}

					var tasks []models.Task
					if err := json.NewDecoder(rr.Body).Decode(&tasks); err != nil {
							t.Fatalf("Ошибка десериализации: %v", err)
					}
					if len(tasks) != tt.wantCount {
							t.Errorf("Ожидалось %d задач, получено %d", tt.wantCount, len(tasks))
					}
			})
	}
}