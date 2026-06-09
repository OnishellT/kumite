package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"kumite/internal/setup"
)

const version = "0.1.1"

func Run(args []string) int {
	return run(args, os.Stdout, os.Stderr)
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stdout)
		return 0
	}

	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, version)
		return 0
	case "setup":
		return runSetup(args[1:], stdout, stderr)
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runSetup(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("setup", flag.ContinueOnError)
	flags.SetOutput(stderr)

	languages := flags.String("languages", "go,python,rust,javascript", "comma-separated languages to install tooling for")
	dryRun := flags.Bool("dry-run", false, "print setup steps without executing installers")
	keepGoing := flags.Bool("keep-going", false, "continue after an installer fails")
	project := registerProjectFlags(flags)
	skipExtensions := flags.Bool("skip-pi-extensions", false, "skip installing pi extension packages")
	globalOnly := flags.Bool("global", false, "install global tooling and pi extensions without writing project files")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	options := setup.Options{
		Languages:      splitCSV(*languages),
		DryRun:         *dryRun,
		KeepGoing:      *keepGoing,
		SkipExtensions: *skipExtensions,
		Stdout:         stdout,
		Stderr:         stderr,
	}
	project.applyTo(&options)
	if *globalOnly {
		options.SkipSkills = true
		options.SkipAgents = true
		options.SkipMemory = true
	}
	if err := setup.Run(options); err != nil {
		fmt.Fprintf(stderr, "setup failed: %v\n", err)
		return 1
	}

	return 0
}

func runInit(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(stderr)

	dryRun := flags.Bool("dry-run", false, "print project files without writing them")
	project := registerProjectFlags(flags)

	if err := flags.Parse(args); err != nil {
		return 2
	}

	options := setup.Options{
		DryRun:         *dryRun,
		SkipTooling:    true,
		SkipExtensions: true,
		Stdout:         stdout,
		Stderr:         stderr,
	}
	project.applyTo(&options)
	if err := setup.Run(options); err != nil {
		fmt.Fprintf(stderr, "init failed: %v\n", err)
		return 1
	}

	return 0
}

type projectFlags struct {
	skillsDir     *string
	piSkillsDir   *string
	agentsDir     *string
	chainsDir     *string
	memoryDir     *string
	mcpConfigPath *string
	piSettings    *string
	piPackage     *string
	instructions  *string
	skipSkills    *bool
	skipAgents    *bool
	skipMemory    *bool
}

func registerProjectFlags(flags *flag.FlagSet) projectFlags {
	return projectFlags{
		skillsDir:     flags.String("skills-dir", ".agents/skills", "directory for generated agent skills"),
		piSkillsDir:   flags.String("pi-skills-dir", ".pi/skills", "directory for generated pi project skills"),
		agentsDir:     flags.String("agents-dir", ".pi/agents", "directory for generated pi subagent definitions"),
		chainsDir:     flags.String("chains-dir", ".pi/chains", "directory for generated pi subagent chains"),
		memoryDir:     flags.String("memory-dir", ".kumite/memory", "directory for generated kumite memory docs"),
		mcpConfigPath: flags.String("mcp-config", ".pi/mcp.json", "path for generated pi MCP config"),
		piSettings:    flags.String("pi-settings", ".pi/settings.json", "path for generated pi project settings"),
		piPackage:     flags.String("pi-package", "npm:pi-kumite", "pi package source to register in generated project settings"),
		instructions:  flags.String("instructions", "agents.md", "project instruction index file to upsert with kumite guidance"),
		skipSkills:    flags.Bool("skip-skills", false, "skip writing agent skill guidance"),
		skipAgents:    flags.Bool("skip-agents", false, "skip writing pi subagent definitions"),
		skipMemory:    flags.Bool("skip-memory", false, "skip writing kumite memory docs"),
	}
}

func (flags projectFlags) applyTo(options *setup.Options) {
	options.SkillsDir = *flags.skillsDir
	options.PiSkillsDir = *flags.piSkillsDir
	options.AgentsDir = *flags.agentsDir
	options.ChainsDir = *flags.chainsDir
	options.MemoryDir = *flags.memoryDir
	options.MCPConfigPath = *flags.mcpConfigPath
	options.PiSettingsPath = *flags.piSettings
	options.PiPackage = *flags.piPackage
	options.Instructions = *flags.instructions
	options.SkipSkills = *flags.skipSkills
	options.SkipAgents = *flags.skipAgents
	options.SkipMemory = *flags.skipMemory
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	languages := make([]string, 0, len(parts))
	for _, part := range parts {
		language := strings.TrimSpace(part)
		if language != "" {
			languages = append(languages, language)
		}
	}

	return languages
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "kumite sets up the pi-agent coding harness.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  kumite init [--dry-run] [--pi-package source]")
	fmt.Fprintln(w, "  kumite setup [--languages go,python,rust,javascript] [--dry-run]")
	fmt.Fprintln(w, "  kumite setup --global [--languages go,python,rust,javascript]")
	fmt.Fprintln(w, "  kumite version")
}
