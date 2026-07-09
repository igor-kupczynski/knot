package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func gitMissingHint() error {
	return fmt.Errorf(`knot: git not found on PATH

Install git: brew install git`)
}

func lookPathGit() (string, error) {
	path, err := exec.LookPath("git")
	if err != nil {
		return "", gitMissingHint()
	}
	return path, nil
}

// safeGitArg rejects empty values and values that look like git options.
func safeGitArg(name, v string) error {
	if v == "" || strings.HasPrefix(v, "-") {
		return fmt.Errorf("knot: invalid %s %q", name, v)
	}
	return nil
}

// gitSafetyArgs are prepended to every git invocation to neutralize
// untrusted repo config and exotic protocols.
var gitSafetyArgs = []string{
	"-c", "core.fsmonitor=",
	"-c", "core.hooksPath=/dev/null",
	"-c", "protocol.ext.allow=never",
}

// gitDangerousEnv prefixes stripped from the child environment so a
// hostile ambient env cannot redirect git state or run helpers.
var gitDangerousEnv = []string{
	"GIT_DIR=",
	"GIT_WORK_TREE=",
	"GIT_OBJECT_DIRECTORY=",
	"GIT_INDEX_FILE=",
	"GIT_CONFIG=",
	"GIT_CONFIG_GLOBAL=",
	"GIT_CONFIG_SYSTEM=",
	"GIT_CONFIG_NOSYSTEM=",
	"GIT_CONFIG_COUNT=",
	"GIT_SSH_COMMAND=",
	"GIT_SSH=",
	"GIT_EXTERNAL_DIFF=",
	"GIT_DIFF_OPTS=",
	"GIT_SEQUENCE_EDITOR=",
	"GIT_EDITOR=",
	"GIT_PAGER=",
	"GIT_TRACE=",
	"GIT_TRACE2=",
	"GIT_EXEC_PATH=",
	"GIT_ALTERNATE_OBJECT_DIRECTORIES=",
	"GIT_TEMPLATE_DIR=",
	"EDITOR=",
	"VISUAL=",
	"PAGER=",
}

func scrubGitEnv() []string {
	env := os.Environ()
	out := make([]string, 0, len(env))
	for _, e := range env {
		drop := false
		for _, p := range gitDangerousEnv {
			if strings.HasPrefix(e, p) {
				drop = true
				break
			}
		}
		// Also drop GIT_CONFIG_KEY_* / GIT_CONFIG_VALUE_* pair vars.
		if strings.HasPrefix(e, "GIT_CONFIG_KEY_") || strings.HasPrefix(e, "GIT_CONFIG_VALUE_") {
			drop = true
		}
		if !drop {
			out = append(out, e)
		}
	}
	return out
}

func runGit(kbRoot string, args ...string) error {
	_, _, err := runGitCapture(kbRoot, args...)
	return err
}

func runGitCapture(kbRoot string, args ...string) (stdout, stderr string, err error) {
	git, err := lookPathGit()
	if err != nil {
		return "", "", err
	}

	full := make([]string, 0, len(gitSafetyArgs)+len(args))
	full = append(full, gitSafetyArgs...)
	full = append(full, args...)

	cmd := exec.Command(git, full...)
	cmd.Dir = kbRoot
	cmd.Env = scrubGitEnv()
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return outBuf.String(), errBuf.String(), fmt.Errorf("%s", strings.TrimSpace(errBuf.String()))
		}
		return outBuf.String(), errBuf.String(), fmt.Errorf("knot: run git %s: %w", strings.Join(args, " "), err)
	}
	return outBuf.String(), errBuf.String(), nil
}

func isGitRepo(kbRoot string) bool {
	gitDir := filepath.Join(kbRoot, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

func isMidRebaseOrMerge(kbRoot string) bool {
	gitDir := filepath.Join(kbRoot, ".git")
	markers := []string{"rebase-merge", "rebase-apply", "MERGE_HEAD"}
	for _, name := range markers {
		if _, err := os.Stat(filepath.Join(gitDir, name)); err == nil {
			return true
		}
	}
	return false
}

// firstGitRemote returns the name of the first configured remote, or "" if none.
func firstGitRemote(kbRoot string) (string, error) {
	stdout, _, err := runGitCapture(kbRoot, "remote")
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(stdout, "\n") {
		if name := strings.TrimSpace(line); name != "" {
			return name, nil
		}
	}
	return "", nil
}

func hasUpstream(kbRoot string) bool {
	_, _, err := runGitCapture(kbRoot, "rev-parse", "--abbrev-ref", "@{upstream}")
	return err == nil
}

func currentBranch(kbRoot string) (string, error) {
	stdout, _, err := runGitCapture(kbRoot, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("knot: resolve current branch: %w", err)
	}
	return strings.TrimSpace(stdout), nil
}

func gitIdentityConfigured(kbRoot string) bool {
	stdout, _, err := runGitCapture(kbRoot, "config", "user.email")
	return err == nil && strings.TrimSpace(stdout) != ""
}

// gitCommit commits with a fallback identity when none is configured,
// never overriding an existing user.name/user.email.
func gitCommit(kbRoot, msg string) error {
	var args []string
	if !gitIdentityConfigured(kbRoot) {
		args = append(args, "-c", "user.name=knot", "-c", "user.email=knot@localhost")
	}
	args = append(args, "commit", "-m", msg)
	return runGit(kbRoot, args...)
}

func isGitDirty(kbRoot string) (bool, error) {
	stdout, _, err := runGitCapture(kbRoot, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(stdout) != "", nil
}

func conflictingPaths(kbRoot string) ([]string, error) {
	stdout, _, err := runGitCapture(kbRoot, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}
	return paths, nil
}
