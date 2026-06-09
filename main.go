package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/justasandbox/my-todo-cli/db"
	"github.com/justasandbox/my-todo-cli/model"
	"github.com/justasandbox/my-todo-cli/todo"
)

func dbPath() string {
	if p := os.Getenv("TODO_DB_PATH"); p != "" {
		if abs, err := filepath.Abs(p); err == nil {
			return abs
		}
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

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	am := finalModel.(model.AppModel)
	summary := am.GetSummary()

	todayTasks, _ := repo.List(todo.FilterToday)
	summary.TodayRemaining = len(todayTasks)

	printSummary(summary)
}

func printSummary(s model.Summary) {
	if s.Completed == 0 && s.Deleted == 0 && s.TodayRemaining == 0 {
		fmt.Println("Nothing to report.")
		return
	}
	fmt.Println("Session summary:")
	fmt.Printf("  Completed today : %d\n", s.Completed)
	fmt.Printf("  Deleted         : %d\n", s.Deleted)
	fmt.Printf("  Still due today : %d\n", s.TodayRemaining)
}
