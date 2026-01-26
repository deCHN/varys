package storage

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// NoteData holds the data for the markdown note
type NoteData struct {
	Title        string
	URL          string
	Language     string
	Summary      string
	KeyPoints    []string
	Tags         []string
	Assessment   map[string]string
	OriginalText string
	Translated   string
	AudioFile    string
	CreatedTime  string
	AssetsFolder string
}

type Manager struct {
	VaultPath string
}

func NewManager(vaultPath string) *Manager {
	return &Manager{VaultPath: vaultPath}
}

func (m *Manager) SanitizeFilename(name string) string {
	// 1. Remove Obsidian link breakers
	reLink := regexp.MustCompile(`[#^[\]]`) // Corrected escaping for [ and ]
	name = reLink.ReplaceAllString(name, "")

	// 2. Remove system illegal chars
	reSys := regexp.MustCompile(`[\\/*?:"><|]`) // Corrected escaping for \ and "
	name = reSys.ReplaceAllString(name, "")

	// 3. Replace spaces and punctuation with underscore
	replacements := []string{" ", "　", "，", ",", "。", ":", "：", "“", "”", "‘", "’"}
	for _, char := range replacements {
		name = strings.ReplaceAll(name, char, "_")
	}

	// 5. Merge underscores
	reUnder := regexp.MustCompile(`_{2,}`)
	name = reUnder.ReplaceAllString(name, "_")

	// 6. Strip
	name = strings.Trim(name, "_")

	// 7. Truncate
	runes := []rune(name)
	if len(runes) > 80 {
		name = string(runes[:80])
	}

	return name
}

// MoveMedia moves the media file (audio/video) to vault/assets
func (m *Manager) MoveMedia(sourcePath string, targetName string) (string, error) {
	// Default assets folder name
	assetsDir := filepath.Join(m.VaultPath, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return "", err
	}

	ext := filepath.Ext(sourcePath)
	finalName := targetName + ext
	destPath := filepath.Join(assetsDir, finalName)

	// Remove dest if exists
	if _, err := os.Stat(destPath); err == nil {
		os.Remove(destPath)
	}

	// Try Rename first
	err := os.Rename(sourcePath, destPath)
	if err != nil {
		// Fallback to Copy
		src, err := os.Open(sourcePath)
		if err != nil { return "", err }
		defer src.Close()

		dst, err := os.Create(destPath)
		if err != nil { return "", err }
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil { return "", err }
		os.Remove(sourcePath)
	}

	return finalName, nil
}

func (m *Manager) SaveNote(data NoteData) (string, error) {
	safeTitle := m.SanitizeFilename(data.Title)
	filename := safeTitle + ".md"
	filePath := filepath.Join(m.VaultPath, filename)

	tmplStr := `---
created: {{.CreatedTime}}
source: "{{.URL}}"
type: auto_clipper
language: {{.Language}}
tags:
{{- range .Tags}}
  - {{.}}
{{- end}}
---

# {{.Title}}

## 智能摘要

{{.Summary}}

### 核心观点

{{- range .KeyPoints}}
- {{.}}
{{- end}}

### 智能评估
| 维度 | 评估内容 |
| :--- | :--- |
| **真实性** | {{index .Assessment "authenticity"}} |
| **有效性** | {{index .Assessment "effectiveness"}} |
| **实时性** | {{index .Assessment "timeliness"}} |
| **替代策略** | {{index .Assessment "alternatives"}} |

---

## 媒体回放
![[{{.AssetsFolder}}/{{.AudioFile}}]]

---
{{if ne .Language "zh"}}
## 全文翻译

{{.Translated}}

---
{{end}}
## 原始内容

{{.OriginalText}}
`

	tt, err := template.New("note").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := tt.Execute(f, data); err != nil {
		return "", err
	}

	return filePath, nil
}
