# Project Guidelines

## Issue Tracking

This project uses GitHub Issues for tracking bugs, feature requests, and tasks.

## Go Development

### Build & Test

```bash
make build      # Build the binary
make test       # Run tests
make coverage   # Run tests with coverage report
make lint       # Run golangci-lint
make clean      # Remove build artifacts
```

### Code Style

- Use `gofmt` for formatting
- Run `go vet` before committing
- Keep functions small and focused
- Prefer explicit error handling over panics

### Project Structure

- `main.go` - Entry point, CLI handling, command routing
- `config/` - Configuration loading and deck ignore logic
- `parser/` - Markdown file parsing and card extraction
- `storage/` - Spaced repetition state (SM-2), persistence
- `tui/` - Interactive terminal UI (Bubble Tea)

### Testing

Run `make coverage` to generate a coverage report. This produces:
- `coverage.out` — raw coverage profile
- `coverage.html` — visual HTML report (open in browser)

Quick coverage check: `go test ./... -cover`

**Readability principles:**
- Use table-driven tests when inputs/outputs vary but structure is the same
- Keep tests with unique setup (temp dirs, complex nested structs) as individual functions
- Every test name should describe the scenario, not the expected result
- Prefer `t.Run` subtests within tables for clear failure output
