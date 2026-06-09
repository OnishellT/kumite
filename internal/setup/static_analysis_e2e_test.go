//go:build e2e

package setup

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestStaticAnalysisToolStacks(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("the static-analysis command matrix uses POSIX shell commands")
	}

	t.Run("go", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		writeFile(t, root, "go.mod", "module example.com/go-fixture\n\ngo 1.26\n")
		writeFile(t, root, "go.sum", "\n")
		writeFile(t, root, "cmd/app/main.go", `package main

import "fmt"

func main() {
	fmt.Println(message())
}

func message() string {
	return "ok"
}
`)

		run(t, root, "git", "init")
		run(t, root, "git", "add", "go.mod", "go.sum")
		runReviewCommands(t, root, "go", nil)
	})

	t.Run("python", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		writeFile(t, root, "pyproject.toml", `[project]
name = "python-fixture"
version = "0.1.0"
dependencies = []

[tool.ruff]
line-length = 100
`)
		writeFile(t, root, "tach.toml", `source_roots = ["src"]

[[modules]]
path = "app"
depends_on = []
`)
		writeFile(t, root, "src/app.py", `def add(left: int, right: int) -> int:
    return left + right


RESULT = add(1, 2)
`)
		writeFile(t, root, "tests/use_app.py", `from app import RESULT

assert RESULT == 3
`)

		runReviewCommands(t, root, "python", nil)
	})

	t.Run("rust", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		writeFile(t, root, "Cargo.toml", `[package]
name = "rust-fixture"
version = "0.1.0"
edition = "2024"
`)
		writeFile(t, root, "pup.ron", `(lints: [],)
`)
		writeFile(t, root, "src/main.rs", `fn main() {
    println!("{}", message());
}

fn message() -> &'static str {
    "ok"
}
`)
		runReviewCommands(t, root, "rust", nil)
	})

	t.Run("javascript", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		writeFile(t, root, "package.json", `{
  "name": "javascript-fixture",
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "start": "node src/index.js"
  },
  "dependencies": {},
  "devDependencies": {}
}
`)
		writeFile(t, root, "src/index.js", `export function add(left, right) {
  return left + right;
}

console.log(add(1, 2));
`)
		run(t, root, "git", "init")
		run(t, root, "git", "add", ".")
		run(t, root, "git", "-c", "user.email=kumite@example.test", "-c", "user.name=Kumite", "commit", "-m", "initial")
		writeFile(t, root, "src/index.js", `export function add(left, right) {
  return left + right;
}

export function multiply(left, right) {
  return left * right;
}

console.log(add(1, 2));
`)

		env := append(os.Environ(), "FALLOW_VERIFY_CACHE_DIR="+filepath.Join(t.TempDir(), "fallow-verify"))
		runReviewCommands(t, root, "javascript", env)
	})
}

func runReviewCommands(t *testing.T, root, language string, env []string) {
	t.Helper()

	commands, err := reviewCommandsForLanguage(language)
	if err != nil {
		t.Fatalf("review commands for %s: %v", language, err)
	}
	if env == nil {
		env = os.Environ()
	}
	for _, command := range append(commands.Primary, commands.Structured...) {
		runShellWithEnv(t, root, env, command)
	}
}

func writeFile(t *testing.T, root, name, content string) {
	t.Helper()

	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent directory for %s: %v", name, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	runWithEnv(t, dir, os.Environ(), name, args...)
}

func runWithEnv(t *testing.T, dir string, env []string, name string, args ...string) {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", shellCommand(name, args), err, string(output))
	}
}

func runShellWithEnv(t *testing.T, dir string, env []string, command string) {
	t.Helper()

	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", command, err, string(output))
	}
}

func shellCommand(name string, args []string) string {
	command := name
	for _, arg := range args {
		command += " " + arg
	}
	return command
}
