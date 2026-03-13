# Invisible Archive

Invisible Archive is a high-performance, self-hosted file management system designed to treat ZIP archives as transparent directories. Browse, stream, and search content inside archives without ever having to extract or download them.

![Architecture](https://img.shields.io/badge/Architecture-VFS--First-blue)
![Backend](https://img.shields.io/badge/Backend-Go-00ADD8)
![Frontend](https://img.shields.io/badge/Frontend-Vue%203-4FC08D)
![Database](https://img.shields.io/badge/Database-SQLite%20FTS5-003B57)

## 🚀 Key Features

- **Transparent Archive Browsing:** Navigate `.zip` files as if they were standard folders. The "Path Peeler" algorithm seamlessly resolves virtual paths.
- **Auto-Enter Archives:** Automatically navigates into archives containing a single root folder, eliminating redundant clicks.
- **NAS-Optimized Performance:** 
    - **Reference-Counted LRU Cache:** Keeps frequently accessed archives "warm" while strictly protecting system memory.
    - **Hybrid Indexing:** Instant search (<5ms for 1M files) using SQLite FTS5. Indexing is lazy and opportunistic to minimize disk IO.
- **High-Performance Streaming:** Native support for HTTP Range requests (206 Partial Content), allowing for $O(1)$ seeking in videos and large media files.
- **Multi-Layout Engine:** Choose between **Grid** (large thumbnails), **List** (compact), or **Details** (list with metadata) views.
- **Global Dark Mode:** Full dark mode support with a single toggle, respecting system preferences and persistent storage.
- **Smart Image Pipeline:** Throttled, pure-Go thumbnail generation with a "Fast Identity" cache system to prevent NAS CPU spikes.
- **Mobile-First Experience:** Fully **installable PWA** with multi-size icon support and native-like touch gestures, including pinch-to-zoom and mobile-optimized video controls.
- **Modern UI:** Responsive "Finder-style" interface built with Vue 3 and Tailwind CSS v4, featuring virtual scrolling for directories with 100,000+ items.

## 🛠 Tech Stack

### Backend (Go)
- **VFS:** `afero` for unified filesystem abstraction.
- **Routing:** `chi` with auto-unescaping for special character support (`[]`, spaces, etc).
- **Database:** Pure-Go `SQLite` with `sqlc` for type-safe queries.
- **Imaging:** `disintegration/imaging` for portable, CGO-free processing.

### Frontend (Vue 3)
- **State Management:** `TanStack Query` for robust server-state synchronization.
- **Virtualization:** `TanStack Virtual` for high-density list rendering.
- **Image Viewer:** `PhotoSwipe v5` for high-performance, gesture-driven browsing.
- **Video Player:** `Plyr` with mobile-optimized progress controls.
- **Styling:** `Tailwind CSS v4` with class-based dark mode.

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

### Environment Variables (.env)

| Variable | Description |
| :--- | :--- |
| `APP_PORT` | The port on your host machine to access the web UI. |
| `LIBRARY_DIR` | The absolute path to your media library on the host machine (mounted as Read-Only). |
| `CACHE_DIR` | The absolute path to the cache directory on the host machine. |

## 🛠 Local Development

1. **Backend:**
   ```bash
   go mod download
   go run cmd/server/main.go
   ```

2. **Frontend:**
   ```bash
   cd frontend
   npm install --legacy-peer-deps
   npm run dev
   ```

## ⚖️ License

MIT
