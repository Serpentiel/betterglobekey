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
  - name: primary
    sources:
      - com.apple.keylayout.US
      - com.apple.keylayout.Spanish
  - name: secondary
    sources:
      - com.apple.keylayout.US
      - com.apple.keylayout.Dvorak
```

All options other than `collections` are optional and fall back to the defaults shown above.

## Trigger Key

Every keyboard has its own trigger key, set under `trigger.keyboards` by the vendor and product ID that
`betterglobekey doctor` and the companion show. A keyboard not listed uses its Globe key, if it has one (a MacBook's
or a Magic Keyboard), and has no trigger key otherwise:

```yaml
trigger:
  keyboards:
    "046d:c33f": # a Logitech keyboard
      key: insert
      mode: pass_through
    "05ac:0281": # the MacBook's own keyboard
      key: caps_lock
      mode: dedicated
```

The keys:

| Key                                           | `key`                                                                    |
| --------------------------------------------- | ------------------------------------------------------------------------ |
| Globe                                         | `globe`                                                                  |
| Caps Lock                                     | `caps_lock`                                                              |
| Right or left Command, Option, Control, Shift | `right_command`, `left_option`, `right_control`, `left_shift`, and so on |
| F13 through F20                               | `f13` … `f20`                                                            |
| Menu (the "Application" key on PC keyboards)  | `menu`                                                                   |
| Insert, Print Screen, Scroll Lock, Pause      | `insert`, `print_screen`, `scroll_lock`, `pause`                         |

A key other than the Globe key is read straight from the keyboard's raw input, which needs **Input Monitoring**
(_System Settings > Privacy & Security > Input Monitoring_; macOS asks the first time, and the service may need a
restart after you grant it). How it is used depends on `mode`:

| `mode`                | The key                                                 |
| --------------------- | ------------------------------------------------------- |
| `dedicated` (default) | is **silenced**: it only switches input sources         |
| `pass_through`        | **keeps its normal function**, and switches when tapped |

- `dedicated` remaps the key on its keyboard while the service runs, the same way `hidutil property --matching … --set`
  does, to F24, which macOS has no key code for: the key produces no event at all and nothing else can react to it.
  Stopping the service puts the key back.
- `pass_through` switches input sources when the key is pressed on its own; in a combination (such as Right Command +
  C) the key works as usual. Caps Lock cannot be passed through.
- No trigger key acts as fn: a third-party keyboard has none of the fn combinations an Apple keyboard has, so a key
  remapped to fn would only half work there.
- The Globe key cannot tell keyboards apart: while any keyboard switches with it, the Globe key switches on every
  keyboard.

Keyboards plugged in while the service runs are remapped within two seconds. Things to know about remapping a
dedicated key:

- Each keyboard's own key mapping table is edited without touching other entries, and an entry it replaced for the same
  key is restored when the service stops. If the service is killed, the next start cleans up.
- It does not apply to keyboards that Karabiner-Elements has taken over; use Karabiner to map the key instead.
- It sits on top of _System Settings > Keyboard > Keyboard Shortcuts > Modifier Keys_. Leave the chosen key set to
  itself there.
- `betterglobekey doctor` lists every connected keyboard with its trigger and mappings, including other tools'.
- To undo every remapping on a keyboard by hand:
  `hidutil property --matching '{"VendorID":0x046d,"ProductID":0xc33f}' --set '{"UserKeyMapping":[]}'`.

Keys pressed together with the trigger key, including media keys such as volume or brightness, never count as a press.

## Collections

`collections` is an **ordered list** of named collections. Each has a `name` (shown in the HUD) and an ordered list of
input source `sources`. A single press of the trigger key cycles forward through the sources of the current collection;
a double press moves to the next collection.

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
