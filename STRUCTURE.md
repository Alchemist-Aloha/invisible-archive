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
- test_mime.go: Small utility program for checking Go MIME type mappings.
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

- frontend/package.json: Frontend scripts and dependency manifest (Vue, Vite, TanStack, PhotoSwipe, Plyr, Tailwind).
- frontend/package-lock.json: npm lockfile for deterministic npm installs.
- frontend/pnpm-lock.yaml: pnpm lockfile for deterministic pnpm installs.
- frontend/tsconfig.json: TypeScript project references (app + node configs).
- frontend/tsconfig.app.json: Browser-side TypeScript compiler options and strictness rules.
- frontend/tsconfig.node.json: Node-side TypeScript options for tooling files (for example Vite config).
- frontend/postcss.config.js: PostCSS pipeline (Tailwind + Autoprefixer).
- frontend/tailwind.config.js: Tailwind content scanning and theme extension entry.
- frontend/vite.config.ts: Vite setup with Vue plugin and PWA plugin/manifest/workbox settings.
- frontend/README.md: Default Vite/Vue template readme.

### App shell and source

- frontend/index.html: SPA host HTML, mobile/PWA meta tags, and app mount script.
- frontend/src/main.ts: Vue bootstrap and plugin registration (Vue Query).
- frontend/src/api.ts: Typed frontend API client, capability constants, URL builders, and fetch helpers.
- frontend/src/App.vue: Main application container (navigation/search/layout modes, previews, gestures, dark mode, history sync).
- frontend/src/style.css: Global Tailwind v4 theme tokens, color variables, base styles, and shared animations.

### Frontend components

- frontend/src/components/Breadcrumbs.vue: Path breadcrumb navigation component.
- frontend/src/components/FileGrid.vue: Virtualized grid/list/details renderer with item interactions and thumbnail fallback.
- frontend/src/components/FileIcon.vue: File-type icon selector from capabilities/extension.
- frontend/src/components/HelloWorld.vue: Unused scaffold component from template.

### Frontend assets

- frontend/src/assets/hero.png: Template/demo hero image asset.
- frontend/src/assets/vite.svg: Vite logo asset.
- frontend/src/assets/vue.svg: Vue logo asset.
- frontend/public/favicon.svg: Site favicon.
- frontend/public/icons.svg: SVG sprite sheet used by template/demo UI.
- frontend/public/manifest.webmanifest: PWA manifest served as static asset.
- frontend/public/pwa-192x192.png: 192px PWA icon.
- frontend/public/pwa-512x512.png: 512px PWA icon.

## test_library

ZIP fixtures used for local/manual validation and test scenarios around archive traversal and media handling.

- test_library/basic.zip: Basic archive fixture for simple path and listing behavior.
- test_library/implicit.zip: Fixture likely covering implicit root-folder behavior.
- test_library/large.zip: Larger archive fixture for performance/streaming checks.
- test_library/media.zip: Media-heavy fixture for render/stream capability paths.
- test_library/mixed.zip: Mixed-content fixture across multiple file types.
- test_library/nested.zip: Nested-directory archive fixture for deep path resolution.
- test_library/photos.zip: Image-focused fixture for thumbnails/gallery behavior.

## Notes

- Generated files: internal/data/db.go, internal/data/models.go, and internal/data/queries.sql.go are generated by sqlc from internal/data/schema.sql and internal/data/queries.sql.
- Runtime data files (for example archive.db and thumbnail cache) are not committed; they are created at runtime in CACHE_DIR.
