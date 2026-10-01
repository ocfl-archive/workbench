# Workbench

Workbench is a terminal-based suite of digital preservation workflow tools designed for archivist operations and automated ingest management. The primary application, `create_container`, manages the scanning of incoming digitization batches, parsing metadata, and packaging and validating standard [OCFL (Oxford Common File Layout)](https://ocfl.io/) containers via [`gocfl`](https://github.com/ocfl-archive/gocfl).

---

### Features

- **Batch & Job Discovery:** Automatically detects batch subfolders and identifies ingest units containing `info.json` metadata files.
- **Interactive TUI:** Multi-pane Terminal User Interface (`tview`) for browsing batches, selecting jobs, viewing metadata details, and executing pipeline tasks.
- **Real-Time Log & Command Streaming:** Dedicated bottom log stream with ANSI color rendering alongside modal dialogs displaying live stdout/stderr during container creation and validation.
- **OCFL Packaging & Reporting:** Direct integration with `gocfl` to create `.zip` containers, produce PDF reports, and run compliance validation.
- **Flexible Path Configuration:** Supports TOML-based configurations with static or dynamic `{batch}` placeholder path expansion.

---

### Prerequisites

- **Go:** Version 1.21 or later (developed and tested with Go 1.27).
- **gocfl:** The [`gocfl`](https://github.com/ocfl-archive/gocfl) CLI executable must be installed and accessible on your system `$PATH`.

---

### Installation & Build

Clone the repository and build the `create_container` executable:

```bash
git clone https://github.com/ocfl-archive/workbench.git
cd workbench
go build ./cmd/create_container
```

To run all unit tests:

```bash
go test -v ./...
```

---

### Quickstart

1. Prepare a `workbench.toml` configuration file (see sample in `config/workbench.toml`):
   ```toml
   batches = "/data/batches/"
   input = "{batch}/incoming/"
   ocfl = "{batch}/ocfl/"
   report = "{batch}/report/"
   error = "{batch}/error/"
   archived = "{batch}/archived/"

   [gocfl]
   config = "/path/to/gocfl.toml"
   ```

2. Run `create_container`:
   ```bash
   ./create_container -config /path/to/workbench.toml
   ```

*(If `-config` is omitted, `create_container` automatically searches `~/wb/workbench.toml`, `/opt/wb/workbench.toml`, and `/etc/wb/workbench.toml`.)*

---

### TUI Navigation & Shortcuts

| Key / Shortcut | Action |
| :--- | :--- |
| `Tab` / `Right Arrow` | Move focus to next pane or action button (Batches → Jobs → Details / Actions) |
| `Shift+Tab` / `Left Arrow` | Move focus to previous pane or action button |
| `Up` / `Down Arrow` | Navigate items in the active list (Batches or Jobs) |
| `Enter` | Trigger highlighted action button (`OCFL erstellen`, `Report erstellen`, `Validate`) |
| `ESC` / `Enter` (Modal) | Close process output modal dialog once execution completes |
| `Ctrl+Q` / `Ctrl+C` | Quit the application |

---

### Project Documentation

- **[System Architecture](docs/ARCHITECTURE.md):** In-depth overview of internal package architecture, data flows, and state transitions.
- **[Configuration Guide](docs/CONFIGURATION.md):** Detailed TOML configuration reference, placeholder resolution rules, and default search paths.
