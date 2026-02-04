package v1

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Okwonks/go-todo/api/v1/controllers"
	"github.com/Okwonks/go-todo/internal/repository"
	"github.com/Okwonks/go-todo/internal/service"
)

func Router(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	todoRepo := repository.NewTodoRepository(db)
	todoService := service.NewTodoService(todoRepo)
	todoControllers := controllers.NewTodoController(todoService)
	todoControllers.RegisterTodos(mux)

	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))

	mux.Handle("/", http.NotFoundHandler())

	return mux
}
