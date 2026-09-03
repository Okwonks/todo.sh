package service_test

import (
	"database/sql"
	"testing"

	"github.com/Okwonks/go-todo/internal/database"
	"github.com/Okwonks/go-todo/internal/model"
	"github.com/Okwonks/go-todo/internal/repository"
	"github.com/Okwonks/go-todo/internal/service"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:?_parseTime=true")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	database.Migrate(db)
	return db
}

func TestTodoService_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewTodoRepository(db)
	svc := service.NewTodoService(repo)

	t.Run("empty description returns error", func(t *testing.T) {
		_, err := svc.Create(&model.Todo{})
		if err == nil {
			t.Fatal("expected error for empty description, got nil")
		}
	})

	t.Run("default project_id is set when empty", func(t *testing.T) {
		todo := &model.Todo{
			Description: "Write tests",
			Priority:    1,
		}
		id, err := svc.Create(todo)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		got, err := svc.GetById(id)
		if err != nil {
			t.Fatalf("GetById() error = %v", err)
		}
		if got.ProjectID != "todos" {
			t.Errorf("got ProjectID %q, want 'todos'", got.ProjectID)
		}
	})

	t.Run("explicit project_id is preserved", func(t *testing.T) {
		todo := &model.Todo{
			Description: "Feature task",
			Priority:    2,
			ProjectID:   "feature-x",
		}
		id, err := svc.Create(todo)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		got, err := svc.GetById(id)
		if err != nil {
			t.Fatalf("GetById() error = %v", err)
		}
		if got.ProjectID != "feature-x" {
			t.Errorf("got ProjectID %q, want 'feature-x'", got.ProjectID)
		}
	})
}
