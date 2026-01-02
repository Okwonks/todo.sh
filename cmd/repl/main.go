package main

import (
	"fmt"
	"strings"

	"github.com/Okwonks/go-todo/internal/client"
	"github.com/Okwonks/go-todo/internal/model"
	"github.com/Okwonks/go-todo/pkg/utils"
	"github.com/c-bata/go-prompt"
)

var apiClient = client.NewClient("http://localhost:8080")

func help() {
	  fmt.Println(`
Commands:
  create [flags] <description>
  list
  help
  exit`)
}

func executor(input string) {
	input = strings.TrimSpace(input)
	parts := strings.Split(input, " ")

	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "exit", "quit":
	  panic("exit")
	case "create":
		remaining := parts[1:]
		if len(remaining) == 0 {
			fmt.Println("A description is required to add a task")
			return
		}

		// TODO: add support on setting flags and other task properties
		// when creating a new task
		todo := model.Todo{Description: strings.Join(remaining, " ")}

		t, err := apiClient.CreateTodo(&todo)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Println("Task added:", t.Description, t.Priority, t.Status)
	  return
	case "list":
	  todos, err := apiClient.List()
	  if err != nil {
			fmt.Println("Error:", err)
	    return
		} 
	  printTable := utils.NewPrintTable(todos)
	  printTable.Standard()
	  return
	case "help":
	  help()
	  return
	}

	fmt.Println("Unknow Command:", input)
	fmt.Println("")
	help()
}

func completer(in prompt.Document) []prompt.Suggest {
	suggestions := []prompt.Suggest{}
	return prompt.FilterHasPrefix(suggestions, in.GetWordAfterCursor(), true)
}

func main() {
	fmt.Println("[ GetItDone ]")
	fmt.Println("Type 'help' for commands.")

	defer func() {
		recover()
	}()

	prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("get-it-done> "),
		prompt.OptionTitle("get-it-done"),
	).Run()
}
