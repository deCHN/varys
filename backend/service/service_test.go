package service

import (
	"os"
	"path/filepath"
	"testing"
	"Varys/backend/dependency"
)

func TestNewCoreService(t *testing.T) {
	dm := &dependency.Manager{}
	svc := NewCoreService(dm)
	if svc.depManager != dm {
		t.Error("NewCoreService did not correctly set depManager")
	}
}

func TestCopyFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "copy_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	src := filepath.Join(tempDir, "src.txt")
	dst := filepath.Join(tempDir, "dst.txt")
	content := "test content"

	if err := os.WriteFile(src, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Errorf("Expected content %s, got %s", content, string(got))
	}
}
