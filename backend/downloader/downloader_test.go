package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"Varys/backend/dependency"
)

func TestDownloadAudio(t *testing.T) {
	// 1. Setup temp env
	tempDir, err := os.MkdirTemp("", "v2k_dl_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 2. Mock yt-dlp binary
	binDir := filepath.Join(tempDir, "bin")
	os.MkdirAll(binDir, 0755)
	
	mockYtPath := filepath.Join(binDir, "yt-dlp")
	
	// Create a script that acts as yt-dlp
	// It needs to parse the -o argument or just create the file in the output dir (which is passed as part of -o)
	// The Downloader passes: -o outputDir/temp_audio.%(ext)s
	// We just need to create outputDir/temp_audio.m4a
	
	var scriptContent string
	if runtime.GOOS == "windows" {
		// Batch script for windows? (Not targeting windows yet)
		t.Skip("Windows mock not implemented")
	} else {
		// Shell script
		// We can't easily parse args in a one-liner without logic.
		// Instead, we can make the script create a file in a known location relative to current dir,
		// OR we rely on the test passing the outputDir which we know.
		// Wait, the test calls DownloadAudio(url, outputDir).
		// The script runs in... arbitrary pwd.
		
		// Let's make the script simple:
		// It receives args. We can verify args if we want.
		// Crucially, it must create the file.
		// We can't easily "know" the outputDir inside the script unless we parse "$7" (the -o arg value).
		// Let's try to extract the -o argument value.
		// Arg structure: ... -o outputTemplate ...
		// We can cheat: The test will use `tempDir` as outputDir too.
		// So we create the file at `tempDir/temp_audio.m4a`.
		
		targetFile := filepath.Join(tempDir, "temp_audio.m4a")
		scriptContent = fmt.Sprintf("#!/bin/sh\ntouch \"%s\"\necho 'Mock Download Finished'", targetFile)
	}

	if err := os.WriteFile(mockYtPath, []byte(scriptContent), 0755); err != nil {
		t.Fatal(err)
	}

	// Update PATH for CheckSystemDependency
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)

	// 3. Initialize Downloader
	depMgr := &dependency.Manager{}
	dl := NewDownloader(depMgr)

	// 4. Run Download
	// We use tempDir as outputDir
	resultPath, err := dl.DownloadAudio("http://fake.url", tempDir, nil)
	if err != nil {
		t.Fatalf("DownloadAudio failed: %v", err)
	}

	// 5. Verify
	expected := filepath.Join(tempDir, "temp_audio.m4a")
	if resultPath != expected {
		t.Errorf("Expected path %s, got %s", expected, resultPath)
	}
	
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}
