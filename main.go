package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/justasandbox/my-todo-cli/db"
	"github.com/justasandbox/my-todo-cli/model"
)

func dbPath() string {
	if p := os.Getenv("TODO_DB_PATH"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".todo.db"
	}
	return home + "/.todo.db"
}

func main() {
	conn, err := db.Open(dbPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	repo := db.NewRepository(conn)
	m := model.New(repo)

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
