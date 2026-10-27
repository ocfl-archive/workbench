# Util Package (`internal/util`)

The `util` package provides low-level filesystem validation, safe directory/file movement with cross-device copy fallbacks, and TCP port allocation helpers.

---

### Functions

- **`EnsureTargetDirectory(dir)`:** Validates that a target path exists or can be created within an existing parent directory. Prevents unintentional deep directory creation if parent paths are misconfigured.
- **`MoveDirectory(src, dst)`:** Moves a directory using `os.Rename`. If `os.Rename` fails (e.g., across separate filesystem mounts/drives), it falls back to recursive copy and removal.
- **`MoveFile(src, dst)`:** Moves a file using `os.Rename`, falling back to byte-stream copy and source deletion across different filesystems.
- **`GetFreePort()`:** Binds to ephemeral port `:0` on `localhost` to discover and return an available local TCP port. Used when serving HTTP endpoints for `gocfl display` PDF report generation.
