# DevDeck

DevDeck is a terminal user interface (TUI) tool designed to manage your local development services. It reads a configuration file to launch multiple processes (like backends, frontends, or databases) and allows you to monitor their logs and status in real-time.

## Features

-   **Process Management**: Start, stop, and monitor multiple services defined in a YAML file.
-   **Real-time Logs**: Stream logs from `stdout` and `stderr` for each process.
-   **Split View**: Pin a process's logs to a secondary pane to watch two services simultaneously (e.g., frontend and backend).
-   **Restart**: Quickly restart any service with a single keypress.
-   **Interactive Input**: Send stdin commands to running processes (press `i`).
-   **Log Filtering**: Search and filter logs in real-time (press `/`).
-   **Cross-Platform**: Works on Windows, macOS, and Linux.

## Usage

1.  **Configure**: Create a `devdeck.yaml` file in the root directory.

    ```yaml
    tasks:
      - name: "Backend API"
        command: "npm start"
        directory: "./backend"
        env:
          - "PORT=3000"

      - name: "Frontend"
        command: "npm run dev"
        directory: "./frontend"
    ```

2.  **Run**:
    ```bash
    go run main.go
    # or build and run
    go build -o devdeck.exe && ./devdeck.exe

    # Optional: Specify a custom configuration file
    ./devdeck.exe -c my-config.yaml
    ```

## Configuration Reference

The `devdeck.yaml` file defines the tasks DevDeck manages. It requires a root `tasks` key containing a list of task definitions.

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `name` | string | **Yes** | The label displayed in the task list. |
| `command` | string | **Yes** | The shell command to execute. |
| `directory` | string | No | The working directory to run the command in. Defaults to the current directory. |
| `env` | list | No | A list of environment variables (`KEY=VALUE`) to inject into the process. |

**Example:**

```yaml
tasks:
  - name: "Database"
    command: "docker-compose up db"

  - name: "Backend"
    command: "go run main.go"
    directory: "./server"
    env:
      - "DB_HOST=localhost"
      - "DB_PORT=5432"
```

## Key Bindings

-   `Tab`: Toggle focus between Task List and Log Pane(s)
-   `j` / `Down`: Move cursor down (Task List) / Scroll down (Log Pane)
-   `k` / `Up`: Move cursor up (Task List) / Scroll up (Log Pane)
-   `r`: Restart the selected process
-   `s`: Toggle split view (Pin/Unpin selected process logs)
-   `i`: Enable interactive input (stdin) for the selected process
-   `/`: Search/Filter logs
-   `q` / `Ctrl+C`: Quit application

## Tech Stack

-   **Language**: Go
-   **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
-   **Styling**: [Lip Gloss](https://github.com/charmbracelet/lipgloss)

## Roadmap

### 🐛 Bugs
- [x] **Split View Layout**: Fix height calculation issues to prevent weird resizing.
- [x] **First Process Visibility**: Fix issue where the first process is not visible in split view.
- [x] **Scroll Behavior**: Improve scrolling in log views (autoscroll vs. manual scroll).
- [x] **Restart Error State**: Fix issue where error messages persist after a successful restart.
- [x] **UI Padding**: Fix issue where the top UI border was cut off by the window edge.

### 🚀 Features
- [x] **CLI Arguments**: Support passing configuration file path (e.g., `-config my-deck.yaml`).
- [x] **JSON Config**: Add support for JSON configuration files alongside YAML.
- [x] **Process Input**: Allow sending interactive input (stdin) to running processes.
- [x] **Log Search**: Implement search functionality within the log views.
- [x] **Hot Reload**: Automatically reload configuration when `devdeck.yaml` changes.

### 🎨 Improvements
- [ ] **Themes**: Allow UI color customization via config.
- [ ] **Process Groups**: Ability to start/stop multiple services at once.
- [ ] **Help Menu**: Add a comprehensive help modal (`?` key).

