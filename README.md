<div align="center">
  <a href="https://github.com/JonasNowak/DankMacShell">
    <img src="assets/danklogo.svg" alt="Dms_GOP" width="200">
  </a>

  # Dms_GOP

  ### Apple Silicon macOS system monitoring CLI, Bubbletea TUI, and OpenAPI REST server

[![GitHub Release](https://img.shields.io/github/v/release/JonasNowak/dms-gop?style=for-the-badge&labelColor=101418&color=a6da95)](https://github.com/JonasNowak/dms-gop/releases)
[![GitHub License](https://img.shields.io/github/license/JonasNowak/dms-gop?style=for-the-badge&labelColor=101418&color=b9c8da)](https://github.com/JonasNowak/dms-gop/blob/master/LICENSE)

</div>

---

<div align="center">
<img src="https://github.com/user-attachments/assets/eb0b4e3e-e6ed-439d-a24b-640d15938510" width="600" alt="dms-gop screenshot" />
</div>

System monitoring tool with CLI, Bubbletea TUI, and REST API built in Go. Fast, single binary, JSON output, OpenAPI spec.

Use standalone for system monitoring, or as a companion for [DankMacShell](https://github.com/JonasNowak/DankMacShell) to unlock system information widgets. Themed dynamically via [dank-theme-engine](https://github.com/JonasNowak/dank-theme-engine).

---

## Installation

### Nix Flake

Run directly with:
```bash
nix run github:JonasNowak/dms-gop
```

Or add to your Nix configuration:

**nix-darwin:**
```nix
{
  imports = [ inputs.dms-gop.darwinModules.default ];
  programs.dms-gop.enable = true;
}
```

**Home Manager:**
```nix
{
  imports = [ inputs.dms-gop.homeManagerModules.default ];
  programs.dms-gop.enable = true;
}
```

### Go Install
```bash
go install github.com/JonasNowak/dms-gop/cmd/dms-gop@master
```

### Build from Source
```bash
# Build it
make

# Install to ~/.local/bin (with dgop alias symlink)
make && make install

# Or just run locally
go run ./cmd/dms-gop [command]
```

## Basic Commands

The binary is `dms-gop`, with a `dgop` alias symlink created on install.

```bash
# See all at once
dms-gop all
# or
dgop all

# Just CPU info
dms-gop cpu

# Memory usage
dms-gop memory

# Network interfaces
dms-gop network

# Disk usage and mounts
dms-gop disk

# Running processes (sorted by CPU usage)
dms-gop processes

# System load and uptime
dms-gop system

# Hardware info
dms-gop hardware

# GPU information (Apple Silicon GPU model; live utilization/power in Scope 2)
dms-gop gpu

# Lifetime I/O, open files, launchd label and bundle ID of one process
dms-gop process --pid 1234

# List available modules
dms-gop modules
```

## Meta Command

Mix and match any modules you want:

```bash
# Just CPU and memory
dms-gop meta --modules cpu,memory

# Everything except processes
dms-gop meta --modules cpu,memory,network,disk,system,hardware,gpu

# Everything (same as 'dms-gop all')
dms-gop meta --modules all
```

## JSON Output

Add `--json` to any command:

```bash
dms-gop cpu --json
dms-gop meta --modules gpu,memory --json
```

## Process Options

```bash
# Sort by memory instead of CPU
dms-gop processes --sort memory

# Limit to top 10
dms-gop processes --limit 10

# Skip CPU calculation for faster results
dms-gop processes --no-cpu

# Add per-process disk I/O, or connections and TCP rates
dms-gop processes --io --net

# List the processes behind each merged row (JSON output)
dms-gop processes --members --json

# Combine options
dms-gop meta --modules processes --sort memory --limit 20 --no-cpu

# Per-process extras in meta
dms-gop meta --modules processes,process-io,process-net,process-members --sort disk
```

## API Server

Start the REST API:

```bash
dms-gop server
```

Then hit these endpoints:

- **GET** `/gops/all` - All metrics combined
- **GET** `/gops/cpu` - CPU info
- **GET** `/gops/memory` - Memory usage
- **GET** `/gops/network` - Network interfaces
- **GET** `/gops/disk` - Disk usage
- **GET** `/gops/disk/mounts` - Disk mount points
- **GET** `/gops/processes?sort_by=memory&limit=10` - Top 10 processes by memory
- **GET** `/gops/processes?io=true&net=true` - Processes with disk I/O and network
- **GET** `/gops/process/{pid}` - Details of one process
- **GET** `/gops/system` - System load and uptime
- **GET** `/gops/hardware` - Hardware info
- **GET** `/gops/gpu` - GPU information
- **GET** `/gops/sensors` - Temperature, fan and power sensors
- **GET** `/gops/net-rate?interface=en0&all_interfaces=true` - Network rates
- **GET** `/gops/disk-rate` - Disk I/O rates
- **GET** `/gops/meta?modules=cpu,memory` - Dynamic modules
- **GET** `/gops/modules` - List available modules

API docs: http://localhost:63484/docs

## Examples

### Monitor system without slow CPU calculations
```bash
dms-gop meta --modules cpu,memory,network --no-cpu
```

### API: Get CPU and memory as JSON
```bash
curl http://localhost:63484/gops/meta?modules=cpu,memory
```

### API: Get all metrics
```bash
curl http://localhost:63484/gops/all
```

## Real-time Monitoring with Cursors

dms-gop supports cursor-based sampling for building real-time monitoring tools. Instead of relying on instantaneous snapshots, you can track system state changes over time for more accurate CPU usage calculations and network/disk rates.

The cursor system works by:
- Taking an initial measurement that establishes baseline metrics and timestamps
- Returning a base64-encoded cursor containing the current state data
- Using that cursor in subsequent calls to calculate precise percentages and rates over the sampling interval

This approach accounts for the actual time elapsed between measurements, making it ideal for monitoring tools that poll every few seconds.

### CPU Usage with Cursors

```bash
# First call - establishes baseline and returns cursor
dms-gop cpu --json
# Returns: {"usage":1.68, ..., "cursor":"eyJ0b3RhbCI6WzE2MjMwLjAzLDUuOTUsNTEyMy4yNV0..."}

# Wait a few seconds, then use cursor for accurate CPU calculations
sleep 3
dms-gop cpu --json --cursor "eyJ0b3RhbCI6WzE2MjMwLjAzLDUuOTUsNTEyMy4yNV0..."
# Returns more accurate usage percentages based on time delta
```

### Process Monitoring with Cursors

```bash
# First call - establishes process baseline
dms-gop processes --json --limit 5
# Returns: {"processes":[...], "cursor":"W3sicGlkIjoyODE2NTYsInRpY2tzIjozOS43Mix9XQ..."}

# Use cursor for accurate per-process CPU calculations
sleep 2
dms-gop processes --json --limit 5 --cursor "W3sicGlkIjoyODE2NTYsInRpY2tzIjozOS43Mix9XQ..."
```

### Network Rate Monitoring

```bash
# First call - establishes network baseline
dms-gop net-rate --json
# Returns: {"interfaces":[...], "cursor":"eyJ0aW1lc3RhbXAiOiIyMDI1LTA4LTExVDE2OjE1OjM1..."}

# Get real-time transfer rates
sleep 3
dms-gop net-rate --json --cursor "eyJ0aW1lc3RhbXAiOiIyMDI1LTA4LTExVDE2OjE1OjM1..."
# Returns: {"interfaces":[{"interface":"en0","rxrate":67771,"txrate":16994}]}

# Include all interfaces
dms-gop net-rate --json --all-interfaces
```

### Disk I/O Rate Monitoring

```bash
# Establish disk I/O baseline
dms-gop disk-rate --json
# Returns cursor for disk rate calculations

# Get real-time disk I/O rates
sleep 2
dms-gop disk-rate --json --cursor "eyJ0aW1lc3RhbXAiOiIyMDI1LTA4LTExVDE2OjE2..."
```

### Combined Monitoring with Meta Command

```bash
# Monitor CPU, processes, and network rates together
dms-gop meta --modules cpu,processes,net-rate --json --limit 10

# Use multiple cursors for comprehensive monitoring
dms-gop meta --modules cpu,processes,net-rate --json --limit 10 \
  --cpu-cursor "eyJ0b3RhbCI6WzE2MjMwLjAz..." \
  --proc-cursor "W3sicGlkIjoyODE2NTYsInRpY2tzIjo..." \
  --net-rate-cursor "eyJ0aW1lc3RhbXAiOiIyMDI1LTA4LTEx..."
```

## Theming (dank-theme-engine Integration)

dms-gop integrates with [dank-theme-engine](https://github.com/JonasNowak/dank-theme-engine) for dynamic theming with the following precedence (`theme.json` > `colors.json` > default):

1. **dank-theme-engine (`theme.json`)**: Watches `~/Library/Application Support/DMS/theme.json` (or a custom path via `DMS_THEME_JSON`) for changes and applies the palette automatically
2. **Legacy local fallback (`colors.json`)**: Uses `~/.config/dms-gop/colors.json` when no valid `theme.json` is present
3. **Built-in defaults**: Falls back cleanly if neither file is present or valid

Theme and keybind watching use `fsnotify` on the parent directories to detect atomic writes and live changes in real time.

## Development

```bash
# Build
make

# Run tests
make test

# Format code
make fmt

# Build and install
make && make install

# Clean build artifacts
make clean
```

## Requirements

- Go 1.27+ (or use the Nix flake)
- macOS on Apple Silicon (`aarch64-darwin`)

## Why Another Monitoring Tool?

Because nothing did what I wanted: no metrics server required, Go for speed and single-binary deployment, cleaner than bash scripts.

TL;DR: single binary CLI and server with JSON output, OpenAPI spec, and comprehensive system data.
