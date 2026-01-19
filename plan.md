# Project Plan: Varys (Auto-Clipper Native App)

**Version:** 0.3.0-draft
**Branch:** `go_wails_mac`
**Target Platform:** macOS (Apple Silicon Optimized)
**Framework:** Go + Wails (v2) + React/TypeScript

---

## 1. 架构概览 (Architecture Overview)

本项目旨在构建一个高性能、本地化的 macOS 桌面应用。

### 1.1 技术栈 (Tech Stack)
*   **Frontend (UI)**: React, TypeScript, TailwindCSS (Via Wails)
*   **Backend (Logic)**: Go (Golang)
*   **AI Inference (Native)**:
    *   **ASR**: `whisper.cpp` (Go Bindings) - 用于语音转文字
    *   **LLM**: `llama.cpp` (via `go-llama.cpp` or compatible bindings) - 用于总结与翻译
*   **External Dependencies**: `yt-dlp`, `ffmpeg` (采用 Sidecar 模式内置)

### 1.2 数据流 (Data Flow)
1.  **UI**: 用户输入 URL -> 点击 "Process"
2.  **Go Bridge**: 接收 URL
3.  **Downloader (Go)**: 调用内置的 `yt-dlp` 下载音频
4.  **Audio Processor (Go/CGO)**: 调用 `whisper.cpp` 进行转录
5.  **Intelligence Engine (Go/CGO)**: 加载 LLM 执行分析与翻译
6.  **File Manager (Go)**: 移动音频到 Obsidian 库，生成 Markdown 笔记
7.  **UI**: 推送进度更新 -> 显示最终结果

---

## 2. 模块划分 (Module Breakdown)

### 2.1 Backend (Go)

| 模块名 | 职责描述 | 依赖 |
| :--- | :--- | :--- |
| `main` | Wails 应用初始化、生命周期管理、菜单配置 | `wails` |
| `app` | Wails 绑定的核心结构体，暴露 API 给前端 | 所有子模块 |
| `dependency` | **[已完成]** 管理内嵌的 `yt-dlp` 和 `ffmpeg`。负责释放二进制文件到本地路径，并执行 `-U` 自动更新。 | `embed`, `os/exec` |
| `downloader` | 执行音频下载，调用 `dependency` 提供的路径。 | `yt-dlp` |
| `transcriber` | 封装 `whisper.cpp`，管理模型加载，执行推理。 | `whisper.cpp` bindings |
| `analyzer` | 封装 `llama.cpp`，管理 Prompt 模板，执行文本分析与翻译 | `llama.cpp` bindings |
| `storage` | 负责 Obsidian 路径解析、文件读写、Markdown 模板渲染 | `text/template` |
| `config` | 管理用户配置 (Vault 路径、模型路径、偏好设置) | `viper` or standard JSON |

### 2.2 Frontend (React)

| 组件名 | 职责描述 |
| :--- | :--- |
| `Dashboard` | 主界面，包含 URL 输入框、任务列表 |
| `TaskCard` | 单个任务的状态显示 (下载中/转录中/完成)，进度条 |
| `Settings` | 配置界面 (Obsidian 路径选择、模型管理、依赖更新检查) |
| `LogViewer` | 实时显示后端日志 (用于调试) |

---

## 3. 接口定义 (Wails API Definitions)

### 3.1 核心业务接口 (Go to JS)
*   `SubmitTask(url string) (string, error)`
*   `GetConfig() (Config, error)`
*   `UpdateConfig(cfg Config) error`
*   `SelectVaultPath() (string, error)`
*   `CheckDependencies() (DependencyStatus, error)`
*   `UpdateDependencies() error`

---

## 4. 详细开发任务 (Development Tasks)

### Phase 1: 基础脚手架与依赖管理 (Infrastructure & Sidecar) - **[Completed]**
- [x] 初始化 Wails 项目 (`wails init -n varys -t react-ts`)
- [x] **Dependency Manager**:
    - [x] 创建 `resources/bin/darwin_arm64/` 目录。
    - [x] 实现 `ReleaseBinaries()`: 将内嵌的 `yt-dlp` 和 `ffmpeg` 释放到 `~/Library/Application Support/Varys/bin/`。
    - [x] 实现 `AutoUpdateDependencies()`: 异步执行 `yt-dlp -U`。
    - [x] 单元测试 `dependency_test.go`。
    - [x] **AI Dependency Script**: 更新 `scripts/deps.go` 支持 `tar.gz` 及特定路径提取，完成 `llama-server` 和 `whisper` 库文件的下载集成。
- [x] 搭建简易 UI (输入框 + 按钮 + 控制台区域)。
- [x] **UI Polish**:
    - [x] 移除 Wails Logo。
    - [x] 调整标题为 "Varys"。
    - [x] 修复按钮样式。
    - [x] 调整初始窗口大小为 800x600。
- [x] **Tests Setup**:
    - [x] 配置 Vitest (happy-dom) 进行单元测试。
    - [x] 配置 Playwright 进行 E2E 测试和截图。

### Phase 2: 核心逻辑迁移 (Core Logic Migration) - **[Completed]**
- [x] **Config Module**: 实现配置文件的读写 (保存 Obsidian 路径)。
- [x] **Downloader Module**:
    - [x] 调用 `dependency` 模块获取 `yt-dlp` 路径并下载音频。
    - [x] 单元测试 `downloader_test.go`。
- [x] **Storage Module**:
    - [x] 实现 `SaveNote` 并移植 `sanitize_filename` 逻辑。

### Phase 3: AI 引擎集成 (AI Integration - BYOAI) - **[Completed]**
- [x] **Strategy Change**: Switch to "Bring Your Own AI" (System Dependencies).
- [x] **Transcriber Module**:
    - [x] Implement `Transcribe` using system `whisper-cli` / `whisper-main`.
    - [x] Implement `convertToWav` using `ffmpeg`.
- [x] **Analyzer Module**:
    - [x] Implement `Analyze` using local Ollama API (`qwen2.5:7b`).
- [x] **App Integration**:
    - [x] Update `SubmitTask` to run full pipeline (DL -> Transcribe -> Analyze -> Save).
    - [x] Add `CheckDependencies` API for frontend.

### Phase 4: UI 完善与任务队列 (UI Polish & Task Queue)
- [ ] **Navigation & Layout**:
    - [ ] 实现侧边栏/Tab切换 (Dashboard vs Settings)。
- [ ] **Settings Page**:
    - [ ] 实现 `CheckDependencies` 可视化反馈 (红/绿灯)。
    - [ ] 实现配置表单 (Vault Path, Model Path, Ollama Model)。
    - [ ] 集成 Wails 原生对话框 (`runtime.OpenDirectoryDialog`, `runtime.OpenFileDialog`) 选择路径。
- [ ] **Task Queue (Dashboard)**:
    - [ ] 重构主界面，支持多任务列表展示。
    - [ ] 优化日志显示，从纯文本改为结构化状态 (下载中 -> 转录中 -> 完成)。

### Phase 5: 打包与发布 (Distribution)
- [ ] 配置应用图标。
- [ ] 编写最终用户指南 (安装 whisper/ollama)。
- [ ] 执行 `wails build` 生成 Release 版本。

---

## 5. 测试用例规划 (Test Plan)

### 5.1 单元测试
- [x] **Dependency Test**: 验证二进制文件释放逻辑。
- [x] **UI Unit Test**: 验证 React 组件渲染 (Vitest)。

### 5.2 E2E 测试
- [x] **Smoke Test**: 启动 App，检查标题，输入 URL，截图 (Playwright)。

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