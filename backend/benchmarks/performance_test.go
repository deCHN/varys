package benchmarks

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"Varys/backend/analyzer"
	"Varys/backend/translation"
)

func TestPerformanceBaseline(t *testing.T) {
	// Setup

	// Test Cases
	testFiles := []string{
		"../../res/test_audio.wav",
		"../../res/testaudio_8000_20s.wav",
		"../../res/testaudio_48000_20s.wav",
	}

	contextSizes := []int{1024, 2048, 4096}

	fmt.Printf("| %-30s | %-10s | %-15s | %-15s | %-15s |\n", "Audio File", "Context", "Transcribe(s)", "Analyze(s)", "Translate(s)")
	fmt.Println("|--------------------------------|------------|-----------------|-----------------|-----------------|")

	for _, fileRel := range testFiles {
		wd, _ := os.Getwd()
		absPath := filepath.Join(wd, fileRel)

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			t.Logf("Skipping %s (not found)", fileRel)
			continue
		}

		for _, ctxSize := range contextSizes {
			// 1. Mock Transcription (Skip crashing binary in test env)
			transcribeDuration := 0.0
			transcript := "This is a dummy transcript about blockchain and finance. It needs to be long enough to test chunking. "
			for i := 0; i < 50; i++ {
				transcript += "Adding more content to simulate a real world transcript of a 20 second video clip. "
			}

			// 2. Benchmark Analysis (LLM)
			an := analyzer.NewAnalyzer("qwen3:8b")

			start := time.Now()
			_, err := an.Analyze(transcript, "", "Simplified Chinese", ctxSize, nil)
			if err != nil {
				t.Errorf("Analyze failed for %s (ctx %d): %v", fileRel, ctxSize, err)
				continue
			}
			analyzeDuration := time.Since(start).Seconds()

			// 3. Benchmark Translation
			tr := translation.NewTranslator("qwen3:0.6b")
			var translateDuration float64 = 0
			if filepath.Base(fileRel) != "test_audio.wav" {
				start = time.Now()
				_, err = tr.Translate(transcript, "Simplified Chinese", ctxSize, nil)
				if err != nil {
					t.Errorf("Translate failed for %s (ctx %d): %v", fileRel, ctxSize, err)
					continue
				}
				translateDuration = time.Since(start).Seconds()
			}

			fmt.Printf("| %-30s | %-10d | %-15.2f | %-15.2f | %-15.2f |\n", filepath.Base(fileRel), ctxSize, transcribeDuration, analyzeDuration, translateDuration)
		}
	}
}
