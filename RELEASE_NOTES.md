# Release Notes

## v0.4.0 (Feb 2026)

**Focus:** GUI Visual Overhaul & Brand Identity.

### Visual & Brand Updates
*   **Integrated Brand Identity:** Extracted a custom color palette (`#290137`) from the project logo to create a cohesive, professional dark theme across the entire application.
*   **Unified Branded Controls:** Integrated the small Varys logo directly into the settings gear icon, featuring a 90-degree rotation animation on hover/active states.
*   **Immersive "About" UI:** Implemented a new full-screen "About" modal with a high-fidelity brand background (`varys.png`),毛玻璃模糊效果 (Backdrop blur), and direct links to the GitHub repository.
*   **Simplified Navigation:** Replaced the legacy tab-style navigation with a clean, header-based layout, providing more screen real estate for core tasks.
*   **UX Stability:** Eliminated layout "jumps" when starting a task by making the headline section static and optimizing vertical spacing.
*   **Contextual Versioning:** Moved the version badge from the global footer into the System Logs console for a more integrated, tidy look.

### Engineering & DX
*   **Quiet Build System:** Optimized the `Makefile` to suppress verbose build logs from Wails and Go. Installation now features a clean progress-based output with a success summary.
*   **Window Optimization:** Updated default window dimensions to 1024x768 and synced the initial background color to prevent flicker during app launch.
*   **Code Quality:** Cleaned up unused imports and refactored navigation state management in `App.tsx`.

---

## v0.3.8 (Jan 2026)

**Focus:** User Control & Aesthetics.

### New Features
*   **Play/Stop UI:** Replaced the "Process" text button with toggleable Play and Stop icons for a modern look.
*   **Task Cancellation:** Users can now abort a running task by clicking the Stop button. The backend now supports graceful interruption via context cancellation, ensuring temporary files are immediately cleaned up.

---

## v0.3.7 (Jan 2026)

**Focus:** Reliability & Data Quality.

### Fixes
*   **Robust Translation:** Fixed an issue where the small 0.6B model would generate empty or malformed translation tables. The new logic uses text-based batching (line-by-line) instead of fragile JSON parsing, ensuring 99% success rate.
*   **Analysis Language:** Stricter prompting now enforces the "Target Language" setting for summaries and key points, preventing English output when Chinese is requested.
*   **Tag Sanitization:** Obsidian tags are now automatically sanitized (spaces replaced with hyphens) to ensure they are clickable and valid.

---

## v0.3.6 (Jan 2026)

**Focus:** Architecture & Maintainability.

### Refactoring
*   **Modular Translation:** Extracted translation logic into a dedicated `backend/translation` module. This decoupling allows for easier maintenance, independent testing, and cleaner service initialization in `app.go`.
*   **Integration Tests:** Added a new integration test suite for the translation module to verify real-world interactions with Ollama.

---

## v0.3.5 (Jan 2026)

**Focus:** Speed and Efficiency.

### New Features
*   **Fast Translation Model:** Introduced a dedicated `Translation Model` setting (default: `qwen3:0.6b`). This smaller, specialized model speeds up translation by **2.5x** without sacrificing analysis quality (which still uses the larger 8B model).

### Performance
*   **Optimized Pipeline:** The dual-model architecture allows for deep analysis and rapid translation in a single pass.
*   **Benchmarks:** Verified ~60s processing time for 10-minute videos (down from ~10m).

---

## v0.3.4 (Jan 2026)

**Focus:** Performance Tuning and Developer Experience.

### New Features
*   **Custom Context Size:** Users can now adjust the LLM context window (4k, 8k, 16k, 32k) via Settings. This allows users with 8GB RAM to run the app safely (using 4k/8k) while power users can maximize analysis depth with 32k.

### Engineering
*   **Performance Benchmarks:** Established a baseline performance test suite (`backend/benchmarks`) to measure transcription and analysis speed across different context sizes.
*   **Test Coverage:** Improved unit test stability and coverage for the core backend modules.

---

## v0.3.3 (Jan 2026)

**Focus:** Stability and Polish.

### Fixes & Improvements
*   **Robust Translation:** Implemented "Chunked Translation" logic (2000-character blocks) to reliably translate long videos without hitting LLM context or output limits.
*   **Log Management:** Application logs are now saved to `~/.varys/logs/` to ensure writability on macOS, solving missing log file issues.
*   **System Logs UI:** Added a "Copy to Clipboard" button and improved text alignment for better debugging experience.
*   **Code Quality:** Cleaned up trailing whitespace across the entire codebase.

---

## v0.3.2 (Jan 2026)

**Focus:** Intelligence and Content Quality.

### New Features
*   **Target Language Selection:** Added a dropdown in Settings to choose the desired output language (e.g., English, Japanese, Spanish). Analysis and translation will now adapt to this setting.
*   **Smart Translation:** Automatically detects the source language of the video. If the source matches your target language (e.g., Chinese video -> Chinese target), translation is skipped to avoid redundancy.
*   **Side-by-Side Translation Table:** Replaced the block-based translation with a sentence-aligned Markdown table for easier comparison.

### Improvements
*   **Structured Analysis:** Enforced strict JSON parsing for LLM outputs, ensuring "Key Points", "Tags", and "Assessment" are always correctly populated in the final note.
*   **Reliability:** Fixed issues where translation tables would appear empty or fail to render.

---

## v0.3.1 (Jan 2026)

**Focus:** Feature enhancement and User Experience.

### New Features
*   **Video Download Toggle:** Added a "Video / Audio Only" switch to the Dashboard. Users can now choose to download the full video (`.mp4`) or just the audio (`.m4a`).
    *   *Default:* Video Mode (Downloads video + extracts audio for analysis).
    *   *Toggle:* Audio Only (Downloads only audio for faster processing).
*   **Integrated UI:** The toggle is seamlessly embedded into the URL input bar with a modern green aesthetic.

### Improvements
*   **Storage Logic:** Updated storage manager to preserve original file extensions, ensuring videos are correctly embedded in Obsidian notes (`![[video.mp4]]`).
*   **Dependencies:** Backend now supports generic media download and processing.
