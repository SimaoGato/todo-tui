package db

import (
	"testing"

	"github.com/justasandbox/my-todo-cli/model"
)

func TestDelete_RemovesRow(t *testing.T) {
	repo := openTestDB(t)

	todo, _ := repo.Create("to be deleted", nil)
	if err := repo.Delete(todo.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	todos, err := repo.List(model.FilterAll)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, td := range todos {
		if td.ID == todo.ID {
			t.Errorf("deleted todo (id=%d) still returned by List", todo.ID)
		}
	}
}

func TestDelete_NotVisibleInAnyFilter(t *testing.T) {
	repo := openTestDB(t)

	todo, _ := repo.Create("gone for good", nil)
	repo.Delete(todo.ID)

	for _, f := range []model.Filter{model.FilterAll, model.FilterToday, model.FilterDone} {
		todos, err := repo.List(f)
		if err != nil {
			t.Fatalf("List(%v): %v", f, err)
		}
		for _, td := range todos {
			if td.ID == todo.ID {
				t.Errorf("deleted todo still visible under filter %v", f)
			}
		}
	}
}

func TestDelete_NonexistentID(t *testing.T) {
	repo := openTestDB(t)
	if err := repo.Delete(99999); err == nil {
		t.Error("expected error for nonexistent ID, got nil")
	}
}

func TestDelete_OnlyRemovesTargetRow(t *testing.T) {
	repo := openTestDB(t)

	keep, _ := repo.Create("keep me", nil)
	remove, _ := repo.Create("remove me", nil)

	if err := repo.Delete(remove.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	todos, _ := repo.List(model.FilterAll)
	if len(todos) != 1 || todos[0].ID != keep.ID {
		t.Errorf("expected only 'keep me' (id=%d) to remain, got %v", keep.ID, todos)
	}
}
