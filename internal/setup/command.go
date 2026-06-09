package setup

import "kumite/internal/shell"

type command struct {
	Name     string
	Args     []string
	Env      []string
	Optional bool
	Note     string
	Resolve  func() (command, error)
}

func shellLine(command command) string {
	return shell.Line(shell.Command{
		Name: command.Name,
		Args: command.Args,
		Env:  command.Env,
	})
}
