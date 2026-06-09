package setup

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/agents/*.md templates/chains/*.chain.md templates/config/*.json templates/handoffs/*.md templates/instructions/*.md templates/memory/*.md templates/plans/*.md templates/skills/*/SKILL.md
var projectTemplates embed.FS

type projectArtifact struct {
	Template         string
	Target           string
	PreserveExisting bool
}

func writeSubagentFiles(options Options) error {
	options = withDefaults(options)
	return writeProjectArtifacts(options, "kumite subagents", subagentArtifacts(options.AgentsDir))
}

func writeChainFiles(options Options) error {
	options = withDefaults(options)
	artifacts := []projectArtifact{
		{Template: "templates/chains/kumite-loop.chain.md", Target: filepath.Join(options.ChainsDir, "kumite-loop.chain.md")},
	}
	if err := writeProjectArtifacts(options, "kumite pi config", artifacts); err != nil {
		return err
	}
	fmt.Fprintf(options.Stdout, "$ merge %s\n", options.MCPConfigPath)
	if !options.DryRun {
		if err := mergeProjectMCPConfig("templates/config/mcp.json", options.MCPConfigPath); err != nil {
			return err
		}
	}
	return writePiSettings(options)
}

func writeProjectInstructions(options Options) error {
	options = withDefaults(options)
	bridgePath := filepath.Join(filepath.Dir(options.Instructions), "AGENTS.md")
	fmt.Fprintf(options.Stdout, "==> kumite parent instructions\n")
	fmt.Fprintf(options.Stdout, "$ upsert %s\n", options.Instructions)
	fmt.Fprintf(options.Stdout, "$ upsert %s\n", bridgePath)
	if options.DryRun {
		return nil
	}

	block, err := fs.ReadFile(projectTemplates, "templates/instructions/AGENTS.kumite.md")
	if err != nil {
		return fmt.Errorf("read project instructions template: %w", err)
	}
	if err := upsertMarkedBlock(options.Instructions, normalizeTrailingNewline(block)); err != nil {
		return fmt.Errorf("write project instructions: %w", err)
	}

	bridge, err := fs.ReadFile(projectTemplates, "templates/instructions/AGENTS.pi-bridge.md")
	if err != nil {
		return fmt.Errorf("read pi bridge instructions template: %w", err)
	}
	if err := upsertMarkedBlock(bridgePath, normalizeTrailingNewline(bridge)); err != nil {
		return fmt.Errorf("write pi bridge instructions: %w", err)
	}

	return nil
}

func writeMemoryFiles(options Options) error {
	options = withDefaults(options)
	return writeProjectArtifacts(options, "kumite memory docs", memoryArtifacts(options.MemoryDir))
}

func subagentArtifacts(agentsDir string) []projectArtifact {
	return []projectArtifact{
		{Template: "templates/agents/orchestrator.md", Target: filepath.Join(agentsDir, "orchestrator.md")},
		{Template: "templates/agents/scout.md", Target: filepath.Join(agentsDir, "scout.md")},
		{Template: "templates/agents/planner.md", Target: filepath.Join(agentsDir, "planner.md")},
		{Template: "templates/agents/planner-fallback.md", Target: filepath.Join(agentsDir, "planner-fallback.md")},
		{Template: "templates/agents/implementer.md", Target: filepath.Join(agentsDir, "implementer.md")},
		{Template: "templates/agents/reviewer.md", Target: filepath.Join(agentsDir, "reviewer.md")},
		{Template: "templates/agents/curator.md", Target: filepath.Join(agentsDir, "curator.md")},
	}
}

func memoryArtifacts(memoryDir string) []projectArtifact {
	kumiteDir := filepath.Dir(memoryDir)
	return []projectArtifact{
		{Template: "templates/memory/architecture.md", Target: filepath.Join(memoryDir, "architecture.md"), PreserveExisting: true},
		{Template: "templates/memory/code-standards.md", Target: filepath.Join(memoryDir, "code-standards.md"), PreserveExisting: true},
		{Template: "templates/memory/business-rules.md", Target: filepath.Join(memoryDir, "business-rules.md"), PreserveExisting: true},
		{Template: "templates/memory/project-status.md", Target: filepath.Join(memoryDir, "project-status.md"), PreserveExisting: true},
		{Template: "templates/plans/README.md", Target: filepath.Join(kumiteDir, "plans", "README.md"), PreserveExisting: true},
		{Template: "templates/handoffs/README.md", Target: filepath.Join(kumiteDir, "handoffs", "current", "README.md"), PreserveExisting: true},
	}
}

func writeProjectArtifacts(options Options, label string, artifacts []projectArtifact) error {
	fmt.Fprintf(options.Stdout, "==> %s\n", label)
	for _, artifact := range artifacts {
		action := "write"
		if artifact.PreserveExisting {
			action = "create-if-missing"
		}
		fmt.Fprintf(options.Stdout, "$ %s %s\n", action, artifact.Target)
		if options.DryRun {
			continue
		}

		if err := writeProjectArtifact(artifact); err != nil {
			return err
		}
	}

	return nil
}

func writeProjectArtifact(artifact projectArtifact) error {
	if artifact.PreserveExisting {
		if _, err := os.Stat(artifact.Target); err == nil {
			return nil
		} else if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("check artifact %s: %w", artifact.Target, err)
		}
	}

	content, err := fs.ReadFile(projectTemplates, artifact.Template)
	if err != nil {
		return fmt.Errorf("read artifact template %s: %w", artifact.Template, err)
	}
	if err := os.MkdirAll(filepath.Dir(artifact.Target), 0o755); err != nil {
		return fmt.Errorf("create artifact directory %s: %w", filepath.Dir(artifact.Target), err)
	}
	if err := os.WriteFile(artifact.Target, normalizeTrailingNewline(content), 0o644); err != nil {
		return fmt.Errorf("write artifact %s: %w", artifact.Target, err)
	}

	return nil
}

func mergeProjectMCPConfig(templatePath string, targetPath string) error {
	templateConfig, err := readTemplateJSONMap(templatePath)
	if err != nil {
		return err
	}

	existingConfig := map[string]any{}
	existingContent, err := os.ReadFile(targetPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read mcp config %s: %w", targetPath, err)
	}
	if err == nil {
		if parseErr := json.Unmarshal(existingContent, &existingConfig); parseErr != nil {
			return fmt.Errorf("parse mcp config %s: %w", targetPath, parseErr)
		}
		if existingConfig == nil {
			existingConfig = map[string]any{}
		}
	}

	merged := mergeJSONMaps(existingConfig, templateConfig)
	content, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return fmt.Errorf("encode mcp config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create mcp config directory %s: %w", filepath.Dir(targetPath), err)
	}
	if err := os.WriteFile(targetPath, append(content, '\n'), 0o644); err != nil {
		return fmt.Errorf("write mcp config %s: %w", targetPath, err)
	}

	return nil
}

func readTemplateJSONMap(templatePath string) (map[string]any, error) {
	content, err := fs.ReadFile(projectTemplates, templatePath)
	if err != nil {
		return nil, fmt.Errorf("read artifact template %s: %w", templatePath, err)
	}

	var value map[string]any
	if err := json.Unmarshal(content, &value); err != nil {
		return nil, fmt.Errorf("parse artifact template %s: %w", templatePath, err)
	}
	if value == nil {
		return map[string]any{}, nil
	}

	return value, nil
}

func mergeJSONMaps(existing map[string]any, template map[string]any) map[string]any {
	merged := copyJSONMap(existing)
	for key, templateValue := range template {
		existingMap, existingOK := merged[key].(map[string]any)
		templateMap, templateOK := templateValue.(map[string]any)
		if existingOK && templateOK {
			merged[key] = mergeJSONMaps(existingMap, templateMap)
			continue
		}
		merged[key] = templateValue
	}

	return merged
}

func copyJSONMap(source map[string]any) map[string]any {
	copied := make(map[string]any, len(source))
	for key, value := range source {
		if nested, ok := value.(map[string]any); ok {
			copied[key] = copyJSONMap(nested)
			continue
		}
		copied[key] = value
	}

	return copied
}

func writePiSettings(options Options) error {
	options = withDefaults(options)
	fmt.Fprintf(options.Stdout, "==> kumite pi package settings\n")
	fmt.Fprintf(options.Stdout, "$ upsert %s\n", options.PiSettingsPath)
	if options.DryRun {
		return nil
	}

	settings, err := readPiSettings(options.PiSettingsPath)
	if err != nil {
		return err
	}
	packages, err := packagesFromSettings(settings)
	if err != nil {
		return err
	}
	settings["packages"] = upsertKumitePackage(packages, options.PiPackage)

	content, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encode pi settings: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(options.PiSettingsPath), 0o755); err != nil {
		return fmt.Errorf("create pi settings directory %s: %w", filepath.Dir(options.PiSettingsPath), err)
	}
	if err := os.WriteFile(options.PiSettingsPath, append(content, '\n'), 0o644); err != nil {
		return fmt.Errorf("write pi settings %s: %w", options.PiSettingsPath, err)
	}

	return nil
}

func readPiSettings(path string) (map[string]any, error) {
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read pi settings %s: %w", path, err)
	}

	var settings map[string]any
	if err := json.Unmarshal(content, &settings); err != nil {
		return nil, fmt.Errorf("parse pi settings %s: %w", path, err)
	}
	if settings == nil {
		return map[string]any{}, nil
	}

	return settings, nil
}

func packagesFromSettings(settings map[string]any) ([]string, error) {
	value, ok := settings["packages"]
	if !ok {
		return nil, nil
	}

	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("pi settings packages must be an array")
	}
	packages := make([]string, 0, len(items))
	for _, item := range items {
		packageSource, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("pi settings packages must contain only strings")
		}
		packages = append(packages, packageSource)
	}

	return packages, nil
}

func upsertKumitePackage(packages []string, desired string) []string {
	desiredIsLocal := isLocalKumitePackageSource(desired)
	result := make([]string, 0, len(packages)+1)
	hasDesired := false
	hasCompatibleKumite := false

	for _, packageSource := range packages {
		if !isKumitePackageSource(packageSource) {
			result = append(result, packageSource)
			continue
		}

		if packageSource == desired {
			if !hasDesired {
				result = append(result, packageSource)
				hasDesired = true
			}
			continue
		}

		if desiredIsLocal {
			continue
		}

		if !hasCompatibleKumite {
			result = append(result, packageSource)
			hasCompatibleKumite = true
		}
	}

	if !hasDesired && (desiredIsLocal || !hasCompatibleKumite) {
		result = append(result, desired)
	}

	return result
}

func isKumitePackageSource(packageSource string) bool {
	if packageSource == "pi-kumite" || isNpmKumitePackageSource(packageSource) {
		return true
	}

	return filepath.Base(packageSource) == "pi-kumite"
}

func isLocalKumitePackageSource(packageSource string) bool {
	return isKumitePackageSource(packageSource) && !isNpmKumitePackageSource(packageSource)
}

func isNpmKumitePackageSource(packageSource string) bool {
	return packageSource == "npm:pi-kumite" || strings.HasPrefix(packageSource, "npm:pi-kumite@")
}

func normalizeTrailingNewline(content []byte) []byte {
	text := strings.TrimRight(string(content), "\n") + "\n"
	return []byte(text)
}

func upsertMarkedBlock(path string, block []byte) error {
	const begin = "<!-- kumite:begin -->"
	const end = "<!-- kumite:end -->"

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return fmt.Errorf("create instructions directory: %w", err)
	}

	blockText := strings.TrimRight(string(block), "\n")
	existingBytes, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if os.IsNotExist(err) {
		return os.WriteFile(path, []byte(blockText+"\n"), 0o644)
	}

	existing := strings.TrimRight(string(existingBytes), "\n")
	start := strings.Index(existing, begin)
	stop := strings.Index(existing, end)
	if start >= 0 && stop > start {
		stop += len(end)
		updated := strings.TrimRight(existing[:start], "\n") + "\n\n" + blockText + "\n\n" + strings.TrimLeft(existing[stop:], "\n")
		return os.WriteFile(path, []byte(strings.TrimRight(updated, "\n")+"\n"), 0o644)
	}

	updated := existing + "\n\n" + blockText + "\n"
	return os.WriteFile(path, []byte(updated), 0o644)
}
