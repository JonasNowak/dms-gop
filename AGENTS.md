# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Dms_GOP (dms-gop) is a system monitoring tool with CLI, Bubbletea TUI, and REST API, designed to provide comprehensive system metrics for Apple Silicon macOS. It's written in Go, compiles to a single binary, and serves as a companion for [DankMacShell](https://github.com/JonasNowak/DankMacShell) with dynamic theming via [dank-theme-engine](https://github.com/JonasNowak/dank-theme-engine).

## Development Commands

### Building
```bash
# Standard build (CGO_ENABLED=0 for maximum portability)
make

# Development build (with debug info)
make dev

# Build and install to ~/.local/bin (with dgop alias symlink)
make && make install
```

### Testing
```bash
# Run all tests
make test

# Run tests for specific package
go test -v ./...

# Run in Nix development shell
nix develop -c make test
```

### Code Quality
```bash
# Format code
make fmt

# Run go vet
make vet

# Update dependencies
make deps
```

### Running the Application
```bash
# Run locally without building
go run ./cmd/dms-gop [command]

# After building
./bin/dms-gop [command]
# or use the dgop alias
./bin/dgop [command]

# Start API server
./bin/dms-gop server
```

## Architecture

### Core Components

1. **CLI Entry Point** (`cmd/dms-gop/`)
   - `main.go`: Sets up Cobra commands and initializes GopsUtil
   - `commands.go`: Implements individual CLI commands (cpu, memory, network, etc.)
   - `server.go`: Starts the REST API server

2. **Data Collection** (`gops/`)
   - `gops.go`: Main utility class that orchestrates all metric collection
   - Individual collectors: `cpu.go`, `memory.go`, `disk.go`, `net.go`, `processes.go`, `system.go`, `hardware.go`
   - Each collector uses `github.com/shirou/gopsutil/v4` for system information

3. **API Layer** (`api/`)
   - `api/gops/`: HTTP handlers that wrap the gops collectors
   - `api/server/`: Server implementation
   - Uses Huma v2 framework for OpenAPI-compliant REST API

4. **Data Models** (`models/`)
   - Defines structures for all system metrics (CPU, Memory, Disk, Network, Process, etc.)
   - Shared between CLI output and API responses

### Key Design Patterns

- **Single Responsibility**: Each file in `gops/` handles one type of metric
- **Dependency Injection**: GopsUtil is passed to commands and API handlers
- **Process Sampling**: CPU and process metrics support sampling for performance optimization
- **Modular Commands**: Meta command allows combining multiple metrics dynamically
- **Cursor-based Monitoring**: Real-time monitoring with accurate rate calculations using base64-encoded state cursors

### API Server Details

- Default port: 63484
- OpenAPI docs available at `/docs`
- All endpoints under `/gops/` prefix
- Supports JSON output and query parameters for filtering/sorting

### Important Implementation Notes

1. **Architecture**: `aarch64-darwin` only (Apple Silicon macOS)
2. **CGO Disabled**: Built with `CGO_ENABLED=0` for maximum portability and consistency
3. **Process CPU Calculation**: Can be disabled with `--no-cpu` flag for faster results
4. **Error Handling**: Non-critical errors are logged but don't stop execution (graceful degradation)
5. **Sampling Support**: CPU, process, network-rate, and disk-rate commands support cursor-based sampling
6. **Theming**: Precedence is `theme.json` (`DMS_THEME_JSON` or `~/Library/Application Support/DMS/theme.json`) > `~/.config/dms-gop/colors.json` fallback > built-in defaults; parent directories are watched via `fsnotify` for live updates
7. **Nix Support**: Full flake support with `darwinModules.default` and `homeManagerModules.default` for system and user-level installation
