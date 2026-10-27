# Runner Package (`internal/runner`)

The `runner` package constructs and configures external CLI commands (`os/exec.Cmd`) for `gocfl` and `ona` subprocess invocations.

---

### Command Builders

| Function | Subprocess & Command | Description |
| :--- | :--- | :--- |
| `CreateOCFLCmd(conf, job)` | `gocfl create <zip> <data> metadata:<meta> -i <sig> ...` | Assembles command to build standard OCFL `.zip` container. |
| `CreateReportCmd(conf, job, port)` | `gocfl display <zip> --display-fullreport <pdf> ...` | Assembles command to generate validation PDF report via local HTTP server on allocated port. |
| `ValidateOCFLCmd(conf, job)` | `gocfl validate <zip>` | Assembles command to validate OCFL container integrity and checksums. |
| `IngestOCFLCmd(conf, job)` | `ona ingest --status <json> -p <zip>` | Assembles command to ingest package into preservation storage. |
| `StatusOCFLCmd(conf, id)` | `ona status -i <id>` | Assembles command to query preservation ingest status for an ID. |

---

### Delegated Lifecycle Operations

For backwards compatibility and convenience, `runner` re-exports key lifecycle operations implemented in `internal/job`:
- `ExtractIngestID(resultJsonPath)`
- `CheckIngestResult(resultJsonPath)`
- `CreateUploadRecord(conf, job)`
- `HandleJobError(conf, job, cmdStr, output)`
- `HandleIngestError(conf, job, cmdStr, output, resultJsonPath)`
- `ArchiveJob(conf, job, resultJsonPath)`
