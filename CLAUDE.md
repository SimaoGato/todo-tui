# CLAUDE.md — my-todo-cli

## What

Keyboard-driven terminal Todo app. Single binary, local-only, vim keybindings.

| Tech | Choice |
|------|--------|
| Language | Go |
| TUI | Bubbletea (Elm architecture) |
| Styling | Lipgloss |
| Input | Bubbles `textinput` |
| DB | SQLite via `modernc.org/sqlite` (pure Go, no CGO) |
| DB access | `database/sql` — no ORM |

### Package layout

```
main.go       → entry point: wires db.Open → db.NewRepository → model.New → tea.Run
todo/         → shared types: Todo struct, Filter enum (FilterAll/FilterToday/FilterDone)
db/           → SQLite: Open, RunMigrations, Repository (Create/List/ToggleDone/Delete)
model/        → Bubbletea Model: AppModel, Update (key handling), View (rendering)
```

## Why — Architectural boundaries

- **No cloud, no accounts, no sync.** Everything is local SQLite at `~/.todo.db`.
- **No new packages** unless a story explicitly requires one. No `ui/` package — views live in `model/view.go`.
- **Repo interface** (`model.Repo`) decouples TUI from DB. `db.Repository` satisfies it implicitly.
- **Shared types** live in `todo/` — never in `db/` or `model/`.
- **Migrations** are numbered structs in `db/migrations.go` with a `schema_migrations` table. Never raw DDL in `Open()`.
- **No mouse support, no subtasks, no priorities, no recurring tasks** (v1 non-goals in PRD).

## How — Commands

```bash
go build -o my-todo-cli .   # build
go test ./...               # all tests (must pass before commit)
go test -cover ./...        # coverage check
./my-todo-cli               # run (uses ~/.todo.db)
TODO_DB_PATH=/tmp/test.db ./my-todo-cli  # custom db path
```

## Workflow rules

1. **Never implement without a linked story** in `TICKETS.md`.
2. Read `PRD.md` + relevant story before coding.
3. **TDD**: write a failing test first, then implement.
4. Story acceptance criteria = definition of done. Nothing more.
5. No refactoring outside story scope.
6. `go test ./...` must pass before any commit.
7. Check `db/migrations.go` before any schema change — never hallucinate columns.

## Key files to read before coding

- `PRD.md` — full product spec, keybindings, data model, UI layout
- `TICKETS.md` — epics and stories
- `db/migrations.go` — current schema (single `todos` table)
- `db/repository.go` — CRUD surface
- `model/model.go` — TUI state + `Update()` key handling
- `model/view.go` — rendering + Lipgloss styles
