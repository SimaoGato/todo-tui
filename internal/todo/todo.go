package todo

import "time"

type Task struct {
	ID        int
	Title     string
	Done      bool
	DueDate   *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Filter int

const (
	FilterAll Filter = iota
	FilterToday
	FilterDone
)
