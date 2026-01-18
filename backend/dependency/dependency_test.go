package dependency

import (
	"testing"
)

func TestDependency(t *testing.T) {
	mgr, _ := NewManager()

	// EnsureBinaries should be no-op and pass
	if err := mgr.EnsureBinaries(); err != nil {
		t.Errorf("EnsureBinaries failed: %v", err)
	}

	// Test CheckSystemDependency with a common binary
	path, found := mgr.CheckSystemDependency("ls") // ls is standard
	if !found {
		t.Log("Warning: 'ls' binary not found in PATH (unlikely on unix)")
	} else {
		if path == "" {
			t.Error("Found 'ls' but path is empty")
		}
	}

	// Test GetBinaryPath
	path2 := mgr.GetBinaryPath("ls")
	if path2 != path {
		t.Errorf("GetBinaryPath mismatch: %s vs %s", path2, path)
	}
}