h2. Epic 1: Project Scaffolding

Set up the Go module, dependencies, and project structure so all subsequent work has a foundation to build on.

---

h3. 1.1 Initialize Go module and dependencies

*Estimation:* 2h

h4. User Story
As a developer I want to initialize a Go module with core TUI dependencies in order to establish the project foundation

h4. Acceptance Criteria
* AC1: @go mod init github.com/user/my-todo-cli@ succeeds without errors
* AC2: Dependencies added: @bubbletea@, @lipgloss@, @bubbles@, @modernc.org/sqlite@
* AC3: Directory structure created: @main.go@, @db/@, @model/@, @ui/@
* AC4: @go build@ compiles without errors

h4. Implementation Details
* Create directory structure at project root
* Use @go get@ to add each dependency to @go.mod@
* Verify @go.sum@ is generated correctly
* No application logic yet, just module setup

h4. Operational Risks
* None for scaffolding phase

h4. Out of Scope
* Application logic
* Database initialization
* TUI rendering

---

h3. 1.2 Create application entry point

*Estimation:* 3h

h4. User Story
As a developer I want to bootstrap a Bubbletea application in @main.go@ in order to confirm the TUI framework is properly configured

h4. Acceptance Criteria
* AC1: @main.go@ initializes a Bubbletea program and calls @tea.Run()@
* AC2: Running the compiled binary displays "Hello, Todo" text
* AC3: Application exits cleanly when pressing @q@ or @Ctrl+C@
* AC4: No panics or runtime errors on startup

h4. Implementation Details
* Create @main.go@ that calls @tea.NewProgram(initialModel).Run()@
* Define a minimal @model@ struct with @Init()@, @Update(msg tea.Msg)@, and @View()@ methods
* @View()@ returns a simple string with "Hello, Todo"
* Handle @tea.KeyMsg@ in @Update()@ for quit logic

h4. Operational Risks
* Terminal might require specific terminal emulator features; test on multiple terminals if issues arise

h4. Out of Scope
* Database connectivity
* Complex UI rendering
* Tab or navigation system

---

h2. Epic 2: Data Layer (SQLite)

Implement the SQLite database and a repository that provides CRUD operations for todos.

---

h3. 2.1 Database initialization and migrations

*Estimation:* 3h

h4. User Story
As a developer I want to initialize an SQLite database on first run in order to persist todo data

h4. Acceptance Criteria
* AC1: On application startup, @~/.todo.db@ is created if it does not exist
* AC2: The @todos@ table is created with columns: @id@ (primary key), @title@ (text), @done@ (boolean), @due_date@ (date nullable), @created_at@ (timestamp), @updated_at@ (timestamp)
* AC3: Subsequent runs detect the existing database and do not recreate it
* AC4: @.todo.db@ is readable and contains valid SQLite schema

h4. Implementation Details
* Create @db/db.go@ with an @Open(path string)@ function
* Use @modernc.org/sqlite@ driver with @database/sql@
* Execute @CREATE TABLE IF NOT EXISTS@ query in @Open()@
* Default path is @~/.todo.db@; support override via environment variable in Epic 6

h4. Operational Risks
* Database migration strategy must be backwards-compatible once more schemas are added; use @IF NOT EXISTS@ for all table/column creation
* File permissions on @~/.todo.db@ may vary by OS; ensure readable/writable

h4. Out of Scope
* Column constraints (indexes, foreign keys)
* Backup or restore functionality
* Multi-user concurrency

---

h3. 2.2 Repository: Create a todo

*Estimation:* 2h

h4. User Story
As a developer I want to implement a repository method to insert todos in order to support todo creation

h4. Acceptance Criteria
* AC1: @Create(title string, dueDate *time.Time)@ returns a @Todo@ struct with auto-generated @id@
* AC2: @created_at@ and @updated_at@ are set to the current time at insert
* AC3: @done@ defaults to @false@
* AC4: Unit test creates a todo, queries it back, and confirms all fields match

h4. Implementation Details
* Create @model/todo.go@ with a @Todo@ struct (id, title, done, dueDate, createdAt, updatedAt)
* Create @db/repository.go@ with a @Repository@ struct holding a @*sql.DB@
* Implement @(r *Repository) Create(title string, dueDate *time.Time) (Todo, error)@
* Use @INSERT INTO todos@ and return the inserted row via @LAST_INSERT_ROWID()@

h4. Operational Risks
* SQLite AUTOINCREMENT behavior differs from other databases; use default @ROWID@ for simplicity

h4. Out of Scope
* Bulk insert
* Transaction handling
* Duplicate detection

---

h3. 2.3 Repository: List todos with filtering

*Estimation:* 4h

h4. User Story
As a developer I want to implement filtered list queries in order to support multiple task views

h4. Acceptance Criteria
* AC1: @Filter@ enum supports @All@, @Today@, @Done@
* AC2: @List(filter Filter)@ returns @[]Todo@ ordered correctly per filter spec:
** @All@: incomplete tasks, ordered by @due_date@ (nulls last), then @created_at@
** @Today@: incomplete tasks where @due_date = TODAY()@, ordered by @created_at@
** @Done@: all completed tasks, ordered by @updated_at@ descending
* AC3: Unit tests verify correct ordering and filtering for each @Filter@ variant
* AC4: Empty list is handled gracefully (returns empty slice, no error)

h4. Implementation Details
* Create @Filter@ type in @model/todo.go@ as an enum (iota or string-based)
* Implement @(r *Repository) List(filter Filter) ([]Todo, error)@
* Build dynamic SQL @WHERE@ clauses per filter
* Use @ORDER BY@ with appropriate columns
* For @Today@, use @DATE(due_date) = DATE('now')@ in SQLite

h4. Operational Risks
* SQLite date functions may behave differently in different timezones; consider storing all dates in UTC
* @NULL@ ordering differs between databases; test explicitly

h4. Out of Scope
* Search or text filtering
* Pagination
* Aggregation or grouping

---

h3. 2.4 Repository: Toggle completion

*Estimation:* 2h

h4. User Story
As a developer I want to toggle a todo's completion status in order to mark tasks as done/undone

h4. Acceptance Criteria
* AC1: @ToggleDone(id int)@ flips the @done@ field from @true@ to @false@ or vice versa
* AC2: @updated_at@ is set to the current time
* AC3: Unit test toggles a task twice and confirms the state changes each time
* AC4: Toggling a nonexistent ID returns an error

h4. Implementation Details
* Implement @(r *Repository) ToggleDone(id int) error@
* Use @UPDATE todos SET done = NOT done, updated_at = CURRENT_TIMESTAMP WHERE id = ?@
* Check affected rows to detect missing IDs; return error if none

h4. Operational Risks
* None specific to this feature

h4. Out of Scope
* Bulk toggle
* Undo/redo functionality

---

h3. 2.5 Repository: Delete a todo

*Estimation:* 1h

h4. User Story
As a developer I want to delete todos from the database in order to remove unwanted tasks

h4. Acceptance Criteria
* AC1: @Delete(id int)@ removes the todo row by id
* AC2: Unit test deletes a row and confirms it is no longer returned by @List()@
* AC3: Deleting a nonexistent ID returns an error (or no-op, to be decided)

h4. Implementation Details
* Implement @(r *Repository) Delete(id int) error@
* Use @DELETE FROM todos WHERE id = ?@
* Check affected rows; return error if none match

h4. Operational Risks
* Soft deletes not implemented; deletion is permanent

h4. Out of Scope
* Cascading deletes
* Audit logging

---

h2. Epic 3: Core TUI Model

Build the Bubbletea model that holds application state, handles keybindings, and delegates to the repository.

---

h3. 3.1 Model struct and Init

*Estimation:* 3h

h4. User Story
As a developer I want to define a Bubbletea model and initialize it with todos from the database in order to connect the UI to the data layer

h4. Acceptance Criteria
* AC1: @model@ struct contains: @tasks []Todo@, @cursor int@, @activeTab Tab@, @inputMode bool@, @repo *Repository@
* AC2: @Init()@ returns @tea.Cmd@ that queries the repository for the default tab (Today)
* AC3: After @Init()@ completes, @tasks@ is populated with todos matching the active filter
* AC4: Model starts with @cursor = 0@ and @inputMode = false@

h4. Implementation Details
* Define @model@ struct in @model/model.go@ with the fields listed above
* Define @Tab@ enum (Today, All, Completed) in @model/model.go@
* Implement @Init()@ to return a cmd that calls @repo.List()@ with the current filter
* Use Bubbletea's @tea.Cmd@ pattern; load tasks asynchronously if needed

h4. Operational Risks
* Blocking database queries in @Init()@ may freeze the UI; consider async loading if database is slow

h4. Out of Scope
* Persistence of cursor or tab state between runs
* Animation or transition effects

---

h3. 3.2 Navigation keybindings (j/k)

*Estimation:* 2h

h4. User Story
As a user I want to navigate the task list with @j@ and @k@ keys in order to move through tasks intuitively

h4. Acceptance Criteria
* AC1: @j@ moves cursor down; @k@ moves cursor up
* AC2: Cursor wraps around at list boundaries (down from last item goes to first; up from first goes to last)
* AC3: Cursor position is never negative or beyond the list length
* AC4: Navigation keys are ignored while in input mode
* AC5: Pressing j/k on an empty list is a no-op

h4. Implementation Details
* In @model.Update(msg tea.Msg)@, handle @tea.KeyMsg@ for @"j"@ and @"k"@
* Update @cursor@ with wrapping logic: @cursor = (cursor + 1) % len(tasks)@ and @cursor = (cursor - 1 + len(tasks)) % len(tasks)@
* Check @inputMode@ flag before processing navigation keys

h4. Operational Risks
* None specific to this feature

h4. Out of Scope
* Support for arrow keys (added in a later story)
* Page-up/page-down navigation

---

h3. 3.3 Tab switching (1/2/3/Tab)

*Estimation:* 3h

h4. User Story
As a user I want to switch between Today, All, and Completed views in order to see different subsets of tasks

h4. Acceptance Criteria
* AC1: @1@, @2@, @3@ keys switch to Today, All, Completed tabs respectively
* AC2: @Tab@ key cycles forward through tabs (Today → All → Completed → Today)
* AC3: Switching tabs reloads @tasks@ from the repository with the appropriate filter
* AC4: Cursor resets to @0@ when switching tabs
* AC5: Active tab state is reflected in the UI

h4. Implementation Details
* Define @Tab@ enum with @Today@, @All@, @Completed@ in @model/model.go@
* In @Update()@, handle key presses for @"1"@, @"2"@, @"3"@, @"tab"@
* Update @activeTab@, call @repo.List(getFilter(activeTab))@, and reset @cursor@
* Return a cmd to fetch the filtered task list

h4. Operational Risks
* None specific to this feature

h4. Out of Scope
* Custom tabs or dynamic tab creation
* Remembering cursor position per tab

---

h3. 3.4 Toggle completion (Space)

*Estimation:* 3h

h4. User Story
As a user I want to toggle the completion status of a task with the Space key in order to mark tasks as done/undone

h4. Acceptance Criteria
* AC1: Pressing @Space@ toggles the @done@ field on the task under the cursor
* AC2: The repository's @ToggleDone()@ is called
* AC3: After toggling, the task list is refreshed
* AC4: If a toggled task leaves the current filter view (e.g., marking done in Today view), it is removed from the list
* AC5: Cursor adjusts if it was on the last item and the list shrinks
* AC6: Pressing @Space@ on an empty list is a no-op

h4. Implementation Details
* In @Update()@, handle @tea.KeyMsg@ for @" "@
* Validate @cursor < len(tasks)@
* Call @repo.ToggleDone(tasks[cursor].ID)@
* Refresh the task list by re-querying @repo.List(currentFilter)@
* Adjust cursor if needed: @if cursor >= len(tasks) { cursor = max(0, len(tasks)-1) }@

h4. Operational Risks
* Task visibility changes may confuse users; the help bar should indicate which view they're in

h4. Out of Scope
* Animated visual feedback
* Undo/redo

---

h3. 3.5 Delete task (d)

*Estimation:* 3h

h4. User Story
As a user I want to delete a task with the @d@ key in order to remove unwanted tasks

h4. Acceptance Criteria
* AC1: Pressing @d@ deletes the task under the cursor
* AC2: The repository's @Delete()@ is called
* AC3: The task is removed from the list immediately
* AC4: Cursor adjusts if it was on the last item
* AC5: Pressing @d@ on an empty list is a no-op
* AC6: The deletion is permanent (no confirmation dialog in v1)

h4. Implementation Details
* In @Update()@, handle @tea.KeyMsg@ for @"d"@
* Validate @cursor < len(tasks)@
* Call @repo.Delete(tasks[cursor].ID)@
* Remove the task from @tasks@ slice: @tasks = append(tasks[:cursor], tasks[cursor+1:]...)@
* Adjust cursor: @if cursor >= len(tasks) { cursor = max(0, len(tasks)-1) }@

h4. Operational Risks
* Permanent deletion without confirmation may lead to accidental data loss; consider adding a confirmation dialog in a future epic

h4. Out of Scope
* Soft deletes or trash bin
* Multi-select delete

---

h3. 3.6 Quit (q/Ctrl+C)

*Estimation:* 1h

h4. User Story
As a user I want to quit the application with @q@ or @Ctrl+C@ in order to exit cleanly

h4. Acceptance Criteria
* AC1: Pressing @q@ quits the application when not in input mode
* AC2: @Ctrl+C@ always quits, regardless of mode
* AC3: Application exits with status code 0
* AC4: No unsaved state; all changes are persisted to the database

h4. Implementation Details
* In @Update()@, handle @tea.KeyMsg@ for @"q"@
* Check @inputMode@; only quit if @false@
* Handle @tea.KeyMsg.Type == tea.KeyCtrlC@ separately; always quit
* Return @tea.Quit()@ cmd in both cases

h4. Operational Risks
* None specific to this feature

h4. Out of Scope
* Confirmation dialogs
* Graceful shutdown hooks

---

h2. Epic 4: Add Task Input Flow

Implement the inline text input for creating new tasks, including due-date entry.

---

h3. 4.1 Enter add mode (a)

*Estimation:* 3h

h4. User Story
As a user I want to press @a@ to enter an inline text input mode in order to create a new task

h4. Acceptance Criteria
* AC1: Pressing @a@ switches @inputMode@ to @true@ and displays a text input field
* AC2: The input field is cleared and ready for text entry
* AC3: While in input mode, navigation keys (j/k/1/2/3/Space/d) are ignored
* AC4: Text input keys add characters to the input buffer
* AC5: Pressing @Esc@ cancels input mode and returns to normal navigation

h4. Implementation Details
* Create @ui/input.go@ with an input model using Bubbletea's @textinput@ component
* In @model.Update()@, detect @"a"@ key and set @inputMode = true@
* Delegate @tea.KeyMsg@ to the input component when @inputMode == true@
* Implement @Esc@ handling to reset @inputMode@ and clear the buffer

h4. Operational Risks
* None specific to this feature

h4. Out of Scope
* Text editing shortcuts (Ctrl+U, Ctrl+W, etc.)
* Copy/paste support

---

h3. 4.2 Due date prompt

*Estimation:* 4h

h4. User Story
As a user I want to enter an optional due date after the task title in order to assign deadlines to tasks

h4. Acceptance Criteria
* AC1: After entering the title and pressing @Enter@, a prompt appears asking for a due date
* AC2: The user can enter @YYYY-MM-DD@ or press @Enter@ to skip
* AC3: Invalid dates (e.g., @2025-13-45@) show an inline error message
* AC4: Valid dates are parsed and stored as @time.Time@
* AC5: Pressing @Esc@ cancels the entire add flow without saving

h4. Implementation Details
* After title @Enter@, set @inputMode@ to a @"date"@ sub-mode
* Use a second input field for date entry
* Validate input with @time.Parse("2006-01-02", input)@
* Show error state in the input field if validation fails
* Store the parsed @time.Time@ for the next step (save)

h4. Operational Risks
* Date parsing is timezone-aware; ensure all dates are stored in UTC

h4. Out of Scope
* Natural language date parsing (e.g., "tomorrow", "next Friday")
* Relative date entry

---

h3. 4.3 Save and return to list

*Estimation:* 3h

h4. User Story
As a user I want to finalize task creation and see the new task in the list in order to confirm the task was created

h4. Acceptance Criteria
* AC1: After entering the title and optional due date, pressing final @Enter@ creates the task
* AC2: @repo.Create(title, dueDate)@ is called with the entered values
* AC3: The task list is refreshed to include the new task
* AC4: Input mode is exited and normal navigation resumes
* AC5: Pressing @Esc@ at any input stage cancels without saving

h4. Implementation Details
* Collect title and optional dueDate from input fields
* Call @repo.Create(title, dueDate)@
* Re-query @repo.List(currentFilter)@ to refresh the task list
* Reset @inputMode@, clear input buffers, reset @cursor@ if needed
* If the new task is not in the current view (e.g., added a future-dated task in Today view), update the UI to reflect this

h4. Operational Risks
* New task might not appear in the current view if it doesn't match the filter; ensure the UI communicates this clearly (or auto-switch to All view)

h4. Out of Scope
* Auto-focus the new task in the list
* Keyboard shortcuts to repeat the last task

---

h2. Epic 5: View Rendering & Styling

Render the full UI with Lipgloss: tab bar, task list with color coding, and help bar.

---

h3. 5.1 Tab bar rendering

*Estimation:* 2h

h4. User Story
As a user I want to see which task view is currently active in order to navigate confidently

h4. Acceptance Criteria
* AC1: Three tabs are rendered at the top: @[ Today ]@, @[ All ]@, @[ Completed ]@
* AC2: The active tab is visually highlighted (bold, underline, inverted colors, or a combination)
* AC3: Inactive tabs are visually distinct but muted
* AC4: Tab styling is consistent across terminal resizes

h4. Implementation Details
* Create @ui/view.go@ with a @View()@ method that renders the UI
* Use @lipgloss@ styles to define tab appearance (@tabActive@, @tabInactive@)
* Render tabs horizontally separated by spaces or dividers
* Regenerate tab styling on each @View()@ call to accommodate dynamic widths

h4. Operational Risks
* Terminal width variations may cause tabs to wrap or overflow; implement text truncation if needed

h4. Out of Scope
* Draggable or reorderable tabs
* Dynamic tab creation

---

h3. 5.2 Task list rendering

*Estimation:* 3h

h4. User Story
As a user I want to see tasks in a clean, aligned list in order to read them comfortably

h4. Acceptance Criteria
* AC1: Each task line contains: cursor indicator (@>@ or @  @), completion icon (@[x]@ or @[ ]@), title, due date (or @—@)
* AC2: Columns are right-aligned and padded consistently
* AC3: Tasks are vertically stacked with no overlap
* AC4: The list respects terminal height and scrolls if needed (to be detailed in a later story)

h4. Implementation Details
* Create columns in @ui/view.go@ for cursor, checkbox, title, and due date
* Use @lipgloss@ @Width()@ and padding to align columns
* Iterate over @model.tasks@ and render each row
* Use @strings.Repeat(" ", width)@ for padding

h4. Operational Risks
* Very long task titles may overflow; truncate with ellipsis if needed
* Unicode characters in titles may affect column widths; use @runewidth@ for accurate width calculation

h4. Out of Scope
* Scrolling or pagination
* Task grouping or sorting beyond the repository layer

---

h3. 5.3 Color coding by due date

*Estimation:* 3h

h4. User Story
As a user I want tasks to be color-coded by urgency in order to quickly identify overdue or urgent tasks

h4. Acceptance Criteria
* AC1: Overdue + incomplete: task line rendered in red
* AC2: Due today + incomplete: task line rendered in yellow
* AC3: Completed: task line rendered dim (e.g., gray) with strikethrough on title
* AC4: Default (future due date or no due date): white/normal text
* AC5: Color coding applies to the entire task line, not just the title

h4. Implementation Details
* In @ui/view.go@, add a helper @getTaskColor(task Todo) lipgloss.Style@ function
* Compare @task.DueDate@ with @time.Now()@ and @time.Now().Add(24h)@
* Apply the style returned from @getTaskColor()@ to each rendered task line
* Use @lipgloss@ styles for red, yellow, dim, and strikethrough

h4. Operational Risks
* Timezone-aware date comparison; ensure all dates are in UTC for consistency

h4. Out of Scope
* Custom color schemes or themes
* Animation or transitions

---

h3. 5.4 Help bar

*Estimation:* 2h

h4. User Story
As a user I want to see contextual keybinding hints at the bottom of the screen in order to learn and remember shortcuts

h4. Acceptance Criteria
* AC1: Help bar is rendered at the bottom of the screen
* AC2: In normal mode: @j/k: navigate  a: add  space: toggle  d: delete  q: quit@
* AC3: In input mode: @enter: confirm  esc: cancel@
* AC4: Help text is muted or styled differently from the task list
* AC5: Help bar updates instantly when switching between modes

h4. Implementation Details
* Add a @renderHelpBar(inputMode bool) string@ function in @ui/view.go@
* Return different strings based on @inputMode@
* Render the help bar as a fixed line at the bottom of @View()@
* Use @lipgloss@ styles to dim or separate the help bar visually

h4. Operational Risks
* Very narrow terminals may truncate the help text; consider abbreviating for small widths

h4. Out of Scope
* Interactive help or detailed documentation in the app

---

h3. 5.5 Empty state

*Estimation:* 2h

h4. User Story
As a user I want to see a message when a tab has no tasks in order to understand why the list is empty

h4. Acceptance Criteria
* AC1: When the current filter returns zero tasks, a centered message is displayed
* AC2: Today tab: shows @"No tasks for today"@
* AC3: All tab: shows @"No tasks"@
* AC4: Completed tab: shows @"No completed tasks"@
* AC5: Empty state message replaces the task list; cursor is hidden

h4. Implementation Details
* In @View()@, check @len(model.tasks) == 0@
* Return an empty-state message instead of rendering the task list
* Center the message using @lipgloss@ alignment
* Vary the message based on @model.activeTab@

h4. Operational Risks
* None specific to this feature

h4. Out of Scope
* Suggestions on how to create a task from the empty state

---

h2. Epic 6: Polish & Edge Cases

Handle remaining UX details for a solid v1 release.

---

h3. 6.1 Graceful handling of empty lists

*Estimation:* 1h

h4. User Story
As a developer I want the application to handle keybinds safely on empty task lists in order to prevent crashes

h4. Acceptance Criteria
* AC1: Pressing @j@, @k@, @Space@, @d@ on an empty list is a no-op
* AC2: No panics, errors, or undefined behavior occurs
* AC3: Help bar and UI remain visible and responsive

h4. Implementation Details
* In @Update()@, validate @cursor < len(tasks)@ before accessing @tasks[cursor]@
* Return @(nil, nil)@ or skip the action if the list is empty
* Handle @ToggleDone()@ and @Delete()@ no-ops gracefully

h4. Operational Risks
* None specific to this feature

h4. Out of Scope
* User feedback when attempting invalid actions (e.g., delete on empty list)

---

h3. 6.2 Cursor persistence after mutations

*Estimation:* 2h

h4. User Story
As a user I want the cursor position to remain stable after toggling or deleting tasks in order to maintain context

h4. Acceptance Criteria
* AC1: After toggling a task that remains visible, cursor stays on the same task
* AC2: After deleting a task, cursor moves to the next task if available, or the previous task if at the end
* AC3: If the list becomes empty after deletion, cursor is hidden and no error occurs
* AC4: Cursor position is adjusted automatically; no manual reset is needed

h4. Implementation Details
* After @ToggleDone()@ or @Delete()@, refresh the task list via @repo.List(currentFilter)@
* If the list shrinks, validate @cursor < len(tasks)@
* Adjust @cursor@ if needed: @if cursor >= len(tasks) { cursor = max(0, len(tasks) - 1) }@

h4. Operational Risks
* None specific to this feature

h4. Out of Scope
* Remembering cursor position across sessions

---

h3. 6.3 Terminal resize handling

*Estimation:* 2h

h4. User Story
As a user I want the UI to adapt when I resize my terminal in order to maintain usability

h4. Acceptance Criteria
* AC1: Application responds to @tea.WindowSizeMsg@
* AC2: UI elements (tabs, task list, help bar) reflow to fit the new terminal size
* AC3: No content is lost or corrupted on resize
* AC4: Text truncation or wrapping adjusts automatically

h4. Implementation Details
* In @Update()@, handle @tea.WindowSizeMsg@ to update @model.width@ and @model.height@
* In @View()@, use @model.width@ and @model.height@ to calculate widths for columns and text wrapping
* Use @lipgloss@ @Width()@ and @MaxWidth()@ to enforce sizing

h4. Operational Risks
* Very small terminal sizes (e.g., < 40 columns) may be unsupported; document minimum size if needed

h4. Out of Scope
* Layout preferences or customization

---

h3. 6.4 Database path configuration

*Estimation:* 1h

h4. User Story
As a developer I want to allow customization of the database path via environment variable in order to support testing and alternative setups

h4. Acceptance Criteria
* AC1: Default database path is @~/.todo.db@
* AC2: Setting @TODO_DB_PATH@ environment variable overrides the default
* AC3: Path is resolved and verified on startup
* AC4: Relative paths are expanded to absolute paths

h4. Implementation Details
* In @main.go@, read @os.Getenv("TODO_DB_PATH")@
* If empty, use @os.ExpandUser("~/.todo.db")@
* Pass the path to @db.Open(path)@
* Document the environment variable in a README (if created)

h4. Operational Risks
* Invalid or inaccessible paths will cause startup to fail; provide clear error messages

h4. Out of Scope
* Configuration files or CLI flags for other settings

---

h3. 6.8 Better Error Handling

*Estimation:* 3h

h4. User Story
As a user I want to receive clear feedback when operations fail so I understand what went wrong.

h4. Acceptance Criteria
* AC1: Database errors (e.g., failed write, read timeout) display a visible inline error message rather than silently failing
* AC2: Error messages are shown in the help bar area or a dedicated status line, styled distinctly (e.g., red text)
* AC3: After displaying the error, the application remains usable — navigation and input are not locked
* AC4: Errors are cleared on the next successful action or when the user presses any navigation key

h4. Implementation Details
* Add an @errorMsg string@ field to @AppModel@
* On any @repo.*@ call that returns an error, set @model.errorMsg@ to a human-readable string
* In @View()@, render @errorMsg@ in the help bar area when non-empty, styled with a red @lipgloss@ color
* Clear @errorMsg@ at the start of each successful command handler

h4. Operational Risks
* Error messages must not obscure key UI elements; keep them to a single line

h4. Out of Scope
* Persistent error logs
* Retry mechanisms

---

h3. 6.9 Database Migrations

*Estimation:* 4h

h4. User Story
As a developer I want database migrations so schema changes can be applied safely over time.

h4. Acceptance Criteria
* AC1: A @schema_migrations@ table tracks which migrations have been applied
* AC2: On startup, any unapplied migrations are run in order before the application opens
* AC3: Each migration is a numbered SQL file or embedded string (e.g., @0001_initial.sql@, @0002_add_priority.sql@)
* AC4: Running the same migration twice is a no-op (idempotent)
* AC5: A failed migration causes startup to abort with a clear error message

h4. Implementation Details
* Create @db/migrations.go@ with a @RunMigrations(db *sql.DB) error@ function
* Create a @schema_migrations@ table with columns: @version@ (int, primary key), @applied_at@ (timestamp)
* Embed migration SQL using @go:embed@ or define them as string constants in order
* On each run, query applied versions and execute any with a higher version number
* Wrap each migration in a transaction; roll back and return an error on failure

h4. Operational Risks
* Existing installations without the @schema_migrations@ table must be handled gracefully on first run
* Irreversible migrations (e.g., @DROP COLUMN@) should be avoided until a down-migration strategy exists

h4. Out of Scope
* Down migrations / rollbacks
* CLI commands to inspect or run migrations manually

---

h3. 6.10 Improve Test Quality

*Estimation:* 5h

h4. User Story
As a developer I want stronger automated tests so regressions are detected before release.

h4. Acceptance Criteria
* AC1: All repository methods (@Create@, @List@, @ToggleDone@, @Delete@) have table-driven unit tests covering happy path and error cases
* AC2: Each filter variant (@All@, @Today@, @Done@) in @List()@ is covered by at least one test
* AC3: Model @Update()@ keybinding logic (navigation, toggle, delete, input mode) is covered by unit tests using a mock or in-memory repository
* AC4: Tests run with @go test ./...@ without requiring external dependencies beyond a temp SQLite file
* AC5: Code coverage for @db/@ and @model/@ packages is at or above 80%

h4. Implementation Details
* Use @testing.T@ and @t.TempDir()@ to create isolated SQLite databases per test
* Define a @Repository@ interface in @db/repository.go@ so the model can be tested with a fake
* Implement @FakeRepository@ in @db/fake_repository.go@ for model-layer tests
* Add subtests with @t.Run()@ for each filter and edge case
* Run @go test -cover ./...@ in CI to enforce the coverage threshold

h4. Operational Risks
* Model tests depend on Bubbletea message types; ensure test helpers send proper @tea.Msg@ values

h4. Out of Scope
* End-to-end or integration tests that launch the full TUI
* Performance benchmarks

---

h3. 6.11 Session Summary on Exit

*Estimation:* 3h

h4. User Story
As a user I want to see a summary in the terminal after closing the app so I know what I accomplished and what still needs attention today.

h4. Acceptance Criteria
* AC1: After quitting (@q@ or @Ctrl+C@), the alt-screen is dismissed and a summary is printed to the terminal
* AC2: Summary shows the number of tasks completed this session (toggled to done)
* AC3: Summary shows the number of tasks deleted this session
* AC4: Summary shows the count of remaining incomplete tasks due today
* AC5: If nothing happened (0 completed, 0 deleted, 0 today), a minimal "Nothing to report" line is shown instead
* AC6: Output is plain text, readable without colour support

h4. Implementation Details
* Add @SessCompleted int@ and @SessDeleted int@ counters to @AppModel@
* Increment @SessCompleted@ in @Update()@ when @ToggleDone@ is called on a task whose @Done == false@
* Increment @SessDeleted@ in @Update()@ when a delete is confirmed (@"y"@ in confirm mode)
* After @p.Run()@ returns in @main.go@, call @repo.List(FilterToday)@ to get remaining today tasks
* Print the summary to @os.Stdout@ using @fmt.Fprintf@; no extra dependencies needed
* Expose a @Summary@ struct from the model package (or return counters via a method) so @main.go@ can read them

h4. Operational Risks
* @ToggleDone@ can toggle a task back to incomplete; @SessCompleted@ should only count net completions (transitions from @done=false@ to @done=true@)

h4. Out of Scope
* Persisting session history across runs
* Showing which specific tasks were completed or deleted
