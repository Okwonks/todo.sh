package utils

import (
	"fmt"
	"reflect"

	"github.com/Okwonks/go-todo/internal/model"
)

type PrintTable interface {
	Standard()
}

type printTable struct {
	todos []model.Todo
}

func NewPrintTable(todos []model.Todo) PrintTable {
	return &printTable{todos}
}

func (p *printTable) Standard() {
	if len(p.todos) == 0 {
		fmt.Println("Empty table")
		return
	}

	headers := getHeaders(p.todos)
	fmt.Println("headers: ", headers)
}

func getHeaders(todos []model.Todo) []string {
	var headers []string
	
	todo := todos[0]
	v := reflect.ValueOf(todo)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		headers = append(headers, t.Field(i).Name)
	}
	return headers
}
