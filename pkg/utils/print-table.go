package utils

import (
	"fmt"
	"reflect"
	"strings"

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
	rows := getRows(p.todos)

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}

	for _, row := range rows {
		for i, val := range row {
	    if len(val) > colWidths[i] {
	      colWidths[i] = len(val)
	    }
	  }
	}

	var rowParts []string
	for _, width := range colWidths {
		rowParts = append(rowParts, fmt.Sprintf(" %%-%ds ", width))
	}
	rowFormat := strings.Join(rowParts, "|")

	headerVals := make([]any, len(headers))
	for i, v := range headers {
		headerVals[i] = v
	}
	fmt.Printf(rowFormat + "\n", headerVals...)

	var separator []string
	for _, width := range colWidths {
		separator = append(separator, strings.Repeat("-", width + 2))
	}
	fmt.Println(strings.Join(separator, "+"))

	for _, row := range rows {
		rowVals := make([]any, len(row))
		for i, v := range row {
			rowVals[i] = v
		}
		fmt.Printf(rowFormat + "\n", rowVals...)
	}

	fmt.Printf("(%d rows) \n", len(rows))
}

func getHeaders(todos []model.Todo) []string {
	var headers []string

	extractJSONName := func (fieldName, tag string) string {
		if tag == "-" {
			return "-"
		}
		if tag == "" {
			return fieldName
		}
		parts := strings.Split(tag, ",")
		name := parts[0]
		if name == "" {
			return fieldName
		}
		return  name
	}
	
	todo := todos[0]
	v := reflect.ValueOf(todo)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		headers = append(headers, extractJSONName(field.Name, tag))
	}
	return headers
}

func getRows(todos []model.Todo) [][]string {
	var rows [][]string

	v := reflect.ValueOf(todos)
	for i := 0; i < v.Len(); i++ {
		var row []string
		item := v.Index(i)
		for j := 0; j < item.NumField(); j++ {
			row = append(row, fmt.Sprintf("%v", item.Field(j).Interface()))
		}
		rows = append(rows, row)
	}
	return  rows
}
