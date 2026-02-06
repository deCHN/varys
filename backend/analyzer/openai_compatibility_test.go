package analyzer_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"Varys/backend/analyzer"
)

// TestOpenAIModelCompatibility 遍历所有 OpenAI 模型，检查哪些与当前的参数配置兼容
func TestOpenAIModelCompatibility(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping: OPENAI_API_KEY not set")
	}

	// 1. 获取模型列表
	p := analyzer.NewOpenAIProvider(apiKey, "")
	ctx := context.Background()
	models, err := p.ListModels(ctx)
	if err != nil {
		t.Fatalf("Failed to list models: %v", err)
	}

	// 2. 遍历并尝试 Analyze
	// 注意：我们使用 t.Errorf 而不是 Fatal，以便跑完所有模型
	for _, model := range models {
		// 过滤掉非聊天模型
		if !strings.HasPrefix(model, "gpt-") && !strings.HasPrefix(model, "o1-") {
			continue
		}

		t.Run(model, func(t *testing.T) {
			an := analyzer.NewAnalyzer("openai", apiKey, model)
			text := "Say 'OK' if you can read this."

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			start := time.Now()
			// 这里的 Analyze 内部会带入 temperature: 0.1 等参数
			_, err := an.Analyze(ctx, text, "", "English", 1024, nil)
			duration := time.Since(start)

			if err != nil {
				t.Errorf("[!] %-25s | ERR | %v", model, err)
			} else {
				t.Logf("[+] %-25s | OK  | %v", model, duration)
			}
		})
	}
}
