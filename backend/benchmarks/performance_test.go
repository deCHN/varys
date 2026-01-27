package benchmarks

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
	"os"

	"Varys/backend/analyzer"
	"Varys/backend/dependency"
	"Varys/backend/transcriber"
)

func TestPerformanceBaseline(t *testing.T) {
	// Setup
	depMgr, err := dependency.NewManager()
	if err != nil {
		t.Fatalf("Failed to init dependency manager: %v", err)
	}
	
	// Ensure external tools are in PATH (since we are running tests, maybe not from .app)
	// We rely on system PATH or what NewManager sets up.
	// We might need to manually inject PATH if NewManager only does it for the App process env.
	// But NewManager sets os.Setenv, so it should affect this process.

	tr := transcriber.NewTranscriber(depMgr)
	// Use default model path (adjust if needed)
	home, _ := os.UserHomeDir()
	modelPath := filepath.Join(home, "Library/Application Support/Varys/ggml-large-v3-turbo.bin")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		// Fallback to local repo res/ if available
		wd, _ := os.Getwd()
		modelPath = filepath.Join(wd, "../../res/ggml-large-v3-turbo.bin")
	}
	
	// Test Cases
	testFiles := []string{
		"../../res/test_audio.wav",
		"../../res/testaudio_8000_20s.wav",
		"../../res/testaudio_48000_20s.wav",
	}
	
	contextSizes := []int{4096, 8192, 16384}

	fmt.Printf("| %-30s | %-10s | %-15s | %-15s |\n", "Audio File", "Context", "Transcribe(s)", "Analyze(s)")
	fmt.Println("|--------------------------------|------------|-----------------|-----------------|")

	for _, fileRel := range testFiles {
		wd, _ := os.Getwd()
		absPath := filepath.Join(wd, fileRel)
		
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			t.Logf("Skipping %s (not found)", fileRel)
			continue
		}

		for _, ctxSize := range contextSizes {
			// 1. Benchmark Transcription
			start := time.Now()
			transcript, _, err := tr.Transcribe(absPath, modelPath, nil)
			if err != nil {
				t.Errorf("Transcribe failed for %s: %v", fileRel, err)
				continue
			}
			transcribeDuration := time.Since(start).Seconds()

			// 2. Benchmark Analysis (LLM)
			// We use a dummy model? No, real model.
			an := analyzer.NewAnalyzer("qwen3:8b") // Updated to available model
			
			start = time.Now()
			_, err = an.Analyze(transcript, "Simplified Chinese", ctxSize, nil)
			if err != nil {
				t.Errorf("Analyze failed for %s (ctx %d): %v", fileRel, ctxSize, err)
				continue
			}
			analyzeDuration := time.Since(start).Seconds()

			fmt.Printf("| %-30s | %-10d | %-15.2f | %-15.2f |\n", filepath.Base(fileRel), ctxSize, transcribeDuration, analyzeDuration)
		}
	}
}
