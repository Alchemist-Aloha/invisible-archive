# Invisible Archive Agent Guide

This document is for AI coding agents working on this repository.

## Project Goal

Build a high-performance, self-hosted file browser that treats ZIP archives as normal folders, so users can browse, search, and stream files inside archives without extraction.

Success criteria:
- ZIP path transparency: physical and archive-internal paths behave consistently in API and UI.
- Seekable media streaming: HTTP Range support works for large files and files inside ZIP archives.
- NAS-friendly operation: low memory pressure, controlled CPU, and low unnecessary disk IO.
- Responsive UI at scale: large directory rendering and smooth preview experience on desktop and mobile.

## Agent Mission

When making changes, optimize for correctness and operational stability first, then performance and UX.

Primary rule:
- Preserve archive transparency end-to-end. If a change breaks archive-as-folder behavior, it is a regression.

## Architecture Snapshot

- Backend: Go service exposing file list, search, thumbnail, and raw stream endpoints.
- VFS core: path peeling plus archive mount cache to bridge OS paths and ZIP virtual paths.
- Indexing: SQLite metadata index with FTS, fed lazily from directory/zip reads and watcher updates.
- Frontend: Vue app with virtualized listing, preview flows, and mobile-first interactions.
- Deployment: multi-stage Docker image and compose setup for self-hosted NAS-like environments.

## Non-Negotiable Invariants

1. Paths
- Input paths may contain spaces, brackets, and encoded characters.
- Decoding and normalization must not break valid filenames.
- Path traversal outside library root must never be possible.

2. Streaming
- Raw endpoint must keep proper content type and Range behavior.
- Avoid loading full media into memory; prefer streaming interfaces.

3. Archive lifecycle
- Mounted archives use cache plus reference counting semantics.
- Active readers must not be invalidated by cache eviction.

4. Indexing
- Index updates should be incremental/lazy where possible.
- Search should remain bounded and performant.

5. UX behavior
- Auto-enter logic for single-root ZIP archives should remain consistent.
- Grid/list/details and preview navigation should stay functional on mobile and desktop.

## Stack and Why It Was Chosen

- Go 1.24+: predictable performance, good concurrency, strong stdlib.
- Chi: lightweight routing and middleware composition.
- Afero: filesystem abstraction for transparent VFS behavior.
- archive/zip: native ZIP support.
- hashicorp/golang-lru/v2: cache eviction with bounded resource usage.
- modernc.org/sqlite + sqlc: portable typed data layer.
- fsnotify: near-real-time metadata refresh.
- imaging: pure-Go thumbnail generation.
- Vue 3 + Vite + TypeScript: fast UI iteration with typed contracts.
- TanStack Query + Virtual: reliable async state and scalable list rendering.
- Tailwind v4: rapid, consistent UI styling.
- PhotoSwipe + Plyr: robust media preview UX.
- Docker multi-stage: reproducible deploy with small runtime image.

## Repository Landmarks

- Server entry: cmd/server/main.go
- API handlers: internal/api
- VFS engine: internal/vfs
- Data/indexing layer: internal/data
- Shared capability flags: pkg/util/capabilities.go
- Frontend app and components: frontend/src
- UI design decisions: UI.md
- Stack summary: STACK.md
- File map: STRUCTURE.md

## Agent Workflows

### A. Backend behavior change

1. Identify invariant touched (path, stream, cache, index).
2. Update minimal code surface in internal/api, internal/vfs, or internal/data.
3. Add or adjust tests nearest to change:
   - internal/api/handlers_test.go
   - internal/vfs/*_test.go
4. Validate with targeted go tests before broad runs.

Definition of done:
- No path transparency regressions.
- Range and preview-related behavior still work.
- Existing tests pass for touched modules.

### B. Frontend UX or API integration change

1. Keep API contract compatibility with frontend/src/api.ts and backend responses.
2. Preserve virtualization and responsive behavior in listing components.
3. Verify preview behavior for image/video/text and back-navigation consistency.

Definition of done:
- No broken navigation states.
- Large list performance characteristics unchanged or improved.
- Mobile interaction paths still functional.

### C. Mobile (Flutter) development

1. **Path Handling**: Always preserve leading slashes for absolute VFS paths. Ensure `Uri.encodeComponent` is used for all path-based API queries.
2. **State Management**: Use `Provider`. Keep business logic in `ExplorerProvider` or `SettingsProvider` and UI in `pages/` or `widgets/`.
3. **Media Integration**:
   - Use `chewie` for video playback to leverage native ExoPlayer/AVPlayer UI.
   - Use `photo_view` for high-performance image zooming and gallery transitions.
   - Use `cached_network_image` for all network-bound visual assets.
4. **Layout**: Maintain support for Grid, List, and Details modes. Ensure the Waterfall View correctly handles recursive archive crawling.

Definition of done:
- `flutter build apk` succeeds using **Java 21**.
- Video controls are functional and fullscreen works.
- ZIP transparency is maintained during navigation.

### D. Data/index/search change

1. Update schema.sql and queries.sql first.
2. Regenerate sqlc outputs when query/schema changes are intentional.
3. Ensure indexer code aligns with new query contracts.

Definition of done:
- Search/list remain fast and correct.
- No mismatch between generated query code and runtime usage.

## Practical Commands

Backend:
- go mod download
- go run cmd/server/main.go
- go test ./...

Frontend:
- cd frontend
- npm install --legacy-peer-deps
- npm run dev
- npm run build

Docker:
- docker-compose up -d

UI Testing (Playwright):
- Use `webapp-testing` skill for browser-based verification.
- To run a test with managed servers:
  ```bash
  python3 scripts/with_server.py \
    --server "go run cmd/server/main.go" --port 8080 \
    --server "cd frontend && npm run dev" --port 5173 \
    -- python3 your_test.py
  ```
- Use `playwright-skill` for custom automation scripts in `/tmp`.
- Always test both **Desktop** and **Mobile (iPhone 13 simulation)** for gesture-heavy features.

## Common Risk Areas

- URL encode/decode mismatches between frontend and backend raw path handling.
- Archive reader lifecycle leaks or premature close during seeks.
- Watcher churn on large trees causing resource pressure.
- UI regressions in history/popstate behavior across image and non-image preview modes.
- Interaction logic (like drag-to-seek) breaking due to missing touch-action rules or threshold logic.

## Decision Rules for Agents

- Prefer small, targeted patches over wide refactors.
- Preserve public API shapes unless explicitly changing contracts.
- Add tests for bug fixes when feasible.
- Use Playwright to verify complex UI interactions or gesture-heavy fixes.
- If behavior is ambiguous, align with project goal: archive transparency plus NAS-friendly performance.

## Future Priorities

- Harden path and archive security checks further.
- Improve watcher strategy for very large libraries.
- Expand test coverage for edge cases in encoded and nested archive paths.
