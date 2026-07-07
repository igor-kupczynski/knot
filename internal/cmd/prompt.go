package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var snippetFlag bool

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Print agent-facing usage instructions",
	Run: func(cmd *cobra.Command, args []string) {
		text := promptFull
		if snippetFlag {
			text = promptSnippet
		}
		fmt.Print(text)
	},
}

func init() {
	promptCmd.Flags().BoolVar(&snippetFlag, "snippet", false, "emit short config paragraph")
}
