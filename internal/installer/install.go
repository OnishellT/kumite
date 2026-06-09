package installer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"kumite/internal/shell"
)

const defaultCommandTimeout = 30 * time.Minute

type Options struct {
	Context        context.Context
	SourceDir      string
	BinDir         string
	RunGlobalSetup bool
	DryRun         bool
	CommandTimeout time.Duration
	Stdout         io.Writer
	Stderr         io.Writer
}

type Result struct {
	BinPath string
	Steps   []StepResult
	Warning string
}

type StepResult struct {
	Name   string
	Detail string
	Done   bool
	Err    error
}

func Run(options Options) (Result, error) {
	options = withDefaults(options)
	target := filepath.Join(options.BinDir, "kumite")
	result := Result{BinPath: target}

	steps := buildSteps(options, target)
	for _, step := range steps {
		stepResult := runStep(options, step)
		result.Steps = append(result.Steps, stepResult)
		if stepResult.Err != nil {
			return result, stepResult.Err
		}
	}

	if !pathContains(options.BinDir, os.Getenv("PATH")) {
		result.Warning = fmt.Sprintf("%s is not on PATH; add it before running kumite from every shell", options.BinDir)
		fmt.Fprintln(options.Stderr, "warning:", result.Warning)
	}

	return result, nil
}

func runStep(options Options, step installStep) StepResult {
	result := StepResult{Name: step.name, Detail: step.detail}
	if options.DryRun {
		result.Done = true
		fmt.Fprintf(options.Stdout, "$ %s\n", step.detail)
		return result
	}

	if err := step.run(options); err != nil {
		result.Err = err
		return result
	}
	result.Done = true

	return result
}

type installStep struct {
	name   string
	detail string
	run    func(Options) error
}

func buildSteps(options Options, target string) []installStep {
	steps := []installStep{
		{
			name:   "Create bin directory",
			detail: fmt.Sprintf("mkdir -p %s", options.BinDir),
			run: func(options Options) error {
				return os.MkdirAll(options.BinDir, 0o755)
			},
		},
		{
			name:   "Build kumite",
			detail: fmt.Sprintf("go build -o %s ./cmd/kumite", target),
			run: func(options Options) error {
				return runCommand(options, "go", "build", "-o", target, "./cmd/kumite")
			},
		},
	}
	if options.RunGlobalSetup {
		steps = append(steps, installStep{
			name:   "Run global setup",
			detail: fmt.Sprintf("%s setup --global --keep-going", target),
			run: func(options Options) error {
				return runCommand(options, target, "setup", "--global", "--keep-going")
			},
		})
	}

	return steps
}

func withDefaults(options Options) Options {
	if options.Context == nil {
		options.Context = context.Background()
	}
	if options.SourceDir == "" {
		options.SourceDir = "."
	}
	if options.BinDir == "" {
		options.BinDir = defaultBinDir()
	}
	if options.CommandTimeout == 0 {
		options.CommandTimeout = defaultCommandTimeout
	}
	if options.Stdout == nil {
		options.Stdout = io.Discard
	}
	if options.Stderr == nil {
		options.Stderr = io.Discard
	}

	return options
}

func defaultBinDir() string {
	if goBin := strings.TrimSpace(os.Getenv("GOBIN")); goBin != "" {
		return goBin
	}

	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".local", "bin")
	}

	if goPath := strings.TrimSpace(os.Getenv("GOPATH")); goPath != "" {
		return filepath.Join(goPath, "bin")
	}

	return filepath.Join(".", "bin")
}

func runCommand(options Options, name string, args ...string) error {
	ctx := options.Context
	cancel := func() {}
	if options.CommandTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, options.CommandTimeout)
	}
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = options.SourceDir
	cmd.Stdout = options.Stdout
	cmd.Stderr = options.Stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("%s timed out after %s", commandLine(name, args), options.CommandTimeout)
		}
		return fmt.Errorf("%s: %w", commandLine(name, args), err)
	}

	return nil
}

func commandLine(name string, args []string) string {
	return shell.Line(shell.Command{Name: name, Args: args})
}

func pathContains(dir string, pathValue string) bool {
	if dir == "" {
		return false
	}
	cleanDir, err := filepath.Abs(dir)
	if err != nil {
		cleanDir = filepath.Clean(dir)
	}
	for _, part := range filepath.SplitList(pathValue) {
		cleanPart, err := filepath.Abs(part)
		if err != nil {
			cleanPart = filepath.Clean(part)
		}
		if cleanPart == cleanDir {
			return true
		}
	}

	return false
}
