# DevDeck

DevDeck is a terminal user interface (TUI) tool designed to manage your local development services. It reads a configuration file to launch multiple processes (like backends, frontends, or databases) and allows you to monitor their logs and status in real-time.

## Features

-   **Process Management**: Start, stop, and monitor multiple services defined in a YAML file.
-   **Real-time Logs**: Stream logs from `stdout` and `stderr` for each process.
-   **Split View**: Pin a process's logs to a secondary pane to watch two services simultaneously (e.g., frontend and backend).
-   **Restart**: Quickly restart any service with a single keypress.
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

-   `j` / `Down`: Move cursor down
-   `k` / `Up`: Move cursor up
-   `r`: Restart the selected process
-   `s`: Toggle split view (Pin/Unpin selected process logs)
-   `q` / `Ctrl+C`: Quit application

## Tech Stack

-   **Language**: Go
-   **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
-   **Styling**: [Lip Gloss](https://github.com/charmbracelet/lipgloss)

## ToDo

-   [ ] Fix height calculation issues in split view (prevent weird resizing).
-   [ ] Fix issue where the first process is not visible in split view.
-   [ ] Support passing configuration file path via CLI arguments (e.g., `-config my-deck.yaml`).
-   [ ] Add support for JSON configuration files (`.json`) alongside YAML.

