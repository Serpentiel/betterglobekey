# Development

This project uses [Task](https://taskfile.dev/) as the entry point for common development workflows, for both the Go
application and the [companion app](companion.md). Run `task` with no arguments to list every available task.

## Prerequisites

The toolchain is pinned in [`mise.toml`](../mise.toml). The recommended setup uses [mise](https://mise.jdx.dev/), which
installs the exact pinned versions of every tool:

```bash
mise trust    # approve this project's mise.toml
mise install  # install the pinned Go, Node.js, Task, buf, and changie
task install   # download Go modules and companion dependencies
```

From there, every workflow runs through `task` (run `task` with no arguments to list them all).

Prefer to manage tools yourself? You will need:

- [Go](https://go.dev/) (see the version in `go.mod`), with `CGO_ENABLED=1` — the application links against macOS
  frameworks.
- [Node.js](https://nodejs.org/) and [rustup](https://rustup.rs/) for the companion app. The Rust toolchain is pinned
  in [`companion/src-tauri/rust-toolchain.toml`](../companion/src-tauri/rust-toolchain.toml); rustup installs it on
  the first `cargo` run.
- [Task](https://taskfile.dev/installation/).
- [buf](https://buf.build/) and the Go protobuf plugins (`protoc-gen-go`, `protoc-gen-go-grpc`) — only needed to
  regenerate code from the protobuf contract.

## Common Tasks

| Task                 | Description                                               |
| -------------------- | --------------------------------------------------------- |
| `task install`       | Download Go modules and install companion dependencies.   |
| `task build`         | Build the `betterglobekey` binary into `./bin`.           |
| `task test`          | Run the Go test suite.                                    |
| `task lint`          | Run all linters via [trunk](https://trunk.io/).           |
| `task fmt`           | Format the codebase via trunk.                            |
| `task generate`      | Regenerate Go code from the protobuf contract.            |
| `task changelog:new` | Add a [changie](https://changie.dev/) changelog fragment. |

## The Companion App

The companion lives in [`companion/`](../companion) and has its own tasks, namespaced under `companion:`:

| Task                       | Description                                             |
| -------------------------- | ------------------------------------------------------- |
| `task companion:install`   | Install companion dependencies.                         |
| `task companion:dev`       | Run the companion against the Vite dev server.          |
| `task companion:build`     | Build the UI into `companion/out`.                      |
| `task companion:lint`      | Lint the UI with ESLint.                                |
| `task companion:lint:rust` | Check formatting and lint the Rust core with Clippy.    |
| `task companion:typecheck` | Type-check the UI.                                      |
| `task companion:test`      | Run the UI unit tests.                                  |
| `task companion:test:rust` | Run the Rust tests, including a stub-daemon round trip. |
| `task companion:dist`      | Build the universal macOS app into `companion/dist`.    |

The companion talks to the running service over the gRPC control API; see [Companion App](companion.md) for the
architecture. The contract is defined in [`proto/`](../proto): the generated Go code is committed under `internal/gen`
(after changing the `.proto` file, run `task generate`), and the companion compiles the same file at build time.

## Releasing

Releases are driven entirely by CI through the manually triggered `goreleaser` workflow, which normalizes the tag,
batches the changelog with changie, tags the commit, and runs goreleaser. The reusable steps are available locally as
`task` targets (`task release:normalize-tag`, `task changelog:batch`, `task release:publish`).
