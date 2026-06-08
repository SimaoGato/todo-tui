package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/justasandbox/my-todo-cli/model"
)

const (
	queryAll = `
		SELECT id, title, done, due_date, created_at, updated_at FROM todos
		WHERE done = 0
		ORDER BY CASE WHEN due_date IS NULL THEN 1 ELSE 0 END, due_date, created_at`

	queryToday = `
		SELECT id, title, done, due_date, created_at, updated_at FROM todos
		WHERE done = 0 AND DATE(due_date) = DATE('now', 'localtime')
		ORDER BY created_at`

	queryDone = `
		SELECT id, title, done, due_date, created_at, updated_at FROM todos
		WHERE done = 1
		ORDER BY updated_at DESC`
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(filter model.Filter) ([]model.Todo, error) {
	var query string
	switch filter {
	case model.FilterToday:
		query = queryToday
	case model.FilterDone:
		query = queryDone
	default:
		query = queryAll
	}

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}
	defer rows.Close()

	var todos []model.Todo
	for rows.Next() {
		var t model.Todo
		var dueDateStr *string
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &dueDateStr, &createdAtStr, &updatedAtStr); err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		if dueDateStr != nil {
			var d time.Time
			var parseErr error
			for _, layout := range []string{"2006-01-02", time.RFC3339, "2006-01-02T15:04:05Z07:00"} {
				d, parseErr = time.Parse(layout, *dueDateStr)
				if parseErr == nil {
					break
				}
			}
			if parseErr != nil {
				return nil, fmt.Errorf("parse due_date: %w", parseErr)
			}
			d = d.Truncate(24 * time.Hour)
			t.DueDate = &d
		}
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		todos = append(todos, t)
	}
	if todos == nil {
		todos = []model.Todo{}
	}
	return todos, rows.Err()
}

func (r *Repository) ToggleDone(id int) error {
	res, err := r.db.Exec(
		`UPDATE todos SET done = NOT done, updated_at = ? WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), id,
	)
	if err != nil {
		return fmt.Errorf("toggle done: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("todo %d not found", id)
	}
	return nil
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
