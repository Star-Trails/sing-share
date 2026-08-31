# sing-share

`sing-share` turns a local sing-box JSON configuration into an official local-profile BPF file and streams that file as QRS fountain-code QR frames in a native desktop window.

```text
sing-box JSON → BPF → QRS/Luby fountain stream → animated SVG QR
```

Everything stays in process. The application does not upload profiles, launch a terminal, run `qifi`, or invoke the sing-box CLI.

## Use

- Start `sing-share`, then choose or drop a `.json` configuration.
- Start with a file directly: `sing-share path/to/config.json`.
- Scan continuously with a QRS-compatible scanner until reconstruction completes.
- **Save BPF** writes the exact bytes used by the active QR stream.
- **Stop** stops frame generation and releases the active buffers.

FPS is adjustable from 1–30; the default is 10. QRS slice size defaults to 500 bytes.

## Architecture

- Go reads and validates UTF-8 JSON, derives the profile name, encodes/verifies BPF, owns native dialogs, and performs all filesystem writes.
- Wails sends one BPF payload plus non-sensitive metadata to the embedded frontend.
- Vue creates one `@qifi/generate` fountain generator and advances it locally. No Go/JavaScript bridge call occurs per frame.
- Vite, TypeScript, Bun, and all frontend packages are build-time dependencies only. Production frontend assets are embedded in the executable.

The BPF codec tracks [`experimental/libbox/profile_import.go`](https://github.com/SagerNet/sing-box/blob/testing/experimental/libbox/profile_import.go). QRS generation uses [`@qifi/generate`](https://github.com/qifi-dev/qrs/tree/main/packages/generate) without reimplementing Luby Transform or changing its wire protocol.

## Build

Prerequisites:

- Go 1.27 or newer
- Bun 1.4.0
- platform desktop build libraries listed below

Install the pinned Wails beta and dependencies:

```text
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.16
cd frontend
bun install --frozen-lockfile
cd ..
```

Build or run in development:

```text
wails3 build
wails3 dev
```

Artifacts are written under `bin/`. Build another supported architecture with, for example, `wails3 build ARCH=arm64`. Native CI builds cover Windows and Linux on amd64/arm64 and macOS on Intel/Apple Silicon.

### Platform requirements

- **Windows 10/11:** WebView2 Runtime. It is normally present; no Node.js/Bun runtime is required.
- **macOS:** system WebKit. Xcode Command Line Tools are required to build.
- **Linux default build:** WebKitGTK 6.0 and GTK4. On Ubuntu 24.04+/Debian 13+, install `build-essential pkg-config libgtk-4-dev libwebkitgtk-6.0-dev`.
- **Linux legacy build:** Ubuntu 22.04/Debian 12 may use GTK3 and WebKit2GTK 4.1, then build with `wails3 build EXTRA_TAGS=gtk3`. This Wails compatibility path is supported through Wails 3.0.x.

Linux users need the corresponding GTK/WebKit shared libraries at runtime. They do not need Go, Bun, Node.js, npm, or qifi.

## Verification

```text
go test ./internal/bpf ./services
cd frontend
bun run qrs:smoke
bun run build
```

The focused Go tests verify message/version bytes, gzip layout, big-endian local profile type, UTF-8 byte-length uvarints, exact configuration-byte preservation, and round-trip decoding. The QRS smoke executes the upstream metadata and SVG fountain APIs.

## Security

- No network API, telemetry, browser storage, or profile logging.
- The original JSON is cleared in Go after encoding and is never exposed to JavaScript.
- BPF data stays in memory; active JavaScript buffers and QRS encoder blocks are cleared when sharing stops or is replaced.
- Saved files are created with owner-only permissions where the operating system supports POSIX modes.

## License

MIT
