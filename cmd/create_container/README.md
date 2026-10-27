# create_container

`create_container` is the primary executable of the Workbench suite. It provides an interactive Terminal User Interface (TUI) to orchestrate digital preservation workflows, including metadata validation, OCFL container creation, PDF report generation, compliance validation, and ingest management.

---

### Responsibilities

- **CLI Flag Parsing:** Accepts optional `-config` flag for custom TOML configuration paths.
- **Log Redirection:** Sets up an operating system pipe (`os.Pipe`) to redirect structured `zerolog` events into the TUI log panel in real time.
- **TUI Initialization:** Initializes the `tview` event loop, displays a startup banner, and transitions into the multi-pane navigation interface.

---

### Build & Run

#### Building the Binary
```bash
go build -o create_container.exe ./cmd/create_container
```

#### Running the Application
```bash
# Using explicit configuration file:
./create_container -config /path/to/workbench.toml

# Using default configuration discovery (~/wb/workbench.toml, /opt/wb/workbench.toml, /etc/wb/workbench.toml):
./create_container
```

---

### Command-Line Arguments

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `-config` | String | `""` | Absolute or relative path to the Workbench TOML configuration file. |

---

### Prerequisites & Dependencies

1. **`gocfl` CLI:** Must be installed and reachable in the system `$PATH` for creating, validating, and reporting OCFL containers.
2. **`ona` CLI:** Must be installed and reachable in the system `$PATH` for ingesting packages and querying status.
3. **Valid Configuration:** A readable `workbench.toml` file defining batch directory locations and runner configurations.
