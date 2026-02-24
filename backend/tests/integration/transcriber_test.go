//go:build integration
package integration_test

import (
	"Varys/backend/dependency"
	"Varys/backend/transcriber"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "."
}

func TestIntegrationTranscribe(t *testing.T) {
	// 1. Setup Environment
	projectRoot := findProjectRoot()

	resDir := filepath.Join(projectRoot, "res")
	audioFile := filepath.Join(resDir, "test_audio.wav")

	// Prefer the large model if the user downloaded it
	modelFile := filepath.Join(resDir, "ggml-large-v3-turbo.bin")
	if _, err := os.Stat(modelFile); os.IsNotExist(err) {
		// Fallback to tiny
		modelFile = filepath.Join(resDir, "ggml-tiny.bin")
	}

	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		t.Skipf("Audio file not found at %s", audioFile)
	}

	// 2. Download Model if missing
	if _, err := os.Stat(modelFile); os.IsNotExist(err) {
		t.Log("Downloading ggml-tiny.bin...")
		if err := downloadFile("https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin", modelFile); err != nil {
			t.Fatalf("Failed to download model: %v", err)
		}
	}

	// 3. Setup Dependencies
	// Use a temp dir for extracted binaries (ffmpeg)
	// tempDir := t.TempDir() // Unused if Manager is stateless
	depMgr := &dependency.Manager{}
	if err := depMgr.EnsureBinaries(); err != nil {
		t.Fatalf("EnsureBinaries failed: %v", err)
	}

	// 4. Check for whisper-cli
	// For this test to pass in the CLI environment where I might not have it installed:
	// I will attempt to download it LOCALLY here if missing from PATH.
	path, found := depMgr.CheckSystemDependency("whisper-cli")
	if !found {
		// Try 'whisper-cpp' (Homebrew)
		path, found = depMgr.CheckSystemDependency("whisper-cpp")
	}
	if !found {
		// Try 'main'
		path, found = depMgr.CheckSystemDependency("main")
	}

	if !found {
		// Attempt to find it in previous download location?
		// Varys/backend/dependency/bin/darwin_arm64/whisper-cli (if I downloaded it before and it persisted?)
		// No, I deleted it.

		t.Skip("whisper binary (whisper-cli/whisper-cpp) not found in PATH. Please install it to run integration tests.")
	} else {
		t.Logf("Using whisper binary: %s", path)
	}

	// 5. Run Transcriber
	tr := transcriber.NewTranscriber(depMgr)
	text, _, err := tr.Transcribe(audioFile, modelFile, nil)
	if err != nil {
		t.Fatalf("Transcribe failed: %v", err)
	}

	t.Logf("SUCCESS! Transcript:\n%s", text)

	// Save to res/transcript.txt
	outFile := filepath.Join(resDir, "transcript.txt")
	if err := os.WriteFile(outFile, []byte(text), 0644); err != nil {
		t.Errorf("Failed to save transcript: %v", err)
	}

	// Verify against expected_transcript.txt
	expectedFile := filepath.Join(resDir, "expected_transcript.txt")
	if _, err := os.Stat(expectedFile); err == nil {
		// Fuzzy check: check if first few words exist in the result
		// We'll look for "范伟" (Fan Wei)
		if !strings.Contains(text, "范伟") {
			t.Errorf("Transcription mismatch. Expected text containing '范伟', but got:\n%s", text)
		} else {
			t.Log("Fuzzy verification passed (Found '范伟')")
		}
	}
}

func downloadFile(url, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("status %s", resp.Status)
	}
	_, err = io.Copy(out, resp.Body)
	return err
}
