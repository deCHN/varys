package dependency

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Manager handles external dependencies
type Manager struct{}

// NewManager creates a new dependency manager
func NewManager() (*Manager, error) {
	// Professional Pragmatic Fix: GUI apps on macOS don't inherit the shell PATH.
	// We manually inject common Homebrew and user paths to ensure dependencies
	// like 'deno' (for yt-dlp challenges) or 'ffmpeg' are found.
	path := os.Getenv("PATH")
	additionalPaths := []string{
		"/opt/homebrew/bin",
		"/usr/local/bin",
		filepath.Join(os.Getenv("HOME"), ".local/bin"),
	}

	for _, p := range additionalPaths {
		if !strings.Contains(path, p) {
			path = p + string(os.PathListSeparator) + path
		}
	}
	os.Setenv("PATH", path)

	return &Manager{}, nil
}

// EnsureBinaries is a no-op as we rely on system binaries
func (m *Manager) EnsureBinaries() error {
	return nil
}

// GetBinaryPath returns the system path to a dependency.

// Returns empty string if not found.

func (m *Manager) GetBinaryPath(name string) string {

	if path, found := m.CheckSystemDependency(name); found {

		return path

	}

	return ""

}

// CheckSystemDependency looks for a binary in the system PATH and common Homebrew locations.

// Returns the path and true if found, empty string and false otherwise.

func (m *Manager) CheckSystemDependency(name string) (string, bool) {

	// 1. Check system PATH

	if path, err := exec.LookPath(name); err == nil {

		return path, true

	}

	// 2. Check common Homebrew locations (macOS)

	commonPaths := []string{

		filepath.Join("/opt/homebrew/bin", name),

		filepath.Join("/usr/local/bin", name),
	}

	for _, p := range commonPaths {

		if _, err := os.Stat(p); err == nil {

			return p, true

		}

	}

	return "", false

}
