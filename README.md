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
      <a href="#about-the-project">1. About this Project</a>
    </li>
    <li>
      <a href="#getting-started">2. Getting Started</a>
      <ul>
        <li>
          <a href="#prerequisites">2.1. Prerequisites</a>
        </li>
        <li>
          <a href="#setting-it-up">2.2. Setting It Up</a>
        </li>
      </ul>
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

This utility completely reworks the way the Globe key functions and provides users with an improved overall user
experience, and I sincerely hope that one day Apple is going to make it this way out-of-box.

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->

## Getting Started

The utility enhances the functionality of the Globe key by introducing two distinct modes of operation and starting
from the currently active input source:

1. **Single Press Mode**

   Single press mode is activated when the Globe key is pressed once.

   In this mode, the utility cycles through a collection of input sources. Each press of the Globe key switches to the
   next input source within the current collection.

   The collections of input sources are defined in the configuration under `input_sources`. Each key-value pair within
   this map represents a named collection of input sources. For example:

   ```yaml
   input_sources:
     foo:
       - com.apple.keylayout.US
       - com.apple.keylayout.Russian
     bar:
       - com.apple.keylayout.Finnish
       - com.apple.keylayout.Ukrainian-PC
       - com.apple.inputmethod.Kotoeri.RomajiTyping.Japanese
   ```

   Upon initialization, the utility determines the current active input source and starts from that particular source
   within its respective collection.

2. **Double Press Mode**

   Double press mode is activated when the Globe key is double-pressed.

   In this mode, the utility switches between different collections of input sources. Each double press of the Globe
   key cycles to the next collection in the configuration.

   The maximum time interval between the first and second press that is considered a double press can be configured
   in the `double_press.maximum_delay` property. This delay is specified in milliseconds.

These enhancements aim to provide a more versatile and user-friendly experience for managing multiple input sources,
especially for users who frequently switch between different languages or keyboard layouts.

See our wiki for more information on
[how to set up and configure the utility](https://github.com/Serpentiel/betterglobekey/wiki).

<!-- markdownlint-disable -->
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

> **N.B.** This project explicitly does not requires its contributors to sign a _Contributor License Agreement_ nor does
> it posses one.

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->
