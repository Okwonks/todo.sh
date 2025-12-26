package repository

import (
	"database/sql"
	"log"
	"reflect"

	"github.com/Okwonks/go-todo/internal/model"
)

type TodoRepository interface {
	Create(t *model.Todo) (int64, error)
	GetAll() ([]model.Todo, error)
	GetById(id int64) (*model.Todo, error)
	Update(t *model.Todo) error
	Delete(id int64) error
}

type todoRepo struct {
	db *sql.DB
}

func NewTodoRepository(db *sql.DB) TodoRepository {
	return &todoRepo{db}
}

func (r *todoRepo) Create(t *model.Todo) (int64, error) {
	res, err := r.db.Exec("INSERT INTO todo(description) VALUES(?)", t.Description)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *todoRepo) GetAll() ([]model.Todo, error) {
	rows, err := r.db.Query("SELECT id, created_at, description, priority, status, due_date FROM todo")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []model.Todo{}
	for rows.Next() {
		var todo model.Todo
		err := rows.Scan(
			&todo.ID,
			&todo.CreatedAt,
			&todo.Description,
			&todo.Priority,
			&todo.Status,
			&todo.DueDate,
		)
		if err != nil {
			log.Printf("ERROR: Scan error: %v", err)
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, nil
}

func (r *todoRepo) GetById(id int64) (*model.Todo, error) {
	var todo model.Todo
	err := r.db.QueryRow("SELECT id, created_at, description, priority, status, due_date FROM todo WHERE id = ?", id).
		Scan(&todo.ID, &todo.CreatedAt, &todo.Description, &todo.Priority, &todo.Status, &todo.DueDate)
	if err != nil {
		log.Printf("ERROR: Scan error: %v", err)
		return nil, err
	}

	return &todo, nil
}

func (r *todoRepo) Update(todo *model.Todo) error {
	stored, err := r.GetById(todo.ID)

	v := reflect.ValueOf(todo) 
	s := reflect.ValueOf(stored)
	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Name
		if fieldName != "Description" && fieldName != "Status" {
	    continue
		}

		todoField := v.FieldByName(fieldName)
		storedField := s.FieldByName(fieldName)
		if todoField.Kind() == reflect.String && todoField.String() == "" {
			if todoField.CanSet() {
				todoField.SetString(storedField.String())
			}
		}
	}

	if todo.DueDate == nil {
		todo.DueDate = stored.DueDate
	}

	// TODO: there's probably a better way to do this :)
	if todo.Priority <= 0 || todo.Priority > 5 {
		todo.Priority = stored.Priority
	}

	_, err = r.db.Exec(
		`UPDATE todo SET description=?, due_date=?, priority=?, status=? WHERE id=?`,
		todo.Description, todo.DueDate, todo.Priority, todo.Status, todo.ID,
	)
	return err
}

func (r* todoRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM todo WHERE id=?", id)
	return err
}
