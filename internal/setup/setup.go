package setup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const defaultCommandTimeout = 10 * time.Minute

type Options struct {
	Context        context.Context
	Languages      []string
	DryRun         bool
	KeepGoing      bool
	CommandTimeout time.Duration
	SkipTooling    bool
	SkillsDir      string
	PiSkillsDir    string
	AgentsDir      string
	ChainsDir      string
	MemoryDir      string
	MCPConfigPath  string
	PiSettingsPath string
	PiPackage      string
	Instructions   string
	SkipExtensions bool
	SkipSkills     bool
	SkipAgents     bool
	SkipMemory     bool
	Stdout         io.Writer
	Stderr         io.Writer
}

func Run(options Options) error {
	options = withDefaults(options)

	plan, err := BuildPlan(options)
	if err != nil {
		return err
	}
	summary := executeCommandPlan(options, plan)
	if summary.hasRequiredFailures() && !options.KeepGoing {
		printExecutionSummary(options, summary)
		return summary.requiredError()
	}

	summary.merge(writeArtifactPlan(options, plan))

	printExecutionSummary(options, summary)
	return summary.requiredError()
}

type Plan struct {
	ExtensionCommands []command
	Languages         []LanguagePlan
	WriteSkill        bool
	WriteAgents       bool
	WriteMemory       bool
}

type LanguagePlan struct {
	Name     string
	Commands []command
}

type artifactWriter func(Options) error

func BuildPlan(options Options) (Plan, error) {
	options = withDefaults(options)

	languages := make([]LanguagePlan, 0, len(options.Languages))
	if !options.SkipTooling {
		for _, language := range options.Languages {
			commands, err := installCommandsForLanguage(language)
			if err != nil {
				return Plan{}, err
			}
			languages = append(languages, LanguagePlan{
				Name:     language,
				Commands: commands,
			})
		}
	}

	plan := Plan{
		Languages:   languages,
		WriteSkill:  !options.SkipSkills,
		WriteAgents: !options.SkipAgents,
		WriteMemory: !options.SkipMemory,
	}
	if !options.SkipExtensions {
		plan.ExtensionCommands = piExtensionInstallCommands()
	}

	return plan, nil
}

func writeArtifactPlan(options Options, plan Plan) executionSummary {
	var summary executionSummary
	for _, write := range artifactWritersForPlan(plan) {
		if err := write(options); err != nil {
			summary.RequiredFailures = append(summary.RequiredFailures, err)
			if !options.KeepGoing {
				return summary
			}
		}
	}

	return summary
}

func artifactWritersForPlan(plan Plan) []artifactWriter {
	writers := make([]artifactWriter, 0, 5)
	if plan.WriteSkill {
		writers = append(writers, writeStaticAnalysisSkill, writePlannerSkill)
	}
	if plan.WriteAgents {
		writers = append(writers, writeProjectInstructions, writeSubagentFiles, writeChainFiles)
	}
	if plan.WriteMemory {
		writers = append(writers, writeMemoryFiles)
	}

	return writers
}

type executionSummary struct {
	RequiredFailures []error
	OptionalFailures []error
}

func (summary executionSummary) hasRequiredFailures() bool {
	return len(summary.RequiredFailures) > 0
}

func (summary executionSummary) requiredError() error {
	return errors.Join(summary.RequiredFailures...)
}

func (summary *executionSummary) merge(other executionSummary) {
	summary.RequiredFailures = append(summary.RequiredFailures, other.RequiredFailures...)
	summary.OptionalFailures = append(summary.OptionalFailures, other.OptionalFailures...)
}

func executeCommandPlan(options Options, plan Plan) executionSummary {
	var summary executionSummary
	if len(plan.ExtensionCommands) > 0 {
		fmt.Fprintln(options.Stdout, "==> pi extensions")
		extensionSummary := runCommands(options, plan.ExtensionCommands)
		summary.merge(extensionSummary)
		if extensionSummary.hasRequiredFailures() && !options.KeepGoing {
			return summary
		}
	}

	for _, language := range plan.Languages {
		languageSummary := runLanguageSetup(options, language)
		summary.merge(languageSummary)
		if languageSummary.hasRequiredFailures() && !options.KeepGoing {
			return summary
		}
	}

	return summary
}

func runLanguageSetup(options Options, language LanguagePlan) executionSummary {
	fmt.Fprintf(options.Stdout, "==> %s static-analysis tooling\n", language.Name)
	return runCommands(options, language.Commands)
}

func runCommands(options Options, commands []command) executionSummary {
	var summary executionSummary
	for _, command := range commands {
		if err := runCommand(options, command); err != nil {
			if command.Optional {
				summary.OptionalFailures = append(summary.OptionalFailures, err)
				continue
			}

			summary.RequiredFailures = append(summary.RequiredFailures, err)
			if shouldStop(options, command) {
				return summary
			}
		}
	}

	return summary
}

func shouldStop(options Options, command command) bool {
	return !options.KeepGoing && !command.Optional
}

func withDefaults(options Options) Options {
	options.Context = defaultContext(options.Context)
	options.Languages = defaultLanguagesIfEmpty(options.Languages)
	options.CommandTimeout = defaultDuration(options.CommandTimeout, defaultCommandTimeout)
	options.SkillsDir = defaultString(options.SkillsDir, ".agents/skills")
	options.PiSkillsDir = defaultString(options.PiSkillsDir, ".pi/skills")
	options.AgentsDir = defaultString(options.AgentsDir, ".pi/agents")
	options.ChainsDir = defaultString(options.ChainsDir, ".pi/chains")
	options.MemoryDir = defaultString(options.MemoryDir, ".kumite/memory")
	options.MCPConfigPath = defaultString(options.MCPConfigPath, ".pi/mcp.json")
	options.PiSettingsPath = defaultString(options.PiSettingsPath, filepath.Join(filepath.Dir(options.MCPConfigPath), "settings.json"))
	options.PiPackage = defaultString(options.PiPackage, "npm:pi-kumite")
	options.Instructions = defaultString(options.Instructions, "agents.md")
	options.Stdout = defaultWriter(options.Stdout)
	options.Stderr = defaultWriter(options.Stderr)

	return options
}

func defaultContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}

	return context.Background()
}

func defaultLanguagesIfEmpty(languages []string) []string {
	if len(languages) > 0 {
		return languages
	}

	return defaultLanguages()
}

func defaultDuration(value, fallback time.Duration) time.Duration {
	if value != 0 {
		return value
	}

	return fallback
}

func defaultString(value, fallback string) string {
	if value != "" {
		return value
	}

	return fallback
}

func defaultWriter(writer io.Writer) io.Writer {
	if writer != nil {
		return writer
	}

	return io.Discard
}

func runCommand(options Options, cmdSpec command) error {
	displayCommand, err := prepareCommand(options, cmdSpec)
	if err != nil {
		return err
	}

	if displayCommand.Note != "" {
		fmt.Fprintf(options.Stdout, "note: %s\n", displayCommand.Note)
	}

	fmt.Fprintf(options.Stdout, "$ %s\n", shellLine(displayCommand))
	if options.DryRun {
		return nil
	}

	return executeCommand(options, displayCommand)
}

func prepareCommand(options Options, cmdSpec command) (command, error) {
	if options.DryRun || cmdSpec.Resolve == nil {
		return cmdSpec, nil
	}

	resolved, err := cmdSpec.Resolve()
	if err != nil {
		return command{}, fmt.Errorf("resolve %s: %w", shellLine(cmdSpec), err)
	}

	return inheritCommandMetadata(cmdSpec, resolved), nil
}

func executeCommand(options Options, cmdSpec command) error {
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}
	timeout := options.CommandTimeout
	if timeout == 0 {
		timeout = defaultCommandTimeout
	}
	cancel := func() {}
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdSpec.Name, cmdSpec.Args...)
	if len(cmdSpec.Env) > 0 {
		cmd.Env = append(os.Environ(), cmdSpec.Env...)
	}
	cmd.Stdout = options.Stdout
	cmd.Stderr = options.Stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("%s timed out after %s", shellLine(cmdSpec), timeout)
		}
		return fmt.Errorf("%s: %w", shellLine(cmdSpec), err)
	}

	return nil
}

func inheritCommandMetadata(source, resolved command) command {
	resolved.Optional = source.Optional
	if resolved.Note == "" {
		resolved.Note = source.Note
	}
	if len(resolved.Env) == 0 {
		resolved.Env = source.Env
	}
	return resolved
}

func printExecutionSummary(options Options, summary executionSummary) {
	if len(summary.OptionalFailures) == 0 {
		return
	}

	fmt.Fprintln(options.Stderr, "optional setup steps failed:")
	for _, failure := range summary.OptionalFailures {
		fmt.Fprintf(options.Stderr, "- %v\n", failure)
	}
}
