# Job Package (`internal/job`)

The `job` package defines the core domain models, filesystem scanning algorithms, path resolution rules, and lifecycle transitions for preservation units.

---

### Core Types

- **`Job`:** Represents a preservation unit within a batch, tracking file paths, signatures, titles, and processing status.
- **`JobStatus`:** Lifecycle states:
  - `StatusIncoming` (`"incoming"`): Item in incoming folder awaiting processing or ingest.
  - `StatusError` (`"error"`): Processing or ingest failed; item relocated to error folder.
  - `StatusArchived` (`"archived"`): Ingest succeeded; item and containers moved to archived folder.
- **`ActionType`:** Actions permitted on a job based on its state (`create_ocfl`, `create_report`, `validate`, `ingest`, `check_status`).

---

### Key Capabilities

1. **Batch & Ingest Scanning (`scanner.go`):**
   - `GetBatches(folder)`: Discovers direct batch directories.
   - `GetSignatures(...)`: Scans incoming, error, and archived directories for jobs, reading `info.json` metadata, and locating associated OCFL `.zip`, `.pdf` reports, `.upload.json`, and `_ona.json` files.

2. **Action State Machine (`job.go`):**
   - `job.AvailableActions()`: Dynamically calculates allowed actions based on current job state, preventing invalid operations.

3. **Standard Path Resolvers (`job.go`):**
   - `ExpectedZipPath(conf)`, `ExpectedReportPath(conf)`, `ExpectedResultJsonPath(conf)`, `ExpectedUploadPath(conf)`, `ExpectedErrorLogPath(conf)`.

4. **Lifecycle & Error Handling (`lifecycle.go`):**
   - `HandleJobError(...)`: Moves failed job directories and companion files to the batch error directory, writes error logs, and removes partial OCFL containers.
   - `HandleIngestError(...)`: Moves job directories and ONA result JSON to the error directory with execution logs.
   - `ArchiveJob(...)`: Relocates incoming job directories, OCFL `.zip` containers, result JSON, and upload markers to the archived directory upon successful ingest.
   - `CheckIngestResult(...)`: Parses ONA result JSON to detect errors or warnings.
   - `CreateUploadRecord(...)`: Generates `.upload.json` containing the RFC3339 timestamp.
