package repository_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/Okwonks/go-todo/internal/database"
	"github.com/Okwonks/go-todo/internal/model"
	"github.com/Okwonks/go-todo/internal/repository"
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

func TestTodoRepo_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewTodoRepository(db)

	t.Run("create with default project_id", func(t *testing.T) {
		todo := &model.Todo{
			Description: "Test default project task",
			Priority:    3,
		}
		id, err := repo.Create(todo)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if id <= 0 {
			t.Fatalf("expected valid id, got %d", id)
		}

		got, err := repo.GetById(id)
		if err != nil {
			t.Fatalf("GetById() error = %v", err)
		}
		if got.Description != todo.Description {
			t.Errorf("got description %q, want %q", got.Description, todo.Description)
		}
		if got.ProjectID != "todos" {
			t.Errorf("got project_id %q, want %q", got.ProjectID, "todos")
		}
		if got.Priority != 3 {
			t.Errorf("got priority %d, want 3", got.Priority)
		}
	})

	t.Run("create with explicit project_id", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		todo := &model.Todo{
			Description: "Test explicit project task",
			Priority:    1,
			DueDate:     &now,
			ProjectID:   "work",
		}
		id, err := repo.Create(todo)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		got, err := repo.GetById(id)
		if err != nil {
			t.Fatalf("GetById() error = %v", err)
		}
		if got.ProjectID != "work" {
			t.Errorf("got project_id %q, want %q", got.ProjectID, "work")
		}
	})
}

func TestTodoRepo_GetAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewTodoRepository(db)

	todos := []model.Todo{
		{Description: "Task 1", Priority: 1, ProjectID: "work"},
		{Description: "Task 2", Priority: 2, ProjectID: "personal"},
	}

	for _, td := range todos {
		item := td
		if _, err := repo.Create(&item); err != nil {
			t.Fatalf("failed to create todo: %v", err)
		}
	}

	all, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(all) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(all))
	}
	if all[0].ProjectID != "work" || all[1].ProjectID != "personal" {
		t.Errorf("unexpected project IDs: %v, %v", all[0].ProjectID, all[1].ProjectID)
	}
}

func TestTodoRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewTodoRepository(db)

	todo := &model.Todo{
		Description: "Initial Task",
		Priority:    3,
		ProjectID:   "todos",
	}
	id, err := repo.Create(todo)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	t.Run("update project_id and description", func(t *testing.T) {
		updateTodo := &model.Todo{
			ID:          id,
			Description: "Updated Task",
			Priority:    2,
			ProjectID:   "shopping",
		}
		if err := repo.Update(updateTodo); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		got, err := repo.GetById(id)
		if err != nil {
			t.Fatalf("GetById() error = %v", err)
		}
		if got.Description != "Updated Task" {
			t.Errorf("got description %q, want %q", got.Description, "Updated Task")
		}
		if got.ProjectID != "shopping" {
			t.Errorf("got project_id %q, want %q", got.ProjectID, "shopping")
		}
		if got.Priority != 2 {
			t.Errorf("got priority %d, want 2", got.Priority)
		}
	})

	t.Run("update without specifying project_id preserves stored project_id", func(t *testing.T) {
		updateTodo := &model.Todo{
			ID:          id,
			Description: "Updated Again",
			Priority:    4,
		}
		if err := repo.Update(updateTodo); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		got, err := repo.GetById(id)
		if err != nil {
			t.Fatalf("GetById() error = %v", err)
		}
		if got.ProjectID != "shopping" {
			t.Errorf("got project_id %q, want preserved %q", got.ProjectID, "shopping")
		}
	})
}

func TestTodoRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewTodoRepository(db)

	todo := &model.Todo{
		Description: "To be deleted",
		Priority:    5,
	}
	id, err := repo.Create(todo)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Delete(id); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = repo.GetById(id)
	if err == nil {
		t.Fatal("expected error getting deleted todo, got nil")
	}
}

