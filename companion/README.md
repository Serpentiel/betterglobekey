# betterglobekey-companion

A desktop app for editing the [betterglobekey](https://github.com/Serpentiel/betterglobekey) configuration. It is a
[Tauri](https://tauri.app/) app: a React and TypeScript UI in the system WebKit view, and a Rust core that talks to the
running betterglobekey service over its local [gRPC](https://grpc.io/) API on a Unix domain socket.

For what the companion does and how it connects to the service, see the
[companion documentation](../docs/companion.md).

## Development

Every workflow runs through [Task](https://taskfile.dev/) from the repository root, namespaced under `companion:`:

```bash
task companion:install    # install dependencies
task companion:dev        # run against the Vite dev server (macOS)
task companion:lint       # lint the UI with ESLint
task companion:lint:rust  # rustfmt and Clippy
task companion:typecheck  # type-check the UI
task companion:test       # UI unit tests
task companion:test:rust  # Rust tests, including a stub-daemon round trip
task companion:dist       # build the universal macOS app into dist/
```

The companion only connects while the betterglobekey service is running, since the service hosts the gRPC API.

## Layout

- `src/renderer` — the React UI. `src/lib/api.ts` is its only link to the Rust core.
- `src/shared` — the UI's types for the control API messages.
- `src-tauri/src/daemon.rs` — the gRPC client, generated at build time from
  [`proto/`](../proto/betterglobekey/control/v1/control.proto) at the repository root.
- `src-tauri/src/lib.rs` — the window, the commands the UI invokes, and the app lifecycle.
- `src-tauri/tests` — the client against a stub daemon on a real socket.
- `scripts/dist.sh` — builds the release bundle and names it the way the release and the Homebrew cask expect.
