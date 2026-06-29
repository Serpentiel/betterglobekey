# Configuration

After the first run, **betterglobekey** creates a configuration file at `~/.betterglobekey.yaml`, seeded with a
`primary` collection containing your currently available input sources. Edit this file to customize how the utility
behaves.

The configuration is reloaded automatically whenever the file changes, so edits take effect immediately — there is no
need to restart the service. Prefer a graphical editor? The [companion app](companion.md) exposes every option below.

```bash
betterglobekey edit
```

## Full Example

```yaml
version: 2

logger:
  path: ~/Library/Logs/betterglobekey.log
  level: info
  retention:
    days: 30
    files: 3

double_press:
  enabled: true
  maximum_delay: 250ms

reverse:
  enabled: true
  modifier: shift

hud:
  enabled: true
  duration: 900ms
  show_collection: true

collections:
  - name: default
    sources:
      - com.apple.keylayout.US
      - com.apple.keylayout.Russian
  - name: coding
    sources:
      - com.apple.keylayout.US
      - com.apple.keylayout.Ukrainian-PC
```

All options other than `collections` are optional and fall back to the defaults shown above.

## Collections

`collections` is an **ordered list** of named collections. Each has a `name` (shown in the HUD) and an ordered list of
input source `sources`. A single press of the Globe key cycles forward through the sources of the current collection; a
double press moves to the next collection.

To list the input source identifiers available on your system, along with their localized names:

```bash
betterglobekey list
```

## Double Press

`double_press.enabled` turns collection switching on or off. When disabled, every press is a single press that cycles
within the current collection.

`double_press.maximum_delay` is the maximum time between two presses for them to count as a double press. It is a Go
duration string such as `250ms` or `1s`. Lower it if deliberate single presses are mistaken for double presses; raise it
if double presses are hard to trigger.

## Reverse

`reverse.enabled` turns the reverse modifier on or off. While held, a single press returns to the previously used input
source and a double press switches to the previous collection.

`reverse.modifier` selects the key that triggers a reverse press: `shift`, `option`, `control`, or `command`.

## HUD

`hud.enabled` controls the on-screen overlay that names the new input source on each change. `hud.duration` is how long
it stays visible (a Go duration), and `hud.show_collection` toggles the collection-name subtitle.

## Logger

`logger.path` is the log file location. `logger.level` is the minimum level to record (`debug`, `info`, `warn`, or
`error`). `logger.retention.days` and `logger.retention.files` control how long and how many rotated logs are kept.

## Migration From Older Versions

Configuration files from before the `version: 2` format (which used an unordered `input_sources` map) are migrated
automatically on load. The original file is preserved next to it as `~/.betterglobekey.yaml.v1.bak`.

Once configured, head over to [Usage](usage.md) to learn about the modes of operation, and [Commands](commands.md) for
the full command reference.
