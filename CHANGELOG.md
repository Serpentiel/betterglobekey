# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

<!-- ### Added -->

<!-- ### Changed -->

<!-- ### Deprecated -->

<!-- ### Removed -->

<!-- ### Fixed -->

<!-- ### Security -->

## [2.1.1] - 2023-04-05

### Changed

- 6e2fd9d785533c7d9bd5ddd8c6d49f493a48a2ae: build(deps): bump DavidAnson/markdownlint-cli2-action from 7 to 8 (#16) (@dependabot[bot])
- c76661d9370267eca0f5bcdb19f879bb3a80c751: build(deps): bump DavidAnson/markdownlint-cli2-action from 8 to 9 (#19) (@dependabot[bot])
- 984f89e3b6491fa19c0d524c25d97facdb03b170: build(deps): bump actions/dependency-review-action from 2 to 3 (#14) (@dependabot[bot])
- 449ecd69eca0908c5daaeef292186c82a8f39f49: build(deps): bump actions/setup-go from 3 to 4 (#22) (@dependabot[bot])
- 983c0482102948e7479faf0fd5ff93b130d80823: build(deps): bump github.com/spf13/cobra from 1.6.1 to 1.7.0 (#23) (@dependabot[bot])
- 5e7763c1080e45a0533dcedc803d8c3f804041a8: build(deps): bump github.com/spf13/viper from 1.14.0 to 1.15.0 (#20) (@dependabot[bot])
- cf7032a1a8477bf25a9a600be74b5809c6573b9c: build(deps): bump go.uber.org/zap from 1.23.0 to 1.24.0 (#17) (@dependabot[bot])
- f4b276b4296a95b4275d9eaa2d57dfa390a98b9d: build(deps): bump goreleaser/goreleaser-action from 3 to 4 (#18) (@dependabot[bot])
- 932a6e7f9bdcc8dde90403025412a80559db7460: build(deps): bump hmarr/auto-approve-action from 2 to 3 (#15) (@dependabot[bot])
- c24ecf28b69ba5383ecab1518ea0776bf2a72b28: chore: migrate to Serpentiel/go-cli template (#25) (@Serpentiel)
- ddccc39e4717dc59000f1230028ec9536c667cea: ci(lint): upgrade trunk (#24) (@Serpentiel)
- 3f75a5476a8d03cf02c0872e84ab916fc68b35cd: docs(changelog): add v2.1.0 entries (#12) (@Serpentiel)
- 584e66060b64bc980b3aad513457d7de68d9f96b: docs(gh): fix PULL_REQUEST_TEMPLATE.md (#13) (@Serpentiel)
- 807170319fe6486cc3ca9a5e91b818d54ee4e584: refactor: internal/pkg (#21) (@Serpentiel)

## [2.1.0] - 2022-11-10

### Added

- 25ebdbb45568a620224bd5acf5e030eec3e95e1c: feat(config): auto pre-fill with existing inputs (@GTANAdam)
- cebb4ed7d4c96e61911db131f1a003286ec3c245: feat(doublepress): disable if none are set (@GTANAdam)

### Changed

- cb9150567bc22ffc3ab95f19b96a90ff6e88c95a: build(deps): bump github.com/spf13/cobra from 1.5.0 to 1.6.0 (#6) (@dependabot[bot])
- 3ae8b85db41d72a3f080c03f7619b7e72151326c: build(deps): bump github.com/spf13/cobra from 1.6.0 to 1.6.1 (#8) (@dependabot[bot])
- 426eb7561bf8d8e7de3a9e18e13311edf6ab550f: build(deps): bump github.com/spf13/viper from 1.13.0 to 1.14.0 (#10) (@dependabot[bot])
- 993d8a6da392a08b820248a1c9b5d39cf20da972: chore: remove inputs from example file (@GTANAdam)
- 39e2358c4d168fead84bd95d65e054b76ae7fb89: docs(README.md): update config instructions (#7) (@Serpentiel)
- 0eb8c5bee5a72db1e040cc55add0a1256ec8d811: docs: update changelog (#5) (@Serpentiel)

### Fixed

- b0d241b2f01142d495e158db85e74d51411791ee: ci(lint): fix block should not start with a whitespace (wsl) (@GTANAdam)

## [2.0.1] - 2022-09-29

### Added

- bf5800239480545f24701f95e424367415339c0b: build: add Homebrew formula, update docs (@Serpentiel)
- 5bc9a59fef02509cf3762313669026ddc22ced91: ci: create latest tag for new releases (@Serpentiel)

### Changed

- 6b0ff7aba47ee14365a3c15bf6fe0aa9dd075f84: build(deps): bump github.com/spf13/viper from 1.12.0 to 1.13.0 (#4) (@dependabot[bot])
- 6caa7634725d3ff6ad7615a0a696227e0a457ef2: docs: update CONTRIBUTING.md (#3) (@Serpentiel)
- bb43e04af4f09951ea8eb6ea820fdd1495eba2ce: docs: update changelog (@Serpentiel)

### Fixed

- 57f1f61925cbd13a876ad494033e2a83ab27a51b: fix: CJK input sources (@Serpentiel)

## [2.0.0] - 2022-08-30

### Changed

- 38f6ec: docs: update changelog (@Serpentiel)
- 12a6d2: refactor!: change config schema, add new option (@Serpentiel)

### Removed

- 1df393: ci: drop Changelog Enforcer workflow (@Serpentiel)

## [1.0.1] - 2022-08-30

### Added

- 839f5a: ci: add codeql workflow (@Serpentiel)
- 4782e7: ci: add commitlint job (@Serpentiel)
- 55c50f: docs: add config file info, minor fixup (@Serpentiel)

### Changed

- 83de3e: chore: fix cgo casing (@Serpentiel)
- 32d780: docs: correct template project attribution (@Serpentiel)
- ffc9df: docs: drop excess headings, minor rephrase (@Serpentiel)
- 091748: refactor: drop Carbon library (@Serpentiel)
- 8d4266: refactor: ensure fnKeyHandler implements handler (@Serpentiel)
- 899fad: refactor: make nolint:nlreturn global (@Serpentiel)

### Fixed

- 9f4a7d: fix: make config to be read properly (@Serpentiel)

## [1.0.0] - 2022-08-30

### Added

- initial release

[unreleased]: https://github.com/Serpentiel/betterglobekey/compare/v2.1.1...HEAD
[2.1.1]: https://github.com/Serpentiel/betterglobekey/releases/tag/v2.1.1
[2.1.0]: https://github.com/Serpentiel/betterglobekey/releases/tag/v2.1.0
[2.0.1]: https://github.com/Serpentiel/betterglobekey/releases/tag/v2.0.1
[2.0.0]: https://github.com/Serpentiel/betterglobekey/releases/tag/v2.0.0
[1.0.1]: https://github.com/Serpentiel/betterglobekey/releases/tag/v1.0.1
[1.0.0]: https://github.com/Serpentiel/betterglobekey/releases/tag/v1.0.0
