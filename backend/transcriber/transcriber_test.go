package transcriber

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"Varys/backend/dependency"
)

func TestTranscribe(t *testing.T) {
	// 1. Setup temp env
	tempDir, err := os.MkdirTemp("", "transcriber_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	binDir := filepath.Join(tempDir, "bin")
	os.MkdirAll(binDir, 0755)

	// Update PATH to verify CheckSystemDependency finds our mocks
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)

	// 2. Mock binaries
	mockWhisper := filepath.Join(binDir, "whisper-cli")
	mockFfmpeg := filepath.Join(binDir, "ffmpeg")

	// Script to handle both
	script := `#!/bin/sh
NAME=$(basename "$0")
if [ "$NAME" = "ffmpeg" ]; then
    # Create the last argument (output file)
    eval LAST=\${$#}
    touch "$LAST"
elif [ "$NAME" = "whisper-cli" ]; then
    # Find input after -f
    INPUT=""
    while [[ $# -gt 0 ]]; do
        if [ "$1" == "-f" ]; then
            INPUT="$2"
            break
        fi
        shift
    done
    if [ -n "$INPUT" ]; then
        printf "Fake Transcript" > "${INPUT}.txt"
    fi
fi
`
	if err := os.WriteFile(mockWhisper, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	// Copy same script to ffmpeg
	if err := os.WriteFile(mockFfmpeg, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	// 3. Init
	depMgr := &dependency.Manager{}
	
	// Create dummy model file
	modelPath := filepath.Join(tempDir, "model.bin")
	os.WriteFile(modelPath, []byte("data"), 0644)

	tr := NewTranscriber(depMgr)

	// 4. Test
	audioPath := filepath.Join(tempDir, "test_audio.m4a")
	os.WriteFile(audioPath, []byte("audio"), 0644)

	text, err := tr.Transcribe(audioPath, modelPath, nil)
	if err != nil {
		t.Fatalf("Transcribe failed: %v", err)
	}

	if strings.TrimSpace(text) != "Fake Transcript" {
		t.Errorf("Unexpected transcript: %q", text)
	}
}
