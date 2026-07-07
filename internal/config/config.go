package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const configRelPath = ".config/knot/config"

// KBRoot returns the validated knowledge base root from KNOT_KB or ~/.config/knot/config.
func KBRoot() (string, error) {
	if v := os.Getenv("KNOT_KB"); v != "" {
		return resolveKBPath(strings.TrimSpace(v))
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("knot: cannot resolve home directory: %w", err)
	}

	configPath := filepath.Join(home, configRelPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", missingConfigError()
		}
		return "", fmt.Errorf("knot: read %s: %w", configPath, err)
	}

	firstLine := strings.TrimSpace(strings.Split(string(data), "\n")[0])
	if firstLine == "" {
		return "", missingConfigError()
	}

	return resolveKBPath(firstLine)
}

func missingConfigError() error {
	return fmt.Errorf(`knot: no knowledge base configured

Set KNOT_KB to your KB path, or run knot init <path> to create one and write ~/.config/knot/config`)
}

func resolveKBPath(path string) (string, error) {
	if path == "" {
		return "", missingConfigError()
	}

	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("knot: cannot expand ~ in KB path: %w", err)
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, path[2:])
		}
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("knot: resolve KB path %q: %w", path, err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("knot: knowledge base %q does not exist", abs)
		}
		return "", fmt.Errorf("knot: stat knowledge base %q: %w", abs, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("knot: knowledge base %q is not a directory", abs)
	}

	return abs, nil
}

// ConfigPath returns the path to the knot config file (~/.config/knot/config).
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("knot: cannot resolve home directory: %w", err)
	}
	return filepath.Join(home, configRelPath), nil
}

// WriteConfig writes the KB path to ~/.config/knot/config if the file does not exist.
func WriteConfig(kbPath string) error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("knot: stat %s: %w", configPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("knot: create config dir: %w", err)
	}

	content := strings.TrimSpace(kbPath) + "\n"
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("knot: write %s: %w", configPath, err)
	}

	return nil
}
