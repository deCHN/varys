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
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
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

func (a *Analyzer) Translate(text string, targetLang string) (string, error) {
	if targetLang == "" {
		targetLang = "Simplified Chinese"
	}
	
	prompt := fmt.Sprintf("Translate the following text into %s. Maintain original formatting:\n\n%s", targetLang, text)

	reqBody := Request{
		Model:  a.modelName,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(a.apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read full response for non-streaming
	body, _ := io.ReadAll(resp.Body)
	// We need to parse the NDJSON or single JSON response. 
	// The /api/generate endpoint with stream:false returns a single JSON object.
	
	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to decode translation: %w", err)
	}

	return result.Response, nil
}
