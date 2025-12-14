package service

import (
	"errors"

	"github.com/Okwonks/go-todo/internal/model"
	"github.com/Okwonks/go-todo/internal/repository"
)

type TodoService struct {
	repo repository.TodoRepository
}

func NewTodoService(r repository.TodoRepository) *TodoService {
	return &TodoService{r}
}

func (s *TodoService) Create(todo *model.Todo) (int64, error) {
	if todo.Description == "" {
		return 0, errors.New("Description of the todo is required")
	}
	return s.repo.Create(todo)
}

func (s *TodoService) GetAll() ([]model.Todo, error) {
	return s.repo.GetAll()
}

func (s *TodoService) GetById(id int64) (*model.Todo, error) {
	return s.repo.GetById(id)
}

func (s *TodoService) Update(u *model.Todo) error {
	return s.repo.Update(u)
}

func (s *TodoService) Delete(id int64) error {
	return s.repo.Delete(id)
}
