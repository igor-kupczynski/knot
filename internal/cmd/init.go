package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/igor-kupczynski/knot/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init <path>",
	Short: "Scaffold a knowledge base",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kbPath, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("knot: resolve path: %w", err)
		}

		if err := os.MkdirAll(kbPath, 0o755); err != nil {
			return fmt.Errorf("knot: create KB root: %w", err)
		}

		scaffold := []struct {
			rel     string
			content string
			isDir   bool
		}{
			{"AGENTS.md", agentsMDTemplate, false},
			{"index.md", "# Index\n\n", false},
			{"log.md", "# Log\n\n", false},
			{"pages", "", true},
			{"inbox", "", true},
		}

		for _, item := range scaffold {
			target := filepath.Join(kbPath, item.rel)
			if _, err := os.Stat(target); err == nil {
				continue
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("knot: stat %s: %w", target, err)
			}

			if item.isDir {
				if err := os.MkdirAll(target, 0o755); err != nil {
					return fmt.Errorf("knot: create %s: %w", target, err)
				}
				continue
			}

			if err := os.WriteFile(target, []byte(item.content), 0o644); err != nil {
				return fmt.Errorf("knot: write %s: %w", target, err)
			}
		}

		if err := config.WriteConfig(kbPath); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Initialized knowledge base at %s\n", kbPath)
		return nil
	},
}
