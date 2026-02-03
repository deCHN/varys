package transcriber

import (
	"Varys/backend/dependency"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Transcriber struct {
	dep *dependency.Manager
}

func NewTranscriber(dep *dependency.Manager) *Transcriber {
	return &Transcriber{dep: dep}
}

func (t *Transcriber) Transcribe(audioPath, modelPath string, onProgress func(string)) (string, string, error) {
	// 1. Find binary
	candidates := []string{"whisper-cli", "whisper-cpp", "whisper-main", "whisper", "main"}
	var binPath string
	for _, name := range candidates {
		if p, found := t.dep.CheckSystemDependency(name); found {
			binPath = p
			break
		}
	}

	if binPath == "" {
		return "", "", fmt.Errorf("whisper binary not found in PATH. Please install whisper.cpp")
	}

	// 2. Check model
	if modelPath == "" {
		return "", "", fmt.Errorf("whisper model path not configured")
	}
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("model file not found at %s", modelPath)
	}

	// 3. Convert to WAV (16kHz, Mono)
	wavPath := audioPath + ".wav"
	if err := t.convertToWav(audioPath, wavPath); err != nil {
		return "", "", err
	}
	defer os.Remove(wavPath)

	// 4. Run Whisper
	// Use --no-timestamps to reduce VRAM usage and prevent OOM on M-series chips for long files.
	// Use --print-progress to maintain a heartbeat in the logs.
	cmd := exec.Command(binPath, "-m", modelPath, "-f", wavPath, "--output-txt", "--no-timestamps", "--print-progress", "--language", "auto")

	// Stream output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("failed to start whisper: %w", err)
	}

	var detectedLang string
	// Regex to match "auto-detected language: zh"
	langRegex := regexp.MustCompile(`auto-detected language:\s+(\w+)`)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		// Try to capture language
		if detectedLang == "" {
			matches := langRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				detectedLang = matches[1]
			}
		}

		if onProgress != nil {
			onProgress(line)
		}
	}

	if err := cmd.Wait(); err != nil {
		return "", "", fmt.Errorf("whisper execution failed: %w", err)
	}

	resultFile := wavPath + ".txt"
	content, err := os.ReadFile(resultFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to read transcript file %s: %w", resultFile, err)
	}
	defer os.Remove(resultFile)

	cleanedContent := t.cleanTimestamps(string(content))
	cleanedContent = t.cleanHallucinations(cleanedContent)
	return cleanedContent, detectedLang, nil
}

func (t *Transcriber) cleanHallucinations(text string) string {
	// 1. Clean "Thank you." repetition (Specific Whisper hallucination)
	// This one is simple and specific, so regex is fine/fast enough usually,
	// but let's be safe and replace it first.
	reThanks := regexp.MustCompile(`(?i)(Thank you\.?(\s*)){2,}`)
	text = reThanks.ReplaceAllString(text, "Thank you.")

	// 2. Clean general repeated phrases/sentences
	// Complex regex on huge text causes hangs (ReDoS).
	// Instead, we split by common delimiters and check for adjacent duplicates.

	// Split by sentence-like boundaries (period, question mark, exclamation, newline)
	// We preserve delimiters to reconstruct somewhat accurately
	reSplit := regexp.MustCompile(`([.?!]|\n)+`)
	segments := reSplit.Split(text, -1)

	var cleanedSegments []string
	if len(segments) == 0 {
		return text
	}

	// Filter out adjacent duplicates
	lastSeg := ""
	dupCount := 0

	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}

		// Simple normalization for comparison (ignore case)
		normalized := strings.ToLower(seg)

		if normalized == lastSeg {
			dupCount++
			if dupCount < 2 { // Allow max 2 repetitions
				cleanedSegments = append(cleanedSegments, seg)
			}
		} else {
			cleanedSegments = append(cleanedSegments, seg)
			lastSeg = normalized
			dupCount = 0
		}
	}

	// Reconstruct (this loses original delimiters, but it's better than a hang)
	// We'll join with ". " as a best-effort approximation for readability
	return strings.Join(cleanedSegments, ". ")
}

func (t *Transcriber) cleanTimestamps(text string) string {
	// Regex to match [00:00:00.000 --> 00:00:05.000]
	re := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\.\d{3} --> \d{2}:\d{2}:\d{2}\.\d{3}\]\s*`)
	return re.ReplaceAllString(text, "")
}

func (t *Transcriber) convertToWav(input, output string) error {
	// Use embedded ffmpeg if available, else system
	ffmpegPath := t.dep.GetBinaryPath("ffmpeg")
	if _, err := os.Stat(ffmpegPath); os.IsNotExist(err) {
		if p, found := t.dep.CheckSystemDependency("ffmpeg"); found {
			ffmpegPath = p
		} else {
			return fmt.Errorf("ffmpeg not found")
		}
	}

	cmd := exec.Command(ffmpegPath, "-i", input, "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", "-y", output)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %s, %w", string(out), err)
	}
	return nil
}
