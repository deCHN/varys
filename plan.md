# Project Plan: v2k-mac (Auto-Clipper Native App)

**Version:** 0.1.0-draft
**Branch:** `go_wails_mac`
**Target Platform:** macOS (Apple Silicon Optimized)
**Framework:** Go + Wails (v2) + React/TypeScript

---

## 1. 架构概览 (Architecture Overview)

本项目旨在将原有的 Python 脚本 (`auto_obsidian.py`) 重构为高性能、本地化的 macOS 桌面应用。

### 1.1 技术栈 (Tech Stack)
*   **Frontend (UI)**: React, TypeScript, TailwindCSS (Via Wails)
*   **Backend (Logic)**: Go (Golang)
*   **AI Inference (Native)**:
    *   **ASR**: `whisper.cpp` (Go Bindings) - 用于语音转文字
    *   **LLM**: `llama.cpp` (via `go-llama.cpp` or compatible bindings) - 用于总结与翻译
*   **External Dependencies**: `yt-dlp` (二进制文件调取)

### 1.2 数据流 (Data Flow)
1.  **UI**: 用户输入 URL -> 点击 "Process"
2.  **Go Bridge**: 接收 URL
3.  **Downloader (Go)**: 调用 `yt-dlp` 下载音频到临时目录
4.  **Audio Processor (Go/CGO)**: 调用 `whisper.cpp` 进行转录 (m4a -> wav -> text/segments)
5.  **Intelligence Engine (Go/CGO)**:
    *   加载 LLM (GGUF)
    *   发送 Prompt + Transcript
    *   获取 JSON 结构的分析结果
6.  **File Manager (Go)**:
    *   移动音频到 Obsidian 库
    *   生成 Markdown 笔记
7.  **UI**: 推送进度更新 -> 显示最终结果

---

## 2. 模块划分 (Module Breakdown)

### 2.1 Backend (Go)

| 模块名 | 职责描述 | 依赖 |
| :--- | :--- | :--- |
| `main` | Wails 应用初始化、生命周期管理、菜单配置 | `wails` |
| `app` | Wails 绑定的核心结构体，暴露 API 给前端 | 所有子模块 |
| `downloader` | 负责检查/更新 `yt-dlp`，执行音频下载，格式转换 (ffmpeg) | `os/exec` |
| `transcriber` | 封装 `whisper.cpp`，管理模型加载，执行推理，清洗幻觉 | `whisper.cpp` bindings |
| `analyzer` | 封装 `llama.cpp`，管理 Prompt 模板，执行文本分析与翻译 | `llama.cpp` bindings |
| `storage` | 负责 Obsidian 路径解析、文件读写、Markdown 模板渲染 | `text/template` |
| `config` | 管理用户配置 (Vault 路径、模型路径、偏好设置) | `viper` or standard JSON |

### 2.2 Frontend (React)

| 组件名 | 职责描述 |
| :--- | :--- |
| `Dashboard` | 主界面，包含 URL 输入框、任务列表 |
| `TaskCard` | 单个任务的状态显示 (下载中/转录中/完成)，进度条 |
| `Settings` | 配置界面 (Obsidian 路径选择、模型管理/下载) |
| `LogViewer` | 实时显示后端日志 (用于调试) |

---

## 3. 接口定义 (Wails API Definitions)

前端通过 `window.go.main.App.<MethodName>` 调用。

### 3.1 核心业务接口

```go
// Task 结构体 (用于前端显示)
type Task struct {
    ID          string  `json:"id"`
    URL         string  `json:"url"`
    Status      string  `json:"status"` // "pending", "downloading", "transcribing", "analyzing", "done", "error"
    Progress    float64 `json:"progress"`
    Title       string  `json:"title"`
    Log         []string `json:"log"`
}

// 提交新任务
// Return: taskID (用于订阅事件) or error
func (a *App) SubmitTask(url string) (string, error)

// 获取当前配置
func (a *App) GetConfig() (Config, error)

// 更新配置
func (a *App) UpdateConfig(cfg Config) error

// 选择文件夹 (调用系统原生对话框)
func (a *App) SelectVaultPath() (string, error)
```

### 3.2 事件订阅 (Wails Runtime Events)

后端通过 `runtime.EventsEmit` 主动推送，前端通过 `EventsOn` 监听。

*   `task:progress`: `{id: "123", status: "transcribing", progress: 0.45, message: "Processing segment 10/24..."}`
*   `task:complete`: `{id: "123", resultPath: "/path/to/obsidian/note.md"}`
*   `task:error`: `{id: "123", error: "Download failed"}`

---

## 4. 详细开发任务 (Development Tasks)

### Phase 1: 基础脚手架 (Infrastructure)
- [ ] 初始化 Wails 项目 (`wails init -n v2k-mac -t react-ts`)
- [ ] 配置 Go mod 依赖
- [ ] 搭建简易 UI (输入框 + 按钮 + 控制台日志区域)
- [ ] 实现 `Log` 系统：Go `fmt.Println` -> Wails Event -> 前端显示

### Phase 2: 核心逻辑迁移 (Core Logic Migration)
- [ ] **Config Module**: 实现配置文件的读写 (保存 Obsidian 路径)。
- [ ] **Downloader Module**:
    - [ ] 检测本地是否安装 `yt-dlp` 和 `ffmpeg`。
    - [ ] 实现 `DownloadAudio(url)` 函数，调用命令下载 m4a。
    - [ ] **Test**: 单元测试 `downloader`，Mock `exec.Command`。
- [ ] **Storage Module**:
    - [ ] 实现 `SaveNote(metadata, content)`。
    - [ ] 移植 Python 中的 `sanitize_filename` 逻辑。

### Phase 3: AI 引擎集成 (AI Integration) - **难点**
- [ ] **Whisper Integration**:
    - [ ] 引入 `whisper.cpp` Go binding。
    - [ ] 实现模型下载器 (首次运行下载 ggml-base.bin)。
    - [ ] 实现 `Transcribe(audioPath)`。
    - [ ] 移植 `clean_hallucinations` 逻辑 (Go string 处理)。
- [ ] **LLM Integration**:
    - [ ] 引入 `go-llama.cpp` binding。
    - [ ] 实现模型下载器 (下载 Qwen2.5 GGUF)。
    - [ ] 实现 `Analyze(text)`，复用 Python 中的 Prompt。

### Phase 4: UI 完善 (UI Polish)
- [ ] 使用 Tailwind 美化界面。
- [ ] 实现任务队列 (允许并发或排队处理多个 URL)。
- [ ] 增加“模型管理”页面 (显示模型下载进度)。

### Phase 5: 打包与发布 (Distribution)
- [ ] 配置 `wails build` 参数 (图标、版权信息)。
- [ ] 测试 `.app` 在无开发环境机器上的运行情况。

---

## 5. 测试用例规划 (Test Plan)

### 5.1 单元测试 (Go Unit Tests)
| 模块 | 测试点 | 预期结果 |
| :--- | :--- | :--- |
| `utils` | `SanitizeFilename("A/B:C")` | `A_B_C` |
| `downloader` | `GetVideoTitle` (Mock Network) | 返回 Mock 标题 |
| `transcriber` | `CleanHallucinations("OK! OK! OK!")` | `OK!` |

### 5.2 集成测试 (Integration Tests)
*   **Audio Pipeline**: 输入一个 30秒的本地测试音频 -> 验证是否生成了非空的文本。
*   **LLM Pipeline**: 输入 "Hello World" -> 验证是否返回 JSON 格式结果。

### 5.3 界面测试 (Manual/E2E)
*   **场景**: 用户首次打开 App。
    *   期望：自动引导设置 Obsidian 路径，提示下载必要模型。
*   **场景**: 输入无效 URL。
    *   期望：UI 弹出错误提示，任务状态变红。

---

## 6. 风险评估 (Risk Assessment)

1.  **CGO 编译兼容性**:
    *   *风险*: `whisper.cpp` / `llama.cpp` 在不同 macOS 版本或架构上的编译参数可能不同。
    *   *对策*: 锁定 Binding 库的版本，使用 Docker 或 CI 确保构建环境一致。优先支持 Apple Silicon。
2.  **模型下载体验**:
    *   *风险*: 国内网络环境下载 HuggingFace 模型慢。
    *   *对策*: 支持自定义模型镜像源 (Mirror)，或允许用户手动导入本地模型文件。
3.  **内存占用**:
    *   *风险*: 同时运行 Whisper 和 LLM 可能导致 OOM。
    *   *对策*: 串行化处理 (先释放 Whisper 内存，再加载 LLM)。

---
