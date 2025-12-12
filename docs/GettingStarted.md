# Getting Started with DevDeck

DevDeck is a terminal-based tool for managing multiple processes.

## Installation

### From Source
1.  Ensure you have **Go 1.21+** installed.
2.  Clone the repository:
    ```bash
    git clone https://github.com/kuo-hm/devdeck.git
    cd devdeck
    ```
3.  Build the binary:
    ```bash
    go build -o devdeck.exe
    ```

## Running DevDeck

1.  Create a configuration file (e.g., `devdeck.yaml`). See [Configuration](Configuration.md) for details.
2.  Run the application:
    ```bash
    ./devdeck.exe -c devdeck.yaml
    ```
3.  Or verify using the Example Environment:
    ```bash
    cd examples
    ../devdeck.exe
    ```

## Key Bindings

| Key | Action |
| :--- | :--- |
| `↑/↓/j/k` | Navigate Checklists |
| `Enter` | Select / Start / Stop |
| `r` | Restart Process |
| `s` | Toggle Split View |
| `g` | Open Group Menu |
| `/` | Search Logs |
| `?` | Help |
| `q` | Quit |
