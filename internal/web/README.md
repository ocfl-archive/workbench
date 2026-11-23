# web

`internal/web` implements the HTTP web server, Server-Sent Events (SSE) broadcasting, execution management, HTML template rendering, and embedded static asset delivery for the `wbweb` Web-GUI.

---

### Key Components

- **`server.go`:** Coordinates HTTP routes (`/`, `/batches`, `/batch/{batch}`, `/batch/{batch}/job/{basename}`, action endpoints, `/events`, and `/static/`), data modeling, and template compilation.
- **`broadcaster.go`:** Manages real-time Server-Sent Events (SSE) subscriptions, ring-buffered log history, and provides an `io.Writer` bridge for `zerolog`.
- **`runner.go`:** Wraps external CLI command executions and the batch pipeline, capturing standard output and error streams in real time for web terminal streaming.
- **`embed.go`:** Embeds static CSS, JS, icons, and `.gohtml` templates using Go's `embed.FS`.
- **`templates/`:** DaisyUI and Tailwind HTML templates (`index.gohtml`, `batches.gohtml`, `jobs.gohtml`, `detail.gohtml`).
- **`static/js/wbweb.js`:** Client-side controller and Web Terminal Emulator with ANSI color parsing, auto-scrolling, HTMX triggers, and theme switching.
