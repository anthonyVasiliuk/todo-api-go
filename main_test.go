package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitTestDB инициализирует тестовую базу SQLite в памяти
func InitTestDB() string {
	// Подключаемся к PostgreSQL
	dsn := "host=localhost user=user password=secret dbname=testing port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Ошибка подключения к PostgreSQL: " + err.Error())
	}

	// Создаём временную схему для тестов
	schemaName := "test_schema_" + fmt.Sprintf("%d", time.Now().UnixNano())
	db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	db.Exec(fmt.Sprintf("CREATE SCHEMA %s", schemaName))
	db.Exec(fmt.Sprintf("SET search_path TO %s", schemaName))

	// Выполняем миграцию в тестовой схеме
	if err := db.AutoMigrate(&Task{}); err != nil {
		panic("Ошибка миграции базы: " + err.Error())
	}

	// Переопределяем глобальную переменную DB
	DB = db
	return schemaName
}

func SeedTasks(count int) []Task {
	tasks := make([]Task, count)
	for i := 0; i < count; i++ {
		tasks[i] = Task{
			Title: "Task " + string(rune('A'+i)), // Task A, Task B, Task C...
			Done:  i%2 == 0,                      // Чередуем true/false
		}
		if err := DB.Create(&tasks[i]).Error; err != nil {
			panic("Ошибка создания тестовой записи: " + err.Error())
		}
	}
	return tasks
}

func TestGetTasks(t *testing.T) {
	// Инициализируем тестовую базу и сохраняем имя схемы
	schemaName := InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
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
	tasksHandler(rr, req)

	// Проверяем статус-код
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Ожидался статус %v, получен %v", http.StatusOK, status)
	}

	// Проверяем тело ответа
	var gotTasks []Task
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
	schemaName := InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
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
	taskHandler(rr, req) // Предполагается, что у вас есть taskHandler для /tasks/{id}

	// Проверяем статус-код
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Ожидался статус %v, получен %v", http.StatusOK, status)
	}

	// Проверяем тело ответа
	var gotTasks Task
	if err := json.NewDecoder(rr.Body).Decode(&gotTasks); err != nil {
		t.Fatalf("Ошибка десериализации ответа: %v", err)
	}

	if gotTasks.ID != expectedTask[0].ID || gotTasks.Title != expectedTask[0].Title || gotTasks.Done != expectedTask[0].Done {
		t.Errorf("Ожидалась задача %+v, получена %+v", expectedTask[0], gotTasks)
	}
}

func TestCreateTask(t *testing.T) {
	schemaName := InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	}()

	// Создаём задачу для отправки
	task := Task{
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
	tasksHandler(rr, req)

	// Проверяем статус-код
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Ожидался статус %v, получен %v", http.StatusCreated, status)
	}

	// Проверяем тело ответа
	var createdTask Task
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

	var savedTask Task
	if err := DB.First(&savedTask, 1).Error; err != nil {
		t.Errorf("Задача не найдена в базе: %v", err)
	}
}
func TestUpdateTask(t *testing.T) {
	// Подготовка: добавляем тестовую задачу
	schemaName := InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	}()

	task1 := SeedTasks(1)
	// Формируем URL с ID
	id := task1[0].ID
	url := fmt.Sprintf("/tasks/%d", id)
	task := Task{
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
	taskHandler(rr, req) // Предполагается, что у вас есть taskHandler для /tasks

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
	schemaName := InitTestDB()
	defer func() {
		// Очищаем тестовую схему после теста
		DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
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
	taskHandler(rr, req) // Предполагается, что у вас есть taskHandler для /tasks/{id}

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
