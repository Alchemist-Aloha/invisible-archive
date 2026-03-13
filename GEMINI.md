# Project Plan: Invisible Archive

## 1. Core Philosophy & Design Principles
*   **Zero-Copy Streaming:** Never load full files into RAM. Use `io.Pipe` and `io.Copy` to move bytes from disk/archive directly to the HTTP response.
*   **Path Transparency:** The UI and API should treat `/path/to/archive.zip/folder/image.jpg` exactly like a standard file system path.
*   **NAS-Friendly Indexing:** Use a single SQLite file (WAL mode) to store the directory structure. Indexing is lazy and opportunistic to minimize IO.
*   **On-Demand Processing:** Generate thumbnails only when a file enters the viewport.
*   **Reference-Counted Resource Management:** Implement a Mount Table using an LRU cache. Open archives remain open as long as a stream is active, ensuring stability for media sessions.

---

## 2. Technical Stack
*   **Backend:** Go (Golang) 1.24+
    *   `afero`: Unified VFS abstraction for transparent archive/OS path handling.
    *   `hashicorp/golang-lru/v2`: LRU cache for Mount Table resource management.
    *   `archive/zip`: Standard library for high-performance ZIP handling.
    *   `modernc.org/sqlite`: CGO-free SQLite for portable metadata indexing.
    *   `disintegration/imaging`: Pure Go image processing for thumbnails (replaced libvips for zero-dependency build).
    *   `Chi (v5)`: Minimalist, standard-compatible HTTP router.
    *   `fsnotify`: Active directory watching for real-time library updates.
    *   `sqlc`: Type-safe Go code generation from SQL schemas and queries.
*   **Frontend:** Vue 3 (Vite)
    *   `Tailwind CSS v4`: Modern utility-first styling.
    *   `TanStack Vue Query`: Robust server-state management and caching.
    *   `TanStack Vue Virtual`: Efficient rendering of large file lists (100k+ items).
    *   `VueUse`: Swipe navigation gestures for mobile devices (`useSwipe`).
    *   `Vite Plugin PWA`: Mobile installability and Service Worker management.
    *   `Plyr`: Modern, accessible video player integration.
    *   `Lucide Vue Next`: Clean and consistent icon set.
*   **Deployment:** Docker (Multi-stage build).
    *   Alpine-based runtime for minimal image footprint.
    *   Optimized build context via `.dockerignore`.

---

## 3. Implementation Phases

### Phase 1: The Smart VFS Engine (Backend MVP)
**Goal:** Create a "Stackable" VFS using Afero.
1.  **Path Resolver (Longest Physical Match):** 
    *   Walk the OS path first. The first segment that is a *file* marks the boundary where the virtual ZIP path begins.
2.  **Mount Table:** 
    *   LRU cache with reference counting for active streams.
3.  **Capabilities-Aware API (`/api/ls`):**
    *   Return a JSON list where each item has a `capabilities` bitmask:
        *   `1 (browse)`: Can be opened as a folder (Directory or ZIP).
        *   `2 (stream)`: Media file supporting Range requests.
        *   `4 (render)`: Image/PDF for thumbnailing.
        *   `8 (edit)`: Text/Code file for Monaco.
4.  **Streaming API (`/api/raw`):**
    *   Standard `http.ServeContent` for robust Range request handling.

### Phase 2: Metadata Indexing (The "Hybrid" Strategy)
**Goal:** Low-IO discovery and search.
1.  **Lazy Discovery:** Index folders and ZIP central directories only when accessed by the user.
2.  **Active Watching (NAS-Optimized):** 
    *   Only use `fsnotify` on the currently viewed directory and its children. 
    *   Dynamically "unwatch" folders as the user navigates away to stay under `max_user_watches` limits.
3.  **SQLite FTS5:** Ultra-fast search across indexed paths.

### Phase 3: The "Finder" UI
**Goal:** Responsive, "seamless" archive browsing.
1.  **Navigation System:**
    *   URL-driven navigation: `app.com/browse/archive.zip/internal/path`.
    *   TanStack Router for state-consistent breadcrumbs.
2.  **Virtual Grid View:**
    *   Render 100k+ items using `TanStack Virtual`.
    *   "Quick Look" triggered by Spacebar using the `capabilities` bitmask to select the player.

### Phase 4: The Image Pipeline (Optimized for Weak CPUs)
**Goal:** Fast previews without CPU spikes.
1.  **Fast Identity Caching:**
    *   Cache keys generated via `Path + Size + ModTime` ($O(1)$ calculation) to avoid heavy hashing.
2.  **Priority Worker Pool:**
    *   **High Priority:** Images in current viewport.
    *   **Low Priority:** Prefetching next/previous items.
3.  **WebP-First Strategy:**
    *   Default to WebP encoding for the best balance of speed and compression on weak CPUs.

### Phase 5: Hardening & Docker
1.  **Security:** Sanitize internal ZIP paths ("Zip Slip") and enforce read-only mounts.
2.  **Dockerization:** Alpine-based multi-stage build with `libvips` dependencies.

---

## 4. NAS-Specific Optimization Strategy

| Feature | Strategy | Why? |
| :--- | :--- | :--- |
| **Search** | SQLite FTS5 | Instant search with minimal CPU/RAM. |
| **Disk IO** | Active Watching | Avoids crashing NAS via inotify limits. |
| **RAM** | LRU + RefCount | Protects memory while keeping archives "warm." |
| **Cache** | Fast Identity | Avoids CPU-heavy SHA256 hashing. |
| **Video** | `ServeContent` | Native Go implementation of seekable streaming. |
