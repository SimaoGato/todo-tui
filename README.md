# my-todo-cli

A keyboard-driven terminal todo app for developers who live in the terminal.

No cloud. No accounts. No bloat. Just a local SQLite-backed task manager with vim-style keybindings, built on [Bubbletea](https://github.com/charmbracelet/bubbletea).

## Features

- Three views: **Today**, **All**, **Completed**
- Color-coded urgency: red (overdue), yellow (due today), dim (done)
- Inline task creation with optional due date (`YYYY-MM-DD`)
- Session summary on exit: tasks completed and deleted this run
- Single binary, no CGO, no external dependencies at runtime

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `k` | Move cursor down / up |
| `a` | Add new task |
| `Space` | Toggle completion |
| `d` | Delete task |
| `1` / `2` / `3` | Switch to Today / All / Completed |
| `Tab` | Cycle tabs |
| `Esc` | Cancel input |
| `q` | Quit |

## Installation

```bash
git clone https://github.com/justasandbox/my-todo-cli
cd my-todo-cli
go build -o my-todo-cli .
./my-todo-cli
```

Requires Go 1.21+. The binary is self-contained — no runtime dependencies.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `TODO_DB_PATH` | `~/.todo.db` | Path to the SQLite database |

## Development

```bash
go build ./...          # compile
go test ./...           # run tests
go test -race ./...     # run tests with race detector (always before commit)
```

## Tech Stack

| Layer | Choice |
|-------|--------|
| Language | Go |
| TUI | [Bubbletea](https://github.com/charmbracelet/bubbletea) |
| Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) |
| Input | [Bubbles](https://github.com/charmbracelet/bubbles) |
| Database | SQLite via [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO) |

## License

MIT
