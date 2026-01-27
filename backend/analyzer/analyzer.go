package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Analyzer struct {
	modelName string
	apiURL    string
}

func NewAnalyzer(model string) *Analyzer {
	if model == "" {
		model = "qwen2.5:7b" // Default
	}
	return &Analyzer{
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

type AnalysisResult struct {
	Summary    string            `json:"summary"`
	KeyPoints  []string          `json:"key_points"`
	Tags       []string          `json:"tags"`
	Assessment map[string]string `json:"assessment"`
}

type TranslationPair struct {
	Original   string `json:"original"`
	Translated string `json:"translated"`
}

func (a *Analyzer) Analyze(text string, targetLang string, onToken func(string)) (*AnalysisResult, error) {
	if targetLang == "" {
		targetLang = "Simplified Chinese"
	}

	prompt := fmt.Sprintf(`You are an expert content analyst.
Task: Analyze the following text and provide a structured analysis in %s.
Format: Return ONLY a valid JSON object with the following structure:
{
  "summary": "Concise summary of the content",
  "key_points": ["Point 1", "Point 2", "Point 3"],
  "tags": ["Tag1", "Tag2", "Tag3"],
  "assessment": {
    "authenticity": "Rating/Comment",
    "effectiveness": "Rating/Comment",
    "timeliness": "Rating/Comment",
    "alternatives": "Rating/Comment"
  }
}

Text to analyze:
%s`, targetLang, text)

	reqBody := Request{
		Model:  a.modelName,
		Prompt: prompt,
		Stream: true,
		Options: map[string]interface{}{
			"num_ctx": 32768, // Increase context window to 32k to handle long transcripts
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(a.apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error %s: %s", resp.Status, string(body))
	}

	var fullResponse strings.Builder
	decoder := json.NewDecoder(resp.Body)

	for {
		var result Response
		if err := decoder.Decode(&result); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode stream: %w", err)
		}

		fullResponse.WriteString(result.Response)
		if onToken != nil {
			onToken(result.Response)
		}

		if result.Done {
			break
		}
	}

	// Clean up markdown code blocks if the LLM wrapped the output
	responseText := fullResponse.String()
	if idx := strings.Index(responseText, "{"); idx != -1 {
		responseText = responseText[idx:]
	}
	if idx := strings.LastIndex(responseText, "}"); idx != -1 {
		responseText = responseText[:idx+1]
	}

	var analysis AnalysisResult
	if err := json.Unmarshal([]byte(responseText), &analysis); err != nil {
		// Fallback: try to return just summary if JSON fails
		return &AnalysisResult{Summary: responseText}, nil
	}

	return &analysis, nil
}

func (a *Analyzer) Translate(text string, targetLang string, onProgress func(int, int)) ([]TranslationPair, error) {

	if targetLang == "" {

		targetLang = "Simplified Chinese"

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
			Model:  a.modelName,
			Prompt: prompt,
			Stream: false,
			Options: map[string]interface{}{
				"num_ctx":     8192, // 8k is enough for a 2k char chunk
				"num_predict": 4096, // Allow enough output tokens
			},
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}

		resp, err := http.Post(a.apiURL, "application/json", bytes.NewBuffer(jsonData))
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
