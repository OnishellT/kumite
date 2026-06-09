package setup

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/static-analysis.SKILL.md.tmpl
var skillTemplates embed.FS

type staticAnalysisSkillView struct {
	Sections []staticAnalysisSkillSection
}

type staticAnalysisSkillSection struct {
	Title              string
	Intro              string
	Commands           []string
	StructuredIntro    string
	StructuredCommands []string
	Guidance           string
}

func renderStaticAnalysisSkill() (string, error) {
	sections, err := staticAnalysisSkillSections()
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("static-analysis.SKILL.md.tmpl").Funcs(template.FuncMap{
		"commandBlock": commandBlock,
	}).ParseFS(skillTemplates, "templates/static-analysis.SKILL.md.tmpl")
	if err != nil {
		return "", fmt.Errorf("parse static-analysis skill template: %w", err)
	}

	var content bytes.Buffer
	err = tmpl.Execute(&content, staticAnalysisSkillView{
		Sections: sections,
	})
	if err != nil {
		return "", fmt.Errorf("render static-analysis skill: %w", err)
	}

	return content.String(), nil
}

func staticAnalysisSkillSections() ([]staticAnalysisSkillSection, error) {
	definitions := skillSectionDefinitions()
	sections := make([]staticAnalysisSkillSection, 0, len(definitions))
	for _, definition := range definitions {
		commands, err := reviewCommandsForLanguage(definition.Language)
		if err != nil {
			return nil, err
		}

		sections = append(sections, staticAnalysisSkillSection{
			Title:              definition.Title,
			Intro:              definition.Intro,
			Commands:           append(definition.Preamble, commands.Primary...),
			StructuredIntro:    definition.StructuredIntro,
			StructuredCommands: commands.Structured,
			Guidance:           definition.Guidance,
		})
	}

	return sections, nil
}

func commandBlock(commands []string) string {
	var content strings.Builder
	for _, command := range commands {
		content.WriteString(command)
		content.WriteByte('\n')
	}

	return strings.TrimRight(content.String(), "\n")
}

func writeStaticAnalysisSkill(options Options) error {
	options = withDefaults(options)

	content, err := renderStaticAnalysisSkill()
	if err != nil {
		return err
	}

	paths := []string{
		filepath.Join(options.PiSkillsDir, "static-analysis-reviewer", "SKILL.md"),
	}

	fmt.Fprintf(options.Stdout, "==> agent skill\n")
	for _, path := range paths {
		fmt.Fprintf(options.Stdout, "$ write %s\n", path)
		if options.DryRun {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("create skill directory: %w", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write static-analysis skill: %w", err)
		}
	}

	return removeObsoleteSkillCopies(options)
}

func writePlannerSkill(options Options) error {
	options = withDefaults(options)
	artifact := projectArtifact{
		Template: "templates/skills/grill-with-docs/SKILL.md",
		Target:   filepath.Join(options.PiSkillsDir, "kumite-grill-with-docs", "SKILL.md"),
	}

	return writeProjectArtifacts(options, "kumite planner skill", []projectArtifact{artifact})
}

func removeObsoleteSkillCopies(options Options) error {
	obsolete := []struct {
		path    string
		markers []string
	}{
		{
			path:    filepath.Join(options.SkillsDir, "static-analysis", "SKILL.md"),
			markers: []string{"# Static Analysis Reviewer", "static-analysis-reviewer"},
		},
		{
			path:    filepath.Join(options.PiSkillsDir, "grill-with-docs", "SKILL.md"),
			markers: []string{"# Grill With Docs", "kumite planning"},
		},
	}

	for _, item := range obsolete {
		fmt.Fprintf(options.Stdout, "$ remove-if-generated %s\n", item.path)
		if options.DryRun {
			continue
		}
		if err := removeIfGenerated(item.path, item.markers); err != nil {
			return err
		}
	}

	return nil
}

func removeIfGenerated(path string, markers []string) error {
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read obsolete skill %s: %w", path, err)
	}

	text := string(content)
	for _, marker := range markers {
		if !strings.Contains(text, marker) {
			return nil
		}
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove obsolete skill %s: %w", path, err)
	}
	_ = os.Remove(filepath.Dir(path))
	return nil
}
