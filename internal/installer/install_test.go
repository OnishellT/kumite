package installer

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDryRunPlansBinaryInstall(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	binDir := filepath.Join(t.TempDir(), "bin")

	result, err := Run(Options{
		SourceDir:      "/src/kumite",
		BinDir:         binDir,
		RunGlobalSetup: true,
		DryRun:         true,
		Stdout:         &stdout,
		Stderr:         &stderr,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.BinPath != filepath.Join(binDir, "kumite") {
		t.Fatalf("BinPath = %q", result.BinPath)
	}
	if len(result.Steps) != 3 {
		t.Fatalf("steps = %d", len(result.Steps))
	}
	for _, want := range []string{
		"mkdir -p " + binDir,
		"go build -o " + filepath.Join(binDir, "kumite") + " ./cmd/kumite",
		filepath.Join(binDir, "kumite") + " setup --global --keep-going",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout missing %q\nstdout:\n%s", want, stdout.String())
		}
	}
}

func TestRunDryRunMarksStepsDoneWithoutInstalling(t *testing.T) {
	t.Parallel()

	result, err := Run(Options{
		BinDir: filepath.Join(t.TempDir(), "bin"),
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	for _, step := range result.Steps {
		if !step.Done {
			t.Fatalf("dry-run step %q was not marked done", step.Name)
		}
		if step.Err != nil {
			t.Fatalf("dry-run step %q error = %v", step.Name, step.Err)
		}
	}
}

func TestRunDryRunCanSkipGlobalSetup(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	binDir := filepath.Join(t.TempDir(), "bin")

	result, err := Run(Options{
		BinDir: binDir,
		DryRun: true,
		Stdout: &stdout,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(result.Steps) != 2 {
		t.Fatalf("steps = %d", len(result.Steps))
	}
	if strings.Contains(stdout.String(), "setup --global") {
		t.Fatalf("stdout unexpectedly includes global setup:\n%s", stdout.String())
	}
}
