---
sessionId: session-260918-090840-ukkn
---

# Requirements

### Overview & Goals
Add clear, comprehensive, and idiomatic English comments (both Go doc comments for exported/unexported symbols and inline comments for complex logic) across all Go source files located in the `cmd/` package (`cmd/create_container/`).

### Scope
- **In Scope**:
  - `cmd/create_container/main.go`: CLI flags, application initialization, logger configuration, and entry point.
  - `cmd/create_container/job.go`: Struct definitions (`Job`, `info`), directory traversal (`getBatches`), and signature/job discovery (`getSignatures`).
  - `cmd/create_container/actions.go`: Modal view creation and asynchronous command execution handlers (`createOCFL`, `createReport`).
  - `cmd/create_container/ui.go`: TUI component setup (`setupUI`), detail view formatting (`showDetail`), dynamic action button handling, keyboard navigation, and background log streaming.
  - `cmd/create_container/banner.go`: Constant documentation for the ASCII banner.
- **Out of Scope**:
  - Functional behavior or architectural changes.
  - Modifying user-facing UI labels or text unless requested.

### Functional Requirements
- Every function and struct must have a standard Go doc comment explaining its purpose, parameters, and behavior.
- Important logic blocks (e.g. streaming process output, keyboard event routing, batch/job state synchronization, background goroutines) must contain concise inline comments explaining *what* and *why*.
- All comments must be written in grammatically correct, professional English.

# Technical Design

### Current Implementation
The `cmd/create_container` package implements a terminal user interface (TUI) application using `rivo/tview` and `gdamore/tcell/v2` to browse batches, inspect container jobs, and trigger OCFL container and report generation actions. Currently, the code contains minimal or no comments explaining the control flow, data models, or TUI event lifecycles.

### Key Decisions
- **Go Doc Comment Conventions**: Follow standard Go doc comment style (e.g., `// setupUI initializes and configures...`).
- **Inline Comment Clarity**: Focus inline comments on architectural flows, asynchronous goroutines, UI focus management, and stream piping to ensure maintainability without cluttering simple code.

### Proposed Changes

#### 1. `cmd/create_container/main.go`
- Add comments explaining CLI flags (`-input-folder`, `-ocfl-folder`, `-report-folder`).
- Add doc comment to `main()`.
- Add inline comments explaining:
  - Pipe creation (`os.Pipe()`) for redirecting zerolog output to the UI log viewer.
  - Application startup with temporary banner display switching to the main pages.

#### 2. `cmd/create_container/job.go`
- Add doc comments to `Job` and `info` structs describing their fields.
- Add doc comment and inline steps to `getBatches(folder string)`.
- Add doc comment and inline steps to `getSignatures(inputF, ocflF, reportF string, logger zerolog.Logger)` detailing:
  - Scanning directory for JSON and subfolder pairs.
  - Unmarshaling signature metadata.
  - Verifying presence of data, metadata folders, and existing artifacts (OCFL zip, report PDF).

#### 3. `cmd/create_container/actions.go`
- Add doc comment and inline steps for `createOCFL`:
  - Creation and centering of the execution modal box.
  - Background goroutine executing external command and streaming stdout/stderr via `io.MultiReader` to `tview.TextView`.
  - Handling completion and attaching input capture to close modal.
- Add doc comment to `createReport` describing its role and expected implementation.

#### 4. `cmd/create_container/ui.go`
- Add doc comments to `showDetail` and `setupUI`.
- Add inline comments for:
  - UI widget hierarchy and layout composition.
  - `updateDetailAndActions` closure for conditional button rendering based on job state.
  - `loadJobsForBatch` and `refreshCurrentBatch` closures for list population and index preservation.
  - Custom keyboard navigation handler (`app.SetInputCapture`) handling `Tab`, `Backtab`, and arrow navigation across panels.
  - Background log consumer scanning from `logReader` and writing ANSI text to `logView`.
  - Delayed timer transitioning root view from banner to main layout.

#### 5. `cmd/create_container/banner.go`
- Add doc comment explaining the ASCII art header.

### Affected Files
- `cmd/create_container/main.go`
- `cmd/create_container/job.go`
- `cmd/create_container/actions.go`
- `cmd/create_container/ui.go`
- `cmd/create_container/banner.go`

# Delivery Steps

### ✓ Step 1: Document data models, batch utilities, and entry point
Documentation comments and inline explanations are added to core models, scanning logic, and application initialization files (`job.go`, `main.go`, `banner.go`).

- Add doc comments to `Job` and `info` structs detailing field meanings and metadata relationships in `cmd/create_container/job.go`.
- Document `getBatches` and `getSignatures` functions in `cmd/create_container/job.go` with descriptive comments explaining directory traversal, JSON parsing, and artifact existence checks.
- Add explanatory comments to CLI flag variables and `main()` in `cmd/create_container/main.go` describing log pipe setup, zerolog initialization, and UI application lifecycle.
- Add descriptive comment to the ASCII art `banner` constant in `cmd/create_container/banner.go`.

### ✓ Step 2: Document UI layout, navigation, and action execution
Comprehensive English doc comments and inline logic explanations are added across all TUI setup, event handling, navigation, and execution action files (`ui.go`, `actions.go`).

- Add doc comments for `showDetail` and `setupUI` in `cmd/create_container/ui.go`.
- Add inline comments within `setupUI` explaining list setup, dynamic action button resolution, batch/job synchronization, keyboard navigation captures (Tab, Shift-Tab, Arrow keys), and background log reading.
- Add doc comments and step-by-step inline comments for `createOCFL` and `createReport` in `cmd/create_container/actions.go` covering modal layout construction, asynchronous command execution, real-time output streaming, and completion callbacks.