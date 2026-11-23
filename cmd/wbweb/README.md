# wbweb

`wbweb` is the web-based interactive interface (Web-GUI) of the Workbench suite. It provides a modern, responsive web dashboard built with **Tailwind CSS**, **DaisyUI**, **HTMX 4**, and **Server-Sent Events (SSE)** to orchestrate archival workflows analogously to the `wb` TUI.

---

### Responsibilities

- **Web Dashboard:** Renders a 3-column layout displaying batches, jobs, and job inspection details with dynamic lifecycle action triggers.
- **Embedded Assets & Templates:** Embeds all HTML templates, Tailwind CSS, DaisyUI themes, icons, and JavaScript libraries with `embed.FS` for zero-dependency deployment.
- **Real-Time Log Streaming & Terminal Emulator:** Streams `zerolog` events and external command executions (`gocfl`, `ona`) line-by-line via Server-Sent Events into an integrated web terminal emulator with ANSI color parsing.
- **Batch Processing & Background Tasks:** Asynchronously executes single job actions and full batch preservation pipelines with real-time UI updates.
- **Theme Support:** Supports light, dark, automatic system color scheme detection, and custom DaisyUI themes.

---

### Build & Run

#### Building the Binary
```bash
go build ./cmd/wbweb
```

#### Running the Application
```bash
# Using explicit configuration file and custom port:
./wbweb -config /path/to/workbench.toml -addr :8765

# Automatically open the browser on startup:
./wbweb -config /path/to/workbench.toml -open
```

---

### Command-Line Arguments

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `-config` | String | `""` | Absolute or relative path to the Workbench TOML configuration file. |
| `-addr` | String | `":8765"` | HTTP listen address (e.g. `:8765` or `localhost:8765`). |
| `-open` | Boolean | `false` | Automatically opens the Web GUI in the default system browser upon launch. |

---

### Prerequisites & Dependencies

1. **`gocfl` CLI:** Must be installed and reachable in the system `$PATH` for creating, validating, and reporting OCFL containers.
2. **`ona` CLI:** Must be installed and reachable in the system `$PATH` for ingesting packages and querying preservation status.
3. **Valid Configuration:** A readable `workbench.toml` file defining batch directory locations and runner configurations.
