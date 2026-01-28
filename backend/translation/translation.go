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

	// 1. Split text into logical chunks (paragraphs/sentences) to avoid context limits
	// We aim for chunks of ~1500 chars to be safe.
	const maxChunkSize = 1500
	sentences := t.splitSentences(text)
	
	var chunks [][]string
	var currentChunk []string
	currentLen := 0

	for _, s := range sentences {
		if currentLen+len(s) > maxChunkSize && len(currentChunk) > 0 {
			chunks = append(chunks, currentChunk)
			currentChunk = []string{}
			currentLen = 0
		}
		currentChunk = append(currentChunk, s)
		currentLen += len(s)
	}
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	var allPairs []TranslationPair
	totalChunks := len(chunks)

	// 2. Process each chunk
	for i, sentenceGroup := range chunks {
		if onProgress != nil {
			onProgress(i, totalChunks)
		}

		// Prepare input block
		inputBlock := strings.Join(sentenceGroup, "\n")

		prompt := fmt.Sprintf(`You are a professional translator.
Task: Translate the following text into %s.
Rules:
1. Translate line-by-line.
2. Maintain the same number of lines as the input.
3. Do not output any notes, explanations, or extra text.
4. Output ONLY the translation.

Input Text:
%s`, targetLang, inputBlock)

		reqBody := Request{
			Model:  t.modelName,
			Prompt: prompt,
			Stream: false,
			Options: map[string]interface{}{
				"num_ctx":     contextSize,
				"num_predict": 4096,
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
			// Fail specific chunk
			continue
		}

		// 3. Parse Output (Line matching)
		translatedLines := strings.Split(strings.TrimSpace(result.Response), "\n")
		
		// Clean empty lines from split if any
		var cleanTranslated []string
		for _, line := range translatedLines {
			line = strings.TrimSpace(line)
			if line != "" {
				cleanTranslated = append(cleanTranslated, line)
			}
		}

		// Zip logic
		// If counts match, perfect.
		// If mismatch, we try to align or just dump the rest.
		count := len(sentenceGroup)
		if len(cleanTranslated) < count {
			count = len(cleanTranslated)
		}

		for j := 0; j < count; j++ {
			allPairs = append(allPairs, TranslationPair{
				Original:   sentenceGroup[j],
				Translated: cleanTranslated[j],
			})
		}
		
		// Handle unmatched lines (if source > translation)
		for j := count; j < len(sentenceGroup); j++ {
			allPairs = append(allPairs, TranslationPair{
				Original:   sentenceGroup[j],
				Translated: "(Translation missing)",
			})
		}
	}

	if len(allPairs) == 0 && len(text) > 0 {
		return nil, fmt.Errorf("translation failed for all chunks")
	}

	return allPairs, nil
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