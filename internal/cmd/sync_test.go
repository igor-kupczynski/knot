package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found on PATH")
	}
}

func initTestGitRepo(t *testing.T, dir string) {
	t.Helper()
	requireGit(t)
	runGitTest(t, dir, "init")
	configGitUser(t, dir)
}

func configGitUser(t *testing.T, dir string) {
	t.Helper()
	runGitTest(t, dir, "config", "user.email", "test@example.com")
	runGitTest(t, dir, "config", "user.name", "Test User")
}

func runGitTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSyncNotARepo(t *testing.T) {
	kbRoot := t.TempDir()
	writeFile(t, filepath.Join(kbRoot, "index.md"), "# Index\n")

	err := runSync(kbRoot, "testhost", time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("expected error for non-git KB")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSyncCommitOnlyNoRemote(t *testing.T) {
	kbRoot := t.TempDir()
	initTestGitRepo(t, kbRoot)
	writeFile(t, filepath.Join(kbRoot, "index.md"), "# Index\n\nchanged\n")

	fixed := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	if err := runSync(kbRoot, "testhost", fixed); err != nil {
		t.Fatal(err)
	}

	log, _, err := runGitCapture(kbRoot, "log", "-1", "--pretty=%B")
	if err != nil {
		t.Fatal(err)
	}
	wantMsg := "sync: testhost 2026-07-07T12:00:00Z"
	if strings.TrimSpace(log) != wantMsg {
		t.Fatalf("commit message = %q, want %q", strings.TrimSpace(log), wantMsg)
	}

	out, _, err := runGitCapture(kbRoot, "remote")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "" {
		t.Fatalf("expected no remote, got %q", out)
	}
}

func TestSyncRoundTrip(t *testing.T) {
	requireGit(t)

	bare := t.TempDir()
	runGitTest(t, bare, "init", "--bare")

	seed := t.TempDir()
	initTestGitRepo(t, seed)
	runGitTest(t, seed, "remote", "add", "origin", bare)
	writeFile(t, filepath.Join(seed, "index.md"), "# Index\n\n")
	runGitTest(t, seed, "add", "index.md")
	runGitTest(t, seed, "commit", "-m", "initial")
	runGitTest(t, seed, "push", "-u", "origin", "HEAD")

	parent := t.TempDir()
	clone := filepath.Join(parent, "kb")
	runGitTest(t, parent, "clone", bare, clone)
	configGitUser(t, clone)

	writeFile(t, filepath.Join(clone, "log.md"), "# Log\n\nentry\n")
	if err := runSync(clone, "machine-a", time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	log, _, err := runGitCapture(clone, "log", "-1", "--pretty=%s")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(log) != "sync: machine-a 2026-07-07T12:00:00Z" {
		t.Fatalf("unexpected commit subject: %q", log)
	}

	out, _, err := runGitCapture(bare, "log", "-1", "--pretty=%s")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "sync: machine-a 2026-07-07T12:00:00Z" {
		t.Fatalf("bare remote missing push: %q", out)
	}
}

func TestSyncConflictAbort(t *testing.T) {
	requireGit(t)

	bare := t.TempDir()
	runGitTest(t, bare, "init", "--bare")

	seed := t.TempDir()
	initTestGitRepo(t, seed)
	runGitTest(t, seed, "remote", "add", "origin", bare)
	writeFile(t, filepath.Join(seed, "pages", "Topic.md"), "base\n")
	runGitTest(t, seed, "add", "pages/Topic.md")
	runGitTest(t, seed, "commit", "-m", "initial")
	runGitTest(t, seed, "push", "-u", "origin", "HEAD")

	parent1 := t.TempDir()
	clone1 := filepath.Join(parent1, "kb")
	runGitTest(t, parent1, "clone", bare, clone1)
	configGitUser(t, clone1)

	parent2 := t.TempDir()
	clone2 := filepath.Join(parent2, "kb")
	runGitTest(t, parent2, "clone", bare, clone2)
	configGitUser(t, clone2)

	writeFile(t, filepath.Join(clone1, "pages", "Topic.md"), "version remote\n")
	if err := runSync(clone1, "machine-a", time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("clone1 sync: %v", err)
	}

	writeFile(t, filepath.Join(clone2, "pages", "Topic.md"), "version local\n")
	runGitTest(t, clone2, "add", "pages/Topic.md")
	runGitTest(t, clone2, "commit", "-m", "local change")

	err := runSync(clone2, "machine-b", time.Date(2026, 7, 7, 14, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if !strings.Contains(err.Error(), "rebase conflict") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "pages/Topic.md") {
		t.Fatalf("error should list conflicting path: %v", err)
	}
	if !strings.Contains(err.Error(), "/git-conflict-resolve") {
		t.Fatalf("error should mention /git-conflict-resolve: %v", err)
	}

	if isMidRebaseOrMerge(clone2) {
		t.Fatal("rebase should have been aborted")
	}

	data, err := os.ReadFile(filepath.Join(clone2, "pages", "Topic.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "version local\n" {
		t.Fatalf("local content lost after abort: %q", data)
	}
}

func TestSyncIdempotentClean(t *testing.T) {
	kbRoot := t.TempDir()
	initTestGitRepo(t, kbRoot)
	writeFile(t, filepath.Join(kbRoot, "index.md"), "# Index\n")
	runGitTest(t, kbRoot, "add", "index.md")
	runGitTest(t, kbRoot, "commit", "-m", "initial")

	before, _, err := runGitCapture(kbRoot, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}

	if err := runSync(kbRoot, "host", time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	after, _, err := runGitCapture(kbRoot, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(before) != strings.TrimSpace(after) {
		t.Fatalf("HEAD changed on clean sync: %s -> %s", before, after)
	}
}

func TestSyncFirstPushSetsUpstream(t *testing.T) {
	requireGit(t)

	bare := t.TempDir()
	runGitTest(t, bare, "init", "--bare")

	kbRoot := t.TempDir()
	initTestGitRepo(t, kbRoot)
	runGitTest(t, kbRoot, "remote", "add", "origin", bare)
	writeFile(t, filepath.Join(kbRoot, "index.md"), "# Index\n")

	// First sync on a fresh repo: no upstream yet, must push -u instead of pulling.
	if err := runSync(kbRoot, "machine-1", time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	up, _, err := runGitCapture(kbRoot, "rev-parse", "--abbrev-ref", "@{upstream}")
	if err != nil {
		t.Fatalf("upstream not set after first sync: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(up), "origin/") {
		t.Fatalf("upstream = %q, want origin/<branch>", strings.TrimSpace(up))
	}

	subj, _, err := runGitCapture(bare, "log", "-1", "--pretty=%s")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(subj) != "sync: machine-1 2026-07-07T12:00:00Z" {
		t.Fatalf("bare remote missing first push: %q", subj)
	}

	// Subsequent sync round-trips normally (pull --rebase + push).
	writeFile(t, filepath.Join(kbRoot, "log.md"), "# Log\n")
	if err := runSync(kbRoot, "machine-1", time.Date(2026, 7, 7, 13, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	subj, _, err = runGitCapture(bare, "log", "-1", "--pretty=%s")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(subj) != "sync: machine-1 2026-07-07T13:00:00Z" {
		t.Fatalf("bare remote missing second push: %q", subj)
	}
}

// clearGitIdentity points git at empty config files so no user.name/user.email resolves.
func clearGitIdentity(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	for _, key := range []string{"EMAIL", "GIT_AUTHOR_NAME", "GIT_AUTHOR_EMAIL", "GIT_COMMITTER_NAME", "GIT_COMMITTER_EMAIL"} {
		if v, ok := os.LookupEnv(key); ok {
			t.Setenv(key, v) // register restore, then unset for the test
			if err := os.Unsetenv(key); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestSyncCommitWithoutGitIdentity(t *testing.T) {
	requireGit(t)
	clearGitIdentity(t)

	kbRoot := t.TempDir()
	runGitTest(t, kbRoot, "init")
	writeFile(t, filepath.Join(kbRoot, "index.md"), "# Index\n")

	if err := runSync(kbRoot, "host", time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	email, _, err := runGitCapture(kbRoot, "log", "-1", "--pretty=%ae")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(email) != "knot@localhost" {
		t.Fatalf("author email = %q, want knot@localhost", strings.TrimSpace(email))
	}
}

func TestInitCommitWithoutGitIdentity(t *testing.T) {
	requireGit(t)
	clearGitIdentity(t)

	kbPath := filepath.Join(t.TempDir(), "kb")
	if err := scaffoldKB(kbPath); err != nil {
		t.Fatal(err)
	}

	email, _, err := runGitCapture(kbPath, "log", "-1", "--pretty=%ae")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(email) != "knot@localhost" {
		t.Fatalf("author email = %q, want knot@localhost", strings.TrimSpace(email))
	}
}

func TestSyncMidRebaseRefused(t *testing.T) {
	kbRoot := t.TempDir()
	initTestGitRepo(t, kbRoot)
	writeFile(t, filepath.Join(kbRoot, "index.md"), "# Index\n")
	runGitTest(t, kbRoot, "add", "index.md")
	runGitTest(t, kbRoot, "commit", "-m", "initial")

	if err := os.MkdirAll(filepath.Join(kbRoot, ".git", "rebase-merge"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := runSync(kbRoot, "host", time.Now())
	if err == nil {
		t.Fatal("expected error for mid-rebase")
	}
	if !strings.Contains(err.Error(), "mid-rebase or mid-merge") {
		t.Fatalf("unexpected error: %v", err)
	}
}
