# DevDeck

**DevDeck** is a modern terminal user interface (TUI) tool to manage, monitor, and orchestrate your local development services.

![DevDeck](https://github.com/kuo-hm/devdeck/assets/placeholder.png)

## ✨ Features

-   **Process Management**: Start, stop, and restart multiple services.
-   **Orchestration**:
    -   **Dependencies**: Ensure services start in order (e.g., Database before Backend).
    -   **Health Checks**: Real TCP/HTTP probes to verify service readiness.
    -   **Process Groups**: Restart related services (e.g., "All Backends") with one key.
-   **Real-time Logs**: Stream, search (`/`), and split-view (`s`) logs.
-   **Resource Monitoring**: Live CPU and Memory usage per process.
-   **Interactive**: Send commands (`stdin`) to running processes (`i`).
-   **Hot Reload**: Modify `devdeck.yaml` and see changes instantly.
-   **Crash Handling**: Robust error recovery and logging.

## 🚀 Quick Start

1.  **Download** the latest release or build from source:
    ```bash
    go build -o devdeck.exe
    ```
2.  **Configure** your services in `devdeck.yaml`:
    ```yaml
    tasks:
      - name: "API"
        command: "npm start"
        health_check:
          type: "http"
          target: "http://localhost:3000"
    ```
3.  **Run**:
    ```bash
    ./devdeck.exe
    ```

👉 **[Getting Started Guide](docs/GettingStarted.md)**
👉 **[Full Configuration Reference](docs/Configuration.md)**

## ⌨️ Key Bindings

| Key | Action |
| :--- | :--- |
| `↑/↓` | Navigate |
| `Enter` | Select / Details |
| `r` | Restart |
| `s` | Split View |
| `g` | Group Menu |
| `/` | Search Logs |
| `i` | Interactive Input |
| `?` | Help |

## 🤝 Contributing

This project is stable. Feel free to open issues or PRs!
