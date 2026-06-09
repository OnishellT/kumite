package installer

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

const asciiLogo = ` _                  _ _       
| | ___   _ _ __ __ (_) |_ ___ 
| |/ / | | | '_ ` + "`" + ` _ \| | __/ _ \
|   <| |_| | | | | | | | ||  __/
|_|\_\\__,_|_| |_| |_|_|\__\___|`

const japaneseLogo = "組手"

type tuiModel struct {
	options Options
	cursor  int
	setup   bool
	state   tuiState
	result  Result
	err     error
}

type tuiState int

const (
	tuiReady tuiState = iota
	tuiRunning
	tuiDone
)

type resultMsg struct {
	result Result
	err    error
}

func RunTUI(options Options) error {
	options = withDefaults(options)
	model := tuiModel{
		options: options,
		setup:   options.RunGlobalSetup,
	}
	_, err := tea.NewProgram(model).Run()
	return err
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleKey(msg.String())
	case resultMsg:
		m.state = tuiDone
		m.result = msg.result
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

func (m tuiModel) handleKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+c", "q", "esc":
		return m, tea.Quit
	case "up", "k":
		return m.moveCursor(-1), nil
	case "down", "j":
		return m.moveCursor(1), nil
	case " ":
		return m.toggleCurrentOption(), nil
	case "enter":
		return m.handleEnter()
	default:
		return m, nil
	}
}

func (m tuiModel) moveCursor(delta int) tuiModel {
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor > 1 {
		m.cursor = 1
	}
	return m
}

func (m tuiModel) toggleCurrentOption() tuiModel {
	if m.cursor == 1 {
		m.setup = !m.setup
		m.options.RunGlobalSetup = m.setup
	}
	return m
}

func (m tuiModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case tuiReady:
		m.state = tuiRunning
		return m, installCmd(m.options)
	case tuiDone:
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m tuiModel) View() tea.View {
	var b strings.Builder
	b.WriteString("\x1b[36m")
	b.WriteString(asciiLogo)
	b.WriteString("\x1b[0m\n")
	b.WriteString("\x1b[2m")
	b.WriteString(japaneseLogo)
	b.WriteString("\x1b[0m\n\n")

	switch m.state {
	case tuiReady:
		b.WriteString("Install kumite globally\n\n")
		b.WriteString(m.optionLine(0, fmt.Sprintf("install binary to %s", m.options.BinDir), true))
		b.WriteString(m.optionLine(1, "run kumite setup --global after install", m.setup))
		b.WriteString("\n↑/↓ move  space toggle  enter install  q quit\n")
	case tuiRunning:
		b.WriteString("Installing...\n\n")
		b.WriteString("This may take a while if global setup is enabled.\n")
	case tuiDone:
		if m.err != nil {
			b.WriteString("\x1b[31mInstall failed\x1b[0m\n\n")
			b.WriteString(m.err.Error())
			b.WriteString("\n\n")
		} else {
			b.WriteString("\x1b[32mInstall complete\x1b[0m\n\n")
			b.WriteString("kumite: ")
			b.WriteString(m.result.BinPath)
			b.WriteString("\n")
			if m.result.Warning != "" {
				b.WriteString("warning: ")
				b.WriteString(m.result.Warning)
				b.WriteString("\n")
			}
			b.WriteString("\nNext:\n")
			b.WriteString("  kumite init\n")
			b.WriteString("  kumite setup --global --keep-going\n")
		}
		b.WriteString("press enter to exit\n")
	}

	return tea.NewView(b.String())
}

func (m tuiModel) optionLine(index int, label string, checked bool) string {
	cursor := " "
	if m.cursor == index {
		cursor = ">"
	}
	mark := " "
	if checked {
		mark = "x"
	}
	return fmt.Sprintf("%s [%s] %s\n", cursor, mark, label)
}

func installCmd(options Options) tea.Cmd {
	return func() tea.Msg {
		result, err := Run(options)
		return resultMsg{result: result, err: err}
	}
}
