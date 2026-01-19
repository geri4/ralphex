# ralphex

[![build](https://github.com/umputun/ralphex/actions/workflows/ci.yml/badge.svg)](https://github.com/umputun/ralphex/actions/workflows/ci.yml)
[![Coverage Status](https://coveralls.io/repos/github/umputun/ralphex/badge.svg?branch=master)](https://coveralls.io/github/umputun/ralphex?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/umputun/ralphex)](https://goreportcard.com/report/github.com/umputun/ralphex)

Autonomous plan execution with Claude Code. A clean Go rewrite of ralph.py.

## Features

- **Task execution loop** - executes plan tasks one at a time with automatic retry
- **Multi-pass code review** - 8 parallel review agents plus codex integration
- **Streaming output** - real-time progress with timestamps and colors
- **Progress logging** - detailed execution logs for debugging
- **Multiple modes** - full execution, review-only, or codex-only

## Installation

### From source

```bash
go install github.com/umputun/ralphex/cmd/ralphex@latest
```

### Using Homebrew

```bash
brew install umputun/tap/ralphex
```

### From releases

Download the appropriate binary from [releases](https://github.com/umputun/ralphex/releases).

## Usage

```bash
# execute plan with task loop + reviews
ralphex docs/plans/feature.md

# use fzf to select plan
ralphex

# review-only mode (skip task execution)
ralphex --review docs/plans/feature.md

# codex-only mode (skip tasks and first claude review)
ralphex --codex-only

# with custom max iterations
ralphex --max-iterations 100 docs/plans/feature.md
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-m, --max-iterations` | Maximum task iterations | 50 |
| `-r, --review` | Skip task execution, run full review pipeline | false |
| `-c, --codex-only` | Skip tasks and first review, run only codex loop | false |
| `-d, --debug` | Enable debug logging | false |

## Requirements

- `claude` - Claude Code CLI
- `git` - for branch management
- `fzf` - for plan selection (optional)
- `codex` - for external review (optional)

## License

MIT License - see [LICENSE](LICENSE) file.
