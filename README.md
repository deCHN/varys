# Varys

Your Personal Knowledge Intelligence Agent.

Varys is a local-first desktop application that automates the process of capturing, transcribing, and analyzing video and audio content. It seamlessly integrates processed insights into your Obsidian Vault, creating a high-fidelity personal knowledge base.

## Project Goal
Varys aims to bridge the gap between online multimedia content and structured personal notes. By leveraging local AI models, it ensures that your data remains private, processing is free, and your knowledge base grows without relying on cloud services.

## Key Features
- One-Click Capture: Download and process video or audio from YouTube, Bilibili, and other platforms via yt-dlp.
- Local Transcription: High-performance, offline speech-to-text powered by whisper.cpp with Metal acceleration.
- AI Intelligence: Automatic summaries, key point extraction, and tagging using local LLMs via Ollama.
- Smart Translation: Automatic translation to your target language with context-aware chunking.
- Obsidian Integration: Direct export to Obsidian Vault with formatted Markdown and embedded media.
- Privacy First: All processing (transcription, analysis, storage) happens locally on your machine.

## How It Works
Varys orchestrates several high-performance tools:
1. yt-dlp: For media extraction and metadata retrieval.
2. whisper.cpp: For high-speed, hardware-accelerated transcription.
3. Ollama: For local LLM-based analysis and translation.
4. Wails: For a lightweight, native desktop experience using Go and React.

## Getting Started

### Prerequisites
- macOS (Apple Silicon recommended)
- Ollama (brew install ollama && ollama serve)
- FFmpeg (brew install ffmpeg)

### Installation
1. Clone the repository:
   git clone https://github.com/yourusername/varys.git
   cd varys

2. Build and Install:
   make install
   
   This command compiles the application and moves it to your /Applications folder.

## Usage
1. Initial Setup: Open Settings to select your Obsidian Vault path and verify that system dependencies are detected.
2. Process Content: Paste a URL into the Dashboard, choose between Audio or Video download, and click Process.
3. Review: Monitor progress through real-time logs. Once finished, the result is automatically saved to your Obsidian Vault.

## Roadmap
- Support for local file drag-and-drop.
- Built-in library view for history management.
- Multi-vault support.
- Custom analysis templates.

## Contributing
Contributions are welcome and appreciated.
1. Fork the Project.
2. Create your Feature Branch (git checkout -b feature/AmazingFeature).
3. Commit your Changes (git commit -m 'Add some AmazingFeature').
4. Push to the Branch (git push origin feature/AmazingFeature).
5. Open a Pull Request.

## License
Varys is released under the MIT License. See LICENSE file for details.
