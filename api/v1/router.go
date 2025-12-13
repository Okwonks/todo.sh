package v1

import (
	"net/http"

	"github.com/Okwonks/go-todo/api/v1/controllers"
)

func Router() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /todos", http.HandlerFunc(controllers.ListTodos))

	mux.Handle("/", http.NotFoundHandler())

	return mux
}
