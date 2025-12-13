package controllers

import (
	"fmt"
	"net/http"
)

func ListTodos(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "TODO api v1 (list)\n")
}

func CreateTodo() {}
