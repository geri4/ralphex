# Add Configuration System with Embedded Defaults

## Overview
Make ralphex fully customizable via configuration files with sensible defaults.
- Directory-based config structure at `~/.config/ralphex/`
- Embedded defaults installed on first run
- Single config location `~/.config/ralphex/`
- Supports prompts, claude/codex settings, timing, paths, and custom review agents

## Context (from discovery)
- Files involved: `cmd/ralphex/main.go`, `pkg/processor/prompts.go`, `pkg/processor/runner.go`, `pkg/executor/executor.go`, `pkg/executor/codex.go`
- Config values found: 4 prompts, claude command/args, codex model/timeout/sandbox, iteration_delay, task_retry_count, plans_dir, custom agents
- Dependencies: embed.FS for defaults, gopkg.in/ini.v1 for config parsing

## Development Approach
- **Testing approach**: TDD - write tests first
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes**

## Configuration Structure

```
~/.config/ralphex/
├── config                  # main config file (key=value)
├── prompts/
│   ├── task.txt           # task execution prompt
│   ├── review_first.txt   # first review prompt (8 agents)
│   ├── review_second.txt  # second review prompt (2 agents)
│   └── codex.txt          # codex evaluation prompt
└── agents/                 # custom review agents (optional)
    ├── agent_1.txt        # custom agent prompt (if present, used instead of built-in)
    └── agent_2.txt        # another custom agent
```

## Config File Format

Uses `gopkg.in/ini.v1` - supports `#` and `;` comments.

```ini
# ralphex configuration
; comments can use # or ;

# claude executor
claude_command = claude
claude_args = --dangerously-skip-permissions --output-format stream-json --verbose

# codex executor
codex_enabled = true
codex_command = codex
codex_model = gpt-5.2-codex
codex_reasoning_effort = xhigh
codex_timeout_ms = 3600000
codex_sandbox = read-only

# timing
iteration_delay_ms = 2000
task_retry_count = 1

# paths
plans_dir = docs/plans
```

## Config Precedence (highest to lowest)
1. CLI flags (--max-iterations, --review, etc.)
2. User config `~/.config/ralphex/config`
3. Embedded defaults

## Implementation Steps

### Task 1: Create config package with types and defaults
- [x] create `pkg/config/config.go` with Config struct
- [x] define all config fields with json/default tags
- [x] create embedded defaults using embed.FS
- [x] write tests for Config struct defaults
- [x] run `go test ./pkg/config` - must pass before task 2

### Task 2: Implement config file loading with ini library
- [x] add `gopkg.in/ini.v1` dependency
- [x] create `pkg/config/parser.go` using ini.Load()
- [x] map ini keys to Config struct fields
- [x] write tests for loading config with various inputs
- [x] run `go test ./pkg/config` - must pass before task 3

### Task 3: Implement config loading with precedence
- [x] create `pkg/config/loader.go` with Load() function
- [x] implement precedence: user config -> embedded defaults
- [x] create user config dir if missing, install defaults
- [x] write tests for loading and first-run installation
- [x] run `go test ./pkg/config` - must pass before task 4

### Task 4: Implement prompt file loading
- [x] create `pkg/config/prompts.go` for loading prompt files
- [x] load from `~/.config/ralphex/prompts/` directory
- [x] fall back to embedded defaults if file missing
- [x] write tests for prompt loading with mock filesystem
- [x] run `go test ./pkg/config` - must pass before task 5

### Task 5: Implement custom agents loading
- [x] create `pkg/config/agents.go` for loading custom agents
- [x] scan `~/.config/ralphex/agents/` directory
- [x] each .txt file becomes an agent with filename as name
- [x] write tests for agent loading
- [x] run `go test ./pkg/config` - must pass before task 6

### Task 6: Integrate config into main.go
- [x] load config at startup before processing
- [x] apply config values to processor and executor configs
- [x] CLI flags override config values
- [x] update main.go tests
- [x] run `go test ./cmd/ralphex` - must pass before task 7

### Task 7: Update processor to use config
- [x] modify Runner to accept prompts from config
- [x] modify Runner to use iteration_delay from config
- [x] modify Runner to use task_retry_count from config
- [x] update processor tests
- [x] run `go test ./pkg/processor` - must pass before task 8

### Task 8: Update executors to use config
- [x] modify ClaudeExecutor to use claude_command, claude_args from config
- [x] modify CodexExecutor to use codex_* settings from config
- [x] add codex_enabled check to skip codex phase
- [x] update executor tests
- [x] run `go test ./pkg/executor` - must pass before task 9

### Task 9: Implement custom agents in review phase
- [x] modify review phase to use custom agents if configured
- [x] run each custom agent prompt via Task subagent
- [x] fall back to built-in agents if no custom agents
- [x] update processor tests for custom agents
- [x] run `go test ./pkg/processor` - must pass before task 10

### Task 10: Create embedded default config files
- [x] create `pkg/config/defaults/config` with all settings
- [x] create `pkg/config/defaults/prompts/task.txt`
- [x] create `pkg/config/defaults/prompts/review_first.txt`
- [x] create `pkg/config/defaults/prompts/review_second.txt`
- [x] create `pkg/config/defaults/prompts/codex.txt`
- [x] add comprehensive comments explaining each option
- [x] verify embed.FS loads all files correctly
- [x] run `go test ./pkg/config` - must pass before task 11

### Task 11: Verify acceptance criteria
- [x] verify ralphex works with no config (embedded defaults)
- [x] verify user config is installed on first run
- [x] verify CLI flags override config values
- [x] verify custom agents work
- [x] run full test suite: `go test ./...`
- [x] run e2e test with toy project

### Task 12: [Final] Update documentation
- [x] update README.md with configuration section
- [x] update CLAUDE.md with config patterns
- [x] move this plan to `docs/plans/completed/`

## Technical Details

### Config struct
```go
type Config struct {
    // claude
    ClaudeCommand string
    ClaudeArgs    string

    // codex
    CodexEnabled         bool
    CodexCommand         string
    CodexModel           string
    CodexReasoningEffort string
    CodexTimeoutMs       int
    CodexSandbox         string

    // timing
    IterationDelayMs int
    TaskRetryCount   int

    // paths
    PlansDir string

    // prompts (loaded separately)
    TaskPrompt         string
    ReviewFirstPrompt  string
    ReviewSecondPrompt string
    CodexPrompt        string

    // custom agents (loaded separately)
    CustomAgents []CustomAgent
}

type CustomAgent struct {
    Name   string
    Prompt string
}
```

### Embedded defaults
```go
//go:embed defaults/*
var defaultsFS embed.FS
```

### Standard paths
- macOS: `~/.config/ralphex/`
- Linux: `~/.config/ralphex/` (XDG_CONFIG_HOME if set)

## Post-Completion

**Manual verification:**
- Test with fresh user (no config exists)
- Test with custom agents
- Verify prompts render correctly with variables
