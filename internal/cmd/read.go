package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/igor-kupczynski/knot/internal/config"
	"github.com/igor-kupczynski/knot/internal/kb"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read <name-or-path>",
	Short: "Print a page by name or KB-relative path",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		kbRoot, err := config.KBRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		query := args[0]

		if path, ok := kb.ResolvePath(kbRoot, query); ok {
			printFile(path)
			return
		}

		if path, ok, err := kb.ResolvePage(kbRoot, query); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		} else if ok {
			printFile(path)
			return
		}

		candidates, err := kb.PageCandidates(kbRoot, query)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		fmt.Fprintf(os.Stderr, "knot: page not found: %s\n", query)
		for _, c := range candidates {
			fmt.Fprintf(os.Stderr, "  %s\n", c)
		}
		os.Exit(1)
	},
}

func printFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "knot: read %s: %v\n", path, err)
		os.Exit(2)
	}
	if _, err := os.Stdout.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "knot: write stdout: %v\n", err)
		os.Exit(2)
	}
	if len(data) > 0 && !strings.HasSuffix(string(data), "\n") {
		fmt.Fprintln(os.Stdout)
	}
}

