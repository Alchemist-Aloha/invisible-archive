# Invisible Archive Tech Stack

This document summarizes the technology stack selected in this project and the key highlights of each choice.

## Backend Runtime

- Go 1.24+
  - Highlight: high-performance, low-overhead server runtime with strong standard library support.
  - Highlight: good fit for streaming, filesystem operations, and concurrency-heavy workloads.

## Backend Libraries

- Chi v5 (HTTP routing)
  - Highlight: minimal router with clean middleware model.
  - Highlight: works well with URL-decoding paths containing spaces and brackets.

- Afero (filesystem abstraction)
  - Highlight: unified interface over OS filesystem and virtualized archive access.
  - Highlight: supports the project goal of archive paths behaving like normal folders.

- archive/zip (Go stdlib)
  - Highlight: native ZIP handling without external runtime dependencies.
  - Highlight: enables transparent traversal and streaming from archive entries.

- hashicorp/golang-lru/v2 (mount cache)
  - Highlight: LRU-based cache for open archive handles.
  - Highlight: pairs with reference counting to keep active media streams stable.

- modernc.org/sqlite (pure-Go SQLite driver)
  - Highlight: CGO-free database access for portable builds and containers.
  - Highlight: used with WAL mode for better concurrent read/write behavior.

- fsnotify (filesystem watcher)
  - Highlight: supports low-latency metadata updates from filesystem events.
  - Highlight: used in a NAS-conscious indexing workflow.

- disintegration/imaging + golang.org/x/image (image pipeline)
  - Highlight: pure-Go image decoding/resizing.
  - Highlight: supports on-demand thumbnail generation without native dependencies.

## Data and Query Layer

- SQLite + FTS5
  - Highlight: single-file metadata index optimized for NAS environments.
  - Highlight: fast search over indexed file and virtual archive paths.

- sqlc
  - Highlight: type-safe Go code generation from SQL schema and query files.
  - Highlight: keeps SQL explicit while reducing query boilerplate and runtime mapping errors.

## Frontend Application

- Vue 3 + TypeScript
  - Highlight: component-based UI with strict typing for API contracts and interactions.
  - Highlight: maintainable stateful UI for browsing, preview, and navigation workflows.

- Vite
  - Highlight: fast local development and modern production bundling.
  - Highlight: clean plugin ecosystem used by this project.

- Tailwind CSS v4 + PostCSS
  - Highlight: utility-first styling with theme variables and dark-mode variant support.
  - Highlight: fast iteration on dense responsive file-browser layouts.

- @tanstack/vue-query
  - Highlight: server-state caching and refetch control for listing/search endpoints.
  - Highlight: reduces manual async state handling complexity.

- @tanstack/vue-virtual
  - Highlight: virtualization support for very large directories (100k+ item scenarios).
  - Highlight: keeps DOM size and scrolling performance stable.

- @vueuse/core
  - Highlight: composables for pointer/swipe interactions used by preview UX.

- PhotoSwipe v5
  - Highlight: high-performance image gallery with gesture-friendly navigation.

- Plyr
  - Highlight: modern media player UI for streamed video playback.

- lucide-vue-next
  - Highlight: consistent icon set across navigation and file-type affordances.

- Axios
  - Highlight: lightweight HTTP client wrapper for API endpoints.

## PWA and Delivery

- vite-plugin-pwa
  - Highlight: installable app behavior with generated service worker + manifest integration.
  - Highlight: improves mobile usability and perceived app-native experience.

- Docker (multi-stage build)
  - Highlight: separate frontend and backend build stages for clean, reproducible images.
  - Highlight: Alpine runtime image keeps deployment footprint small.

- Docker Compose
  - Highlight: simple local/self-hosted deployment with mounted media library and cache volumes.

## Architecture Highlights (Why this stack works here)

- Path transparency first: archive paths are treated as browseable filesystem paths.
- Streaming-first backend: raw file endpoint supports HTTP Range for seekable media playback.
- NAS-friendly indexing: opportunistic indexing with SQLite WAL and targeted watching strategy.
- Resource control: LRU + refcount archive mounting avoids memory blowups while preserving stream stability.
- On-demand media pipeline: thumbnail generation and heavy work only happen when needed.
