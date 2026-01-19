package downloader

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"Varys/backend/dependency"
)

type Downloader struct {
	dep *dependency.Manager
}

func NewDownloader(dep *dependency.Manager) *Downloader {
	return &Downloader{dep: dep}
}

// GetVideoTitle fetches the title of the video
func (d *Downloader) GetVideoTitle(url string) (string, error) {
	ytPath := d.dep.GetBinaryPath("yt-dlp")
	if ytPath == "" {
		return "", fmt.Errorf("yt-dlp not found")
	}

	cmd := exec.Command(ytPath, "--get-title", "--cookies-from-browser", "chrome", url)
	// Output() returns standard output only.
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("failed to get title: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to get title: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// DownloadAudio downloads the audio from the given URL to the output directory.
// Returns the absolute path to the downloaded file.
func (d *Downloader) DownloadAudio(url string, outputDir string, onProgress func(string)) (string, error) {
	ytPath := d.dep.GetBinaryPath("yt-dlp")
	if ytPath == "" {
		return "", fmt.Errorf("yt-dlp binary not found")
	}

	// Output template: outputDir/audio.%(ext)s
	tempBase := "temp_audio"
	outputTemplate := filepath.Join(outputDir, tempBase+".%(ext)s")

	args := []string{
		"-x", "--audio-format", "m4a",
		"--cookies-from-browser", "chrome",
		"--no-playlist", "--newline", // --newline helps with progress parsing
		"-o", outputTemplate,
		url,
	}

	cmd := exec.Command(ytPath, args...)
	
	// Capture stdout for progress
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	
	// Merge stderr into stdout usually helps catch errors in stream, but for progress just stdout
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start download: %w", err)
	}

	// Read output line by line
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if onProgress != nil {
			// Filter noise? or just send raw
			// yt-dlp with --newline sends [download] ...
			if strings.Contains(line, "[download]") {
				onProgress(line)
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("download command failed: %w", err)
	}

	// Verify file exists. Expected to be .m4a
	expectedFile := filepath.Join(outputDir, tempBase+".m4a")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		return "", fmt.Errorf("download success but file not found at %s", expectedFile)
	}

	return expectedFile, nil
}
