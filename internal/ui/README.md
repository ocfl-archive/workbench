# UI Package (`internal/ui`)

The `ui` package builds and manages the Terminal User Interface (TUI) using [`tview`](https://github.com/rivo/tview) and [`tcell`](https://github.com/gdamore/tcell).

---

### Layout Components

```
┌─────────────────────────────────────────────────────────────────┐
│ Batches (List) │ Jobs (List) │ Detail (TextView)                │
│                │             │ ──────────────────────────────── │
│                │             │ Action Buttons (Button Flex)     │
├─────────────────────────────────────────────────────────────────┤
│ Logs (TextView with ANSI color streaming)                       │
├─────────────────────────────────────────────────────────────────┤
│ Footer (Navigation & Shortcut hints)                            │
└─────────────────────────────────────────────────────────────────┘
```

---

### Key Components

- **`SetupUI(app, conf, logger, logReader)`:** Constructs lists, detail views, flex splits, global input captures, and asynchronous log streaming routines.
- **`ShowDetail(job, textView)`:** Formats and renders metadata and artifact paths with color highlights.
- **`RunCommandInModal(app, pages, title, startMsg, cmd, onFinish)`:** Launches a subprocess in a background goroutine and streams live output (`stdout`/`stderr`) into a centered modal dialog.
- **`Banner`:** Embedded ASCII art logo displayed on application launch.

---

### Navigation & Keybindings

- **`Tab` / `Right Arrow`:** Focus next pane (Batches → Jobs → Detail / Action Buttons).
- **`Shift+Tab` / `Left Arrow`:** Focus previous pane.
- **`Up` / `Down Arrow`:** Navigate entries in the active list.
- **`Enter`:** Select item or trigger active action button.
- **`ESC` / `Enter` (inside Modal):** Dismiss process execution modal dialog once finished.
- **`Ctrl+Q` / `Ctrl+C`:** Terminate application.
