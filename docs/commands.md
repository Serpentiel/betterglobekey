# Commands

Running **betterglobekey** with no arguments starts the daemon (this is what the Homebrew service does). The following
subcommands are also available.

## `betterglobekey list`

Lists every selectable input source as its identifier alongside its localized name. Use the identifiers to populate
`collections` in your [configuration](configuration.md).

```bash
betterglobekey list
```

## `betterglobekey current`

Prints the currently active input source (identifier and localized name).

## `betterglobekey doctor`

Diagnoses your setup and prints:

- whether Accessibility permission is granted;
- the configuration file path and whether it is valid;
- any configured sources that are not currently available (e.g. typos or removed layouts);
- the current input source.

This is the first thing to run if switching is not working.

## `betterglobekey edit`

Opens the configuration file in `$EDITOR` (falling back to `open`), creating it first if necessary.

## Global Flags

- `--config <path>` (`-c`) — use a configuration file other than `~/.betterglobekey.yaml`.
- `--version` — print the version.
- `--help` (`-h`) — show help for any command.
