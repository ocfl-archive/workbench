# Workbench Configuration Guide

This document describes the configuration file format, parameter specifications, placeholder resolution rules, and default search paths used by Workbench (`create_container`).

---

### Configuration File Format

Workbench uses [TOML](https://toml.io/) for its configuration files. A typical configuration looks like this:

```toml
# Root batch directory containing individual batch folders
batches = "C:/temp/zhbluzern/raphael_upload/"

# Ingest and destination folder templates (supports {batch} token)
input = "{batch}/incoming/"
ocfl = "{batch}/ocfl/"
report = "{batch}/report/"
error = "{batch}/error/"
archived = "{batch}/archived/"

[gocfl]
config = "C:/daten/go/dev/gocfl-cli/config/gocfl2.toml"

[ona]
config = "C:/daten/go/dev/workbench/config/ona.yml"
```

---

### Configuration Parameters

| Parameter | Type | Required | Description | Default Fallback |
| :--- | :--- | :--- | :--- | :--- |
| `batches` | String | Yes | Base folder containing one or more batch subdirectories. | None |
| `input` | String | No | Path or template for incoming ingest data. | `{batches}/{batch}/incoming` |
| `ocfl` | String | No | Path or template for generated OCFL `.zip` containers. | `{batches}/{batch}/ocfl` |
| `report` | String | No | Path or template for generated `.pdf` reports. | `{batches}/{batch}/report` |
| `error` | String | No | Path or template for error logs and failed items. | `{batches}/{batch}/error` |
| `archived` | String | No | Path or template for finalized/archived packages. | `{batches}/{batch}/archived` |
| `gocfl.config` | String | No | Path to the `gocfl` TOML configuration file. | None |
| `ona.config` | String | No | Path to the `ona` YAML configuration file. | None |

---

### Dynamic Path Resolution (`{batch}`)

The folder parameters (`input`, `ocfl`, `report`, `error`, `archived`) support dynamic batch token expansion:

1. **Token Replacement:**
   If a parameter contains the literal substring `{batch}`, Workbench replaces all occurrences of `{batch}` with the active batch name:
   ```
   Pattern:  C:/data/batches/{batch}/incoming
   Batch:    batch_2026_01
   Result:   C:/data/batches/batch_2026_01/incoming
   ```

2. **Static Paths:**
   If a path is explicitly configured without `{batch}`, the static path is used directly across all batches:
   ```
   Pattern:  C:/shared/global_reports
   Batch:    batch_2026_01
   Result:   C:/shared/global_reports
   ```

3. **Fallback Convention:**
   If a folder parameter is omitted or left empty (`""`), Workbench defaults to joining the root `batches` directory, the current batch name, and the standard subfolder name:
   ```
   batches:  C:/data/batches
   Batch:    batch_2026_01
   Result:   C:/data/batches/batch_2026_01/incoming (for input)
             C:/data/batches/batch_2026_01/ocfl     (for ocfl)
             C:/data/batches/batch_2026_01/report   (for report)
   ```

---

### Configuration Discovery Order

When launching `create_container`, configuration loading follows this precedence order:

1. **Explicit CLI Flag:**
   Specified via the `-config` flag:
   ```bash
   create_container -config /path/to/custom_workbench.toml
   ```
2. **User Home Directory:**
   `~/wb/workbench.toml` (or `%USERPROFILE%\wb\workbench.toml` on Windows).
3. **System Opt Directory:**
   `/opt/wb/workbench.toml`
4. **System Etc Directory:**
   `/etc/wb/workbench.toml`

If no valid configuration file is found in any of these locations, `create_container` exits with an error.
