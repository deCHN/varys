<div align="center">
  <img src="build/appicon.png" width="128" height="128" alt="Varys Logo">
  <h1>Varys</h1>
  <p>
    <strong>Your Personal Knowledge Intelligence Agent</strong>
  </p>
  <p>
    <a href="#english">English</a> | <a href="#simplified-chinese">简体中文</a>
  </p>
  <p>
    <img src="https://img.shields.io/badge/Platform-macOS-black?logo=apple" alt="Platform">
    <img src="https://img.shields.io/badge/Built%20with-Wails%20%2B%20Go-cyan?logo=go" alt="Wails">
    <img src="https://img.shields.io/badge/AI-Local%20LLM-violet?logo=openai" alt="AI">
    <img src="https://img.shields.io/badge/License-MIT-green" alt="License">
  </p>
</div>

<hr>

<a name="english"></a>
## English

**Varys** is a powerful desktop application designed to bridge the gap between video content and your personal knowledge base. It automates the process of capturing, transcribing, and analyzing videos from platforms like **YouTube** and **Bilibili**, seamlessly integrating the insights into **Obsidian**.

Optimized for **Apple Silicon**, Varys leverages local AI models for privacy, speed, and zero cost.

### Key Features

*   **One-Click Capture:** Instantly download and process audio from YouTube, Bilibili, and other supported platforms using `yt-dlp`.
*   **Local Transcription:** High-performance, offline speech-to-text powered by `whisper.cpp` (Metal accelerated). No cloud API keys required.
*   **AI Intelligence:** Automatically summarizes content, extracts key points, and tags notes using local LLMs (via **Ollama**). Supports real-time streaming analysis.
*   **Smart Translation:** Automatically translates content to your selected **Target Language** (configured in Settings). Smartly skips translation if source matches target, and uses chunking for reliable long-video support.
*   **Obsidian Integration:** Direct export to your Obsidian Vault with properly formatted Markdown, frontmatter metadata, and embedded audio.
*   **Modern UI:** A beautiful, dark-themed interface built with React and Tailwind CSS for a premium desktop experience.

### ⚡️ Performance
Varys is optimized for speed on Apple Silicon.
[View Performance Benchmarks & History](docs/PERFORMANCE.md)

### Getting Started

#### Prerequisites
*   **macOS** (Apple Silicon recommended for best performance)
*   **Ollama** installed and running (`brew install ollama && ollama serve`)
*   **FFmpeg** (`brew install ffmpeg`)

#### Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/yourusername/varys.git
    cd varys
    ```

2.  **Install & Build:**
    ```bash
    make install
    ```
    This will build the app and install it to your `/Applications` folder.

3.  **Run:**
    Open **Varys** from your Applications folder or Spotlight.

### Usage

1.  **Configure:** On first run, go to the **Settings** tab to select your Obsidian Vault path and ensure all dependencies (Ollama, Whisper) are detected.
2.  **Capture:** Paste a video URL into the **Task** dashboard and click **Process**.
3.  **Review:** Watch the real-time logs and AI analysis. Once complete, your new note will appear instantly in Obsidian.

---

<a name="simplified-chinese"></a>
## 简体中文

**Varys** 是一款强大的桌面应用，旨在连接视频内容与您的个人知识库。它可以自动化地从 **YouTube** 和 **Bilibili** 等平台抓取、转录并分析视频内容，将洞察无缝集成到 **Obsidian** 中。

Varys 专为 **Apple Silicon** 优化，利用本地 AI 模型，确保隐私安全、极速响应且零成本。

### 核心功能

*   **一键捕获：** 使用 `yt-dlp` 瞬间下载并处理来自 YouTube、Bilibili 等平台的音频。
*   **本地转录：** 基于 `whisper.cpp` (Metal 加速) 的高性能离线语音转文字。无需任何云端 API Key。
*   **AI 智能分析：** 使用本地大模型 (通过 **Ollama**) 自动生成摘要、提取核心观点并打标签。支持实时流式输出分析结果。
*   **智能翻译：** 自动将内容翻译为您选择的**目标语言**（在设置中配置）。智能跳过源语言匹配的翻译，并支持长视频分块翻译。
*   **Obsidian 集成：** 直接导出到您的 Obsidian 仓库，包含格式完美的 Markdown、Frontmatter 元数据和内嵌音频文件。
*   **现代界面：** 基于 React 和 Tailwind CSS 构建的精美暗色主题界面，提供顶级的桌面应用体验。

### 快速开始

#### 前置要求
*   **macOS** (推荐使用 Apple Silicon 芯片以获得最佳性能)
*   **Ollama** 已安装并运行 (`brew install ollama && ollama serve`)
*   **FFmpeg** (`brew install ffmpeg`)

#### 安装指南

1.  **克隆代码仓库：**
    ```bash
    git clone https://github.com/yourusername/varys.git
    cd varys
    ```

2.  **安装与构建：**
    ```bash
    make install
    ```
    此命令将构建应用并自动将其安装到您的 `/Applications` 文件夹。

3.  **运行：**
    从应用程序文件夹或通过 Spotlight 启动 **Varys**。

### 使用说明

1.  **配置：** 首次运行时，进入 **设置 (Settings)** 页面选择您的 Obsidian 仓库路径，并确保所有依赖项 (Ollama, Whisper) 均已检测通过。
2.  **捕获：** 在 **任务 (Task)** 面板中粘贴视频链接，点击 **Process**。
3.  **回顾：** 观看实时的系统日志和 AI 分析流。完成后，新的笔记将立即出现在您的 Obsidian 中。

---

<div align="center">
  <p>Built by deChn</p>
</div>
