package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runInitFrom(t *testing.T, kbPath, url string) error {
	t.Helper()
	initFrom = url
	t.Cleanup(func() { initFrom = "" })
	rootCmd.SetArgs([]string{"init", kbPath, "--from", url})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })
	return rootCmd.Execute()
}

func seedBareKB(t *testing.T) string {
	t.Helper()
	requireGit(t)

	bare := t.TempDir()
	runGitTest(t, bare, "init", "--bare")

	seed := t.TempDir()
	initTestGitRepo(t, seed)
	runGitTest(t, seed, "remote", "add", "origin", bare)
	writeFile(t, filepath.Join(seed, "AGENTS.md"), "SEED AGENTS\n")
	writeFile(t, filepath.Join(seed, "index.md"), "# Seed Index\n")
	runGitTest(t, seed, "add", "-A")
	runGitTest(t, seed, "commit", "-m", "seed KB")
	runGitTest(t, seed, "push", "-u", "origin", "HEAD")

	return bare
}

func TestInitFromClone(t *testing.T) {
	bare := seedBareKB(t)

	home := t.TempDir()
	t.Setenv("HOME", home)

	kbPath := filepath.Join(t.TempDir(), "kb")
	if err := runInitFrom(t, kbPath, bare); err != nil {
		t.Fatal(err)
	}

	// Cloned content, not scaffold: AGENTS.md must be the seeded version.
	data, err := os.ReadFile(filepath.Join(kbPath, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "SEED AGENTS\n" {
		t.Fatalf("AGENTS.md overwritten by scaffold: %q", data)
	}

	if !isGitRepo(kbPath) {
		t.Fatal("cloned KB is not a git repository")
	}

	remotes, _, err := runGitCapture(kbPath, "remote")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(remotes, "origin") {
		t.Fatalf("expected origin remote, got %q", remotes)
	}

	configData, err := os.ReadFile(filepath.Join(home, ".config", "knot", "config"))
	if err != nil {
		t.Fatalf("config not written: %v", err)
	}
	if strings.TrimSpace(string(configData)) != kbPath {
		t.Fatalf("config = %q, want %q", strings.TrimSpace(string(configData)), kbPath)
	}
}

func TestInitFromRefusesNonEmptyTarget(t *testing.T) {
	bare := seedBareKB(t)

	home := t.TempDir()
	t.Setenv("HOME", home)

	kbPath := t.TempDir()
	writeFile(t, filepath.Join(kbPath, "existing.md"), "already here\n")

	err := runInitFrom(t, kbPath, bare)
	if err == nil {
		t.Fatal("expected error for non-empty target")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("unexpected error: %v", err)
	}

	// Target untouched.
	data, err := os.ReadFile(filepath.Join(kbPath, "existing.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "already here\n" {
		t.Fatalf("existing file modified: %q", data)
	}
}
