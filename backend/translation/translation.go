package translation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
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

	// 1. Split text into individual sentences
	sentences := t.splitSentences(text)
	if len(sentences) == 0 {
		return nil, nil
	}

	// 2. Process in small numbered batches (5-8 sentences)
	// Small models (0.6b) are much more accurate with numbering and small batches.
	const batchSize = 7
	var allPairs []TranslationPair

	totalBatches := (len(sentences) + batchSize - 1) / batchSize

	for i := 0; i < len(sentences); i += batchSize {
		batchIdx := i / batchSize
		if onProgress != nil {
			onProgress(batchIdx, totalBatches)
		}

		end := i + batchSize
		if end > len(sentences) {
			end = len(sentences)
		}
		currentBatch := sentences[i:end]

		// Prepare numbered input
		var inputBuilder strings.Builder
		for j, s := range currentBatch {
			inputBuilder.WriteString(fmt.Sprintf("%d. %s\n", j+1, s))
		}

		prompt := fmt.Sprintf(`You are a professional translator.
Task: Translate the following numbered sentences into %s.
Rules:
1. Output exactly %d translated sentences.
2. Use the same numbering format: "1. [translation]\n2. [translation]..."
3. Do not include any introductory text, notes, or explanations.
4. If a line is just punctuation, keep it as is.

Input:
%s`, targetLang, len(currentBatch), inputBuilder.String())

		reqBody := Request{
			Model:  t.modelName,
			Prompt: prompt,
			Stream: false,
			Options: map[string]interface{}{
				"num_ctx":     contextSize,
				"num_predict": 2048, // Smaller output expected per batch
				"temperature": 0.1,  // Lower temperature for stricter adherence
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
			continue
		}

		// 3. Parse Numbered Output
		translatedLines := t.parseNumberedOutput(result.Response, len(currentBatch))

		for j := 0; j < len(currentBatch); j++ {
			trans := "(Translation missing)"
			if j < len(translatedLines) {
				trans = translatedLines[j]
			}
			allPairs = append(allPairs, TranslationPair{
				Original:   currentBatch[j],
				Translated: trans,
			})
		}
	}

	if len(allPairs) == 0 && len(text) > 0 {
		return nil, fmt.Errorf("translation failed for all chunks")
	}

	return allPairs, nil
}

// parseNumberedOutput extracts text from lines starting with "1. ", "2. ", etc.
func (t *Translator) parseNumberedOutput(output string, expectedCount int) []string {
	lines := strings.Split(output, "\n")
	results := make([]string, 0, expectedCount)

	// Regex to match "1. Text" or "1: Text" or just the text if model forgot numbering
	reNumber := regexp.MustCompile(`^\d+[\.\:\s]+(.*)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := reNumber.FindStringSubmatch(line)
		if len(matches) > 1 {
			results = append(results, strings.TrimSpace(matches[1]))
		} else {
			// If model didn't use numbering but gave a line of text, take it
			results = append(results, line)
		}
	}

	// If the model produced a single block instead of lines,
	// this parser might return too few items.
	// The caller handles padding with "missing".
	return results
}

// splitSentences splits text into sentences or logical segments using regex.
func (t *Translator) splitSentences(text string) []string {
	// Split by common sentence terminators (. ? ! \n) followed by space or end of string
	// We want to keep the delimiter attached to the previous sentence if possible,
	// but Go's regex split is simple.
	// Alternative: Walk through and split.

	// Simple approach: Use regex to find sentences.
	// This regex matches non-empty sequences ending in punctuation or newline.
	// re := regexp.MustCompile(`[^.!?\n]+[.!?\n]+`)
	// matches := re.FindAllString(text, -1)

	// If text has no punctuation, it might be one huge block.
	// Let's use a simpler split for robustness: split by newlines first (whisper segments),
	// then maybe split long lines.

	// Whisper output is already segmented by logic, often separated by ". "

	var segments []string

	// 1. Split by hard newlines first
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 2. Further split by ". " "? " "! " if the line is very long (>200 chars)
		// Otherwise keep it as one unit for context.
		if len(line) > 200 {
			// Simple split by ". "
			subParts := strings.Split(line, ". ")
			for i, part := range subParts {
				if i < len(subParts)-1 {
					part += "."
				}
				segments = append(segments, part)
			}
		} else {
			segments = append(segments, line)
		}
	}

	return segments
}
