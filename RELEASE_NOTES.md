# Release Notes

## v0.4.5 (March 2026)

**Focus:** Search Visibility & UX.

### Improvements
*   **Upload Date in Search Results:** Search results (yt-dlp) now display the actual upload date of the content, helping users identify the most relevant and recent videos.
*   **Animated Search Interface:** The CLI search interface now features an animated progress bar and spinner, providing real-time feedback during detailed metadata extraction.
*   **Robust Search Parsing:** Improved yt-dlp JSON parsing with increased buffers and support for multiple timestamp formats to handle complex metadata reliably.
*   **Session-Aware Extraction:** Integrated browser cookies support for search to ensure reliable access to detailed video metadata without bot interference.

## v0.4.4 (March 2026)

**Focus:** Workflow Integration & Ease of Access.

### Improvements
*   **One-Click Note Opening (GUI):** Successfully completed tasks now show a clickable result message. Clicking it immediately opens the generated Markdown note in Obsidian (or your default editor).
*   **CLI Auto-Open:** Added a new `--open` (or `-o`) flag to the CLI. When enabled, the generated note will automatically open upon task completion.
*   **Cross-Platform File Opening:** Implemented a robust, cross-platform file opening mechanism for both GUI and CLI.

## v0.4.3 (March 2026)

**Focus:** Flexible Prompting & Global Consistency.

### Improvements
*   **Unified Default Language:** Established "English" as the global default target language in `backend/config`, ensuring consistency across all modules.
*   **Flexible Prompt Templates:** Replaced fragile `fmt.Sprintf` with a robust template system. Users can now use `{{.Language}}` and `{{.Content}}` placeholders in custom prompts.
*   **Enhanced Prompt Logging:** The CLI and GUI now log the fully rendered prompt before starting AI analysis, significantly easing the debugging of custom prompts.
*   **Smart Custom Prompts:** Custom prompts now respect the user's intent without forced global language injection, while still supporting automatic content placement.

---

## v0.4.2 (March 2026)

**Focus:** Security & Pipeline Robustness.

### Security
*   **Secure Secret Storage:** Integrated OS Keyring (Keychain on macOS, Credential Manager on Windows) to store API keys. Secrets are no longer stored in plain text in `config.json`.
*   **Automatic Migration:** Existing plain-text keys are automatically moved to secure storage upon first run.

### Ingestion & Translation
*   **Multi-Provider Translation:** Refactored the translation engine to support any AI provider (OpenAI or Ollama), aligning with the analysis settings.
*   **Smart Language Detection:** Scraper now detects the source language of web pages via HTML tags, automatically skipping redundant translations.
*   **Whisper Optimization:** Added entropy and logprob thresholds to suppress hallucinations and infinite loops during transcription.

### Testing
*   **Mock Storage:** Implemented a mockable secret store to ensure unit tests run safely in non-interactive/CI environments.

---

## v0.4.1 (Feb 2026)

**Focus:** UX Consistency & CLI Messaging.

### Improvements
*   **Unified Health UX:** Removed the duplicated startup dependency modal and kept dependency checks in Settings as the single source of truth.
*   **About Modal Consistency:** Made version badges in both Dashboard and Settings open the same About modal.
*   **Clickable GitHub CTA:** Ensured the GitHub button in About reliably opens the repository in the default browser.
*   **CLI Copy Refresh:** Updated CLI help text to reflect desktop+CLI scope and refined AI provider wording.

---

## v0.4.0 (Feb 2026)

**Focus:** GUI Visual Overhaul & Brand Identity.

### Visual & Brand Updates
*   **Integrated Brand Identity:** Extracted a custom color palette (`#290137`) from the project logo to create a cohesive, professional dark theme across the entire application.
*   **Unified Branded Controls:** Integrated the small Varys logo directly into the settings gear icon, featuring a 90-degree rotation animation on hover/active states.
*   **Immersive "About" UI:** Implemented a new full-screen "About" modal with a high-fidelity brand background (`varys.png`), Backdrop blur, and direct links to the GitHub repository.
*   **Simplified Navigation:** Replaced the legacy tab-style navigation with a clean, header-based layout, providing more screen real estate for core tasks.
*   **UX Stability:** Eliminated layout "jumps" when starting a task by making the headline section static and optimizing vertical spacing.
*   **Contextual Versioning:** Moved the version badge from the global footer into the System Logs console for a more integrated, tidy look.

### Engineering & DX
*   **Quiet Build System:** Optimized the `Makefile` to suppress verbose build logs from Wails and Go. Installation now features a clean progress-based output with a success summary.
*   **Window Optimization:** Updated default window dimensions to 1024x768 and synced the initial background color to prevent flicker during app launch.
*   **Code Quality:** Cleaned up unused imports and refactored navigation state management in `App.tsx`.
