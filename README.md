# DevDeck

**DevDeck** is a modern terminal user interface (TUI) tool designed to manage and monitor your local development services (backends, frontends, databases, scripts). It allows you to run multiple processes defined in a simple YAML configuration, stream their logs in real-time, and control them with keyboard shortcuts or mouse interactions.

## ✨ Features

-   **Process Management**: Start, stop, and monitor multiple services from a single UI.
-   **Real-time Logs**: Stream logs (`stdout`/`stderr`) for each process.
-   **Split View**: Pin a process to a secondary pane to watch two logs simultaneously (e.g., Frontend + Backend).
-   **Resource Monitoring**:
    -   **Global**: Status bar showing total System CPU & RAM usage.
    -   **Per-Process**: Live CPU and Memory usage displayed in the task list.
-   **Mouse Support**:
    -   **Click**: Select tasks or focus log panes.
    -   **Scroll**: Use the mouse wheel to scroll through logs.
-   **Interactive Input**: Send commands (`stdin`) to running processes (press `i`).
-   **Log Search**: Filter and search logs in real-time (press `/`).
-   **Theming**: Fully customizable color themes via configuration.
-   **Hot Reload**:
    -   **Config**: Auto-reloads `devdeck.yaml` on change.
    -   **Development**: Supports `air` for hot-reloading the DevDeck app itself.

## 🚀 Usage

### 1. Installation

```bash
# Clone the repository
git clone https://github.com/kuo-hm/devdeck.git
cd devdeck

# Build and Run
go build -o devdeck.exe main.go
./devdeck.exe
```

### 2. Configuration

Create a `devdeck.yaml` file in your project root:

```yaml
tasks:
  - name: "Backend API"
    command: "go run main.go"
    directory: "./backend"
    env:
      - "PORT=8080"

  - name: "Frontend"
    command: "npm start"
    directory: "./frontend"

# Optional: Custom Theme
theme:
  primary: "#FF00FF"    # Focused elements, borders
  secondary: "#00FFFF"  # Backgrounds, accents
  border: "#444444"     # Default borders
  text: "#FFFFFF"       # Default text
```

### 3. Key Bindings

| Key | Action |
| :--- | :--- |
| `Tab` | Switch focus between Task List and Log Panes |
| `↑` / `k` | Move cursor up / Scroll log up |
| `↓` / `j` | Move cursor down / Scroll log down |
| `Mouse` | **Click** to select/focus, **Wheel** to scroll |
| `Enter` | Select task / Confirm input |
| `r` | Restart the selected process |
| `s` | Toggle **Split View** (Pin selected process) |
| `G` | Open **Group Menu** (Restart named groups) |
| `i` | Enable **Input Mode** (Send text to stdin) |
| `/` | **Search** logs (Enter to jump to matches) |
| `?` | Toggle **Help Menu** |
| `q` | Quit |

## 🛠️ Development

### Hot Reload (Dev Mode)

To work on DevDeck source code with hot reloading:

1.  Install **Air**:
    ```bash
    go install github.com/air-verse/air@latest
    ```
2.  Run Air:
    ```bash
    air
    ```
    This will auto-build and restart DevDeck whenever you save a `.go` file.

## 🗺️ Roadmap

### Phase 1: Core (Completed) ✅
- [x] Process Management (Start/Stop/Restart)
- [x] Log Streaming & Search
- [x] Split View
- [x] Config Hot Reload

### Phase 2: UX (Completed) ✅
- [x] Mouse Support (Click/Scroll)
- [x] Themes
- [x] Help Menu
- [x] Status Bar (CPU/RAM Stats)

### Phase 3: Advanced Control (Planned) 🚧
- [ ] **Process Groups**: Control multiple services as a unit.
- [ ] **Dependencies**: "Wait for Database" before starting Backend.
- [ ] **Health Checks**: HTTP/TCP probes for "Ready" status.

## 📝 Configuration Reference

| Field | Type | Description |
| :--- | :--- | :--- |
| `tasks` | list | List of task objects (Required). |
| `tasks[].name` | string | Display name. |
| `tasks[].command` | string | Shell command to execute. |
| `tasks[].directory` | string | Working directory. |
| `tasks[].env` | list | Environment variables (`KEY=VAL`). |
| `tasks[].health_check` | object | `{ type: "tcp/http", target: "...", interval: 2000 }` |
| `tasks[].depends_on` | list | List of task names to wait for before starting. |
| `tasks[].groups` | list | List of group tags (e.g. `["backend"]`). |
| `theme` | object | Color overrides (Optional). |
