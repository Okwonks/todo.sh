package controllers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Okwonks/go-todo/api/v1/controllers"
	"github.com/Okwonks/go-todo/internal/database"
	"github.com/Okwonks/go-todo/internal/model"
	"github.com/Okwonks/go-todo/internal/repository"
	"github.com/Okwonks/go-todo/internal/service"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestServer(t *testing.T) (*http.ServeMux, *service.TodoService) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:?_parseTime=true")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	database.Migrate(db)

	repo := repository.NewTodoRepository(db)
	svc := service.NewTodoService(repo)
	ctrl := controllers.NewTodoController(svc)

	mux := http.NewServeMux()
	ctrl.RegisterTodos(mux)

	return mux, svc
}

func TestTodoHandler_CreateAndList(t *testing.T) {
	mux, _ := setupTestServer(t)

	// 1. Create with default project
	reqBody, _ := json.Marshal(map[string]any{
		"description": "API Test Task",
		"priority":    2,
	})
	req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /todos status = %d, want %d", rec.Code, http.StatusOK)
	}

	var created model.Todo
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode created todo: %v", err)
	}
	if created.ProjectID != "todos" {
		t.Errorf("got ProjectID %q, want 'todos'", created.ProjectID)
	}

	// 2. Create with custom project
	reqBodyCustom, _ := json.Marshal(map[string]any{
		"description": "API Custom Project Task",
		"priority":    1,
		"project_id":  "backend",
	})
	req2 := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(reqBodyCustom))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("POST /todos status = %d, want %d", rec2.Code, http.StatusOK)
	}

	var createdCustom model.Todo
	if err := json.NewDecoder(rec2.Body).Decode(&createdCustom); err != nil {
		t.Fatalf("failed to decode created todo: %v", err)
	}
	if createdCustom.ProjectID != "backend" {
		t.Errorf("got ProjectID %q, want 'backend'", createdCustom.ProjectID)
	}

	// 3. List
	listReq := httptest.NewRequest(http.MethodGet, "/todos", nil)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("GET /todos status = %d, want %d", listRec.Code, http.StatusOK)
	}

	var list []model.Todo
	if err := json.NewDecoder(listRec.Body).Decode(&list); err != nil {
		t.Fatalf("failed to decode list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}
}
