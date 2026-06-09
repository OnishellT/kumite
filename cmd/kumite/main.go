package main

import (
	"os"

	"kumite/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
