package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var deniedRgFlags = map[string]string{
	"--pre":          "program execution (--pre)",
	"--pre-glob":     "program execution (--pre-glob)",
	"--hostname-bin": "program execution (--hostname-bin)",
	"--follow":       "symlink escape (--follow)",
	"-L":             "symlink escape (-L)",
}

func rgMissingHint() error {
	return fmt.Errorf(`knot: ripgrep (rg) not found on PATH

Install ripgrep: brew install ripgrep`)
}

func lookPathRg() (string, error) {
	path, err := exec.LookPath("rg")
	if err != nil {
		return "", rgMissingHint()
	}
	return path, nil
}

func checkDeniedRgFlags(args []string) error {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		if name, value, ok := strings.Cut(arg, "="); ok {
			if msg, denied := deniedRgFlags[name]; denied {
				return fmt.Errorf("knot: rejected ripgrep flag %s (%s)", name, msg)
			}
			_ = value
			continue
		}
		if msg, denied := deniedRgFlags[arg]; denied {
			return fmt.Errorf("knot: rejected ripgrep flag %s (%s)", arg, msg)
		}
	}
	return nil
}

func buildSearchArgs(userArgs []string) ([]string, error) {
	if err := checkDeniedRgFlags(userArgs); err != nil {
		return nil, err
	}
	out := []string{"--smart-case"}
	out = append(out, userArgs...)
	return out, nil
}

func runRg(kbRoot string, args ...string) int {
	rg, err := lookPathRg()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	cmd := exec.Command(rg, args...)
	cmd.Dir = kbRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			if code < 0 {
				return 2
			}
			return code
		}
		fmt.Fprintf(os.Stderr, "knot: run rg: %v\n", err)
		return 2
	}
	return 0
}
