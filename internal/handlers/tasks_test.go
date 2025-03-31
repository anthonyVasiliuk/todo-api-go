package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"todo-api/pkg/db"
	"todo-api/internal/handlers"
	"todo-api/internal/models"
)

func SeedTasks(count int) []models.Task {
	tasks := make([]models.Task, count)
	for i := 0; i < count; i++ {
		tasks[i] = models.Task{
			Title: "Task " + string(rune('A'+i)), // Task A, Task B, Task C...
			Done:  i%2 == 0,                      // Чередуем true/false
		}
		if err := db.DB.Create(&tasks[i]).Error; err != nil {
			panic("Ошибка создания тестовой записи: " + err.Error())
		}
	}
	return tasks
}

func TestGetTasks(t *testing.T) {
	// Инициализируем тестовую базу и сохраняем имя схемы
	schemaName := db.InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		db.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	}()

	// Генерируем 3 тестовые записи
	expectedTasks := SeedTasks(3)

	// Создаём HTTP-запрос
	req, err := http.NewRequest("GET", "/tasks", nil)
	if err != nil {
		t.Fatalf("Ошибка создания запроса: %v", err)
	}

	// Создаём ResponseRecorder
	rr := httptest.NewRecorder()

	// Вызываем обработчик (оригинальный, без параметров db)
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

	// Проверяем количество задач
	if len(gotTasks) != len(expectedTasks) {
		t.Errorf("Ожидалось %d задач, получено %d", len(expectedTasks), len(gotTasks))
	}

	// Проверяем содержимое задач
	for i, task := range gotTasks {
		if task.ID != expectedTasks[i].ID || task.Title != expectedTasks[i].Title || task.Done != expectedTasks[i].Done {
			t.Errorf("Ожидалась задача %+v, получена %+v", expectedTasks[i], task)
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

	var savedTask models.Task
	if err := db.DB.First(&savedTask, 1).Error; err != nil {
		t.Errorf("Задача не найдена в базе: %v", err)
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
		ID:    id,
		Title: "task updated",
		Done:  true,
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
	got := strings.TrimSpace(rr.Body.String())
	if got != string(body) {
		t.Errorf("Ожидался ответ %v, получен %v", body, got)
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
