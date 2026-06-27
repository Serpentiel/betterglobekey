# betterglobekey companion

A desktop companion application for editing the [betterglobekey](https://github.com/Serpentiel/betterglobekey)
configuration through a graphical interface. It is built with [Electron](https://www.electronjs.org/),
[React](https://react.dev/), and TypeScript, and talks to the running betterglobekey service over a local
[gRPC](https://grpc.io/) API on a Unix domain socket.

For what the companion does and how it connects to the service, see the
[companion documentation](../docs/companion.md).

## Development

All workflows are exposed as [Task](https://taskfile.dev/) targets from the repository root, namespaced under
`companion:`:

```bash
task companion:install    # install dependencies
task companion:dev        # run in development mode
task companion:lint       # lint with ESLint
task companion:typecheck  # type-check
task companion:build      # build main, preload, and renderer
task companion:dist       # build a distributable macOS app
```

The betterglobekey service must be running for the companion to connect, as the service hosts the gRPC control API.

## Layout

- `src/main` — the Electron main process: window lifecycle and the gRPC client.
- `src/preload` — the context-isolated bridge exposing a typed API to the renderer.
- `src/renderer` — the React application.
- `src/shared` — types shared between the processes.
- `proto` — a synced copy of the control contract, loaded at runtime and bundled into the packaged app.
