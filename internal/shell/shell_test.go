package shell

import "testing"

func TestLineQuotesShellSensitiveValues(t *testing.T) {
	t.Parallel()

	got := Line(Command{
		Env:  []string{"A=value with space"},
		Name: "cmd",
		Args: []string{"plain", "has space", "has'quote"},
	})
	want := "'A=value with space' cmd plain 'has space' 'has'\\''quote'"
	if got != want {
		t.Fatalf("Line() = %q, want %q", got, want)
	}
}
