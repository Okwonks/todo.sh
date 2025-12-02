package main

import (
	"fmt"
	"log"
	"net/http"
)

func todoHandler() *http.ServeMux {
	router := http.NewServeMux()

	router.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "TODO api v1\n")
	}))

	router.Handle("/", http.NotFoundHandler())

	return router;
}

type HealthHandler struct{}
func (h *HealthHandler) handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Method:", r.Method)
	w.Write([]byte("OK"))
}

func main() {
	api := http.NewServeMux()
	api.Handle("/todos", todoHandler())

	mux := http.NewServeMux()

	health := &HealthHandler{}

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))
	mux.HandleFunc("GET /health", health.handler)
	mux.Handle("/", http.NotFoundHandler())

	srv := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	fmt.Printf("TODO api v1 listening on port %s\n", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
