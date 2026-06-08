package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/justasandbox/my-todo-cli/model"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(title string, dueDate *time.Time) (model.Todo, error) {
	now := time.Now().UTC()

	var dueDateStr *string
	if dueDate != nil {
		s := dueDate.UTC().Format("2006-01-02")
		dueDateStr = &s
	}

	res, err := r.db.Exec(
		`INSERT INTO todos (title, done, due_date, created_at, updated_at)
		 VALUES (?, 0, ?, ?, ?)`,
		title, dueDateStr, now.Format(time.RFC3339), now.Format(time.RFC3339),
	)
	if err != nil {
		return model.Todo{}, fmt.Errorf("create todo: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return model.Todo{}, fmt.Errorf("last insert id: %w", err)
	}

	return model.Todo{
		ID:        int(id),
		Title:     title,
		Done:      false,
		DueDate:   dueDate,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
