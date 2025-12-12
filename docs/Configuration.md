# Configuration Reference

DevDeck is configured via a YAML (or JSON) file.

## Structure

```yaml
tasks:
  - name: "My Service"
    command: "npm start"
    directory: "./frontend"
    env: ["PORT=3000"]
    env_file: ".env"
    health_check:
      type: "http"
      target: "http://localhost:3000"
    depends_on: ["Backend"]
    groups: ["frontend"]

theme:
  primary: "#BD93F9"
```

## Field Reference

### Tasks (`tasks[]`)

| Field | Type | Description |
| :--- | :--- | :--- |
| `name` | string | **Required**. Display name. |
| `command` | string | **Required**. Command to execute. |
| `directory` | string | Working directory (relative to config file). |
| `env` | list | Environment variables (`key=value`). |
| `env_file` | string | Path to `.env` file to load. |
| `groups` | list | Tags for group management. |
| `depends_on` | list | Wait for these task names to be healthy. |
| `health_check` | object | See below. |

### Health Checks (`health_check`)

| Field | Type | Description |
| :--- | :--- | :--- |
| `type` | string | `tcp` or `http`. |
| `target` | string | Port (`localhost:8080`) or URL (`http://...`). |
| `interval` | int | Milliseconds between checks (default 2000). |
| `timeout` | int | Timeout for check (default 1000). |

### Theme (`theme`)

Customize the UI colors. All fields expect hex codes (e.g. `#FFFFFF`).
- `primary`, `secondary`, `border`, `text`
