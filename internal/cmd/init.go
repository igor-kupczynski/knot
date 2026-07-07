package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/igor-kupczynski/knot/internal/config"
	"github.com/spf13/cobra"
)

//go:embed assets/gitattributes
var gitattributesTemplate string

//go:embed assets/gitignore
var gitignoreTemplate string

//go:embed assets/skills/ingest/SKILL.md
var ingestSkillTemplate string

//go:embed assets/skills/git-conflict-resolve/SKILL.md
var gitConflictResolveSkillTemplate string

var initFrom string

var initCmd = &cobra.Command{
	Use:   "init <path>",
	Short: "Scaffold a knowledge base, or clone an existing one with --from",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kbPath, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("knot: resolve path: %w", err)
		}

		if initFrom != "" {
			if err := cloneKB(kbPath, initFrom); err != nil {
				return err
			}
			if err := config.WriteConfig(kbPath); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Cloned knowledge base into %s\n", kbPath)
			return nil
		}

		if err := scaffoldKB(kbPath); err != nil {
			return err
		}
		if err := config.WriteConfig(kbPath); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Initialized knowledge base at %s\n", kbPath)
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initFrom, "from", "", "git URL of an existing KB to clone instead of scaffolding")
}

func cloneKB(kbPath, url string) error {
	if _, err := lookPathGit(); err != nil {
		return err
	}

	if info, err := os.Stat(kbPath); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("knot: %s exists and is not a directory", kbPath)
		}
		entries, err := os.ReadDir(kbPath)
		if err != nil {
			return fmt.Errorf("knot: read %s: %w", kbPath, err)
		}
		if len(entries) > 0 {
			return fmt.Errorf("knot: %s already exists and is not empty; refusing to clone into it", kbPath)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("knot: stat %s: %w", kbPath, err)
	}

	parent := filepath.Dir(kbPath)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("knot: create parent dir %s: %w", parent, err)
	}

	if err := runGit(parent, "clone", url, kbPath); err != nil {
		return fmt.Errorf("knot: git clone: %w", err)
	}

	return nil
}

func scaffoldKB(kbPath string) error {
	if err := os.MkdirAll(kbPath, 0o755); err != nil {
		return fmt.Errorf("knot: create KB root: %w", err)
	}

	scaffold := []struct {
		rel     string
		content string
		isDir   bool
	}{
		{"AGENTS.md", agentsMDTemplate, false},
		{".gitattributes", gitattributesTemplate, false},
		{".gitignore", gitignoreTemplate, false},
		{"index.md", "# Index\n\n", false},
		{"log.md", "# Log\n\n", false},
		{".agents/skills/ingest/SKILL.md", ingestSkillTemplate, false},
		{".agents/skills/git-conflict-resolve/SKILL.md", gitConflictResolveSkillTemplate, false},
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

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("knot: create parent dir for %s: %w", target, err)
		}

		if err := os.WriteFile(target, []byte(item.content), 0o644); err != nil {
			return fmt.Errorf("knot: write %s: %w", target, err)
		}
	}

	claudeSkills := filepath.Join(kbPath, ".claude", "skills")
	if _, err := os.Lstat(claudeSkills); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(claudeSkills), 0o755); err != nil {
			return fmt.Errorf("knot: create .claude: %w", err)
		}
		if err := os.Symlink("../.agents/skills", claudeSkills); err != nil {
			return fmt.Errorf("knot: create .claude/skills symlink: %w", err)
		}
	}

	return initGitRepo(kbPath)
}

func initGitRepo(kbPath string) error {
	if _, err := lookPathGit(); err != nil {
		fmt.Fprintln(os.Stdout, "Note: git not found; KB was not initialized as a git repository.")
		return nil
	}

	if isGitRepo(kbPath) {
		return nil
	}

	if err := runGit(kbPath, "init"); err != nil {
		return fmt.Errorf("knot: git init: %w", err)
	}

	if err := runGit(kbPath, "add", "-A"); err != nil {
		return fmt.Errorf("knot: git add: %w", err)
	}

	if err := gitCommit(kbPath, "init: knot scaffold"); err != nil {
		return fmt.Errorf("knot: git commit: %w", err)
	}

	return nil
}
