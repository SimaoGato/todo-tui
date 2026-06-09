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
cmd/my-todo-cli/main.go   → entry point: wires db.Open → db.NewRepository → ui.New → tea.Run
internal/todo/            → shared types: Todo struct, Filter enum (FilterAll/FilterToday/FilterDone)
internal/db/              → SQLite: Open, RunMigrations, Repository (Create/List/ToggleDone/Delete)
internal/ui/              → Bubbletea Model: AppModel, Update (key handling), View (rendering)
```

## Why — Architectural boundaries

- **No cloud, no accounts, no sync.** Everything is local SQLite at `~/.todo.db`.
- **No new packages** unless a story explicitly requires one. Views live in `internal/ui/view.go`.
- **Repo interface** (`ui.Repo`) decouples TUI from DB. `db.Repository` satisfies it implicitly.
- **Shared types** live in `internal/todo/` — never in `internal/db/` or `internal/ui/`.
- **Migrations** are numbered structs in `internal/db/migrations.go` with a `schema_migrations` table. Never raw DDL in `Open()`.
- **No mouse support, no subtasks, no priorities, no recurring tasks** (v1 non-goals in PRD).

## How — Commands

```bash
go build -o my-todo-cli ./cmd/my-todo-cli  # build
go vet ./...                               # required before commit
go test ./... -race -count=1               # all tests with race detector (must pass before commit)
go test -cover -race ./...                 # with coverage
./my-todo-cli                              # run (uses ~/.todo.db)
TODO_DB_PATH=/tmp/test.db ./my-todo-cli    # custom db path
```

## CI

`.github/workflows/ci.yml` runs on every push and PR to `main`:
steps are `go build ./cmd/my-todo-cli` → `go vet ./...` → `go test ./... -race -count=1`.
The pipeline must stay green. The `-race` flag is critical — the async `loadTodos` pattern is concurrent.

## Workflow rules

1. **Never implement without a linked story** in `TICKETS.md`.
2. Read `PRD.md` + relevant story before coding.
3. **TDD**: write a failing test first, then implement.
4. Story acceptance criteria = definition of done. Nothing more.
5. No refactoring outside story scope.
6. `go vet ./...` and `go test ./... -race -count=1` must pass before any commit.
7. Check `internal/db/migrations.go` before any schema change — never hallucinate columns.

## Key files to read before coding

- `PRD.md` — full product spec, keybindings, data model, UI layout
- `TICKETS.md` — epics and stories
- `internal/db/migrations.go` — current schema (single `todos` table)
- `internal/db/repository.go` — CRUD surface
- `internal/ui/model.go` — TUI state + `Update()` key handling
- `internal/ui/view.go` — rendering + Lipgloss styles
