# Requirements

## Overview
AudioDrive is a lightweight, single-binary podcast server designed for self-hosting audiobooks or personal podcasts. It serves an RSS feed compatible with major podcast apps (Apple Podcasts, Overcast, etc.) and provides a simple web dashboard for file management.

## Core Features

### 1. Security
*   **RSS Feed Protection:**
    *   The RSS feed (`/rss.xml`) and audio files must be protected by a "Secret Token".
    *   Access is granted via a query parameter: `?token=YOUR_SECRET`.
    *   Requests without the valid token must be rejected (401/403).
*   **Admin Dashboard Protection:**
    *   The dashboard (`/`) and API endpoints (`/upload`, `/api/*`) must be protected via HTTP Basic Auth.
    *   Admin credentials (password) are set via a command-line flag or environment variable.
*   **Rate Limiting:**
    *   Implement basic in-memory rate limiting (e.g., per IP) to prevent brute-force attacks on tokens or passwords.

### 2. Storage & Metadata
*   **Filesystem First:** Audio files are stored in a local directory (`./audio` by default).
*   **Metadata Storage:**
    *   Use a JSON sidecar file (`metadata.json`) in the audio directory to store editable fields (Title, Description).
    *   **Abstraction:** The code must use a `Store` interface to allow future swapping to SQLite without refactoring the business logic.
*   **RSS Generation:**
    *   Iterate over files in the directory.
    *   If metadata exists in the `Store` for a file, use it.
    *   If no metadata exists, fall back to the filename for the Title/Description.
    *   Support large files (Audiobooks).

### 3. File Management (Dashboard)
*   **Technology:** Pure HTML/CSS/JS embedded in the Go binary (`go:embed`). No React, no external CDNs.
*   **Uploads:**
    *   Support uploading large files (limit set to **800MB**).
    *   Show a progress bar during upload.
*   **Editing:**
    *   Allow the user to edit the `Title` and `Description` of an existing file.

## Technical Constraints
*   **Single Binary:** The final artifact must be a single executable.
*   **Lightweight:** Minimal dependencies.
*   **Concurrency:** The metadata store and server handles must be concurrency-safe.
