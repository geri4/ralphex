// Package main provides ralphex - autonomous plan execution with Claude Code.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
)

// opts holds all command-line options.
type opts struct {
	MaxIterations int  `short:"m" long:"max-iterations" default:"50" description:"maximum task iterations"`
	Review        bool `short:"r" long:"review" description:"skip task execution, run full review pipeline"`
	CodexOnly     bool `short:"c" long:"codex-only" description:"skip tasks and first review, run only codex loop"`
	Debug         bool `short:"d" long:"debug" description:"enable debug logging"`

	PlanFile string `positional-arg-name:"plan-file" description:"path to plan file (optional, uses fzf if omitted)"`
}

var revision = "unknown"

const plansDir = "docs/plans"

func main() {
	fmt.Printf("ralphex %s\n", revision)

	var o opts
	parser := flags.NewParser(&o, flags.Default)
	parser.Usage = "[OPTIONS] [plan-file]"

	args, err := parser.Parse()
	if err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	// handle positional argument
	if len(args) > 0 {
		o.PlanFile = args[0]
	}

	// setup context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, o); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, o opts) error {
	// check dependencies
	for _, dep := range []string{"claude", "git"} {
		if _, err := exec.LookPath(dep); err != nil {
			return fmt.Errorf("%s not found in PATH", dep)
		}
	}

	// select plan file
	planFile, err := selectPlan(ctx, o.PlanFile, o.Review || o.CodexOnly)
	if err != nil {
		return err
	}

	skipTasks := o.Review || o.CodexOnly
	if planFile == "" && !skipTasks {
		return errors.New("plan file required for task execution")
	}

	// create branch if on main/master
	if planFile != "" {
		if err := createBranchIfNeeded(ctx, planFile); err != nil {
			return err
		}
	}

	// run main loop
	return runLoop(ctx, o, planFile)
}

func selectPlan(ctx context.Context, planFile string, optional bool) (string, error) {
	if planFile != "" {
		if _, err := os.Stat(planFile); err != nil {
			return "", fmt.Errorf("plan file not found: %s", planFile)
		}
		return planFile, nil
	}

	// for review-only modes, plan is optional
	if optional {
		return "", nil
	}

	// use fzf to select plan
	return selectPlanWithFzf(ctx)
}

func selectPlanWithFzf(ctx context.Context) (string, error) {
	if _, err := os.Stat(plansDir); err != nil {
		return "", fmt.Errorf("plans directory not found: %s", plansDir)
	}

	if _, err := exec.LookPath("fzf"); err != nil {
		return "", errors.New("fzf not found, please provide plan file as argument")
	}

	// find plan files (excluding completed/)
	plans, err := filepath.Glob(filepath.Join(plansDir, "*.md"))
	if err != nil || len(plans) == 0 {
		return "", fmt.Errorf("no plans found in %s", plansDir)
	}

	// auto-select if single plan
	if len(plans) == 1 {
		fmt.Printf("auto-selected: %s\n", plans[0])
		return plans[0], nil
	}

	// use fzf for selection
	cmd := exec.CommandContext(ctx, "fzf",
		"--prompt=select plan: ",
		"--preview=head -50 {}",
		"--preview-window=right:60%",
	)
	cmd.Stdin = strings.NewReader(strings.Join(plans, "\n"))
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return "", errors.New("no plan selected")
	}

	return strings.TrimSpace(string(out)), nil
}

func createBranchIfNeeded(ctx context.Context, planFile string) error {
	// get current branch
	out, err := exec.CommandContext(ctx, "git", "branch", "--show-current").Output()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	currentBranch := strings.TrimSpace(string(out))
	if currentBranch != "main" && currentBranch != "master" {
		return nil // already on feature branch
	}

	// extract branch name from filename
	name := strings.TrimSuffix(filepath.Base(planFile), ".md")
	// remove date prefix like "2024-01-15-"
	re := regexp.MustCompile(`^[\d-]+`)
	branchName := strings.TrimLeft(re.ReplaceAllString(name, ""), "-")
	if branchName == "" {
		branchName = name
	}

	fmt.Printf("creating branch: %s\n", branchName)
	if err := exec.CommandContext(ctx, "git", "checkout", "-b", branchName).Run(); err != nil { //nolint:gosec // branch name from plan filename
		return fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}

	return nil
}

func runLoop(ctx context.Context, o opts, planFile string) error {
	startTime := time.Now()

	// get current branch for logging
	out, _ := exec.CommandContext(ctx, "git", "branch", "--show-current").Output()
	branch := strings.TrimSpace(string(out))

	// ensure progress files are gitignored
	if err := ensureGitignore(ctx); err != nil {
		return err
	}

	// determine mode
	mode := "full"
	if o.CodexOnly {
		mode = "codex-only"
	} else if o.Review {
		mode = "review"
	}

	// create progress file
	progressPath := getProgressFilename(planFile, mode)
	progressFile, err := os.Create(progressPath) //nolint:gosec // path derived from plan filename
	if err != nil {
		return fmt.Errorf("failed to create progress file: %w", err)
	}
	defer progressFile.Close()

	// write progress header
	planStr := planFile
	if planStr == "" {
		planStr = "(no plan - review only)"
	}
	fmt.Fprintf(progressFile, "# Ralph Progress Log\n")
	fmt.Fprintf(progressFile, "Plan: %s\n", planStr)
	fmt.Fprintf(progressFile, "Branch: %s\n", branch)
	fmt.Fprintf(progressFile, "Mode: %s\n", mode)
	fmt.Fprintf(progressFile, "Started: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(progressFile, "%s\n\n", strings.Repeat("-", 60))

	modeStr := ""
	if mode != "full" {
		modeStr = fmt.Sprintf(" (%s mode)", mode)
	}
	fmt.Printf("starting ralph loop: %s (max %d iterations)%s\n", planStr, o.MaxIterations, modeStr)
	fmt.Printf("branch: %s\n", branch)
	fmt.Printf("progress log: %s\n\n", progressPath)

	// TODO: implement task execution loop
	// TODO: implement review passes
	// TODO: implement codex integration

	elapsed := time.Since(startTime)
	fmt.Printf("\ncompleted in %s\n", formatElapsed(elapsed))
	fmt.Fprintf(progressFile, "\n%s\n", strings.Repeat("-", 60))
	fmt.Fprintf(progressFile, "Completed: %s (%s)\n", time.Now().Format("2006-01-02 15:04:05"), formatElapsed(elapsed))

	return nil
}

func ensureGitignore(ctx context.Context) error {
	// check if already ignored
	if err := exec.CommandContext(ctx, "git", "check-ignore", "-q", "progress-test.txt").Run(); err == nil {
		return nil // already ignored
	}

	// add to .gitignore
	f, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644) //nolint:gosec // .gitignore needs world-readable
	if err != nil {
		return fmt.Errorf("failed to open .gitignore: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n# ralph progress logs\nprogress-*.txt\n"); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	fmt.Println("added progress-*.txt to .gitignore")
	return nil
}

func getProgressFilename(planFile, mode string) string {
	if planFile != "" {
		stem := strings.TrimSuffix(filepath.Base(planFile), ".md")
		switch mode {
		case "codex-only":
			return fmt.Sprintf("progress-%s-codex.txt", stem)
		case "review":
			return fmt.Sprintf("progress-%s-review.txt", stem)
		default:
			return fmt.Sprintf("progress-%s.txt", stem)
		}
	}

	switch mode {
	case "codex-only":
		return "progress-codex.txt"
	case "review":
		return "progress-review.txt"
	default:
		return "progress.txt"
	}
}

func formatElapsed(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	secs := seconds % 60
	if minutes < 60 {
		return fmt.Sprintf("%dm%ds", minutes, secs)
	}
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%dh%dm%ds", hours, mins, secs)
}
