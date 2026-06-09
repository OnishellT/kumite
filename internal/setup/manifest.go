package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type reviewCommandSet struct {
	Primary    []string
	Structured []string
}

type skillSectionDefinition struct {
	Language        string
	Title           string
	Intro           string
	Preamble        []string
	StructuredIntro string
	Guidance        string
}

func defaultLanguages() []string {
	return []string{"go", "python", "rust", "javascript"}
}

func piExtensionInstallCommands() []command {
	return []command{
		{Name: "pi", Args: []string{"install", "npm:pi-subagents"}, Env: npmQuietEnv(), Note: "Installs the subagent runtime used by the kumite delegation chain."},
		{Name: "pi", Args: []string{"install", "npm:pi-intercom"}, Env: npmQuietEnv(), Note: "Enables parent/child session coordination for subagent results and attention signals."},
		{Name: "pi", Args: []string{"install", "npm:pi-mcp-adapter"}, Env: npmQuietEnv(), Note: "Adds the MCP gateway used by Memo and other project-local MCP servers."},
		{Name: "pi", Args: []string{"install", "npm:pi-web-access"}, Env: npmQuietEnv(), Note: "Adds web search, URL fetch, GitHub, PDF, and video research tools."},
		{Name: "pi", Args: []string{"install", "npm:@juicesharp/rpiv-todo"}, Env: npmQuietEnv(), Note: "Adds a persistent model-visible todo overlay for long setup and implementation loops."},
		{Name: "pi", Args: []string{"install", "npm:@juicesharp/rpiv-ask-user-question"}, Env: npmQuietEnv(), Note: "Adds structured clarification dialogs for the planner."},
		{Name: "npm", Args: []string{"install", "-g", "context-mode"}, Env: npmQuietEnv(), Note: "Installs the context-mode binary required by the Pi MCP server entry."},
		{Name: "pi", Args: []string{"install", "npm:context-mode"}, Env: npmQuietEnv(), Note: "Adds context routing and session-continuity support where the host Pi/OpenClaw setup supports it."},
		{Name: "context-mode", Args: []string{"--help"}, Optional: true, Note: "Verifies the context-mode binary is discoverable; run `ctx stats` after restarting Pi."},
		{
			Name:     "memo",
			Args:     []string{"version"},
			Optional: true,
			Note:     "Installs Memo from MEMO_INSTALLER, MEMO_SOURCE_DIR, or ~/projects/memo when the Memo CLI is not already available.",
			Resolve:  resolveMemoInstallCommand,
		},
		{Name: "memo", Args: []string{"init"}, Optional: true, Note: "Initializes Memo local SQLite memory when the Memo CLI is available."},
		{Name: "memo", Args: []string{"setup", "path"}, Optional: true, Note: "Ensures the Memo CLI is available on PATH for the Memo MCP server entry."},
		{Name: "memo", Args: []string{"setup", "doctor"}, Optional: true, Note: "Verifies Memo local memory and MCP setup surface."},
	}
}

func npmQuietEnv() []string {
	return []string{
		"npm_config_audit=false",
		"npm_config_fund=false",
		"NPM_CONFIG_AUDIT=false",
		"NPM_CONFIG_FUND=false",
	}
}

func installCommandsForLanguage(language string) ([]command, error) {
	switch normalizeLanguage(language) {
	case "go":
		commands := []command{
			{Name: "go", Args: []string{"install", "golang.org/x/tools/cmd/deadcode@latest"}},
			{Name: "go", Args: []string{"install", "github.com/roblaszczak/go-cleanarch@latest"}},
			{Name: "go", Args: []string{"install", "github.com/loov/goda@latest"}},
			{Name: "go", Args: []string{"install", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"}},
			{Name: "dot", Args: []string{"-V"}, Note: "Graphviz is required for rendering Goda graphs."},
		}
		return append(commands, goPathLinkCommands()...), nil
	case "python":
		commands := []command{
			{Name: "uv", Args: []string{"tool", "install", "deadcode"}},
			{Name: "uv", Args: []string{"tool", "install", "deptry"}},
			{Name: "uv", Args: []string{"tool", "install", "tach"}},
			{Name: "uv", Args: []string{"tool", "install", "vulture"}},
			{Name: "uv", Args: []string{"tool", "install", "radon"}},
			{Name: "uv", Args: []string{"tool", "install", "ruff"}},
			{Name: "uv", Args: []string{"tool", "install", "pylint"}},
		}
		return append(commands, pythonToolLinkCommands()...), nil
	case "javascript":
		commands := []command{
			{
				Name: "npm",
				Args: []string{"install", "-g", "fallow"},
				Note: "Fallow provides Fallow-like static analysis for JavaScript and TypeScript projects.",
			},
		}
		commands = append(commands, fallowVerifyCommands()...)
		return commands, nil
	case "rust":
		return []command{
			{Name: "rustup", Args: []string{"component", "add", "clippy"}},
			{Name: "rustup", Args: []string{"component", "add", "rust-analyzer"}},
			{Name: "cargo", Args: []string{"install", "cargo-shear"}},
			{Name: "cargo", Args: []string{"install", "cargo-workspace-unused-pub"}},
			{Name: "cargo", Args: []string{"install", "cargo-duplicated"}},
			{
				Name: "rustup",
				Args: []string{
					"component", "add", "--toolchain", "nightly-2026-01-22",
					"rust-src", "rustc-dev", "llvm-tools-preview",
				},
				Optional: true,
				Note:     "cargo-pup is experimental and may require a specific nightly-compatible release.",
			},
			{Name: "cargo", Args: []string{"+nightly-2026-01-22", "install", "cargo_pup"}, Optional: true},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported language %q", language)
	}
}

func reviewCommandsForLanguage(language string) (reviewCommandSet, error) {
	switch normalizeLanguage(language) {
	case "go":
		return reviewCommandSet{Primary: []string{
			"deadcode -test ./...",
			"golangci-lint run ./...",
			"go mod tidy",
			"git diff --exit-code -- go.mod go.sum",
			"go-cleanarch",
			`goda graph "./..." | dot -Tsvg -o graph.svg`,
		}}, nil
	case "python":
		return reviewCommandSet{Primary: []string{
			"python-deadcode . --fix --dry",
			"deptry .",
			"tach check",
			"vulture src tests",
			"radon cc -s -a src",
			"pylint --disable=all --enable=duplicate-code src",
			"ruff check .",
		}}, nil
	case "javascript":
		return reviewCommandSet{
			Primary: []string{
				"fallow audit --changed-since HEAD",
				"fallow dead-code --changed-since HEAD",
				"fallow dupes --changed-since HEAD",
				"fallow health --changed-since HEAD",
				"fallow fix --dry-run --changed-since HEAD",
			},
			Structured: []string{
				"fallow audit --changed-since HEAD --format json --quiet --explain",
				"fallow dead-code --changed-since HEAD --format json --quiet",
				"fallow fix --dry-run --changed-since HEAD --format json --quiet",
			},
		}, nil
	case "rust":
		return reviewCommandSet{Primary: []string{
			"cargo shear --deny-warnings",
			"cargo clippy --workspace --all-targets --all-features -- -D warnings",
			"cargo workspace-unused-pub",
			"cargo-duplicated .",
			"cargo pup",
		}}, nil
	default:
		return reviewCommandSet{}, fmt.Errorf("unsupported language %q", language)
	}
}

func skillSectionDefinitions() []skillSectionDefinition {
	return []skillSectionDefinition{
		{
			Language: "go",
			Title:    "Go",
			Intro:    "Run from the Go module root:",
			Guidance: "Use `deadcode` for unreachable functions, `golangci-lint` for unused symbols, duplication, complexity, and import/module policy, `go mod tidy` for dependency drift, `go-cleanarch` for package boundaries, and `goda` for package graph analysis.\n\nFor Go review in a Git worktree, run the six-command set above verbatim. Do not substitute `go vet`, direct `staticcheck`, or `go test` for the static-analysis command set. In scratch projects without `.git`, run the non-Git commands plus `go mod tidy`, record `git diff --exit-code -- go.mod go.sum` as `SKIPPED: no git worktree`, and do not fail the review solely because Git metadata is absent. In your review response, list the exact Go static-analysis commands run or skipped.",
		},
		{
			Language: "python",
			Title:    "Python",
			Intro:    "Run from the Python project root. For dependency checks, prefer the project's virtual environment because `deptry` needs project dependency metadata:",
			Guidance: "Use `python-deadcode` and `vulture` for unused/dead code, `deptry` for dependency hygiene, `tach` for module graph boundaries and cycles, `radon` for complexity, `pylint duplicate-code` for duplication, and `ruff` for local cleanup evidence. `python-deadcode` is intentionally named to avoid colliding with Go's `deadcode` command.\n\nFor Python review, run the seven-command set above verbatim. In your review response, list the exact Python static-analysis commands run.",
		},
		{
			Language: "javascript",
			Title:    "JavaScript and TypeScript",
			Intro:    "Run from the JS/TS project root. Pick a base ref for changed-file review; use the PR base branch when available, or `HEAD` for uncommitted local changes:",
			Preamble: []string{
				`export FALLOW_VERIFY_CACHE_DIR="${FALLOW_VERIFY_CACHE_DIR:-/tmp/fallow-verify-${USER:-agent}}"`,
				`mkdir -p "$FALLOW_VERIFY_CACHE_DIR"`,
			},
			StructuredIntro: "For agent-readable evidence, prefer structured output:",
			Guidance:        "For JS/TS review, run the five-command minimum set above verbatim: `audit`, `dead-code`, `dupes`, `health`, and `fix --dry-run`. Replace `HEAD` only when the user or repository context gives a better base ref. Do not substitute `--diff-stdin`, bare `fallow`, `fallow security`, or any other Fallow command for the minimum set. Use `fallow audit` as the changed-file quality gate, `fallow dead-code` for unused files, exports, dependencies, circular dependencies, and boundary violations, `fallow dupes` for duplicate code, `fallow health` for complexity hotspots and maintainability targets, and `fallow fix --dry-run` only as a preview before any cleanup. Run `fallow list --plugins` when framework conventions affect the graph. In your review response, list the exact Fallow commands run.",
		},
		{
			Language: "rust",
			Title:    "Rust",
			Intro:    "Run from the Cargo workspace root. `cargo pup` requires a `pup.ron` architecture lint file:",
			Guidance: "Use `cargo-shear` for unused dependencies and unlinked source files, Rust compiler/Clippy lints for private dead code and complexity, `cargo-workspace-unused-pub` for unused public workspace API, `cargo-duplicated` for duplicate Rust blocks, and `cargo-pup` for explicit architecture assertions.\n\nFor Rust review, run the five-command set above verbatim. In your review response, list the exact Rust static-analysis commands run.",
		},
	}
}

func normalizeLanguage(language string) string {
	switch strings.ToLower(language) {
	case "go", "golang":
		return "go"
	case "python", "py":
		return "python"
	case "javascript", "typescript", "js", "ts", "node":
		return "javascript"
	case "rust", "rs":
		return "rust"
	default:
		return strings.ToLower(language)
	}
}

func goPathLinkCommands() []command {
	targetDir := "${HOME}/.local/bin"
	commands := []command{{
		Name:    "mkdir",
		Args:    []string{"-p", targetDir},
		Resolve: resolveHomeBinCommand("mkdir", "-p"),
	}}
	for _, tool := range []string{"deadcode", "go-cleanarch", "goda", "golangci-lint"} {
		commands = append(commands, command{
			Name:    "ln",
			Args:    []string{"-sf", filepath.Join("${GOPATH}", "bin", tool), filepath.Join(targetDir, tool)},
			Note:    "Expose Go-installed analyzers on the user PATH.",
			Resolve: resolveGoToolLinkCommand(tool),
		})
	}

	return commands
}

func resolveGoToolLinkCommand(tool string) func() (command, error) {
	return func() (command, error) {
		goPath, err := exec.Command("go", "env", "GOPATH").Output()
		if err != nil {
			return command{}, fmt.Errorf("discover GOPATH: %w", err)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return command{}, fmt.Errorf("discover user home: %w", err)
		}

		return command{
			Name: "ln",
			Args: []string{
				"-sf",
				filepath.Join(strings.TrimSpace(string(goPath)), "bin", tool),
				filepath.Join(home, ".local", "bin", tool),
			},
			Note: "Expose Go-installed analyzers on the user PATH.",
		}, nil
	}
}

func pythonToolLinkCommands() []command {
	targetDir := "${HOME}/.local/bin"
	return []command{
		{
			Name:    "mkdir",
			Args:    []string{"-p", targetDir},
			Resolve: resolveHomeBinCommand("mkdir", "-p"),
		},
		{
			Name: "ln",
			Args: []string{
				"-sf",
				filepath.Join("${UV_TOOL_DIR}", "deadcode", "bin", "deadcode"),
				filepath.Join(targetDir, "python-deadcode"),
			},
			Note:    "Expose Python deadcode without colliding with Go deadcode.",
			Resolve: resolvePythonDeadcodeLinkCommand,
		},
	}
}

func resolvePythonDeadcodeLinkCommand() (command, error) {
	toolDir, err := exec.Command("uv", "tool", "dir").Output()
	if err != nil {
		return command{}, fmt.Errorf("discover uv tool directory: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return command{}, fmt.Errorf("discover user home: %w", err)
	}

	return command{
		Name: "ln",
		Args: []string{
			"-sf",
			filepath.Join(strings.TrimSpace(string(toolDir)), "deadcode", "bin", "deadcode"),
			filepath.Join(home, ".local", "bin", "python-deadcode"),
		},
		Note: "Expose Python deadcode without colliding with Go deadcode.",
	}, nil
}

func resolveHomeBinCommand(name string, args ...string) func() (command, error) {
	return func() (command, error) {
		home, err := os.UserHomeDir()
		if err != nil {
			return command{}, fmt.Errorf("discover user home: %w", err)
		}

		return command{
			Name: name,
			Args: append(args, filepath.Join(home, ".local", "bin")),
		}, nil
	}
}

func fallowVerifyCommands() []command {
	cacheDir := filepath.Join(os.TempDir(), "fallow-verify-kumite")
	return []command{
		{Name: "mkdir", Args: []string{"-p", cacheDir}},
		{
			Name: "fallow",
			Args: []string{"--version"},
			Env:  []string{"FALLOW_VERIFY_CACHE_DIR=" + cacheDir},
		},
	}
}

func resolveMemoInstallCommand() (command, error) {
	if _, err := exec.LookPath("memo"); err == nil {
		return command{
			Name: "memo",
			Args: []string{"version"},
			Note: "Memo CLI is already available.",
		}, nil
	}

	if installer := os.Getenv("MEMO_INSTALLER"); installer != "" {
		return command{
			Name: installer,
			Args: []string{"--agent", "pi", "--quiet"},
			Note: "Install Memo from MEMO_INSTALLER.",
		}, nil
	}

	sourceDir := os.Getenv("MEMO_SOURCE_DIR")
	if sourceDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return command{}, fmt.Errorf("discover user home for Memo install: %w", err)
		}
		sourceDir = filepath.Join(home, "projects", "memo")
	}

	installer := filepath.Join(sourceDir, "scripts", "install.sh")
	if info, err := os.Stat(installer); err == nil && !info.IsDir() {
		return command{
			Name: installer,
			Args: []string{"--agent", "pi", "--quiet"},
			Note: "Install Memo from local source checkout.",
		}, nil
	}

	return command{}, fmt.Errorf("memo CLI not found and Memo installer not found; set MEMO_INSTALLER or MEMO_SOURCE_DIR")
}
