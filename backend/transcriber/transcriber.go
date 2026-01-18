package transcriber

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"Varys/backend/dependency"
)

type Transcriber struct {
	dep       *dependency.Manager
}

func NewTranscriber(dep *dependency.Manager) *Transcriber {
	return &Transcriber{dep: dep}
}

func (t *Transcriber) Transcribe(audioPath, modelPath string, onProgress func(string)) (string, error) {
	// 1. Find binary
	candidates := []string{"whisper-cpp", "whisper-cli", "whisper-main", "whisper", "main"}
	var binPath string
	for _, name := range candidates {
		if p, found := t.dep.CheckSystemDependency(name); found {
			binPath = p
			break
		}
	}

	if binPath == "" {
		return "", fmt.Errorf("whisper binary not found in PATH. Please install whisper.cpp")
	}

	// 2. Check model
	if modelPath == "" {
		return "", fmt.Errorf("whisper model path not configured")
	}
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return "", fmt.Errorf("model file not found at %s", modelPath)
	}

	// 3. Convert to WAV (16kHz, Mono)
	wavPath := audioPath + ".wav"
	if err := t.convertToWav(audioPath, wavPath); err != nil {
		return "", err
	}
	defer os.Remove(wavPath)

	// 4. Run Whisper
	// Expects output: wavPath + ".txt"
	cmd := exec.Command(binPath, "-m", modelPath, "-f", wavPath, "--output-txt", "--no-timestamps", "--language", "auto")
	
	// Stream output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	cmd.Stderr = cmd.Stdout // Capture stderr too (whisper prints progress there usually)

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start whisper: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if onProgress != nil {
			onProgress(line)
		}
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("whisper execution failed: %w", err)
	}

	resultFile := wavPath + ".txt"
	content, err := os.ReadFile(resultFile)
	if err != nil {
		return "", fmt.Errorf("failed to read transcript file %s: %w", resultFile, err)
	}
	defer os.Remove(resultFile)

	return string(content), nil
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
