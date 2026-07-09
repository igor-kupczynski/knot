package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeGitArg(t *testing.T) {
	if err := safeGitArg("x", "origin"); err != nil {
		t.Fatalf("origin: %v", err)
	}
	if err := safeGitArg("x", "main"); err != nil {
		t.Fatalf("main: %v", err)
	}
	if err := safeGitArg("x", ""); err == nil {
		t.Fatal("empty should fail")
	}
	if err := safeGitArg("x", "-evil"); err == nil {
		t.Fatal("dash-prefixed should fail")
	}
	if err := safeGitArg("x", "--config=core.fsmonitor=touch /tmp/pwned"); err == nil {
		t.Fatal("long option should fail")
	}
}

func TestValidateCloneURL(t *testing.T) {
	ok := []string{
		"https://github.com/user/repo.git",
		"http://example.com/repo.git",
		"git@github.com:user/repo.git",
		"ssh://git@github.com/user/repo.git",
		"file:///tmp/repo.git",
		"/tmp/local-bare",
		"./relative-repo",
	}
	for _, u := range ok {
		if err := validateCloneURL(u); err != nil {
			t.Errorf("validateCloneURL(%q): %v", u, err)
		}
	}

	bad := []string{
		"",
		"--config=core.fsmonitor=touch /tmp/pwned",
		"-u",
		"ext::sh -c id",
		"fd::3",
	}
	for _, u := range bad {
		if err := validateCloneURL(u); err == nil {
			t.Errorf("validateCloneURL(%q) should fail", u)
		}
	}
}

func TestInitFromRejectsDashConfig(t *testing.T) {
	requireGit(t)

	home := t.TempDir()
	t.Setenv("HOME", home)

	kbPath := filepath.Join(t.TempDir(), "kb")
	// Marker that must not be created by a successful fsmonitor side effect.
	marker := filepath.Join(t.TempDir(), "pwned")

	evil := "--config=core.fsmonitor=touch " + marker
	err := runInitFrom(t, kbPath, evil)
	if err == nil {
		t.Fatal("expected rejection of dash-prefixed --from")
	}
	if _, statErr := os.Stat(marker); statErr == nil {
		t.Fatal("fsmonitor side effect ran; arg injection not blocked")
	}
	if _, statErr := os.Stat(kbPath); statErr == nil {
		// Clone must not have succeeded into kbPath.
		entries, _ := os.ReadDir(kbPath)
		if len(entries) > 0 {
			t.Fatalf("kb path populated after rejected clone: %v", entries)
		}
	}
}

func TestInitFromRejectsExtScheme(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	kbPath := filepath.Join(t.TempDir(), "kb")
	err := runInitFrom(t, kbPath, "ext::sh -c 'echo pwned'")
	if err == nil {
		t.Fatal("expected rejection of ext:: scheme")
	}
	if !strings.Contains(err.Error(), "unsupported") && !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckDeniedRgFlagsClustered(t *testing.T) {
	if err := checkDeniedRgFlags([]string{"-L"}); err == nil {
		t.Fatal("-L should be denied")
	}
	if err := checkDeniedRgFlags([]string{"-Li"}); err == nil {
		t.Fatal("-Li should be denied via clustered expansion")
	}
	if err := checkDeniedRgFlags([]string{"--follow"}); err == nil {
		t.Fatal("--follow should be denied")
	}
	if err := checkDeniedRgFlags([]string{"-i", "pattern"}); err != nil {
		t.Fatalf("-i should be allowed: %v", err)
	}
	if err := checkDeniedRgFlags([]string{"-p", "pattern"}); err != nil {
		t.Fatalf("-p (--pretty) should be allowed: %v", err)
	}
}

func TestValidateRgPaths(t *testing.T) {
	if err := validateRgPaths([]string{"pattern"}); err != nil {
		t.Fatalf("pattern only: %v", err)
	}
	if err := validateRgPaths([]string{"pattern", "pages"}); err != nil {
		t.Fatalf("relative path: %v", err)
	}
	if err := validateRgPaths([]string{"pattern", "../etc"}); err == nil {
		t.Fatal(".. path should fail")
	}
	if err := validateRgPaths([]string{"pattern", "/etc/passwd"}); err == nil {
		t.Fatal("absolute path should fail")
	}
	// Pattern that looks absolute is still a pattern, not a path.
	if err := validateRgPaths([]string{"/regex/"}); err != nil {
		t.Fatalf("regex pattern should be allowed: %v", err)
	}
	if err := validateRgPaths([]string{"-e", "pat", "/etc/passwd"}); err == nil {
		t.Fatal("abs path with -e should fail")
	}
	if err := validateRgPaths([]string{"pat", "--", "../x"}); err == nil {
		t.Fatal("path after -- with .. should fail")
	}
}

func TestBuildSearchArgsNoConfig(t *testing.T) {
	args, err := buildSearchArgs([]string{"foo"})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, a := range args {
		if a == "--no-config" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected --no-config in %v", args)
	}
}

func TestScrubGitEnv(t *testing.T) {
	t.Setenv("GIT_DIR", "/tmp/evil")
	t.Setenv("GIT_SSH_COMMAND", "touch /tmp/pwned")
	t.Setenv("PATH", os.Getenv("PATH"))
	env := scrubGitEnv()
	for _, e := range env {
		if strings.HasPrefix(e, "GIT_DIR=") {
			t.Fatal("GIT_DIR should be scrubbed")
		}
		if strings.HasPrefix(e, "GIT_SSH_COMMAND=") {
			t.Fatal("GIT_SSH_COMMAND should be scrubbed")
		}
	}
}

func TestScrubRgEnv(t *testing.T) {
	t.Setenv("RIPGREP_CONFIG_PATH", "/tmp/evil")
	env := scrubRgEnv()
	for _, e := range env {
		if strings.HasPrefix(e, "RIPGREP_") {
			t.Fatalf("RIPGREP_* should be scrubbed: %s", e)
		}
	}
}
