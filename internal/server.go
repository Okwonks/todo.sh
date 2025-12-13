package internal

import (
	"encoding/json"
	"net/http"

	v1 "github.com/Okwonks/go-todo/api/v1"
)

type health struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	GitDescription string `json:"git_description"`
}

// TODO: pass config object to handle ports and other
// configs
func Server() *http.Server {
	apiv1 := v1.Router()

	mux := http.NewServeMux()

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiv1))

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		res := &health{
			Name: "TODO API",
			Version: "v1.0.0",
			GitDescription: "...",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	srv := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	return srv
}
