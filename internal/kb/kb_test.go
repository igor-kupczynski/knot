package kb

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Deploys", "deploys"},
		{"my-page", "my-page"},
		{"My Page", "my-page"},
		{"My-Page", "my-page"},
	}
	for _, tc := range tests {
		if got := NormalizeName(tc.in); got != tc.want {
			t.Errorf("NormalizeName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestWrapGlobPattern(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"postgres", "*postgres*"},
		{"pages/*.md", "pages/*.md"},
		{"inbox?", "inbox?"},
		{"", ""},
		{"!foo", "!foo"},
	}
	for _, tc := range tests {
		if got := WrapGlobPattern(tc.in); got != tc.want {
			t.Errorf("WrapGlobPattern(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSlugFromText(t *testing.T) {
	if got := SlugFromText("CI only deploys from main"); got != "ci-only-deploys-from" {
		t.Errorf("SlugFromText = %q", got)
	}
	if got := SlugFromText(""); got != "note" {
		t.Errorf("empty SlugFromText = %q", got)
	}
	longText := strings.Repeat("word ", 2000000)
	if got := SlugFromText(longText); len(got) > maxSlugLen {
		t.Errorf("SlugFromText on huge input = len %d, want <= %d", len(got), maxSlugLen)
	}
}

func TestSlugFromPage(t *testing.T) {
	if got := SlugFromPage("Deploy Pipeline"); got != "deploy-pipeline" {
		t.Errorf("SlugFromPage = %q", got)
	}
}

func TestIsPathEscape(t *testing.T) {
	if isPathEscape("..notes.md") {
		t.Fatal("..notes.md should not escape")
	}
	if !isPathEscape("..") {
		t.Fatal(".. should escape")
	}
	if !isPathEscape("../etc/hosts") {
		t.Fatal("../etc/hosts should escape")
	}
}

func TestPageNamesMatch(t *testing.T) {
	if !pageNamesMatch("mixedCase", "mixedcase") {
		t.Fatal("mixedCase should match mixedcase")
	}
	if !pageNamesMatch("Deploy Pipeline", "deploy-pipeline") {
		t.Fatal("Deploy Pipeline should match deploy-pipeline")
	}
}

func TestResolvePathRejectsOutsideSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(secret, []byte("secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "leak.md")
	if err := os.Symlink(secret, link); err != nil {
		t.Fatal(err)
	}

	if _, ok := ResolvePath(root, "leak.md"); ok {
		t.Fatal("symlink escaping KB should not resolve")
	}
}

func TestResolvePathAllowsInsideSymlink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "pages", "real.md")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "alias.md")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	got, ok := ResolvePath(root, "alias.md")
	if !ok {
		t.Fatal("in-KB symlink should resolve")
	}
	if got != link {
		// Return value is the logical path under root; either is fine if under root.
		if _, err := os.Stat(got); err != nil {
			t.Fatalf("resolved path unreadable: %v", err)
		}
	}
}

func TestResolvePathRegularFile(t *testing.T) {
	root := t.TempDir()
	f := filepath.Join(root, "index.md")
	if err := os.WriteFile(f, []byte("# hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := ResolvePath(root, "index.md")
	if !ok || got != f {
		t.Fatalf("ResolvePath = %q, %v; want %q, true", got, ok, f)
	}
}

func TestResolvePageRejectsOutsideSymlink(t *testing.T) {
	root := t.TempDir()
	pages := filepath.Join(root, "pages")
	if err := os.MkdirAll(pages, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.md")
	if err := os.WriteFile(secret, []byte("secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secret, filepath.Join(pages, "Topic.md")); err != nil {
		t.Fatal(err)
	}

	if _, ok, err := ResolvePage(root, "Topic"); err != nil {
		t.Fatal(err)
	} else if ok {
		t.Fatal("page symlink escaping KB should not resolve")
	}
}
