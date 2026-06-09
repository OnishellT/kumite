package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunSetupDryRun(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"setup", "--languages", "go", "--dry-run"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() exit code = %d, stderr = %q", code, stderr.String())
	}

	output := stdout.String()
	for _, want := range []string{
		"==> pi extensions",
		"pi install npm:pi-kumite",
		"pi install npm:pi-subagents",
		"pi install npm:pi-mcp-adapter",
		"npm install -g context-mode",
		"pi install npm:context-mode",
		"==> go static-analysis tooling",
		"go install golang.org/x/tools/cmd/deadcode@latest",
		"go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest",
		"$ write .pi/skills/static-analysis-reviewer/SKILL.md",
		"$ write .pi/skills/kumite-grill-with-docs/SKILL.md",
		"$ upsert agents.md",
		"$ upsert AGENTS.md",
		"$ write .pi/agents/orchestrator.md",
		"$ write .pi/chains/kumite-loop.chain.md",
		"$ merge .pi/mcp.json",
		"$ upsert .pi/settings.json",
		"$ create-if-missing .kumite/memory/architecture.md",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout missing %q\nstdout:\n%s", want, output)
		}
	}
}

func TestRunSetupGlobalDryRunSkipsProjectArtifacts(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"setup", "--global", "--languages", "go", "--dry-run"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() exit code = %d, stderr = %q", code, stderr.String())
	}

	output := stdout.String()
	for _, want := range []string{
		"==> pi extensions",
		"pi install npm:pi-kumite",
		"pi install npm:pi-subagents",
		"==> go static-analysis tooling",
		"go install golang.org/x/tools/cmd/deadcode@latest",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout missing %q\nstdout:\n%s", want, output)
		}
	}
	for _, unwanted := range []string{
		"$ write .pi/agents/orchestrator.md",
		"$ create-if-missing .kumite/memory/architecture.md",
		"$ upsert .pi/settings.json",
		"$ upsert agents.md",
		"$ upsert AGENTS.md",
	} {
		if strings.Contains(output, unwanted) {
			t.Fatalf("stdout unexpectedly contains %q\nstdout:\n%s", unwanted, output)
		}
	}
}

func TestRunInitDryRunWritesOnlyProjectArtifacts(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"init", "--dry-run"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() exit code = %d, stderr = %q", code, stderr.String())
	}

	output := stdout.String()
	for _, want := range []string{
		"$ write .pi/skills/static-analysis-reviewer/SKILL.md",
		"$ upsert agents.md",
		"$ upsert AGENTS.md",
		"$ write .pi/agents/orchestrator.md",
		"$ write .pi/chains/kumite-loop.chain.md",
		"$ merge .pi/mcp.json",
		"$ upsert .pi/settings.json",
		"$ create-if-missing .kumite/memory/architecture.md",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout missing %q\nstdout:\n%s", want, output)
		}
	}
	for _, unwanted := range []string{
		"pi install npm:pi-subagents",
		"go install golang.org/x/tools/cmd/deadcode@latest",
		"npm install -g fallow",
	} {
		if strings.Contains(output, unwanted) {
			t.Fatalf("stdout unexpectedly contains %q\nstdout:\n%s", unwanted, output)
		}
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"bogus"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("run() exit code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), `unknown command "bogus"`) {
		t.Fatalf("stderr missing unknown command message: %q", stderr.String())
	}
}
