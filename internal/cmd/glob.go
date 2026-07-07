package cmd

import (
	"fmt"
	"os"

	"github.com/igor-kupczynski/knot/internal/config"
	"github.com/igor-kupczynski/knot/internal/kb"
	"github.com/spf13/cobra"
)

var globCmd = &cobra.Command{
	Use:                "glob [pattern]",
	Short:              "List knowledge base files matching a pattern",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 1 {
			fmt.Fprintln(os.Stderr, "knot: glob accepts at most one pattern")
			os.Exit(2)
		}

		kbRoot, err := config.KBRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		rgArgs := []string{"--files"}
		if len(args) == 1 {
			pattern := kb.WrapGlobPattern(args[0])
			rgArgs = append(rgArgs, "-g", pattern)
		}

		os.Exit(runRg(kbRoot, rgArgs...))
	},
}
