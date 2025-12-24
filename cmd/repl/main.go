package main

import (
	"fmt"
	"strings"

	"github.com/Okwonks/go-todo/internal/client"
	"github.com/Okwonks/go-todo/pkg/utils"
	"github.com/c-bata/go-prompt"
)

var apiClient = client.NewClient("http://localhost:8080")

func help() {
	  fmt.Println(`
Commands:
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
	  fmt.Println("...exiting")
	  panic("exit")
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
