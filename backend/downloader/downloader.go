package downloader

import (
	"Varys/backend/dependency"
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
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

// GetVideoDescription fetches the description of the video
func (d *Downloader) GetVideoDescription(url string) (string, error) {
	ytPath := d.dep.GetBinaryPath("yt-dlp")
	if ytPath == "" {
		return "", fmt.Errorf("yt-dlp not found")
	}

	cmd := exec.Command(ytPath, "--get-description", "--cookies-from-browser", "chrome", url)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("failed to get description: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to get description: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// DownloadMedia downloads the media (audio/video) from the given URL to the output directory.
// Returns the absolute path to the downloaded file.
func (d *Downloader) DownloadMedia(url string, outputDir string, audioOnly bool, onProgress func(string)) (string, error) {
	ytPath := d.dep.GetBinaryPath("yt-dlp")
	if ytPath == "" {
		return "", fmt.Errorf("yt-dlp binary not found")
	}

	tempBase := "temp_media"
	var outputTemplate string
	var args []string

	if audioOnly {
		// Audio Mode: Force m4a
		outputTemplate = filepath.Join(outputDir, tempBase+".%(ext)s")
		args = []string{
			"-x", "--audio-format", "m4a",
			"--cookies-from-browser", "chrome",
			"--no-playlist", "--newline",
			"-o", outputTemplate,
			url,
		}
	} else {
		// Video Mode: Best mp4
		// We use -S "ext" to prefer mp4 container for better compatibility
		outputTemplate = filepath.Join(outputDir, tempBase+".%(ext)s")
		args = []string{
			"-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
			"--cookies-from-browser", "chrome",
			"--no-playlist", "--newline",
			"-o", outputTemplate,
			url,
		}
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

	// Verify file exists
	// For video, we might get .mp4 or .mkv depending on fallback, but we requested mp4 preference.
	// We'll search for the file pattern.
	files, err := filepath.Glob(filepath.Join(outputDir, tempBase+".*"))
	if err != nil || len(files) == 0 {
		return "", fmt.Errorf("download success but output file not found")
	}

	// Return the first match (likely the only one)
	return files[0], nil
}
