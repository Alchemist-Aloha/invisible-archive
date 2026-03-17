# Invisible Archive Project Structure

This document maps the repository layout and explains the responsibility of each file.

## Root

- README.md: Main project overview, architecture summary, feature list, and local/deployment run instructions.
- GEMINI.md: High-level engineering plan and design goals (VFS-first browsing, indexing strategy, UI phases, NAS optimizations).
- UI.md: Frontend design log with visual, interaction, performance, and accessibility decisions.
- docker-compose.yml: Single-service deployment definition with volume mounts for library/cache and runtime env vars.
- Dockerfile: Multi-stage build (Vue frontend build + Go backend build + Alpine runtime image).
- go.mod: Go module definition and direct/indirect backend dependencies.
- go.sum: Dependency checksums for reproducible Go builds.
- sqlc.yaml: sqlc generation config pointing to schema/queries and Go output package.
- STRUCTURE.md: This file.

## cmd/server

- cmd/server/main.go: Backend entrypoint; initializes DB schema, indexer watcher, VFS manager, thumbnailer, middleware, API routes, and static file serving.

## internal/api

- internal/api/handlers.go: HTTP handlers for directory listing, search, and raw file streaming (with Range support via ServeContent).
- internal/api/thumb.go: Thumbnail generation and cache pipeline (identity hash key + bounded worker semaphore).
- internal/api/handlers_test.go: API behavior tests for special-character raw paths and ZIP auto-enter listing.

## internal/data

- internal/data/schema.sql: SQLite schema for items table, indexes, FTS5 virtual table, and sync triggers.
- internal/data/queries.sql: Source SQL queries for sqlc (upsert/list/search/delete/get-by-path).
- internal/data/db.go: sqlc-generated DB abstraction (DBTX interface, Queries wrapper, transaction helper).
- internal/data/models.go: sqlc-generated Go structs for DB rows.
- internal/data/queries.sql.go: sqlc-generated typed query implementations.
- internal/data/indexer.go: Filesystem/ZIP indexer, SQLite WAL setup, fsnotify watch loop, and metadata upsert logic.

## internal/vfs

- internal/vfs/peeler.go: Longest-physical-match path resolver that splits physical path vs virtual path-inside-zip.
- internal/vfs/mount_table.go: Reference-counted LRU cache for open ZIP archives.
- internal/vfs/vfs.go: Core virtual filesystem manager (open/stat/readdir, ZIP auto-enter, indexed search, seekable ZIP streaming).
- internal/vfs/peeler_test.go: Unit tests for path peeling across normal paths and nested archive paths.
- internal/vfs/mount_table_test.go: Unit tests for cache hit/eviction and archive lifecycle semantics.
- internal/vfs/vfs_test.go: End-to-end VFS tests for listing, reading, and seeking within ZIP content.

## pkg/util

- pkg/util/capabilities.go: File capability bitmask detection (browse/stream/render/edit) based on type and extension.

## frontend (Vite + Vue 3 + TypeScript)

### Build and toolchain

- frontend/package.json: Frontend scripts and dependency manifest.
- frontend/package-lock.json: npm lockfile for deterministic npm installs.
- frontend/pnpm-lock.yaml: pnpm lockfile for deterministic pnpm installs.
- frontend/tsconfig.json: TypeScript project references.
- frontend/tsconfig.app.json: Browser-side TypeScript compiler options.
- frontend/tsconfig.node.json: Node-side TypeScript options for tooling.
- frontend/postcss.config.js: PostCSS pipeline.
- frontend/tailwind.config.js: Tailwind configuration.
- frontend/vite.config.ts: Vite setup with Vue and PWA plugins.

### App shell and source

- frontend/index.html: SPA host HTML.
- frontend/src/main.ts: Vue bootstrap and plugin registration.
- frontend/src/api/index.ts: Typed frontend API client and URL builders.
- frontend/src/App.vue: Main application container (navigation/search/layout modes).
- frontend/src/style.css: Global Tailwind v4 theme and base styles.
- frontend/src/composables/: Shared logic (useLayout, useNavigation, usePreview, useTheme).
- frontend/src/types/index.ts: TypeScript interface definitions for API responses.

### Frontend components

#### Explorer
- frontend/src/components/explorer/Breadcrumbs.vue: Path breadcrumb navigation.
- frontend/src/components/explorer/ExplorerHeader.vue: Search and layout control header.
- frontend/src/components/explorer/FileGrid.vue: Virtualized grid/list/details renderer.
- frontend/src/components/explorer/FileItemGrid.vue: Individual grid item component.
- frontend/src/components/explorer/FileItemRow.vue: Individual list/details row component.
- frontend/src/components/explorer/WaterfallView.vue: Random discovery view with masonry layout.
- frontend/src/components/explorer/WaterfallItem.vue: Viewport-aware item for Waterfall view (Swap & Unload).

#### Preview
- frontend/src/components/preview/FilePreview.vue: Main preview overlay controller.
- frontend/src/components/preview/TextPreview.vue: Raw text content renderer.
- frontend/src/components/preview/VideoPreview.vue: Video player (Plyr) wrapper.

#### UI
- frontend/src/components/ui/FileIcon.vue: File-type icon component.

## mobile (Flutter)

- mobile/lib/main.dart: Flutter app entrypoint.
- mobile/lib/api.dart: Flutter API client wrapper.
- mobile/lib/models.dart: Data models for file items and API responses.
- mobile/lib/pages/: Application screens (Explorer, Preview, Settings, Waterfall).
- mobile/lib/providers/: State management (ExplorerProvider, SettingsProvider).
- mobile/lib/widgets/breadcrumbs_widget.dart: Mobile breadcrumb navigation.
- mobile/lib/widgets/file_item_widget.dart: Item renderer for explorer grid/list.
- mobile/lib/widgets/waterfall_item.dart: Viewport-aware item for Waterfall view (Swap & Unload).

## test_library

ZIP fixtures used for local/manual validation and test scenarios.
