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

The utility replaces the default behavior of the Globe key and adds two new modes to it:

### Single Press Mode

Single press mode is the mode that is activated when the Globe key is pressed once.

Single press mode cycles between your primary input sources—I believe most of the users out there will not even need
the other available mode as it is probably only useful if you have more than average input sources.

Single press mode uses the input sources defined in the config's `primary_input_sources` array.

### Double Press Mode

Double press mode is the mode that is activated when the Globe key is double pressed.

Double press mode cycles between your additional input sources. If you use multiple input sources, you
probably use only several input sources frequently—you might consider putting those that you use the least under
additional input sources.

Double press mode uses the input sources defined in the config's `additional_input_sources` array.

> **N.B.** This is not working as designed at the moment—this is supposed to open the original input source popup, but
> implementing it requires some reverse engineering. There is probably a function in macOS private API that can be used
> to open the popup.

### Prerequisites

- A macOS-powered computer
- A keyboard that features the Globe key, e.g. MacBook's builtin keyboard
- A will to use this utility

### Setting It Up

> **N.B.** The next release of the utility is going to feature a more convenient way to set it up and I really hope
> that it is going to be a Homebrew Formulae.

- Download the [latest prebuilt binary](https://github.com/Serpentiel/betterglobekey/releases/latest). You are also
  free to build it from the source, e.g. by running `go install`.
- Move it to a reasonable and secluded place, e.g. under `/usr/local/bin`. Make sure to rename it to `betterglobekey`
  for convenience.
- Touch `~/Library/LaunchAgents/me.serpentiel.betterglobekey.plist` and fill it with the following contents:

  ```xml
  <?xml version="1.0" encoding="UTF-8"?>
  <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
  <plist version="1.0">
      <dict>
          <key>Label</key>
          <string>me.serpentiel.betterglobekey</string>
          <key>ProgramArguments</key>
          <array>
              <string>/usr/local/bin/betterglobekey</string>
          </array>
          <key>RunAtLoad</key>
          <true/>
      </dict>
  </plist>
  ```

- Tired yet? Feel free to [contribute](#contributing) to the project by providing us with a more convenient way to set
  it up.
- Now, evaluate `launchctl load -w ~/Library/LaunchAgents/me.serpentiel.betterglobekey.plist` in your favorite
  terminal app and you are all set and good to go!

> **P.S.** Do not forget to disable the default Globe key behavior under **_System Preferences > Keyboard_**.

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
