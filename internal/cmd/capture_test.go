package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/igor-kupczynski/knot/internal/kb"
	"gopkg.in/yaml.v3"
)

func TestCaptureConcurrentSameSlug(t *testing.T) {
	kbRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(kbRoot, "inbox"), 0o755); err != nil {
		t.Fatal(err)
	}

	baseName := "2026-01-01-120000-same"
	bodies := []string{"concurrent body alpha", "concurrent body beta"}
	paths := make([]string, len(bodies))

	var wg sync.WaitGroup
	for i, body := range bodies {
		wg.Add(1)
		go func(i int, body string) {
			defer wg.Done()
			content := []byte("---\ncaptured: 2026-01-01T12:00:00Z\n---\n" + body + "\n")
			path, err := writeCaptureExclusive(kbRoot, baseName, content)
			if err != nil {
				t.Errorf("goroutine %d: %v", i, err)
				return
			}
			paths[i] = path
		}(i, body)
	}
	wg.Wait()

	if paths[0] == "" || paths[1] == "" {
		t.Fatal("expected both captures to succeed")
	}
	if paths[0] == paths[1] {
		t.Fatalf("expected distinct files, got same path %q", paths[0])
	}

	for i, body := range bodies {
		data, err := os.ReadFile(paths[i])
		if err != nil {
			t.Fatalf("read %q: %v", paths[i], err)
		}
		if !strings.Contains(string(data), body) {
			t.Fatalf("file %q missing body %q: %s", paths[i], body, data)
		}
	}
}

func TestCaptureYAMLEscaping(t *testing.T) {
	capturePage = "a: b"
	captureSources = []string{`- https://example.com?q="x"`}
	t.Cleanup(func() {
		capturePage = ""
		captureSources = nil
	})

	content, err := buildCaptureContent(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC), "body")
	if err != nil {
		t.Fatal(err)
	}

	front := strings.Split(string(content), "---")
	if len(front) < 3 {
		t.Fatalf("unexpected content: %s", content)
	}

	var fm captureFrontmatter
	if err := yaml.Unmarshal([]byte(front[1]), &fm); err != nil {
		t.Fatalf("invalid YAML frontmatter: %v\n%s", err, front[1])
	}
	if fm.Page != "a: b" {
		t.Fatalf("page = %q, want %q", fm.Page, "a: b")
	}
	if len(fm.Sources) != 1 || fm.Sources[0] != `- https://example.com?q="x"` {
		t.Fatalf("sources = %#v", fm.Sources)
	}
}

func TestCaptureLargeStdin(t *testing.T) {
	kbRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(kbRoot, "inbox"), 0o755); err != nil {
		t.Fatal(err)
	}

	capturePage = ""
	captureSources = nil

	huge := strings.Repeat("x", 10*1024*1024)
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		_, _ = io.Copy(pw, strings.NewReader(huge))
		_ = pw.Close()
	}()

	oldStdin := os.Stdin
	os.Stdin = pr
	t.Cleanup(func() { os.Stdin = oldStdin })

	text, err := readCaptureText(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(text) != len(huge) {
		t.Fatalf("read %d bytes, want %d", len(text), len(huge))
	}

	content, err := buildCaptureContent(time.Now(), text)
	if err != nil {
		t.Fatal(err)
	}

	slug := kb.SlugFromText(text)
	if len(slug) > 50 {
		t.Fatalf("slug len %d exceeds cap", len(slug))
	}

	path, err := writeCaptureExclusive(kbRoot, "2026-01-01-120000-"+slug, content)
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() <= int64(len(huge)) {
		t.Fatalf("capture file too small: %d", info.Size())
	}
}
