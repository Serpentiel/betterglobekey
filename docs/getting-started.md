# Getting Started

## Installation

First, ensure you have Homebrew installed. If not, install it by running the following command in your terminal:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

See <https://brew.sh> for more information.

With Homebrew installed, install **betterglobekey**:

```bash
brew tap Serpentiel/tools
brew install betterglobekey
```

## Accessibility Permissions

**betterglobekey** needs Accessibility permission to observe the Globe key. The first time it runs it prompts you
automatically; grant it under _System Settings > Privacy & Security > Accessibility_ by enabling the switch next to
**betterglobekey**.

You can check the permission (and the rest of your setup) at any time with:

```bash
betterglobekey doctor
```

## Freeing the Globe Key

So that macOS does not also switch the input source on its own — which would conflict with **betterglobekey** — set the
Globe key to do nothing: open _System Settings > Keyboard_ and set **"Press 🌐 key to"** to **"Do Nothing"**.

## Starting the Service

Start the **betterglobekey** service with Homebrew:

```bash
brew services start betterglobekey
```

It will then start automatically on login. Next, see [Configuration](configuration.md) and [Usage](usage.md).
