package v1

import (
	"database/sql"
	"log"
	"net/http"
	"time"

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

	mux.Handle("/", http.NotFoundHandler())

	return logger(mux)
}

func logger(next http.Handler) http.Handler {
	return  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Printf("Received Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.Printf("Completed Request: %s %s took %v", r.Method, r.URL.Path, duration)
	})
}
