# Invisible Archive

Invisible Archive is a high-performance, self-hosted file management system designed to treat ZIP archives as transparent directories. Browse, stream, and search content inside archives without ever having to extract or download them.

![Architecture](https://img.shields.io/badge/Architecture-VFS--First-blue)
![Backend](https://img.shields.io/badge/Backend-Go-00ADD8)
![Frontend](https://img.shields.io/badge/Frontend-Vue%203-4FC08D)
![Database](https://img.shields.io/badge/Database-SQLite%20FTS5-003B57)

## 🚀 Key Features

- **Transparent Archive Browsing:** Navigate `.zip` files as if they were standard folders. The **Path Peeler** algorithm seamlessly resolves virtual paths.
- **Auto-Enter Archives:** Automatically navigates into archives containing a single root folder, eliminating redundant clicks.
- **Waterfall Discovery Mode:** Recursive, visual browsing of images under any path, with full support for crawling inside archives.
- **NAS-Optimized Performance:** 
    - **Reference-Counted LRU Cache:** Keeps frequently accessed archives "warm" while strictly protecting system memory.
    - **Hybrid Indexing:** Instant search (<5ms for 1M files) using SQLite FTS5. Indexing is lazy and opportunistic to minimize disk IO.
- **High-Performance Streaming:** Native support for HTTP Range requests (206 Partial Content), allowing for $O(1)$ seeking in uncompressed media inside ZIPs.
- **Multi-Layout Engine:** Choose between **Grid** (large thumbnails), **List** (compact), or **Details** (list with metadata) views.
- **Global Dark Mode:** Full dark mode support with a single toggle, respecting system preferences and persistent storage.
- **Android Mobile App:** Dedicated Flutter-based Android application with advanced media playback and ZIP path transparency.
- **Smart Image Pipeline:** Throttled, `libvips`-powered thumbnail generation with a **Fast Identity** cache system to prevent NAS CPU spikes.
- **Mobile-First Experience:** Fully **installable PWA** with multi-size icon support and native-like touch gestures, including pinch-to-zoom and mobile-optimized video controls.
- **Modern UI:** Responsive "Finder-style" interface built with Vue 3 and Tailwind CSS v4, featuring virtual scrolling for directories with 100,000+ items.

## ⚙️ How It Works: Path Peeling

Invisible Archive uses a custom "Path Peeling" algorithm to bridge the gap between physical and virtual filesystems. When you request a path like `/photos/2023.zip/summer/beach.jpg`, the engine "peels" the path from left to right:
1. It identifies `/photos/2023.zip` as a physical file on disk.
2. It treats the remaining part `/summer/beach.jpg` as a virtual path inside the archive.
3. It mounts the ZIP into an LRU cache and streams the file directly to the client.

This process is recursive and transparent, allowing for infinite nesting (e.g., a ZIP inside another ZIP) without extraction.

## 🛠 Tech Stack

### Backend (Go)
- **VFS:** `afero` for unified filesystem abstraction.
- **Routing:** `chi` with auto-unescaping for special character support (`[]`, spaces, etc).
- **Database:** Pure-Go `SQLite` with `sqlc` for type-safe queries.
- **Imaging:** `libvips` via `govips` for high-speed, memory-efficient processing.

### Frontend (Vue 3)
- **State Management:** `TanStack Query` for robust server-state synchronization.
- **Virtualization:** `TanStack Virtual` for high-density list rendering.
- **Image Viewer:** `PhotoSwipe v5` for high-performance, gesture-driven browsing.
- **Video Player:** Native HTML5 player with a custom **Gesture Seek Engine** (via `@vueuse/core`). Supports horizontal swipe-to-seek with a real-time HUD and visual time-delta overlays.
- **Styling:** `Tailwind CSS v4` with class-based dark mode.

### Mobile (Flutter)
- **Framework:** `Flutter 3.x` for high-performance Android builds.
- **State:** `Provider` for reactive UI updates and settings persistence.
- **Video:** `chewie` + `video_player` for a feature-rich playback UI using native ExoPlayer/AVPlayer.
- **Image:** `photo_view` + `cached_network_image` for native-like gallery behavior and zoom.

## 📦 Deployment

### Using Docker Compose (Recommended)

The easiest way to run Invisible Archive is via Docker Compose.

1.  **Configure environment:** Create a `.env` file in the root directory:
    ```env
    APP_PORT=8881
    LIBRARY_DIR=/path/to/your/media
    CACHE_DIR=/path/to/your/cache
    ```
2.  **Start the application:**
    ```bash
    docker-compose up -d
    ```

### Environment Variables

| Variable | Description | Default |
| :--- | :--- | :--- |
| `APP_PORT` | The port on your host machine to access the web UI. | `8881` |
| `LIBRARY_DIR` | The absolute path to your media library on the host machine. | (Required) |
| `CACHE_DIR` | The absolute path to the cache directory for thumbnails. | (Required) |
| `DB_PATH` | Path to the SQLite database file inside the container. | `/cache/archive.db` |
| `THUMB_WORKERS` | Maximum concurrent thumbnail generation workers. | `1` |
| `PORT` | The internal port the server listens on inside the container. | `8080` |

## 🔒 Security

Invisible Archive is built with security as a priority:
- **Path Sanitization:** All input paths are normalized and checked for traversal attacks (`..`).
- **Read-Only VFS:** The media library is mounted as read-only, ensuring no accidental modifications to your source files.
- **Scoped Access:** Path resolution is strictly bounded by the defined `LIBRARY_PATH`.

## 🛠 Local Development

### Prerequisites
- Go 1.24+
- Node.js 20+
- libvips (required for backend)

### Backend
```bash
go mod download
go run cmd/server/main.go
```

### Frontend
```bash
cd frontend
npm install --legacy-peer-deps
npm run dev
```

### Mobile
```bash
cd mobile
flutter pub get
flutter build apk --release
```

## ⚖️ License

MIT
