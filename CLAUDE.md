# ralphex

Autonomous plan execution with Claude Code - Go rewrite of ralph.py.

## Build Commands

```bash
make build      # build binary to .bin/ralphex
make test       # run tests with coverage
make lint       # run golangci-lint
make fmt        # format code
make install    # install to ~/.local/bin
```

## Project Structure

```
cmd/ralphex/    # main entry point with all logic
docs/plans/     # plan files location
```

## Code Style

- Use jessevdk/go-flags for CLI parsing
- All comments lowercase except godoc
- Table-driven tests with testify
- 80%+ test coverage target

## Key Patterns

- Signal-based completion detection (COMPLETED, FAILED, REVIEW_DONE signals)
- Streaming output with timestamps
- Progress logging to files
- Multiple execution modes: full, review-only, codex-only

## Testing

```bash
go test ./...           # run all tests
go test -cover ./...    # with coverage
```
