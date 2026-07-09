package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
		// Expand clustered short flags: -Li → -L, -i
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && len(arg) > 2 {
			for _, c := range arg[1:] {
				short := "-" + string(c)
				if msg, denied := deniedRgFlags[short]; denied {
					return fmt.Errorf("knot: rejected ripgrep flag %s (%s)", short, msg)
				}
			}
		}
	}
	return nil
}

// validateRgPaths rejects absolute paths and path escapes in path positionals.
// The first non-flag argument is the pattern (unless -e/--regexp was used) and
// is not treated as a path. Everything after "--" is a path.
func validateRgPaths(args []string) error {
	afterDash := false
	seenPattern := false
	explicitRegexp := false
	for _, a := range args {
		if a == "-e" || a == "--regexp" || strings.HasPrefix(a, "--regexp=") || strings.HasPrefix(a, "-e=") {
			explicitRegexp = true
			break
		}
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !afterDash {
			if arg == "--" {
				afterDash = true
				continue
			}
			if strings.HasPrefix(arg, "-") {
				if !strings.Contains(arg, "=") && rgFlagTakesValue(arg) && i+1 < len(args) {
					i++
				}
				continue
			}
			if !explicitRegexp && !seenPattern {
				seenPattern = true
				continue
			}
		}
		if err := checkSearchPath(arg); err != nil {
			return err
		}
	}
	return nil
}

func rgFlagTakesValue(flag string) bool {
	// Clustered shorts that take a value are unusual; treat the whole token.
	if strings.HasPrefix(flag, "--") {
		switch flag {
		case "--regexp", "--file", "--files-from", "--type", "--type-not",
			"--glob", "--iglob", "--ignore-file", "--max-filesize",
			"--max-columns", "--max-count", "--max-depth", "--threads",
			"--context", "--after-context", "--before-context",
			"--colors", "--heading", "--field-context-separator",
			"--field-match-separator", "--replace", "--sort", "--sortr",
			"--engine", "--type-add", "--type-clear", "--pre", "--pre-glob",
			"--hostname-bin", "--encoding", "--dfa-size-limit",
			"--regex-size-limit", "--stop-on-nonmatch":
			return true
		default:
			return false
		}
	}
	// Short flags that take a value: -e -f -t -g -A -B -C -m -j -r
	if strings.HasPrefix(flag, "-") && !strings.HasPrefix(flag, "--") && len(flag) == 2 {
		switch flag[1] {
		case 'e', 'f', 't', 'g', 'A', 'B', 'C', 'm', 'j', 'r':
			return true
		}
	}
	return false
}

func checkSearchPath(p string) error {
	if p == "" {
		return fmt.Errorf("knot: invalid search path %q", p)
	}
	if filepath.IsAbs(p) {
		return fmt.Errorf("knot: search path must be relative to the knowledge base: %q", p)
	}
	clean := filepath.Clean(p)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("knot: search path escapes knowledge base: %q", p)
	}
	return nil
}

func buildSearchArgs(userArgs []string) ([]string, error) {
	if err := checkDeniedRgFlags(userArgs); err != nil {
		return nil, err
	}
	if err := validateRgPaths(userArgs); err != nil {
		return nil, err
	}
	out := []string{"--smart-case", "--no-config"}
	out = append(out, userArgs...)
	return out, nil
}

func scrubRgEnv() []string {
	env := os.Environ()
	out := make([]string, 0, len(env))
	for _, e := range env {
		if strings.HasPrefix(e, "RIPGREP_") {
			continue
		}
		out = append(out, e)
	}
	return out
}

func runRg(kbRoot string, args ...string) int {
	rg, err := lookPathRg()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	cmd := exec.Command(rg, args...)
	cmd.Dir = kbRoot
	cmd.Env = scrubRgEnv()
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
