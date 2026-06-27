# Companion App

The companion is an optional desktop application for editing your **betterglobekey** configuration with a graphical
interface, instead of editing the YAML file by hand. It is a small [Electron](https://www.electronjs.org/) application
built with React and TypeScript.

## How It Works

The companion does not read or write the configuration file directly. Instead, the running **betterglobekey** service
hosts a local [gRPC](https://grpc.io/) server over a Unix domain socket at `~/.betterglobekey.sock`, and the companion
connects to it. This means:

- the socket is local-only — it is never exposed on the network, and its file permissions restrict it to your user;
- saving in the companion writes the configuration through the service, which then reloads it automatically (the same
  [live reload](configuration.md) that applies to manual edits);
- the service must be running for the companion to connect. If it is not, the companion shows a clear message and a
  retry button.

The companion can read your input sources (with their localized names) and the currently active source directly from the
service, so building collections is a matter of picking from a list rather than copying identifiers by hand.

## Using It

1. Make sure the **betterglobekey** service is running (see [Getting Started](getting-started.md)).
2. Launch the companion application.
3. Edit your behavior, collections, and logging settings.
4. Click **Save**. The change is validated, written, and applied immediately.

If a setting is invalid (for example, a malformed double-press delay), the save is rejected and the existing
configuration is left untouched.

## Installing It

The companion is distributed as a Homebrew cask. Installing it also pulls in the **betterglobekey** formula (the
command-line tool and its background service):

```bash
brew install --cask serpentiel/tools/betterglobekey-companion
```

The app is placed in your Applications folder. Start the service with `brew services start betterglobekey`, then open
the companion.

To build it from source instead, see the [development guide](development.md).
