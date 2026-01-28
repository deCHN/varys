package translation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Translator struct {
	modelName string
	apiURL    string
}

func NewTranslator(model string) *Translator {
	if model == "" {
		model = "qwen3:0.6b" // Default for translation
	}
	return &Translator{
		modelName: model,
		apiURL:    "http://localhost:11434/api/generate",
	}
}

type Request struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type Response struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type TranslationPair struct {
	Original   string `json:"original"`
	Translated string `json:"translated"`
}

func (t *Translator) Translate(text string, targetLang string, contextSize int, onProgress func(int, int)) ([]TranslationPair, error) {
	if targetLang == "" {
		targetLang = "Simplified Chinese"
	}
	if contextSize == 0 {
		contextSize = 8192 // Fallback default
	}

	// Chunking logic: Translate in blocks of ~2000 characters to ensure stability
	const chunkSize = 2000
	var allPairs []TranslationPair

	chunks := (len(text) + chunkSize - 1) / chunkSize

	// Simple split by characters (could be improved to split by paragraph)
	for i := 0; i < len(text); i += chunkSize {
		chunkIdx := i / chunkSize
		if onProgress != nil {
			onProgress(chunkIdx, chunks)
		}

		end := i + chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunk := text[i:end]

		prompt := fmt.Sprintf(`You are a professional translator.
Task: Translate the following text into %s.
Format: Return ONLY a valid JSON array of objects. Each object represents a sentence or logical segment.
Structure: [{"original": "source sentence", "translated": "translated sentence"}]

Text to translate:
%s`, targetLang, chunk)

		reqBody := Request{
			Model:  t.modelName,
			Prompt: prompt,
			Stream: false,
			Options: map[string]interface{}{
				"num_ctx":     contextSize,
				"num_predict": 4096, // Allow enough output tokens
			},
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}

		resp, err := http.Post(t.apiURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("ollama error %s: %s", resp.Status, string(body))
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var result Response
		if err := json.Unmarshal(body, &result); err != nil {
			// If a single chunk fails, we log and continue or return error
			continue
		}

		responseText := result.Response
		if idx := strings.Index(responseText, "["); idx != -1 {
			responseText = responseText[idx:]
		}
		if idx := strings.LastIndex(responseText, "]"); idx != -1 {
			responseText = responseText[:idx+1]
		}

		var pairs []TranslationPair
		if err := json.Unmarshal([]byte(responseText), &pairs); err != nil {
			// Fallback for this specific chunk
			allPairs = append(allPairs, TranslationPair{
				Original:   chunk,
				Translated: responseText,
			})
		} else {
			allPairs = append(allPairs, pairs...)
		}
	}

	if len(allPairs) == 0 && len(text) > 0 {
		return nil, fmt.Errorf("translation failed for all chunks")
	}

	return allPairs, nil
}
