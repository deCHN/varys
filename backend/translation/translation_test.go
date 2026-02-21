package translation

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestTranslate(t *testing.T) {
	// 1. Init translator with mock HTTP client (no real socket listener required).
	tr := NewTranslator("test-model")
	tr.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodPost {
				t.Errorf("Expected POST, got %s", req.Method)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBufferString(`{"response":"1. 你好.\n2. 世界。","done":true}`)),
			}, nil
		}),
	}

	// 2. Run
	input := "Hello.\nWorld."
	results, err := tr.Translate(input, "Simplified Chinese", 4096, nil)
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if strings.TrimRight(results[0].Translated, "。.") != "你好" {

		t.Errorf("Expected '你好', got: %s", results[0].Translated)

	}

	if strings.TrimRight(results[1].Translated, "。.") != "世界" {

		t.Errorf("Expected '世界', got: %s", results[1].Translated)

	}

}
