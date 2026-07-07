package cmd

import (
	"fmt"
	"os"

	"github.com/igor-kupczynski/knot/internal/config"
	"github.com/spf13/cobra"
)

var grepCmd = &cobra.Command{
	Use:                "grep [pattern] [flags...]",
	Short:              "Search the knowledge base with ripgrep",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "knot: grep requires a pattern")
			os.Exit(2)
		}

		kbRoot, err := config.KBRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		rgArgs, err := buildSearchArgs(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		os.Exit(runRg(kbRoot, rgArgs...))
	},
}
