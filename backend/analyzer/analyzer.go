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
		model = "qwen3:8b" // Default
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

func (a *Analyzer) Analyze(text string, targetLang string, contextSize int, onToken func(string)) (*AnalysisResult, error) {
	if targetLang == "" {
		targetLang = "Simplified Chinese"
	}
	if contextSize == 0 {
		contextSize = 8192 // Fallback default
	}

		prompt := fmt.Sprintf(`You are an expert content analyst.

	Task: Analyze the following text and provide a structured analysis in %s.

	Rules:

	1. OUTPUT MUST BE IN %s.

	2. If the input text is in English, TRANSLATE your analysis to %s.

	3. Tags must be single words or hyphenated (no spaces).

	

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

	%s`, targetLang, targetLang, targetLang, text)

	

		reqBody := Request{

			Model:  a.modelName,

			Prompt: prompt,

			Stream: true,

			Options: map[string]interface{}{

				"num_ctx": contextSize,

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

	

		// Post-process: Sanitize Tags

		for i, tag := range analysis.Tags {

			// Replace spaces with hyphens

			analysis.Tags[i] = strings.ReplaceAll(tag, " ", "-")

		}

	

		return &analysis, nil

	}

	