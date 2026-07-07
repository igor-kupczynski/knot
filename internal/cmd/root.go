package cmd

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

//go:embed assets/AGENTS.md
var agentsMDTemplate string

//go:embed assets/prompt.txt
var promptFull string

//go:embed assets/prompt_snippet.txt
var promptSnippet string

var rootCmd = &cobra.Command{
	Use:   "knot",
	Short: "Personal knowledge base CLI",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	return 0
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(grepCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(globCmd)
	rootCmd.AddCommand(captureCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(promptCmd)
}
