package kb

import (
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
