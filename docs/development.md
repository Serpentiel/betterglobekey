# Development

This project uses [Task](https://taskfile.dev/) as the entry point for common development workflows, for both the Go
application and the [companion app](companion.md). Run `task` with no arguments to list every available task.

## Prerequisites

- [Go](https://go.dev/) (see the version in `go.mod`), with `CGO_ENABLED=1` — the application links against macOS
  frameworks.
- [Node.js](https://nodejs.org/) for the companion app.
- [Task](https://taskfile.dev/installation/).
- [buf](https://buf.build/) and the Go protobuf plugins (`protoc-gen-go`, `protoc-gen-go-grpc`) — only needed to
  regenerate code from the protobuf contract.

If you use [mise](https://mise.jdx.dev/), the pinned toolchain versions are declared in `mise.toml`.

## Common Tasks

| Task                 | Description                                                       |
| -------------------- | ----------------------------------------------------------------- |
| `task install`       | Download Go modules and install companion dependencies.           |
| `task build`         | Build the `betterglobekey` binary into `./bin`.                   |
| `task test`          | Run the Go test suite.                                            |
| `task lint`          | Run all linters via [trunk](https://trunk.io/).                   |
| `task fmt`           | Format the codebase via trunk.                                    |
| `task generate`      | Regenerate Go code from the protobuf contract and sync it across. |
| `task changelog:new` | Add a [changie](https://changie.dev/) changelog fragment.         |

## The Companion App

The companion lives in [`companion/`](../companion) and has its own tasks, namespaced under `companion:`:

| Task                       | Description                                    |
| -------------------------- | ---------------------------------------------- |
| `task companion:install`   | Install companion dependencies.                |
| `task companion:dev`       | Run the companion in development mode.         |
| `task companion:build`     | Build the companion (main, preload, renderer). |
| `task companion:lint`      | Lint the companion with ESLint.                |
| `task companion:typecheck` | Type-check the companion.                      |
| `task companion:dist`      | Build a distributable macOS application.       |

The companion talks to the running service over the gRPC control API; see [Companion App](companion.md) for the
architecture. The contract is defined in [`proto/`](../proto) and the generated Go code is committed under
`internal/gen`. After changing the `.proto` file, run `task generate`.

## Releasing

Releases are driven entirely by CI through the manually triggered `goreleaser` workflow, which normalizes the tag,
batches the changelog with changie, tags the commit, and runs goreleaser. The reusable steps are available locally as
`task` targets (`task release:normalize-tag`, `task changelog:batch`, `task release:publish`).
