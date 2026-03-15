# Invisible Archive Mobile (Flutter)

This application is a companion mobile client for the Invisible Archive service, designed to provide a smooth browsing and viewing experience for files stored locally or inside ZIP archives.

## Architecture

The app follows a **Provider-based state management** pattern combined with a clean separation of concerns:

- **API Layer (`lib/api.dart`)**: Handles all communication with the Go backend. It includes methods for listing directories, searching, fetching random items, and generating URLs for raw files and thumbnails.
- **Models (`lib/models.dart`)**: Type-safe Dart classes for `FileItem` and `ListResponse`, mapping directly to the backend's JSON structure.
- **Providers (`lib/providers/`)**:
    - `SettingsProvider`: Manages user preferences (Server URL, Dark Mode, Layout) using `shared_preferences` for persistence.
    - `ExplorerProvider`: Manages the current folder state, file list, search results, and navigation history.
- **Pages (`lib/pages/`)**:
    - `ExplorerPage`: The main entry point for browsing files with support for multiple layout modes (Grid, List, Details).
    - `PreviewPage`: High-performance viewer for images, videos, and text files. Uses **Chewie** for a full-featured video experience.
    - `WaterfallPage`: A specialized "discovery" view that displays random images from the current directory and its subdirectories.
    - `SettingsPage`: UI for configuring server connectivity and app theme.
- **Widgets (`lib/widgets/`)**: Reusable components like `FileItemWidget` (handles rendering logic for different file types) and `BreadcrumbsWidget`.

## Key Features

### 1. ZIP Path Transparency
The mobile app mirrors the backend's core feature: archives (ZIP files) are treated as folders. Users can tap into a ZIP file and browse its contents just like a physical directory.

### 2. High-Performance Image Viewing
- **Dynamic Scrolling**: Uses `PhotoViewGallery` to allow swiping through all images within a directory or search result.
- **Intelligent Preloading**: Implements a manual preloading strategy in `PreviewPage`. When viewing an image, the app automatically precaches the next and previous images in the background to ensure instantaneous swiping.
- **Lazy Loading**: Utilizes `CachedNetworkImage` for all thumbnails and full-size images, reducing data usage and ensuring smooth scrolling by only loading images as they enter the viewport.

### 3. Advanced Video Playback
- **Chewie Integration**: Powered by `chewie`, providing a native-like playback interface on top of `video_player`.
- **Full Controls**: Includes play/pause, seek bar, volume, playback speed, and fullscreen support.
- **ExoPlayer/AVPlayer**: Leverages the best native engines for high-performance streaming and wide format support.

### 4. Waterfall Discovery Mode
Introduces a "Waterfall View" button in the file explorer. This mode:
- Fetches a random selection of images recursively from the current path using the backend's `/api/random` endpoint.
- **Full Archive Support**: Correctly crawls inside ZIP files and nested virtual folders to find images.
- Displays them in a masonry grid (using `flutter_staggered_grid_view`), making it ideal for visual browsing.

### 5. Adaptive UI
- **Layout Modes**: Users can switch between Grid, List, and Detailed List views.
- **Theme Support**: Full support for Light and Dark modes, persisted across app restarts.
- **Responsive Design**: Built with Flutter's flexible layout system to ensure compatibility across various screen sizes.

## Tech Stack

- **Framework**: Flutter 3.x
- **State Management**: `provider`
- **Navigation**: `MaterialPageRoute`
- **Caching**: `cached_network_image`
- **Video Interaction**: `chewie`, `video_player` (ExoPlayer/AVPlayer)
- **Image Interaction**: `photo_view`
- **Persistence**: `shared_preferences`
- **Layout Utilities**: `flutter_staggered_grid_view`

## Build Instructions

To build the application for Android:

1. Ensure Flutter SDK and **Java 21** are installed (Java 25+ may cause build script compatibility issues).
2. Set `ANDROID_HOME` to your Android SDK path.
3. Navigate to the `mobile` directory.
4. Run `flutter pub get`.
5. Run `flutter build apk --release`.

The resulting APK will be found at `mobile/build/app/outputs/flutter-apk/app-release.apk`.
