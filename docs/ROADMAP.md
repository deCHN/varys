# Varys Roadmap

This document outlines the strategic milestones for Varys, moving from a developer tool to a polished commercial desktop application.

## Milestones Overview

| Milestone | Version | Focus | Status | Target Date |
| :--- | :--- | :--- | :--- | :--- |
| **Foundation & Rebranding** | `v0.3.0` | Core Pipeline, UI Modernization, Rebranding | Completed | Jan 2026 |
| **GUI & Visual Overhaul** | `v0.4.0` | Brand Integration, Logo, Simplified Navigation | Completed | Feb 2026 |
| **Packaging & UX Polish** | `v0.5.0` | Dependency Bundling, Auto-Updates, Signed Build | In Progress | Mar 2026 |

---

## 1. Milestone: Foundation & Rebranding (v0.3.0) - COMPLETED

Establish the core architecture and complete the transition from "v2k" to "Varys".

*   **Status:** 100% Complete
*   **Key Features:**
    *   One-Click Capture (YouTube/Bilibili).
    *   Local Transcription (Whisper) & Analysis (Ollama).
    *   Initial Tailwind CSS UI.
    *   Obsidian Integration.

---

## 2. Milestone: GUI & Visual Overhaul (v0.4.0) - COMPLETED

A major visual and brand identity update to provide a professional, seamless experience.

*   **Status:** 100% Complete
*   **Key Features:**
    *   **Brand Integration:** Extracted custom color palette (`#290137`) from logo for a unified look.
    *   **Integrated Logo:** Small Varys logo integrated into the settings gear for a unique branded control.
    *   **Simplified Navigation:** Removed tab-style buttons in favor of a clean, header-based navigation.
    *   **Immersive About UI:** Created a high-fidelity "About" modal with full-window brand background.
    *   **UX Stability:** Eliminated layout jumps during processing by keeping headlines static.
    *   **Version Tracking:** Integrated version numbers directly into the System Logs console.

---

## 3. Milestone: Packaging & UX Polish (v0.5.0)

Eliminate "Developer Friction" to make the app installable and usable by non-technical users.

*   **Status:** In Progress
*   **Key Tasks:**
    *   **Dependency Bundling:** Embed `ffmpeg` and `yt-dlp` binaries so Homebrew is not required.
    *   **Onboarding Wizard:** First-run guide to check/install Ollama and download Whisper models automatically.
    *   **Code Signing:** Notarize the macOS app to remove "Unverified Developer" warnings.
    *   **Auto-Update:** Integrate GitHub Releases poller for seamless background updates.
