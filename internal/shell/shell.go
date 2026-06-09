package shell

import "strings"

type Command struct {
	Name string
	Args []string
	Env  []string
}

func Line(command Command) string {
	parts := append([]string{}, command.Env...)
	parts = append(parts, command.Name)
	parts = append(parts, command.Args...)
	for index, part := range parts {
		parts[index] = Quote(part)
	}

	return strings.Join(parts, " ")
}

func Quote(value string) string {
	if value == "" {
		return "''"
	}
	if strings.IndexFunc(value, requiresQuote) == -1 {
		return value
	}

	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func requiresQuote(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
		return false
	case r >= 'A' && r <= 'Z':
		return false
	case r >= '0' && r <= '9':
		return false
	}

	return !strings.ContainsRune("-_./:=@", r)
}
