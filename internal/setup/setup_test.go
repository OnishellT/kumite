package setup

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunWritesStaticAnalysisSkill(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	skillsDir := filepath.Join(t.TempDir(), "skills")

	err := Run(Options{
		Languages: []string{"python"},
		DryRun:    true,
		SkillsDir: skillsDir,
		Stdout:    &stdout,
		Stderr:    &stderr,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(skillsDir, "static-analysis", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("dry-run wrote skill file, stat err = %v", err)
	}
	if !strings.Contains(stdout.String(), "uv tool install deadcode") {
		t.Fatalf("stdout missing python installer: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "python-deadcode") {
		t.Fatalf("stdout missing python deadcode alias: %s", stdout.String())
	}
}

func TestRunJavaScriptSetupDryRun(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run(Options{
		Languages: []string{"typescript"},
		DryRun:    true,
		Stdout:    &stdout,
		Stderr:    &stderr,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	for _, want := range []string{
		"==> typescript static-analysis tooling",
		"npm install -g fallow",
		"fallow --version",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout missing %q\nstdout:\n%s", want, stdout.String())
		}
	}
}

func TestRunRejectsUnsupportedLanguage(t *testing.T) {
	t.Parallel()

	err := Run(Options{
		Languages: []string{"ruby"},
		DryRun:    true,
		Stdout:    &bytes.Buffer{},
		Stderr:    &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("Run() error = nil")
	}
	if !strings.Contains(err.Error(), `unsupported language "ruby"`) {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestBuildPlanIncludesLanguageCommandsAndSkill(t *testing.T) {
	t.Parallel()

	plan, err := BuildPlan(Options{
		Languages: []string{"go"},
		DryRun:    true,
		Stdout:    &bytes.Buffer{},
		Stderr:    &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	if !plan.WriteSkill {
		t.Fatal("BuildPlan() did not plan skill write")
	}
	if !plan.WriteAgents {
		t.Fatal("BuildPlan() did not plan subagent write")
	}
	if !plan.WriteMemory {
		t.Fatal("BuildPlan() did not plan memory docs write")
	}
	if len(plan.ExtensionCommands) == 0 {
		t.Fatal("BuildPlan() did not include pi extension commands")
	}
	if len(plan.Languages) != 1 {
		t.Fatalf("language plans = %d", len(plan.Languages))
	}
	if len(plan.Languages[0].Commands) == 0 {
		t.Fatal("go language plan has no commands")
	}
}

func TestBuildPlanDryRunDoesNotRequireDiscoveryTools(t *testing.T) {
	t.Setenv("PATH", "")

	plan, err := BuildPlan(Options{
		Languages: []string{"go", "python"},
		DryRun:    true,
		Stdout:    &bytes.Buffer{},
		Stderr:    &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	var output bytes.Buffer
	summary := executeCommandPlan(Options{
		DryRun: true,
		Stdout: &output,
		Stderr: &bytes.Buffer{},
	}, plan)
	if summary.hasRequiredFailures() {
		t.Fatalf("dry-run summary required failures = %v", summary.RequiredFailures)
	}
	for _, want := range []string{"${GOPATH}", "${UV_TOOL_DIR}", "${HOME}/.local/bin"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("dry-run output missing %q\nstdout:\n%s", want, output.String())
		}
	}
}

func TestBuildPlanCanSkipPiExtensions(t *testing.T) {
	t.Parallel()

	plan, err := BuildPlan(Options{
		Languages:      []string{"go"},
		DryRun:         true,
		SkipExtensions: true,
		Stdout:         &bytes.Buffer{},
		Stderr:         &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	if len(plan.ExtensionCommands) != 0 {
		t.Fatalf("extension commands = %d", len(plan.ExtensionCommands))
	}
}

func TestRunWritesSkillFile(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	piSkillsDir := filepath.Join(root, "pi-skills")
	oldStaticPath, oldPlannerPath := writeObsoleteSkillCopies(t, skillsDir, piSkillsDir)

	err := writeStaticAnalysisSkill(Options{
		DryRun:      false,
		SkillsDir:   skillsDir,
		PiSkillsDir: piSkillsDir,
		Stdout:      &stdout,
		Stderr:      &stderr,
	})
	if err != nil {
		t.Fatalf("writeStaticAnalysisSkill() error = %v", err)
	}

	piContent, err := os.ReadFile(filepath.Join(piSkillsDir, "static-analysis-reviewer", "SKILL.md"))
	if err != nil {
		t.Fatalf("read pi skill file: %v", err)
	}
	want, err := renderStaticAnalysisSkill()
	if err != nil {
		t.Fatalf("render skill: %v", err)
	}
	if _, err := os.Stat(oldStaticPath); !os.IsNotExist(err) {
		t.Fatalf("static analysis skill should not be duplicated into .agents skills, stat err = %v", err)
	}
	if _, err := os.Stat(oldPlannerPath); !os.IsNotExist(err) {
		t.Fatalf("old planner skill should be removed, stat err = %v", err)
	}
	if string(piContent) != want {
		t.Fatal("written pi skill did not match embedded template output")
	}
	for _, want := range []string{"# Static Analysis Reviewer", "deadcode -test ./...", "python-deadcode . --fix --dry", "fallow audit", "cargo shear --deny-warnings", "SKIPPED: no git worktree"} {
		if !strings.Contains(string(piContent), want) {
			t.Fatalf("skill content missing %q", want)
		}
	}
}

func writeObsoleteSkillCopies(t *testing.T, skillsDir string, piSkillsDir string) (string, string) {
	t.Helper()

	oldStaticPath := filepath.Join(skillsDir, "static-analysis", "SKILL.md")
	writeTestFile(t, oldStaticPath, "# Static Analysis Reviewer\n\nname: static-analysis-reviewer\n")

	oldPlannerPath := filepath.Join(piSkillsDir, "grill-with-docs", "SKILL.md")
	writeTestFile(t, oldPlannerPath, "# Grill With Docs\n\nkumite planning\n")

	return oldStaticPath, oldPlannerPath
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunWritesPlannerSkillFile(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := t.TempDir()
	piSkillsDir := filepath.Join(root, "pi-skills")

	err := writePlannerSkill(Options{
		DryRun:      false,
		PiSkillsDir: piSkillsDir,
		Stdout:      &stdout,
		Stderr:      &stderr,
	})
	if err != nil {
		t.Fatalf("writePlannerSkill() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(piSkillsDir, "kumite-grill-with-docs", "SKILL.md"))
	if err != nil {
		t.Fatalf("read planner skill file: %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"name: kumite-grill-with-docs",
		"# Grill With Docs",
		"return planner-generated grill questions after reading memory and targeted project context",
		".kumite/memory/architecture.md",
		"Memory documents used",
		"Before unresolved questions are answered and the draft plan is accepted/refined",
		"STATUS: DRAFT_PLAN_FOR_GRILL",
		"Draft plan",
		"Bounded Chain Use",
		"Use the scout handoff as the primary source of code context",
		"Do not perform broad code reconnaissance, tests, static analysis",
		"run the grill discovery before writing final spec or Gherkin artifacts",
		"Do not write final plan or Gherkin artifacts in that state",
		"`Grill questions and gaps`",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("planner skill content missing %q\ncontent:\n%s", want, text)
		}
	}
}

func TestRunWritesSubagentsAndMemoryFiles(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := t.TempDir()
	agentsDir := filepath.Join(root, ".pi", "agents")
	chainsDir := filepath.Join(root, ".pi", "chains")
	memoryDir := filepath.Join(root, ".kumite", "memory")
	instructions := filepath.Join(root, "agents.md")
	piInstructions := filepath.Join(root, "AGENTS.md")
	mcpConfig := filepath.Join(root, ".pi", "mcp.json")
	piSettings := filepath.Join(root, ".pi", "settings.json")

	if err := writeSubagentFiles(Options{
		AgentsDir: agentsDir,
		Stdout:    &stdout,
		Stderr:    &stderr,
	}); err != nil {
		t.Fatalf("writeSubagentFiles() error = %v", err)
	}
	if err := writeChainFiles(Options{
		ChainsDir:      chainsDir,
		MCPConfigPath:  mcpConfig,
		PiSettingsPath: piSettings,
		Stdout:         &stdout,
		Stderr:         &stderr,
	}); err != nil {
		t.Fatalf("writeChainFiles() error = %v", err)
	}
	if err := writeProjectInstructions(Options{
		Instructions: instructions,
		Stdout:       &stdout,
		Stderr:       &stderr,
	}); err != nil {
		t.Fatalf("writeProjectInstructions() error = %v", err)
	}
	if err := writeMemoryFiles(Options{
		MemoryDir: memoryDir,
		Stdout:    &stdout,
		Stderr:    &stderr,
	}); err != nil {
		t.Fatalf("writeMemoryFiles() error = %v", err)
	}

	checks := map[string]string{
		filepath.Join(agentsDir, "orchestrator.md"):      "ctx_stats",
		filepath.Join(agentsDir, "scout.md"):             "ctx_stats",
		filepath.Join(agentsDir, "planner.md"):           "Parallelization Plan",
		filepath.Join(agentsDir, "planner-fallback.md"):  "noninteractive planner",
		filepath.Join(agentsDir, "implementer.md"):       "ctx_batch_execute",
		filepath.Join(agentsDir, "reviewer.md"):          "Parallel workstream review",
		filepath.Join(agentsDir, "curator.md"):           "canonical project truth",
		filepath.Join(chainsDir, "kumite-loop.chain.md"): "review-summary-round-1.md",
		mcpConfig:      `"context-mode"`,
		piSettings:     `"npm:pi-kumite"`,
		instructions:   "Kumite Agent Index",
		piInstructions: "Before any code-modifying or planning task",
		filepath.Join(memoryDir, "architecture.md"):                                "last-update: unset",
		filepath.Join(memoryDir, "code-standards.md"):                              "Review Gates",
		filepath.Join(memoryDir, "business-rules.md"):                              "Glossary",
		filepath.Join(memoryDir, "project-status.md"):                              "durable project handoff",
		filepath.Join(filepath.Dir(memoryDir), "plans", "README.md"):               "Kumite planner agents write dated spec-driven plans",
		filepath.Join(filepath.Dir(memoryDir), "handoffs", "current", "README.md"): "durable project-local mirrors",
	}
	for path, want := range checks {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if !strings.Contains(string(content), want) {
			t.Fatalf("%s missing %q", path, want)
		}
	}
	piInstructionsContent, err := os.ReadFile(piInstructions)
	if err != nil {
		t.Fatalf("read pi instructions bridge: %v", err)
	}
	if strings.Contains(string(piInstructionsContent), "## Project Wiki Map") {
		t.Fatal("AGENTS.md bridge should not duplicate the canonical agents.md index")
	}
	assertGeneratedWorkflowInstructions(t, instructions, filepath.Join(agentsDir, "orchestrator.md"))
	assertGeneratedContextModeInstructions(t, instructions, agentsDir, filepath.Join(chainsDir, "kumite-loop.chain.md"), mcpConfig)
	assertReviewerAvoidsNestedSubagent(t, filepath.Join(agentsDir, "reviewer.md"))
	assertKumiteLoopChain(t, filepath.Join(chainsDir, "kumite-loop.chain.md"))
}

func TestWriteProjectInstructionsIsIdempotent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	instructions := filepath.Join(root, "agents.md")
	options := Options{
		Instructions: instructions,
		Stdout:       &bytes.Buffer{},
		Stderr:       &bytes.Buffer{},
	}

	for range 2 {
		if err := writeProjectInstructions(options); err != nil {
			t.Fatalf("writeProjectInstructions() error = %v", err)
		}
	}

	bridge, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md bridge: %v", err)
	}
	if count := strings.Count(string(bridge), "# Kumite Pi Context Bridge"); count != 1 {
		t.Fatalf("bridge heading count = %d\ncontent:\n%s", count, bridge)
	}

	index, err := os.ReadFile(instructions)
	if err != nil {
		t.Fatalf("read agents.md index: %v", err)
	}
	if count := strings.Count(string(index), "# Kumite Agent Index"); count != 1 {
		t.Fatalf("index heading count = %d\ncontent:\n%s", count, index)
	}
}

func TestWritePiSettingsMergesExistingLocalPackage(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	settingsPath := filepath.Join(root, ".pi", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("create settings dir: %v", err)
	}
	existing := `{
  "quietStartup": true,
  "packages": [
    "../pi-kumite"
  ]
}
`
	if err := os.WriteFile(settingsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write settings: %v", err)
	}

	var stdout bytes.Buffer
	if err := writePiSettings(Options{
		PiSettingsPath: settingsPath,
		PiPackage:      "npm:pi-kumite",
		Stdout:         &stdout,
	}); err != nil {
		t.Fatalf("writePiSettings() error = %v", err)
	}

	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, `"quietStartup": true`) {
		t.Fatalf("existing setting was not preserved:\n%s", text)
	}
	if strings.Count(text, "pi-kumite") != 1 {
		t.Fatalf("kumite package should not be duplicated:\n%s", text)
	}
	if !strings.Contains(text, `"../pi-kumite"`) {
		t.Fatalf("existing local package should be preserved:\n%s", text)
	}
	if strings.Contains(text, `"npm:pi-kumite"`) {
		t.Fatalf("default package should not replace existing local package:\n%s", text)
	}
}

func TestWritePiSettingsExplicitLocalPackageReplacesNPM(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	settingsPath := filepath.Join(root, ".pi", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("create settings dir: %v", err)
	}
	existing := `{
  "quietStartup": true,
  "packages": [
    "npm:pi-kumite"
  ]
}
`
	if err := os.WriteFile(settingsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write settings: %v", err)
	}

	var stdout bytes.Buffer
	if err := writePiSettings(Options{
		PiSettingsPath: settingsPath,
		PiPackage:      "/home/dev/projects/pi-kumite",
		Stdout:         &stdout,
	}); err != nil {
		t.Fatalf("writePiSettings() error = %v", err)
	}

	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, `"/home/dev/projects/pi-kumite"`) {
		t.Fatalf("explicit local package should replace npm package:\n%s", text)
	}
	if strings.Contains(text, `"npm:pi-kumite"`) {
		t.Fatalf("stale npm package should be removed:\n%s", text)
	}
}

func TestWriteMemoryFilesPreservesExistingProjectDocs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	memoryDir := filepath.Join(root, ".kumite", "memory")
	if err := os.MkdirAll(memoryDir, 0o755); err != nil {
		t.Fatalf("create memory dir: %v", err)
	}

	architecturePath := filepath.Join(memoryDir, "architecture.md")
	architecture := "# Existing architecture\n\nProject-specific rules.\n"
	if err := os.WriteFile(architecturePath, []byte(architecture), 0o644); err != nil {
		t.Fatalf("write architecture: %v", err)
	}

	var stdout bytes.Buffer
	if err := writeMemoryFiles(Options{
		MemoryDir: memoryDir,
		Stdout:    &stdout,
	}); err != nil {
		t.Fatalf("writeMemoryFiles() error = %v", err)
	}

	content, err := os.ReadFile(architecturePath)
	if err != nil {
		t.Fatalf("read architecture: %v", err)
	}
	if string(content) != architecture {
		t.Fatalf("existing architecture doc was overwritten:\n%s", content)
	}

	standards, err := os.ReadFile(filepath.Join(memoryDir, "code-standards.md"))
	if err != nil {
		t.Fatalf("read code standards: %v", err)
	}
	if !strings.Contains(string(standards), "Review Gates") {
		t.Fatalf("missing seed code standards content:\n%s", standards)
	}
	if !strings.Contains(stdout.String(), "$ create-if-missing ") {
		t.Fatalf("memory writer did not report create-if-missing:\n%s", stdout.String())
	}
}

func TestWriteChainFilesMergesExistingMCPConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mcpPath := filepath.Join(root, ".pi", "mcp.json")
	if err := os.MkdirAll(filepath.Dir(mcpPath), 0o755); err != nil {
		t.Fatalf("create mcp dir: %v", err)
	}
	existing := `{
  "settings": {
    "idleTimeout": 5000
  },
  "mcpServers": {
    "custom": {
      "command": "custom-server"
    }
  }
}
`
	if err := os.WriteFile(mcpPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write mcp: %v", err)
	}

	var stdout bytes.Buffer
	if err := writeChainFiles(Options{
		ChainsDir:      filepath.Join(root, ".pi", "chains"),
		MCPConfigPath:  mcpPath,
		PiSettingsPath: filepath.Join(root, ".pi", "settings.json"),
		Stdout:         &stdout,
	}); err != nil {
		t.Fatalf("writeChainFiles() error = %v", err)
	}

	content, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("read mcp: %v", err)
	}

	assertMergedMCPConfig(t, content)
	if !strings.Contains(stdout.String(), "$ merge ") {
		t.Fatalf("chain writer did not report merge:\n%s", stdout.String())
	}
}

func assertMergedMCPConfig(t *testing.T, content []byte) {
	t.Helper()

	var config map[string]any
	if err := json.Unmarshal(content, &config); err != nil {
		t.Fatalf("parse mcp: %v\n%s", err, content)
	}

	settings := config["settings"].(map[string]any)
	assertJSONValue(t, settings, "idleTimeout", float64(5000), content)
	assertJSONValue(t, settings, "toolPrefix", "none", content)

	servers := config["mcpServers"].(map[string]any)
	assertJSONValue(t, servers["custom"].(map[string]any), "command", "custom-server", content)
	assertKumiteMCPServer(t, servers, "context-mode", content)
	assertKumiteMCPServer(t, servers, "memo", content)
}

func assertKumiteMCPServer(t *testing.T, servers map[string]any, name string, content []byte) {
	t.Helper()

	server := servers[name].(map[string]any)
	assertJSONValue(t, server, "directTools", true, content)
	assertJSONValue(t, server, "lifecycle", "keep-alive", content)
}

func assertJSONValue(t *testing.T, values map[string]any, key string, want any, content []byte) {
	t.Helper()

	if values[key] != want {
		t.Fatalf("json value %q = %v, want %v:\n%s", key, values[key], want, content)
	}
}

func assertGeneratedContextModeInstructions(t *testing.T, instructionsPath string, agentsDir string, chainPath string, mcpPath string) {
	t.Helper()

	mcpContent, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("read mcp config: %v", err)
	}
	for _, want := range []string{`"mcpServers"`, `"context-mode"`, `"command": "context-mode"`, `"memo"`, `"command": "memo"`, `"directTools": true`, `"lifecycle": "keep-alive"`} {
		if !strings.Contains(string(mcpContent), want) {
			t.Fatalf("mcp config missing %q\ncontent:\n%s", want, mcpContent)
		}
	}

	checks := map[string][]string{
		instructionsPath: {
			"ask the user whether they want the full Kumite workflow or a normal Pi workflow",
			"dynamic orchestrator flow",
			"Verify context-mode after setup/restart with `ctx stats`",
			"`ctx_batch_execute`, `ctx_execute`, `ctx_execute_file`, `ctx_index`, `ctx_search`, and `ctx_fetch_and_index`",
			"Markdown files under `.kumite/memory/` are canonical project memory",
			"Project Wiki Map",
			"Architecture Documentation Index",
			"Business Documentation Index",
			"`.kumite/memory/code-standards.md` requires explicit user approval",
			"do not use `output`, `outputMode: \"file-only\"`, or file-only routing for implementer subagent calls",
		},
		filepath.Join(agentsDir, "orchestrator.md"): {
			"tools: subagent, todo, ask_user_question, read, ctx_stats, kumite_next_step",
			"Start-of-task choice",
			"Default interactive loop",
			"Require the planner to send back `STATUS: DRAFT_PLAN_FOR_GRILL` before any Gherkin or final plan is written",
			"Ask the draft-plan grill questions in the parent session with `ask_user_question`",
			"Child completion handling",
			"After every scout, planner, implementer, reviewer, or rework child completion, call `kumite_next_step`",
			"If it says `agent: reviewer` after an implementer handoff, launch reviewer immediately",
			"Do not leave the parent parked behind a child that is waiting on intercom",
			"Ignore stale `subagent needs attention` notifications",
			"STATUS: IMPLEMENTED",
			"run scout within the first few parent actions",
			"`subagent list` is forbidden before scout",
			"planner's `Execution Plan` and `Parallelization Plan`",
			"If context-mode tools are available, call `ctx_stats`",
			"Treat `agents.md` as the primary project index",
			"Never use `output`, `outputMode: \"file-only\"`, or file-only routing for interactive implementer calls",
		},
		filepath.Join(agentsDir, "scout.md"): {
			"tools: write, find, ls, ctx_stats",
			"Use `ctx_stats` once",
			"Do not use context-mode execution/search/indexing from scout",
			"Memo is supporting searchable memory, not the source of truth",
			"first verify the current Memo project path matches the repository root",
			"bounded `kumite-loop` handoff",
			".kumite/handoffs/current/scout-context.md",
			"write the compact scout handoff to `.kumite/handoffs/current/scout-context.md`",
			"This is a scouting step, not planning or implementation",
			"Do not read source files or documentation files",
			"Do not grep source code contents",
			"Use at most four tool rounds total",
			"Original task, copied exactly from the task text when available",
			"Relevant `agents.md`, architecture index, and business-rule index paths",
		},
		filepath.Join(agentsDir, "planner.md"): {
			"In the standard `kumite-loop`, planner is file-backed and should not use context-mode",
			"assume the planner is running interactively",
			"grill-discovery",
			"STATUS: DRAFT_PLAN_FOR_GRILL",
			"Draft plan:",
			"Do not write final spec, Gherkin, or plan handoff artifacts while returning `STATUS: DRAFT_PLAN_FOR_GRILL` or `STATUS: NEED_USER_INPUT`",
			"Treat the scout handoff as the primary source of code context",
			"Do not run tests, static analysis, package graph tools, or implementation probes",
			"Do not read source files, test files, parser/evaluator/runtime files, package manifests, or repository docs in the saved planner step",
			"Run a `kumite-grill-with-docs` discovery pass before writing final plan or Gherkin artifacts",
			"`Grill questions and gaps`",
			"Execution Plan",
			"STATUS: NEED_USER_INPUT",
			"Keep the first plan, Gherkin artifact, and handoff compact enough to fit a saved-chain step",
			"treat that as the durable task artifact",
			"do not wait or search for one",
			"write `plan-handoff.md` in the first response",
			".kumite/handoffs/current/plan-handoff.md",
			"Project index and memory used",
		},
		filepath.Join(agentsDir, "planner-fallback.md"): {
			"name: planner-fallback",
			"prefer completion over exhaustive planning",
			"Do not ask the user questions",
			"do not use Memo",
			"do not use context-mode",
			"tools: read, write",
			"Do not read any files. Use the inline original task from the chain step.",
			"Write that handoff to `.kumite/handoffs/current/plan-handoff.md`",
			"Do not write additional files in this fallback step",
			"Include a compact `Spec Plan` section",
			"Include a compact `Gherkin Scenarios` section",
			"Return the exact same handoff content from step 3 as your final answer",
			"deferred to implementer/reviewer in noninteractive fallback",
			"Grill questions and gaps",
			"Deferred questions",
			"Execution Plan",
			"Never read or reuse an existing `.kumite/handoffs/current/plan-handoff.md`",
		},
		filepath.Join(agentsDir, "implementer.md"): {
			"ctx_execute",
			"ctx_batch_execute",
			"Acceptance Contract",
			"acceptance-report",
			"Do not finish, stop, or return `STATUS: IMPLEMENTED`",
			"Writing the `acceptance-report` only to a handoff file is not sufficient",
			"Pi validates the visible subagent result artifact",
			"that file must also include the required fenced `acceptance-report` block",
			"visible final response and any requested output/handoff file",
			"Required acceptance-report shape when requested",
			"criteriaSatisfied",
			"residualRisks",
			"when test/build/search output is expected to be large",
			"Do not use Memo for routine implementation",
			"Read `agents.md`",
			"After reading the plan handoff and task, use at most six targeted source-inspection tool calls before the first test edit",
			"Do not call `contact_supervisor`, wait on intercom, or leave the child run active",
			"return `STATUS: BLOCKED` with the exact file list instead of waiting for a supervisor decision",
			".kumite/handoffs/current/implementation-summary.md",
		},
		filepath.Join(agentsDir, "reviewer.md"): {
			"ctx_stats",
			"ctx_batch_execute",
			"Context-mode status when `ctx_stats` is available",
			"Bounded review protocol",
			"Let the plan determine review depth",
			"Use at most three extra manual behavior probes",
			"Immediately write the requested review summary file",
			"SKIPPED: no git worktree",
			"Retry carry-forward steps must be cheap",
			"check whether an index exists for the current project",
			".kumite/handoffs/current/review-summary-round-1.md",
			"Project index and memory compliance",
		},
		filepath.Join(agentsDir, "curator.md"): {
			"ctx_index",
			"Do not treat context-mode indexes as canonical memory",
			"Do not update `code-standards.md` without explicit user approval",
		},
		chainPath: {
			"This is a bounded saved-chain scout step",
			"The handoff must start with this exact original task:",
			"Produce compact saved-chain planning artifacts",
			"prefer completion over exhaustive planning",
			"Do not read any files in this fallback planner step",
			"Do not use context-mode, Memo, web tools, supervisor/intercom, user-question tools, or read tools",
			"write it to:",
			"Do not write `.kumite/plans/current-kumite-loop-plan.md`",
			"Return the same plan handoff content as the final answer",
			".kumite/handoffs/current/plan-handoff.md",
			"SKIPPED: no git worktree",
			"not used by scout; planner should stay file-backed and may use one Memo retrieval pass",
			"Do not use context-mode, Memo, web tools, supervisor/intercom, user-question tools, or read tools in this fallback planner step",
			"context-mode status",
			"include `ctx_stats` status in the review summary",
			"Execution Plan",
			"FOLLOW_UP_PLAN",
		},
	}

	for path, wants := range checks {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		text := string(content)
		for _, want := range wants {
			if !strings.Contains(text, want) {
				t.Fatalf("%s missing context-mode instruction %q\ncontent:\n%s", path, want, text)
			}
		}
		if strings.Contains(text, "model:") || strings.Contains(text, "thinking:") {
			t.Fatalf("%s hardcodes model selection; Kumite agents must inherit Pi defaults:\n%s", path, text)
		}
	}
}

func assertReviewerAvoidsNestedSubagent(t *testing.T, path string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read reviewer: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "tools: read, bash, grep, find, ls, subagent") {
		t.Fatalf("reviewer declares nested subagent tool, which conflicts in chain child runs:\n%s", text)
	}
	if !strings.Contains(text, "the next rework implementer step will consume this directly") {
		t.Fatalf("reviewer missing orchestrator rework handoff instruction:\n%s", text)
	}
	for _, want := range []string{"STATUS: REWORK_REQUIRED", "REWORK_TASK:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("reviewer missing strict review contract %q:\n%s", want, text)
		}
	}
}

func assertGeneratedWorkflowInstructions(t *testing.T, instructionsPath string, orchestratorPath string) {
	t.Helper()

	for _, path := range []string{instructionsPath, orchestratorPath} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		text := string(content)
		for _, want := range []string{
			`Do not call ` + "`subagent({agent: \"kumite-loop\"})`",
			`If the user explicitly says to use ` + "`kumite-loop`",
			`subagent({action: "get", chainName: "kumite-loop", agentScope: "both"})`,
			`replace every literal ` + "`{task}`",
			`Do not pass a chain containing ` + "`{task}`" + ` placeholders`,
			`subagent({chain: [...], task: "<task>", clarify: false, agentScope: "both"})`,
			`The scout handoff must include an ` + "`Original task`" + ` section`,
			`.kumite/handoffs/current/`,
			"`grill-discovery` mode",
			"`kumite-grill-with-docs`",
			`STATUS: REWORK_REQUIRED`,
			`REWORK_TASK`,
			`Parallelization Plan`,
			`worktree: true`,
			`After any child subagent returns a completed result`,
		} {
			if !strings.Contains(text, want) {
				t.Fatalf("%s missing workflow instruction %q\ncontent:\n%s", path, want, text)
			}
		}
	}
}

func assertKumiteLoopChain(t *testing.T, path string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read chain file: %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"---\nname: kumite-loop\n",
		"description: Non-interactive fallback chain",
		"## scout",
		"This is a bounded saved-chain scout step",
		"Immediately write a compact handoff to `.kumite/handoffs/current/scout-context.md`",
		"Do not read files, search, use context-mode, use Memo, or inspect source",
		"not used by scout; planner should stay file-backed and may use one Memo retrieval pass",
		"planner should stay file-backed and may use one Memo retrieval pass",
		"Do not use Memo in this fallback planner step",
		"Memo status: not used in noninteractive fallback planner",
		"Use Memo when available to check project knowledge",
		"## planner-fallback",
		"## implementer",
		"## reviewer",
		"output: scout-context.md",
		"reads: scout-context.md",
		"output: review-summary-round-1.md",
		".kumite/handoffs/current/implementation-summary.md",
		".kumite/handoffs/current/review-summary-round-1.md",
		"SKIPPED: no git worktree",
		"Implement the approved plan for the original task described in `scout-context.md` and `plan-handoff.md`",
		"Review the implementation for the original task described in `scout-context.md` and `plan-handoff.md`",
		"Include a `Parallelization Plan` even if the result is `serial-only`",
		"Include an `Execution Plan`",
		"FOLLOW_UP_PLAN",
		"Produce compact saved-chain planning artifacts",
		"prefer completion over exhaustive planning",
		"Do not read any files in this fallback planner step",
		"Do not use context-mode, Memo, web tools, supervisor/intercom, user-question tools, or read tools in this fallback planner step",
		"`.kumite/handoffs/current/plan-handoff.md`",
		"Return the same plan handoff content as the final answer",
		"Do not read scout handoff files, memory files, source files, test files",
		"write it to:",
		"The handoff itself must include compact `Spec Plan` and `Gherkin Scenarios` sections",
		"Do not write `.kumite/plans/current-kumite-loop-plan.md`",
		"Run a noninteractive equivalent of the `kumite-grill-with-docs` planning protocol",
		"`Grill questions and gaps`",
		"This saved chain is executing the first serial implementation pass only",
		"use at most six targeted source-inspection tool calls before the first test edit",
		"`STATUS: BLOCKED`",
		"merge order",
		"owned files and off-limits files",
		"verify ownership, off-limits files, merge order, and cross-workstream integration",
		"Use a bounded review protocol",
		"at most three manual behavior probes",
		"immediately write `review-summary-round-1.md`",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("chain file missing %q\ncontent:\n%s", want, text)
		}
	}
	if strings.Contains(text, "{chain_dir}") {
		t.Fatalf("chain file contains unexpanded chain_dir variable:\n%s", text)
	}
	if strings.Contains(text, "skills: static-analysis-reviewer") {
		t.Fatalf("chain file contains step-level skills override that markdown chains do not execute reliably:\n%s", text)
	}
}

func TestOptionalFailureIsSummarized(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	summary := runCommands(Options{
		Stdout: &stdout,
		Stderr: &stderr,
	}, []command{{
		Name:     "kumite-missing-optional-command",
		Optional: true,
	}})

	if len(summary.RequiredFailures) != 0 {
		t.Fatalf("required failures = %d", len(summary.RequiredFailures))
	}
	if len(summary.OptionalFailures) != 1 {
		t.Fatalf("optional failures = %d", len(summary.OptionalFailures))
	}

	printExecutionSummary(Options{Stderr: &stderr}, summary)
	if !strings.Contains(stderr.String(), "optional setup steps failed:") {
		t.Fatalf("stderr missing optional summary: %q", stderr.String())
	}
}

func TestCommandTimeoutIsReported(t *testing.T) {
	t.Parallel()

	summary := runCommands(Options{
		CommandTimeout: time.Nanosecond,
		Stdout:         &bytes.Buffer{},
		Stderr:         &bytes.Buffer{},
	}, []command{{
		Name: "sleep",
		Args: []string{"1"},
	}})

	if len(summary.RequiredFailures) != 1 {
		t.Fatalf("required failures = %d", len(summary.RequiredFailures))
	}
	if !strings.Contains(summary.RequiredFailures[0].Error(), "timed out") {
		t.Fatalf("timeout failure missing timed out message: %v", summary.RequiredFailures[0])
	}
}

func TestShellLineQuotesUnsafeParts(t *testing.T) {
	t.Parallel()

	got := shellLine(command{
		Name: "ln",
		Args: []string{"-sf", "/tmp/source path/tool", "/tmp/target's path/tool"},
		Env:  []string{"NAME=value with spaces"},
	})
	for _, want := range []string{
		"'NAME=value with spaces'",
		"'/tmp/source path/tool'",
		"'/tmp/target'\\''s path/tool'",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("shellLine() = %q, missing %q", got, want)
		}
	}
}
