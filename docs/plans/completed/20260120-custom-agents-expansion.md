# Custom Agents with Template Expansion

## Overview
Create a set of general-purpose agent files that can be referenced in prompts via `{{agent:name}}` syntax. When expanded, these agents get wrapped with Task tool invocation instructions, telling Claude to run them as separate parallel tasks.

**This completely replaces the existing custom agents implementation.** The old system (hardcoded `buildFirstReviewPromptWithCustomAgents()` that bypassed prompt files) is removed. The new system uses `{{agent:name}}` template syntax in prompt files, which get expanded during normal template processing.

Users can reference both:
- Custom ralphex agents via `{{agent:name}}` - expands to Task tool instructions with agent content
- Built-in Claude Code agents (like `qa-expert`, `go-smells-expert`) - referenced directly by name in prompts

## Context (from discovery)
- Files involved: `pkg/config/config.go`, `pkg/processor/prompts.go`, `pkg/config/defaults/`
- Current template variables: `{{PLAN_FILE}}`, `{{PROGRESS_FILE}}`, `{{GOAL}}`, `{{CODEX_OUTPUT}}`
- Agent loading: `loadAgents()` in config.go scans `~/.config/ralphex/agents/*.txt`
- Prompt replacement: `replacePromptVariables()` in prompts.go
- **To remove**: `buildFirstReviewPromptWithCustomAgents()` function and conditional in `buildFirstReviewPrompt()`

## Development Approach
- **Testing approach**: TDD - write tests first
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes**

## Design

### Agent Syntax
In any prompt file, use `{{agent:name}}` to reference an agent:
```
## Step 2: Run Review Agents

{{agent:quality}}
{{agent:testing}}
{{agent:documentation}}
```

### Expansion Output
Each `{{agent:name}}` expands to:
```
Use the Task tool to launch a general-purpose agent with this prompt:
"[contents of name.txt agent file]"
Report findings only - no positive observations.
```

### Default Agents (5 universal agents)
Installed to `~/.config/ralphex/agents/` on first run.

Base these on Claude Code's built-in agents but make them **language-agnostic**:

| ralphex agent | Based on Claude Code agent | Purpose |
|---------------|---------------------------|---------|
| **implementation.txt** | `implementation-reviewer` | Verifies code achieves stated goals, checks proper wiring |
| **quality.txt** | `qa-expert` | Reviews for bugs, security issues, race conditions, logic errors |
| **documentation.txt** | `documentation-expert` | Checks if README.md, CLAUDE.md need updates |
| **simplification.txt** | `go-simplify-expert` | Detects over-engineering, unnecessary complexity (remove Go-specific parts) |
| **testing.txt** | `go-test-expert` | Reviews test coverage, quality, missing edge cases (remove Go-specific parts) |

Reference the Claude Code agent descriptions from the Task tool documentation:
- `qa-expert` - "Reviews code for bugs, logic errors, security vulnerabilities, code quality issues"
- `implementation-reviewer` - "Verifies correctness of approach, proper wiring, complete coverage of requirements"
- `documentation-expert` - "Checks README.md, CLAUDE.md, and plan files for required updates"
- `go-simplify-expert` - "Detect over-engineered and overcomplicated code - unnecessary abstractions, excessive layers"
- `go-test-expert` - "Write, review, enhance, or modernize tests - coverage, table-driven tests, mocks"

Strip Go-specific references (like "table-driven tests", "moq", "testify") and keep universal review criteria

### Loading Priority
1. On first run: embedded defaults from `pkg/config/defaults/agents/` are copied to `~/.config/ralphex/agents/`
2. At runtime: agents are loaded from `~/.config/ralphex/agents/` only
3. Users can edit/delete/add agent files - changes take effect on next run
4. To restore defaults: delete agent files and restart ralphex

## Implementation Steps

### Task 1: Create embedded default agent files
- [x] create `pkg/config/defaults/agents/implementation.txt`
- [x] create `pkg/config/defaults/agents/quality.txt`
- [x] create `pkg/config/defaults/agents/documentation.txt`
- [x] create `pkg/config/defaults/agents/simplification.txt`
- [x] create `pkg/config/defaults/agents/testing.txt`
- [x] update embed directive in config.go to include agents/
- [x] write tests verifying embedded agents load correctly
- [x] run `go test ./pkg/config` - must pass before task 2

### Task 2: Implement agent installation on first run
- [x] modify `installDefaults()` to copy embedded agents to user dir
- [x] skip installation if agent file already exists (don't overwrite user edits)
- [x] write tests for agent installation with mock filesystem
- [x] run `go test ./pkg/config` - must pass before task 3

### Task 3: Implement agent template expansion
- [x] add `expandAgentReferences()` function in prompts.go
- [x] parse `{{agent:name}}` patterns and look up agent content
- [x] wrap agent content with Task tool instruction prefix
- [x] handle missing agent (log warning, leave reference unexpanded)
- [x] write tests for agent expansion with various patterns
- [x] run `go test ./pkg/processor` - must pass before task 4

### Task 4: Integrate agent expansion into prompt processing
- [x] call `expandAgentReferences()` in `replacePromptVariables()`
- [x] **remove `buildFirstReviewPromptWithCustomAgents()` function entirely**
- [x] **remove conditional check in `buildFirstReviewPrompt()` - always use template file**
- [x] update existing tests to expect agent expansion
- [x] write tests for end-to-end prompt building with agents
- [x] run `go test ./pkg/processor` - must pass before task 5

### Task 5: Update review prompts to use agent syntax
- [x] update `review_first.txt` to use `{{agent:*}}` syntax for custom agents (e.g., `{{agent:quality}}`, `{{agent:testing}}`)
- [x] default prompts reference only custom ralphex agents (quality, implementation, documentation, simplification, testing)
- [x] users can edit prompts to add built-in Claude agents if desired (qa-expert, go-smells-expert, etc.)
- [x] update `review_second.txt` if applicable
- [x] run `go test ./pkg/processor` - must pass before task 6

### Task 6: Verify acceptance criteria
- [x] write integration test: verify default agents installed on first run (config package)
- [x] write integration test: verify `{{agent:name}}` expands correctly (processor package)
- [x] write integration test: verify user agent overrides work (config package)
- [x] write integration test: verify missing agent references handled gracefully (processor package)
- [x] run full test suite: `go test ./...`
- [x] run linter: `golangci-lint run`

### Task 7: [Final] Update documentation
- [x] update README.md with custom agents section
- [x] update CLAUDE.md:
  - [x] change "Custom agents: `~/.config/ralphex/agents/*.txt` (replaces built-in 8 agents)" to describe new `{{agent:name}}` syntax
  - [x] document that default agents are installed on first run
  - [x] explain users can override defaults or add custom agents
  - [x] note built-in Claude agents can still be referenced directly in prompts
- [x] move this plan to `docs/plans/completed/`

## Technical Details

### Agent File Format
Plain text, no special formatting required. The entire file content becomes the agent prompt:
```
Review code for bugs, security issues, race conditions, and logic errors.
Focus on:
- Potential nil pointer dereferences
- Resource leaks (unclosed files, connections)
- Concurrency issues (data races, deadlocks)
- Security vulnerabilities (injection, improper validation)
Report problems only - no positive observations.
```
Note: Comments are not supported - all lines become part of the agent prompt.

### Expansion Template
```go
const agentExpansionTemplate = `Use the Task tool to launch a general-purpose agent with this prompt:
%q

Report findings only - no positive observations.`
```

### Agent Reference Regex
```go
var agentRefPattern = regexp.MustCompile(`\{\{agent:([a-zA-Z0-9_-]+)\}\}`)
```

## Post-Completion

**E2E verification with toy project (see CLAUDE.md for instructions):**
- Test with fresh config dir (agents get installed)
- Verify review phase uses expanded agent prompts
- Verify agents run in parallel as Task tool calls
