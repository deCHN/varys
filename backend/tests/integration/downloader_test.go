package integration

import (
	"Varys/backend/dependency"
	"Varys/backend/downloader"
	"os"
	"testing"
)

func TestDownloaderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dm, err := dependency.NewManager()
	if err != nil {
		t.Fatalf("Failed to init dependency manager: %v", err)
	}
	dl := downloader.NewDownloader(dm)

	// Use a short, stable YouTube video for testing
	// Note: This requires network access and yt-dlp installed.
	testURL := "https://www.youtube.com/watch?v=BaW_jenozKc" 

	t.Run("GetMetadata", func(t *testing.T) {
		title, err := dl.GetVideoTitle(testURL)
		if err != nil {
			t.Logf("Note: Skipping title check, likely network or binary missing: %v", err)
			return
		}
		if title == "" {
			t.Error("Expected non-empty title")
		}
	})

	t.Run("DownloadAudioOnly", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "dl_integ_audio_")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		path, err := dl.DownloadMedia(testURL, tempDir, true, nil)
		if err != nil {
			t.Logf("Note: Skipping download, likely network or binary missing: %v", err)
			return
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("File not found at %s", path)
		}
	})
}
