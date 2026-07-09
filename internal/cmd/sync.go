package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/igor-kupczynski/knot/internal/config"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Commit, pull, and push the knowledge base via git",
	Run: func(cmd *cobra.Command, args []string) {
		kbRoot, err := config.KBRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		hostname, err := os.Hostname()
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("knot: hostname: %w", err))
			os.Exit(2)
		}

		if err := runSync(kbRoot, hostname, time.Now()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func runSync(kbRoot, hostname string, now time.Time) error {
	if _, err := lookPathGit(); err != nil {
		return err
	}

	if !isGitRepo(kbRoot) {
		return fmt.Errorf(`knot: %s is not a git repository

Run git init in the KB root, or knot init <path> to scaffold a new knowledge base with git initialized`, kbRoot)
	}

	if isMidRebaseOrMerge(kbRoot) {
		return fmt.Errorf(`knot: KB is mid-rebase or mid-merge

Finish or abort the in-progress git operation before syncing:
  git rebase --abort   # if a rebase is stuck
  git merge --abort    # if a merge is stuck

Then resolve any conflicts with /git-conflict-resolve in a resident agent session.`)
	}

	if err := runGit(kbRoot, "add", "-A"); err != nil {
		return fmt.Errorf("knot: git add: %w", err)
	}

	dirty, err := isGitDirty(kbRoot)
	if err != nil {
		return err
	}

	if dirty {
		msg := fmt.Sprintf("sync: %s %s", hostname, now.Format(time.RFC3339))
		if err := gitCommit(kbRoot, msg); err != nil {
			return fmt.Errorf("knot: git commit: %w", err)
		}
		fmt.Fprintln(os.Stdout, "Committed local changes.")
	} else {
		fmt.Fprintln(os.Stdout, "Nothing to commit.")
	}

	remote, err := firstGitRemote(kbRoot)
	if err != nil {
		return err
	}

	if remote == "" {
		fmt.Fprintln(os.Stdout, "No git remote configured; sync is local-only until you add one (git remote add origin <url>).")
		return nil
	}

	if !hasUpstream(kbRoot) {
		// First sync to this remote: nothing to pull, push and set the upstream.
		branch, err := currentBranch(kbRoot)
		if err != nil {
			return err
		}
		if err := safeGitArg("remote", remote); err != nil {
			return err
		}
		if err := safeGitArg("branch", branch); err != nil {
			return err
		}
		if err := runGit(kbRoot, "push", "-u", "--", remote, branch); err != nil {
			return fmt.Errorf("knot: git push -u %s %s: %w", remote, branch, err)
		}
		fmt.Fprintln(os.Stdout, "Synced with remote (first push; upstream set).")
		return nil
	}

	if err := runGit(kbRoot, "pull", "--rebase"); err != nil {
		paths, listErr := conflictingPaths(kbRoot)
		if listErr == nil && len(paths) > 0 {
			_ = runGit(kbRoot, "rebase", "--abort")
			var b strings.Builder
			b.WriteString("knot: rebase conflict while syncing\n\nConflicting files:\n")
			for _, p := range paths {
				b.WriteString("  ")
				b.WriteString(p)
				b.WriteByte('\n')
			}
			b.WriteString("\nResolve conflicts in a resident agent session with /git-conflict-resolve, then run knot sync again.")
			return fmt.Errorf("%s", b.String())
		}
		return fmt.Errorf("knot: git pull --rebase: %w", err)
	}

	if err := runGit(kbRoot, "push"); err != nil {
		return fmt.Errorf("knot: git push: %w", err)
	}

	fmt.Fprintln(os.Stdout, "Synced with remote.")
	return nil
}
