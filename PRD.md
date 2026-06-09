# Todo TUI - Product Requirements Document

## Vision

A fast, keyboard-driven terminal Todo application that stays out of your way. No accounts, no cloud sync, no bloat — just a local SQLite-backed task manager with vim-style keybindings and a clean, color-coded interface built on Bubbletea.

## Problem Statement

Existing todo tools either live in the browser (slow context-switch from the terminal), require cloud accounts, or lack the keyboard ergonomics that terminal-native developers expect. A TUI todo app that speaks vim and stores everything locally fills this gap.

## Target User

Developers and terminal-power-users who live in the command line and want task management without leaving it.

## Architecture

```
┌─────────────────────────────────────────────┐
│                 TUI Layer                    │
│          (Bubbletea / Lipgloss)             │
│                                             │
│  ┌─────────┐  ┌─────────┐  ┌────────────┐  │
│  │  Today   │  │   All   │  │ Completed  │  │
│  │   Tab    │  │   Tab   │  │    Tab     │  │
│  └─────────┘  └─────────┘  └────────────┘  │
│                     │                        │
│            ┌────────┴────────┐              │
│            │  Todo Model     │              │
│            │  (state, keys,  │              │
│            │   filtering)    │              │
│            └────────┬────────┘              │
│                     │                        │
├─────────────────────┼───────────────────────┤
│              Data Layer                      │
│            ┌────────┴────────┐              │
│            │   Repository    │              │
│            │  (CRUD ops)     │              │
│            └────────┬────────┘              │
│                     │                        │
│            ┌────────┴────────┐              │
│            │     SQLite      │              │
│            │  (~/.todo.db)   │              │
│            └─────────────────┘              │
└─────────────────────────────────────────────┘
```

### Component Breakdown

**TUI Layer** — Bubbletea `tea.Model` implementation with Lipgloss styling. Handles input, rendering, and view switching between tabs.

**Todo Model** — Core application state: the task list, cursor position, active tab, and input mode. All filtering (today/completed/all) happens here.

**Repository** — Thin data-access layer over SQLite. Exposes `Create`, `List`, `Update`, `Delete`. Keeps SQL out of the TUI code.

**SQLite** — Single-file database at `~/.todo.db`. One `todos` table.

## Tech Stack

| Component       | Choice                         | Rationale                                                  |
|-----------------|--------------------------------|------------------------------------------------------------|
| Language        | Go                             | Fast compilation, single binary, strong TUI ecosystem      |
| TUI Framework   | [Bubbletea](https://github.com/charmbracelet/bubbletea) | Elm-architecture for terminals; composable, testable |
| Styling         | [Lipgloss](https://github.com/charmbracelet/lipgloss)   | Declarative terminal styling, pairs with Bubbletea   |
| Text Input      | [Bubbles](https://github.com/charmbracelet/bubbles)     | Pre-built text input component for the add-task flow |
| Database        | SQLite via [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) | Pure-Go SQLite, no CGO required — simplifies cross-compilation |
| DB Access       | `database/sql`                 | Standard library; no ORM overhead for one table            |

## Data Model

### `todos` Table

| Column       | Type      | Description                              |
|-------------|-----------|------------------------------------------|
| `id`        | INTEGER   | Primary key, autoincrement               |
| `title`     | TEXT      | Task description (required)              |
| `done`      | BOOLEAN   | Completion status, default `false`       |
| `due_date`  | TEXT      | ISO-8601 date (`YYYY-MM-DD`), nullable   |
| `created_at`| DATETIME  | Row creation timestamp                   |
| `updated_at`| DATETIME  | Last modification timestamp              |

## Keybindings

| Key          | Action                              |
|-------------|-------------------------------------|
| `j` / `Down` | Move cursor down                   |
| `k` / `Up`   | Move cursor up                     |
| `a`          | Add new task (opens inline input)   |
| `Space`      | Toggle task completion              |
| `d`          | Delete task under cursor            |
| `1`          | Switch to Today tab                 |
| `2`          | Switch to All tab                   |
| `3`          | Switch to Completed tab             |
| `Tab`        | Cycle to next tab                   |
| `Enter`      | Confirm input (when adding a task)  |
| `Esc`        | Cancel input / exit add mode        |
| `q`          | Quit                                |

## Views (Tabs)

1. **Today** — Tasks with `due_date` equal to today, not yet completed. Default view on launch.
2. **All** — Every incomplete task, sorted by due date (soonest first), then by creation date.
3. **Completed** — All tasks where `done = true`, most recently completed first.

## Color Coding

| Condition              | Color  |
|-----------------------|--------|
| Overdue (`due_date < today` and not done) | Red    |
| Due today (`due_date == today`)           | Yellow |
| Completed tasks                           | Dim / strikethrough |
| Default                                   | White  |

## UI Layout

```
 Todo ─────────────────────────────────────
 [ Today ] [ All ] [ Completed ]

   ● Buy groceries              2026-06-07   <- red (overdue)
   ● Finish PR review           2026-06-08   <- yellow (today)
 > ○ Write tests                2026-06-10
   ○ Update README              —

 j/k: navigate  a: add  space: toggle  d: delete  q: quit
```

- `●` = incomplete, `○` = completed (dimmed)
- `>` = cursor position
- Bottom bar shows contextual help

## Non-Goals (v1)

- Cloud sync or multi-device support
- Task priorities or tags
- Subtasks or dependencies
- Recurring tasks
- Edit-in-place for existing task titles
- Mouse support
