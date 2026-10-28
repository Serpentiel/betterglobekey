<!-- markdownlint-disable -->
<div id="top"></div>

<div align="center">
  <a href="https://github.com/Serpentiel/betterglobekey/graphs/contributors">
    <img src="https://img.shields.io/github/contributors/Serpentiel/betterglobekey.svg?style=for-the-badge" alt="Contributors" height="28">
  </a>
  <a href="https://github.com/Serpentiel/betterglobekey/network/members">
    <img src="https://img.shields.io/github/forks/Serpentiel/betterglobekey.svg?style=for-the-badge" alt="Forks" height="28">
  </a>
  <a href="https://github.com/Serpentiel/betterglobekey/stargazers">
    <img src="https://img.shields.io/github/stars/Serpentiel/betterglobekey.svg?style=for-the-badge" alt="Stars" height="28">
  </a>
  <a href="https://github.com/Serpentiel/betterglobekey/issues">
    <img src="https://img.shields.io/github/issues/Serpentiel/betterglobekey.svg?style=for-the-badge" alt="Issues" height="28">
  </a>
  <a href="https://github.com/Serpentiel/betterglobekey/blob/main/LICENSE.md">
    <img src="https://img.shields.io/github/license/Serpentiel/betterglobekey.svg?style=for-the-badge" alt="License" height="28">
  </a>
  <br>
  <br>
  <a href="https://github.com/Serpentiel/betterglobekey">
    <img src="https://github.com/Serpentiel/betterglobekey/blob/repo-assets/README.md/logo.png" alt="Logo" width="427" height="256">
  </a>
  <h3>betterglobekey</h3>
  <p>Make macOS Globe key great again!</p>
  <br>
  <br>
  <p>
    <a href="https://github.com/Serpentiel/betterglobekey/issues/new?labels=question&template=01_question.md">Ask a Question</a>
    &bullet;
    <a href="https://github.com/Serpentiel/betterglobekey/issues/new?labels=bug&template=02_bug.md">Report a Bug</a>
    &bullet;
    <a href="https://github.com/Serpentiel/betterglobekey/issues/new?labels=enhancement&template=03_feature.md">Request a Feature</a>
  </p>
</div>
<details>
  <summary>Table of Contents</summary>
  <ul>
    <li>
      <a href="#about-the-project">1. About the Project</a>
    </li>
    <li>
      <a href="#quick-start">2. Quick Start</a>
    </li>
    <li>
      <a href="#contributing">3. Contributing</a>
    </li>
    <li>
      <a href="#license">4. License</a>
    </li>
  </ul>
</details>
<!-- markdownlint-restore -->

## About the Project

macOS features a brilliant idea for the way to switch your keyboard's input source—the Globe key. While the idea is
brilliant in design, in practice, however, the key is coded in a very intrusive and impractical way, which might
prevent some users—especially those who tend to switch their input source often—from enjoying their overall user
experience.

This utility completely reworks the way the Globe key functions:

- **Press** switches to the next input source in the current collection: your own ordered list of input sources.
- **Double press** switches to the next collection, such as from your everyday layouts to the ones you use less often.
- **Shift** goes back: to the previous input source, or with a double press, to the previous collection.
- An on-screen HUD shows the new input source and its collection.
- Any keyboard can switch with keys of its own as well, such as Caps Lock or Right Command, which also covers keyboards
  without a Globe key.

The configuration is a YAML file, reloaded as soon as it changes; an optional companion app edits it for you. I
sincerely hope that one day Apple is going to make it this way out-of-box.

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->

## Quick Start

1. Install **betterglobekey**:

   ```bash
   brew install betterglobekey
   ```

2. Write the configuration to `~/.betterglobekey.yaml`, with your own input sources (`betterglobekey list` shows
   them):

   ```yaml
   version: 2
   trigger:
     globe: true
   collections:
     - name: primary
       sources:
         - com.apple.keylayout.US
         - com.apple.keylayout.Russian
     - name: secondary
       sources:
         - com.apple.keylayout.US
         - com.apple.inputmethod.Kotoeri.RomajiTyping.Japanese
   ```

3. Start the service, or restart it if it is already running:

   ```bash
   brew services restart betterglobekey
   ```

4. Grant Accessibility permission when prompted, and set **"Press 🌐 key to"** to **"Do Nothing"** under _System
   Settings > Keyboard_ so macOS stops switching the input source on its own. `betterglobekey doctor` checks both.

Done. For other trigger keys, every option, and the optional [companion app](docs/companion.md), see
[Getting Started](docs/getting-started.md).

<!-- markdownlint-disable -->
<div align="center">
  <img src="https://github.com/Serpentiel/betterglobekey/blob/repo-assets/README.md/companion.png" alt="betterglobekey-companion" width="760">
</div>
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->

## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create, so any
contributions you make are greatly appreciated.

If you would like to contribute, please make sure to take a look
at [this guideline](https://github.com/Serpentiel/betterglobekey/blob/main/CONTRIBUTING.md) beforehand.

Please see our [security policy](https://github.com/Serpentiel/betterglobekey/blob/main/SECURITY.md) to report any possible
vulnerabilities or serious issues.

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->

## License

Distributed under the MIT License. See
[`LICENSE.md`](https://github.com/Serpentiel/betterglobekey/blob/main/LICENSE.md) for more information.

> **N.B.** This project explicitly does not require its contributors to sign a _Contributor License Agreement_ nor does
> it possess one.

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->
