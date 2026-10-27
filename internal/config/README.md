# Config Package (`internal/config`)

The `config` package manages TOML configuration loading, default configuration discovery, dynamic path interpolation, and conditional execution profiles.

---

### Key Structs

- **`WBConfig`:** Represents the root configuration structure mapped from `workbench.toml`.
- **`GOCFLConfig`:** Holds parameters for the `gocfl` CLI tool (`config`, `largeconfig`, `largelimit`).
- **`ONAConfig`:** Holds parameters for the `ona` preservation tool (`config`).

---

### Features & Behavior

1. **Automatic Config Discovery:**
   `LoadConfig(path)` checks the explicit path argument if provided. When empty, it sequentially searches:
   - `~/wb/workbench.toml` (or `%USERPROFILE%\wb\workbench.toml`)
   - `/opt/wb/workbench.toml`
   - `/etc/wb/workbench.toml`

2. **Dynamic Path Expansion (`{batch}`):**
   Methods like `GetInputFolder(batch)`, `GetOcflFolder(batch)`, `GetReportFolder(batch)`, `GetErrorFolder(batch)`, and `GetArchivedFolder(batch)` replace `{batch}` tokens with the active batch name or fall back to joining `batches + batch + subfolder`.

3. **GOCFL Profile Switching:**
   `GetGOCFLConfig(dataFolder)` dynamically switches from standard `config` to `largeconfig` when the recursive file count in `dataFolder` exceeds `largelimit`.
