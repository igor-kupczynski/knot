package cmd

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/igor-kupczynski/knot/internal/config"
	"github.com/igor-kupczynski/knot/internal/kb"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	capturePage    string
	captureSources []string
)

var captureCmd = &cobra.Command{
	Use:   "capture [text]",
	Short: "Append a timestamped note to the KB inbox",
	Run: func(cmd *cobra.Command, args []string) {
		kbRoot, err := config.KBRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		text, err := readCaptureText(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		if err := ensureInbox(kbRoot); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		now := time.Now()
		slug := kb.SlugFromPage(capturePage)
		if capturePage == "" {
			slug = kb.SlugFromText(text)
		}

		baseName := fmt.Sprintf("%s-%s", now.Format("2006-01-02-150405"), slug)
		content, err := buildCaptureContent(now, text)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		path, err := writeCaptureExclusive(kbRoot, baseName, content)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		rel, err := filepath.Rel(kbRoot, path)
		if err != nil {
			fmt.Fprintln(os.Stdout, path)
			return
		}
		fmt.Fprintln(os.Stdout, rel)
	},
}

func init() {
	captureCmd.Flags().StringVar(&capturePage, "page", "", "page hint for filing")
	captureCmd.Flags().StringArrayVar(&captureSources, "source", nil, "source URL (repeatable)")
}

type captureFrontmatter struct {
	Captured string   `yaml:"captured"`
	Page     string   `yaml:"page,omitempty"`
	Sources  []string `yaml:"sources,omitempty"`
}

func readCaptureText(args []string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("knot: stat stdin: %w", err)
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", fmt.Errorf("knot: capture requires text argument or stdin")
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("knot: read stdin: %w", err)
	}
	return strings.TrimRight(string(data), "\n"), nil
}

func sanitizeFrontmatterValue(field, value string) (string, error) {
	if strings.ContainsAny(value, "\r\n") {
		return "", fmt.Errorf("knot: %s must not contain newlines", field)
	}
	return value, nil
}

func buildCaptureContent(now time.Time, text string) ([]byte, error) {
	fm := captureFrontmatter{
		Captured: now.Format(time.RFC3339),
	}
	if capturePage != "" {
		page, err := sanitizeFrontmatterValue("page", capturePage)
		if err != nil {
			return nil, err
		}
		fm.Page = page
	}
	if len(captureSources) > 0 {
		fm.Sources = make([]string, 0, len(captureSources))
		for i, src := range captureSources {
			clean, err := sanitizeFrontmatterValue(fmt.Sprintf("source[%d]", i), src)
			if err != nil {
				return nil, err
			}
			fm.Sources = append(fm.Sources, clean)
		}
	}

	var header bytes.Buffer
	header.WriteString("---\n")
	enc := yaml.NewEncoder(&header)
	enc.SetIndent(2)
	if err := enc.Encode(fm); err != nil {
		return nil, fmt.Errorf("knot: encode frontmatter: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("knot: encode frontmatter: %w", err)
	}
	header.WriteString("---\n")

	var body []byte
	body = append(body, header.Bytes()...)
	body = append(body, text...)
	if text != "" && !strings.HasSuffix(text, "\n") {
		body = append(body, '\n')
	}
	return body, nil
}

const maxCaptureCreateAttempts = 16

func writeCaptureExclusive(kbRoot, baseName string, content []byte) (string, error) {
	inbox := filepath.Join(kbRoot, "inbox")

	candidate := filepath.Join(inbox, baseName+".md")
	path, err := createCaptureFile(candidate, content)
	if err == nil {
		return path, nil
	}
	if !errors.Is(err, os.ErrExist) {
		return "", err
	}

	for range maxCaptureCreateAttempts {
		suffix, err := randomSuffix(4)
		if err != nil {
			return "", fmt.Errorf("knot: generate suffix: %w", err)
		}
		candidate = filepath.Join(inbox, baseName+"-"+suffix+".md")
		path, err := createCaptureFile(candidate, content)
		if err == nil {
			return path, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return "", err
		}
	}
	return "", fmt.Errorf("knot: could not create unique capture file after %d attempts", maxCaptureCreateAttempts)
}

func createCaptureFile(path string, content []byte) (string, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	if _, err := f.Write(content); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("knot: write capture: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("knot: close capture: %w", err)
	}
	return path, nil
}

func randomSuffix(n int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	var b strings.Builder
	for range n {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		b.WriteByte(alphabet[idx.Int64()])
	}
	return b.String(), nil
}

func ensureInbox(kbRoot string) error {
	inbox := filepath.Join(kbRoot, "inbox")
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		return fmt.Errorf("knot: create inbox: %w", err)
	}
	return nil
}
