package service

import (
	"Varys/backend/analyzer"
	"Varys/backend/dependency"
	"Varys/backend/downloader"
	"Varys/backend/scraper"
	"Varys/backend/storage"
	"Varys/backend/transcriber"
	"Varys/backend/translation"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CoreService implements the Processor interface.
type CoreService struct {
	depManager *dependency.Manager
	scraper    *scraper.Scraper
}

// NewCoreService creates a new instance of CoreService.
func NewCoreService(dm *dependency.Manager) *CoreService {
	return &CoreService{
		depManager: dm,
		scraper:    scraper.NewScraper(),
	}
}

// ProcessTask runs the full pipeline: download, transcribe, analyze, and save.
func (s *CoreService) ProcessTask(ctx context.Context, url string, opts Options, logger EventLogger) (*TaskResult, error) {
	tempDir, err := os.MkdirTemp("", "varys_task_")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Detect if input is a local file
	isLocalFile := false
	if info, err := os.Stat(url); err == nil && !info.IsDir() {
		isLocalFile = true
		logger.Log("Local file detected: " + url)
	}

	var videoTitle, videoDescription string
	var transcript, sourceLang string
	var mediaPath string
	isArticle := false

	dl := downloader.NewDownloader(s.depManager)

	if isLocalFile {
		videoTitle = strings.TrimSuffix(filepath.Base(url), filepath.Ext(url))
		videoDescription = "Local file: " + url
		logger.Log(fmt.Sprintf("Using filename as title: %s", videoTitle))
	} else {
		// Attempt to get media info
		logger.Log("Fetching media metadata...")
		title, err := dl.GetVideoTitle(url)
		if err != nil || title == "" {
			logger.Log("Media not detected. Attempting to scrape as article...")
			art, sErr := s.scraper.Scrape(url)
			if sErr != nil {
				return nil, fmt.Errorf("content ingestion failed (tried media and article): %v", sErr)
			}
			isArticle = true
			videoTitle = art.Title
			videoDescription = "Article: " + url
			transcript = art.Content
			sourceLang = "auto" // Scraper usually doesn't give precise lang code
			logger.Log(fmt.Sprintf("Article detected: %s", videoTitle))
		} else {
			videoTitle = title
			logger.Log(fmt.Sprintf("Media found: %s", videoTitle))
			videoDescription, _ = dl.GetVideoDescription(url)
		}
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 2. Download/Prepare Media (Only for non-articles)
	if !isArticle {
		if isLocalFile {
			logger.Log("Preparing local file...")
			destPath := filepath.Join(tempDir, filepath.Base(url))
			if err := copyFile(url, destPath); err != nil {
				return nil, fmt.Errorf("failed to copy local file: %w", err)
			}
			mediaPath = destPath
		} else {
			logger.Log(fmt.Sprintf("Downloading media from %s...", url))
			mediaPath, err = dl.DownloadMedia(url, tempDir, opts.AudioOnly, func(msg string) {
				if ctx.Err() == nil {
					logger.Log("[DL] " + msg)
				}
			})
			if err != nil {
				return nil, fmt.Errorf("download failed: %w", err)
			}
		}
		logger.Log(fmt.Sprintf("Media ready: %s", mediaPath))

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// 3. Transcribe (Only for non-articles)
		logger.Log("Transcribing audio...")
		tr := transcriber.NewTranscriber(s.depManager)
		transcript, sourceLang, err = tr.Transcribe(mediaPath, opts.ModelPath, func(msg string) {
			if ctx.Err() == nil {
				logger.Log("[Whisper] " + msg)
			}
		})
		if err != nil {
			logger.Log("Transcription failed. Analysis will be skipped.")
			transcript = "Transcription failed."
		} else {
			logger.Log(fmt.Sprintf("Transcription complete (Language: %s).", sourceLang))
		}
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 4. Translate & Analyze (Common for both)
	analysis := &analyzer.AnalysisResult{}
	var translationPairs []translation.TranslationPair
	summary := "No analysis performed."

	if transcript != "Transcription failed." {
		targetLang := opts.TargetLanguage
		if targetLang == "" {
			targetLang = "English"
		}

		// Smart Translation Logic
		shouldTranslate := true
		isChineseSource := sourceLang == "zh"
		isChineseTarget := strings.Contains(targetLang, "Chinese")
		isEnglishSource := sourceLang == "en"
		isEnglishTarget := strings.Contains(targetLang, "English")

		if (isChineseSource && isChineseTarget) || (isEnglishSource && isEnglishTarget) {
			shouldTranslate = false
			logger.Log("Source language matches target. Skipping translation.")
		}

		if shouldTranslate {
			logger.Log(fmt.Sprintf("Translating to %s...", targetLang))
			translator := translation.NewTranslator(opts.TranslationMod)
			translationPairs, err = translator.Translate(transcript, targetLang, opts.ContextSize, func(current, total int) {
				if ctx.Err() == nil {
					percent := float64(current+1) / float64(total) * 100
					logger.Progress(percent)
				}
			})
			if err != nil {
				logger.Log(fmt.Sprintf("Translation failed: %v", err))
			} else {
				logger.Progress(100.0)
				logger.Log("Translation complete.")
			}
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Analysis
		logger.Log("Analyzing content...")
		provider := opts.AIProvider
		apiKey := opts.OpenAIKey
		model := opts.LLMModel
		if provider == "openai" {
			model = opts.OpenAIModel
		}

		az := analyzer.NewAnalyzer(provider, apiKey, model)
		analysis, err = az.Analyze(ctx, transcript, opts.CustomPrompt, targetLang, opts.ContextSize, func(token string) {
			if ctx.Err() == nil {
				logger.AnalysisChunk(token)
			}
		})
		if err != nil {
			logger.Log(fmt.Sprintf("Analysis failed: %v", err))
			analysis = &analyzer.AnalysisResult{}
		} else {
			summary = analysis.Summary
			logger.Log("Analysis complete.")
		}
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 5. Save to Storage
	vaultPath := opts.VaultPath
	if vaultPath == "" {
		home, _ := os.UserHomeDir()
		vaultPath = home
	}
	sm := storage.NewManager(vaultPath)
	safeTitle := sm.SanitizeFilename(videoTitle)

	var finalMedia string
	if !isArticle {
		finalMedia, err = sm.MoveMedia(mediaPath, safeTitle)
		if err != nil {
			return nil, fmt.Errorf("failed to move media: %w", err)
		}
	}

	noteData := storage.NoteData{
		Title:            safeTitle,
		URL:              url,
		Language:         opts.TargetLanguage,
		Description:      videoDescription,
		Summary:          summary,
		KeyPoints:        analysis.KeyPoints,
		Tags:             analysis.Tags,
		Assessment:       analysis.Assessment,
		OriginalText:     transcript,
		TranslationPairs: translationPairs,
		AudioFile:        finalMedia,
		AssetsFolder:     "assets",
		CreatedTime:      time.Now().Format("2006-01-02 15:04"),
		AIProvider:       analysis.Provider,
		AIModel:          analysis.Model,
	}

	notePath, err := sm.SaveNote(noteData)
	if err != nil {
		return nil, fmt.Errorf("failed to save note: %w", err)
	}

	return &TaskResult{
		NotePath:  notePath,
		MediaFile: finalMedia,
		Title:     safeTitle,
	}, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
