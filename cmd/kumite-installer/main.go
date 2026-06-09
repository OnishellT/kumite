package main

import (
	"flag"
	"fmt"
	"os"

	"kumite/internal/installer"
)

func main() {
	flags := flag.NewFlagSet("kumite-installer", flag.ExitOnError)
	sourceDir := flags.String("source-dir", ".", "kumite source directory")
	binDir := flags.String("bin-dir", "", "directory where the kumite binary will be installed")
	runGlobalSetup := flags.Bool("setup-global", false, "run kumite setup --global --keep-going after installing the binary")
	yes := flags.Bool("yes", false, "run non-interactively")
	dryRun := flags.Bool("dry-run", false, "print installer steps without executing them")
	if err := flags.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}

	options := installer.Options{
		SourceDir:      *sourceDir,
		BinDir:         *binDir,
		RunGlobalSetup: *runGlobalSetup,
		DryRun:         *dryRun,
		Stdout:         os.Stdout,
		Stderr:         os.Stderr,
	}

	var err error
	if *yes || *dryRun {
		var result installer.Result
		result, err = installer.Run(options)
		if err == nil {
			if *dryRun {
				fmt.Fprintf(os.Stdout, "planned kumite install at %s\n", result.BinPath)
			} else {
				fmt.Fprintf(os.Stdout, "installed kumite at %s\n", result.BinPath)
			}
		}
	} else {
		err = installer.RunTUI(options)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "kumite installer failed: %v\n", err)
		os.Exit(1)
	}
}
