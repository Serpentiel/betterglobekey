# Getting Started

## Installation

First, ensure you have Homebrew installed. If not, install it by running the following command in your terminal:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

See <https://brew.sh> for more information.

With Homebrew installed, install **betterglobekey** from `homebrew/core`:

```bash
brew install betterglobekey
```

> **N.B.** Existing `Serpentiel/tools` tap installations migrate to `homebrew/core` automatically on `brew update`.

## Accessibility Permissions

**betterglobekey** needs Accessibility permission to observe the Globe key. The first time it runs it prompts you
automatically; grant it under _System Settings > Privacy & Security > Accessibility_ by enabling the switch next to
**betterglobekey**.

A keyboard's own [trigger keys](configuration.md#trigger-keys), and recording keys in the companion, also need Input
Monitoring, under _System Settings > Privacy & Security > Input Monitoring_. If **betterglobekey** is not listed there,
click **+** and add `$(brew --prefix)/opt/betterglobekey/bin/betterglobekey` (in the file picker, ⌘⇧G opens a path).
The service picks it up without a restart.

You can check the permissions (and the rest of your setup) at any time with:

```bash
betterglobekey doctor
```

## Freeing the Globe Key

So that macOS does not also switch the input source on its own — which would conflict with **betterglobekey** — set the
Globe key to do nothing: open _System Settings > Keyboard_ and set **"Press 🌐 key to"** to **"Do Nothing"**.

No Globe key on your keyboard? Choose another [trigger key](configuration.md#trigger-keys), such as Caps Lock or Right
Command.

## Starting the Service

Start the **betterglobekey** service with Homebrew:

```bash
brew services start betterglobekey
```

It will then start automatically on login. Next, see [Configuration](configuration.md) and [Usage](usage.md).
