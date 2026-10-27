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

trigger:
  key: globe
  mode: fn
  keyboards:
    apply_to: without_globe

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

`trigger.key` selects the key that switches input sources. The default, `globe`, is the Globe (fn) key on Apple
keyboards. For a keyboard without one, pick another key:

| Key                                           | `trigger.key`                                                            |
| --------------------------------------------- | ------------------------------------------------------------------------ |
| Globe (fn)                                    | `globe`                                                                  |
| Caps Lock                                     | `caps_lock`                                                              |
| Right or left Command, Option, Control, Shift | `right_command`, `left_option`, `right_control`, `left_shift`, and so on |
| F13 through F20                               | `f13` … `f20`                                                            |
| Menu (the "Application" key on PC keyboards)  | `menu`                                                                   |

The Globe key is also the fn key: tapped alone it switches input sources, held it is the fn modifier. How a key other
than Globe is used depends on `trigger.mode`:

| `trigger.mode` | The key                                                   | Globe also switches | Per keyboard |
| -------------- | --------------------------------------------------------- | ------------------- | ------------ |
| `fn` (default) | is **remapped to fn**: fn+arrows, fn+Delete, fn+E, … work | yes                 | yes          |
| `dedicated`    | is **remapped to F20**, which only the service watches    | no                  | yes          |
| `pass_through` | **keeps its normal function**                             | no                  | no           |

- `fn` and `dedicated` remap the key at the hardware level while the service runs, the same way
  `hidutil property --matching … --set` does, so it loses its normal function. Stopping the service puts the key back.
  In `dedicated` mode the Globe (fn) key keeps only what macOS does with it. F20 is used because no Apple keyboard
  has it and macOS binds nothing to it; F21 to F24 have no key codes on macOS.
- `pass_through` watches the key directly: pressing it on its own switches input sources; using it in a combination
  (such as Right Command + C) does not. Caps Lock cannot be passed through, and F20 cannot be `dedicated`.

### Keyboards

A remapped key applies per keyboard. `trigger.keyboards.apply_to` picks the default:

- `without_globe` (the default): only keyboards without a Globe key of their own, that is, anything but Apple's built-in
  and Magic keyboards. Caps Lock on a MacBook stays Caps Lock while an external keyboard's Caps Lock becomes fn.
- `all`: every keyboard.

`trigger.keyboards.overrides` includes (`true`) or excludes (`false`) single keyboards by vendor and product ID, as
`betterglobekey doctor` and the companion show them:

```yaml
trigger:
  key: caps_lock
  mode: dedicated
  keyboards:
    apply_to: without_globe
    overrides:
      "05ac:029c": true # this Magic Keyboard too
      "046d:c33f": false # but not this one
```

Keyboards plugged in while the service runs are remapped within two seconds.

Things to know about remapping:

- Each keyboard's own key mapping table is edited without touching other entries, and an entry it replaced for the same
  key is restored when the service stops. If the service is killed, the next start cleans up.
- It does not apply to keyboards that Karabiner-Elements has taken over; use Karabiner to map the key instead.
- It sits on top of _System Settings > Keyboard > Keyboard Shortcuts > Modifier Keys_. Leave the chosen key set to
  itself there.
- `betterglobekey doctor` lists every connected keyboard with its mappings, including other tools'.
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
