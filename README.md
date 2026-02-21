<p align="center">
  <img src="res/varys_banner.png" width="100%" alt="Varys Banner">
</p>

<h1 align="center">Varys</h1>

<p align="center">
  <strong>Your Personal Knowledge Intelligence Agent</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Platform-macOS-black?logo=apple" alt="Platform">
  <img src="https://img.shields.io/badge/Built%20with-Wails%20%2B%20Go-cyan?logo=go" alt="Wails">
  <img src="https://img.shields.io/badge/AI-Local%20LLM-violet?logo=ollama" alt="AI">
  <img src="https://img.shields.io/badge/License-MIT-green" alt="License">
</p>

---

Varys is a local-first desktop application designed to automate the capture, transcription, and analysis of video and audio content. It seamlessly transforms multimedia into structured insights and integrates them directly into your Obsidian Vault, building a high-fidelity personal knowledge base with zero cloud dependency.

## Project Goal

Varys bridges the gap between online multimedia and structured personal knowledge. By utilizing local AI models, it ensures your data remains private, processing is free, and your intelligence grows without reliance on external subscriptions or cloud providers.

## Key Features

- **One-Click Capture**: Extract and process video or audio from YouTube, Bilibili, and other major platforms via yt-dlp.
- **Local Transcription**: High-performance, hardware-accelerated speech-to-text powered by whisper.cpp with Metal optimization.
- **AI-Powered Analysis**: Automatic generation of summaries, key points, and tags using local LLMs via Ollama.
- **Intelligent Translation**: Context-aware translation to your target language with smart chunking for long-form content.
- **Obsidian Integration**: Direct export to your vault with formatted Markdown, frontmatter metadata, and embedded media files.
- **Privacy Centric**: All processing (transcription, analysis, and storage) is performed locally on your hardware.

## Technical Architecture

Varys orchestrates a stack of industry-leading local intelligence tools:

1. **yt-dlp**: For robust media extraction and metadata retrieval.
2. **whisper.cpp**: For high-speed, offline transcription with Apple Silicon optimization.
3. **Ollama**: For running high-quality LLMs (e.g., Qwen, Llama) for analysis and translation.
4. **Wails**: For a lightweight, native desktop experience using Go and React.

## Getting Started

### Prerequisites

- **macOS** (Apple Silicon M1/M2/M3 recommended for optimal performance).
- **Ollama**: Installed and running (`brew install ollama && ollama serve`).
- **FFmpeg**: Required for audio processing (`brew install ffmpeg`).

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/yourusername/varys.git
   cd varys
   ```

2. **Build and Install**:
   ```bash
   make install
   ```
   This command compiles the application and installs it into your `/Applications` folder.

## Usage

1. **Initial Setup**: Open the **Settings** tab to configure your Obsidian Vault path and verify that system dependencies are detected.
2. **Process Content**: Paste a video URL into the dashboard, select between **Audio** or **Video** mode, and click **Process**.
3. **Review**: Monitor real-time logs and AI analysis. Once the task is complete, the note will appear immediately in your Obsidian Vault.

## Roadmap

- Native support for local file drag-and-drop.
- Built-in library view for historical task management.
- Multi-vault configuration and management.
- User-definable analysis templates and prompts.

## Contributing

We welcome contributions to Varys. To contribute:

1. Fork the Project.
2. Create your Feature Branch (`git checkout -b feature/NewFeature`).
3. Commit your changes (`git commit -m 'Add NewFeature'`).
4. Push to the branch (`git push origin feature/NewFeature`).
5. Open a Pull Request.

## License

Varys is released under the **MIT License**. See the `LICENSE` file for full details.

---

<p align="center">
  Built with passion for the Personal Knowledge Management community.
</p>
