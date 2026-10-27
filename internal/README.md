# Internal Packages

The `internal/` directory contains all private domain logic, configuration handlers, subprocess runners, user interfaces, and shared utilities for the Workbench application suite.

Following Go conventions, packages in `internal/` cannot be imported by external modules, enforcing clean encapsulation.

---

### Package Hierarchy & Architecture

```
internal/
├── config/       # TOML configuration loading, path interpolation, and gocfl profile heuristics
├── job/          # Domain model, batch filesystem scanner, path resolvers, and lifecycle operations
├── runner/       # External command builders for gocfl and ona subprocess execution
├── ui/           # Terminal User Interface (tview/tcell) layout, navigation, and modal runners
└── util/         # Filesystem safety checks, atomic move/copy fallbacks, and port allocation
```

---

### Module Summary

| Package | Purpose | Key Types / Functions |
| :--- | :--- | :--- |
| **`config`** | Parses `workbench.toml`, resolves `{batch}` path templates, determines `gocfl` profiles. | `WBConfig`, `GOCFLConfig`, `ONAConfig`, `LoadConfig`, `CountFiles` |
| **`job`** | Defines preservation job states, scans batch folders, and executes lifecycle transitions. | `Job`, `JobStatus`, `ActionType`, `GetSignatures`, `ArchiveJob`, `HandleJobError`, `CheckIngestResult` |
| **`runner`** | Assembles `*exec.Cmd` objects for CLI tool invocations (`gocfl`, `ona`). | `CreateOCFLCmd`, `CreateReportCmd`, `ValidateOCFLCmd`, `IngestOCFLCmd`, `StatusOCFLCmd` |
| **`ui`** | Renders the multi-pane TUI, streams logs, and executes commands in modal dialogs. | `SetupUI`, `ShowDetail`, `RunCommandInModal` |
| **`util`** | Provides directory validation and safe cross-device move/copy operations. | `EnsureTargetDirectory`, `MoveDirectory`, `MoveFile`, `GetFreePort` |

---

### Dependency Flow

```
cmd/create_container
        │
        ├──► internal/config
        └──► internal/ui
                 │
                 ├──► internal/config
                 ├──► internal/job
                 ├──► internal/runner ──► internal/job
                 └──► internal/util
```
