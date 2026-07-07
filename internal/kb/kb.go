package kb

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const maxSlugLen = 50

// NormalizeName lowercases and treats spaces and dashes as equivalent.
func NormalizeName(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r == ' ' || r == '-' {
			b.WriteRune('-')
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isPathEscape(clean string) bool {
	if clean == ".." {
		return true
	}
	sep := string(filepath.Separator)
	return strings.HasPrefix(clean, ".."+sep)
}

// ResolvePath returns the absolute path for a KB-relative regular file if it exists.
func ResolvePath(kbRoot, relPath string) (string, bool) {
	if filepath.IsAbs(relPath) {
		return "", false
	}
	clean := filepath.Clean(relPath)
	if clean == "." || isPathEscape(clean) {
		return "", false
	}
	abs := filepath.Join(kbRoot, clean)
	info, err := os.Stat(abs)
	if err != nil || !info.Mode().IsRegular() {
		return "", false
	}
	return abs, true
}

func pageStem(name string) (string, bool) {
	if !strings.EqualFold(filepath.Ext(name), ".md") {
		return "", false
	}
	ext := filepath.Ext(name)
	return name[:len(name)-len(ext)], true
}

func pageNamesMatch(stem, query string) bool {
	if strings.EqualFold(stem, query) {
		return true
	}
	return NormalizeName(stem) == NormalizeName(query)
}

// ResolvePage finds a page by case-insensitive name with space/dash normalization.
func ResolvePage(kbRoot, name string) (string, bool, error) {
	pagesDir := filepath.Join(kbRoot, "pages")
	entries, err := os.ReadDir(pagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("knot: read pages: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		stem, ok := pageStem(entry.Name())
		if !ok {
			continue
		}
		if pageNamesMatch(stem, name) {
			return filepath.Join(pagesDir, entry.Name()), true, nil
		}
	}
	return "", false, nil
}

// PageCandidates returns page names containing the query as a substring (normalized).
func PageCandidates(kbRoot, query string) ([]string, error) {
	pagesDir := filepath.Join(kbRoot, "pages")
	entries, err := os.ReadDir(pagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("knot: read pages: %w", err)
	}

	q := NormalizeName(query)
	if q == "" {
		return nil, nil
	}

	var matches []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		stem, ok := pageStem(entry.Name())
		if !ok {
			continue
		}
		if strings.Contains(NormalizeName(stem), q) {
			matches = append(matches, stem)
		}
	}
	return matches, nil
}

// SlugFromText derives a filename slug from capture text (first few words).
func SlugFromText(text string) string {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return "note"
	}
	if len(fields) > 4 {
		fields = fields[:4]
	}
	return slugify(strings.Join(fields, " "))
}

// SlugFromPage derives a filename slug from a page hint.
func SlugFromPage(page string) string {
	page = strings.TrimSpace(page)
	if page == "" {
		return "note"
	}
	return slugify(page)
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '-' || r == '_':
			if !lastDash && b.Len() > 0 {
				b.WriteRune('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "note"
	}
	if len(out) > maxSlugLen {
		out = strings.Trim(out[:maxSlugLen], "-")
		if out == "" {
			return "note"
		}
	}
	return out
}

// WrapGlobPattern wraps bare terms as *term* for partial matching.
func WrapGlobPattern(pattern string) string {
	if pattern == "" {
		return pattern
	}
	if strings.HasPrefix(pattern, "!") {
		return pattern
	}
	for _, r := range pattern {
		switch r {
		case '*', '?', '[', '{':
			return pattern
		}
	}
	return "*" + pattern + "*"
}
