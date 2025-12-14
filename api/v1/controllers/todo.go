package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Okwonks/go-todo/internal/model"
	"github.com/Okwonks/go-todo/internal/service"
)

type TodoHandler struct {
	service *service.TodoService
}

func NewTodoController(s *service.TodoService) *TodoHandler {
	return &TodoHandler{s}
}

func (h *TodoHandler) RegisterTodos(mux *http.ServeMux) http.Handler {
	mux.Handle("GET /todos", http.HandlerFunc(h.List))
	mux.Handle("GET /todos/{id}", http.HandlerFunc(h.FindById))
	mux.Handle("POST /todos", http.HandlerFunc(h.Create))

	return mux
}

func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	todo := model.Todo{}
	json.NewDecoder(r.Body).Decode(&todo)

	id, err := h.service.Create(&todo)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	todo.ID = id
	json.NewEncoder(w).Encode(todo)
}

func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	todos, _ := h.service.GetAll()
	json.NewEncoder(w).Encode(todos)
}

func (h *TodoHandler) FindById(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	todo, err := h.service.GetById(int64(id))
	if err != nil {
		http.Error(w, "Not found", 404)
		return
	}
	json.NewEncoder(w).Encode(todo)
}
