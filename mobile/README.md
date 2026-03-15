# Invisible Archive Mobile (Android)

This is the Flutter-based mobile application for Invisible Archive.

## Prerequisites

1.  **Flutter SDK**: [Install Flutter](https://docs.flutter.dev/get-started/install)
2.  **Android SDK**: Required for building the APK. Usually installed via Android Studio.

## Building the APK

To install dependencies and build the production APK, run the following commands from this directory:

```bash
# Install dependencies
flutter pub get

# Build production APK
flutter build apk --release
```

The resulting APK will be located at:
`build/app/outputs/flutter-apk/app-release.apk`

## Configuration

When you first launch the app, go to **Settings** and configure the **Server URL** to point to your Invisible Archive instance.

- If running on an **Android Emulator** and the server is on your host machine, use `http://10.0.2.2:8080`.
- If running on a **Physical Device**, use the IP address of your server (e.g., `http://192.168.1.100:8080`).
