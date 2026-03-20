# Changelog

All notable changes to this project will be documented in this file.

## Unreleased (2026-03-20)

### Features

- add snapshot and diff commands for environment comparison ([644cace](https://github.com/somaz94/bash-pilot/commit/644cace92e47c999e2eff6dc399dcffa4b70a81b))

### Bug Fixes

- prevent Header panic when title exceeds 50 characters ([fcf6d2e](https://github.com/somaz94/bash-pilot/commit/fcf6d2e17699f6fa616ee376495f7693acc0037c))

### Contributors

- somaz

<br/>

## [v0.5.0](https://github.com/somaz94/bash-pilot/compare/v0.4.0...v0.5.0) (2026-03-20)

### Features

- add doctor command for full system diagnostics ([05b07a7](https://github.com/somaz94/bash-pilot/commit/05b07a775d64be20c5fd766307b032e6693b4b60))

### Bug Fixes

- show key name in SSH audit messages ([00f3ee6](https://github.com/somaz94/bash-pilot/commit/00f3ee65633030b2d40f2c540fdcb6aeddddf093))

### Contributors

- somaz

<br/>

## [v0.4.0](https://github.com/somaz94/bash-pilot/compare/v0.3.0...v0.4.0) (2026-03-20)

### Features

- add prompt module with init and show subcommands ([20e690c](https://github.com/somaz94/bash-pilot/commit/20e690c81fb392dc5c09b75784f3e10c28c0a0ab))

### Bug Fixes

- suppress SIGPIPE error in demo prompt init phase ([9cc6ffa](https://github.com/somaz94/bash-pilot/commit/9cc6ffae3097a589073cb75e172b9787c067237b))
- apply gofmt formatting to prompt helpers ([093f3e5](https://github.com/somaz94/bash-pilot/commit/093f3e540d3ce04999b0f0788d87f5fea123c1af))

### Contributors

- somaz

<br/>

## [v0.3.0](https://github.com/somaz94/bash-pilot/compare/v0.2.0...v0.3.0) (2026-03-20)

### Features

- add env module with check and path subcommands ([0dfca6f](https://github.com/somaz94/bash-pilot/commit/0dfca6f53025d10b20c2f1afd49861641f19754e))

### Tests

- improve env module coverage to 99.2% ([9630f14](https://github.com/somaz94/bash-pilot/commit/9630f1411fb64fa8233de89a6b2b255153fb246c))

### Contributors

- somaz

<br/>

## [v0.2.0](https://github.com/somaz94/bash-pilot/compare/v0.1.1...v0.2.0) (2026-03-20)

### Features

- add git module (profiles, doctor, clean) ([631f499](https://github.com/somaz94/bash-pilot/commit/631f499401893d2a89f06f949b072155f1c562c6))

### Bug Fixes

- handle case-insensitive includeIf in gitconfig parser ([33c0027](https://github.com/somaz94/bash-pilot/commit/33c002709d2e8751a63578dbd9b717b53f641f92))

### Documentation

- add git module commands to Homebrew caveats ([af4f576](https://github.com/somaz94/bash-pilot/commit/af4f576b5a954211deafd0bf6745044b1e034805))

### Contributors

- somaz

<br/>

## [v0.1.1](https://github.com/somaz94/bash-pilot/compare/v0.1.0...v0.1.1) (2026-03-20)

### Features

- generate wildcard patterns in init command ([a5d8ffb](https://github.com/somaz94/bash-pilot/commit/a5d8ffbd6bccff681c7db372f3177ff8c0dfa217))

### Bug Fixes

- Shell Completion ([1407594](https://github.com/somaz94/bash-pilot/commit/1407594cb86dde45a51543c76cc1e980695d4af5))
- update caveats with init command and remove unreleased modules ([996b69d](https://github.com/somaz94/bash-pilot/commit/996b69da3f1f58fd897f68d06efaa3a07fc571fb))

### Documentation

- add macOS zsh to bash switch guide ([a97af15](https://github.com/somaz94/bash-pilot/commit/a97af1519d334461483d6522b6ffba4f94792a5f))
- note bash-completion@2 requirement for macOS bash completion ([295abe5](https://github.com/somaz94/bash-pilot/commit/295abe5453673cf8420d80900aeb2cf08259da65))
- fix bash completion path to use brew --prefix on macOS ([7320807](https://github.com/somaz94/bash-pilot/commit/732080753e14b4d1f09b34cce4a7387094444170))
- add sudo for bash completion on macOS ([5e9a21c](https://github.com/somaz94/bash-pilot/commit/5e9a21c92b121a6edc2d5c3bf03aafd41f5aec08))
- add mkdir -p for bash completion directory on macOS ([1ed9f4d](https://github.com/somaz94/bash-pilot/commit/1ed9f4d528ba9ddbed41961efbb4b9a3e6ad67b0))
- separate bash completion instructions for macOS and Linux ([565d666](https://github.com/somaz94/bash-pilot/commit/565d666cc4196452a0ffd6d5abaa2a4834607029))
- add shell completion setup instructions ([1c59386](https://github.com/somaz94/bash-pilot/commit/1c593868c4a5957592b26650e0f18cf5b9579335))
- update DEVELOPMENT.md with workflow targets and init command ([5c37a7a](https://github.com/somaz94/bash-pilot/commit/5c37a7a9c96a426ac9bbaf1c2164719454eab0b6))

### Continuous Integration

- add GitLab mirror workflow ([a60f727](https://github.com/somaz94/bash-pilot/commit/a60f72771d9bc830a8e3c610dcb175c78ff4c029))

### Contributors

- somaz

<br/>

## [v0.1.0](https://github.com/somaz94/bash-pilot/releases/tag/v0.1.0) (2026-03-20)

### Features

- add Scoop bucket support for Windows distribution ([6a43d5b](https://github.com/somaz94/bash-pilot/commit/6a43d5b051aef3ae5b1bb4c927d91581c7199517))
- add init command to auto-generate config from SSH config ([a86041d](https://github.com/somaz94/bash-pilot/commit/a86041d387afc217677bfebfcb01ace61d56d827))
- add gh CLI pre-check before PR creation ([61e67bd](https://github.com/somaz94/bash-pilot/commit/61e67bd7665f85f477c499a132cd1ebd1815ecb3))
- add branch and pr workflow targets to Makefile ([5e3765c](https://github.com/somaz94/bash-pilot/commit/5e3765c8dbd652c4b82bae5c45fb3f90a265a7a2))
- add post-install message to curl installer ([99ca83c](https://github.com/somaz94/bash-pilot/commit/99ca83c490747ba103bfae66e5f2fe903c74f224))
- add demo scripts with make demo/demo-clean/demo-all ([907e887](https://github.com/somaz94/bash-pilot/commit/907e88763c9f24e052759ffab2b49ff19b1059db))
- implement SSH module (list, ping, audit) ([67ad715](https://github.com/somaz94/bash-pilot/commit/67ad715031827060c22a0e34ea65510362986a47))
- initialize project structure with CLI framework ([7b0c86c](https://github.com/somaz94/bash-pilot/commit/7b0c86cdf34505761ecdb0e6c57624b831a5b440))

### Bug Fixes

- remove broken pipe in demo init phase ([ed6032e](https://github.com/somaz94/bash-pilot/commit/ed6032eb4f8aab04314a2295507696184f3cd0f9))
- apply gofmt formatting to root.go and output.go ([137ea31](https://github.com/somaz94/bash-pilot/commit/137ea310f51eca5694e607e7573dacab4f661dd5))

### Documentation

- add CI, license, tag, and language badges to README ([9f0d942](https://github.com/somaz94/bash-pilot/commit/9f0d9422d7ad338293ef503104d88f3e76d0a453))
- replace real environment values with sample placeholders ([14372a1](https://github.com/somaz94/bash-pilot/commit/14372a1888afda849a7f08c6c8c795971651960c))
- add initial setup guide and improve .gitignore ([fcdfdc6](https://github.com/somaz94/bash-pilot/commit/fcdfdc6787a3e1b5987dfd528a2be3d99dd1c909))
- add curl installer, docs folder, and README restructure ([e2f3ec9](https://github.com/somaz94/bash-pilot/commit/e2f3ec9e29a9f9e7d3bc788e5eee9273377b0f5d))
- update README with features, Homebrew install, and CLAUDE.md workflow rules ([891c1e1](https://github.com/somaz94/bash-pilot/commit/891c1e108e2faf7fe6284fce3e3d4efed5e6e991))

### Tests

- expand SSH test coverage to 96% and refactor parseKeyValue ([76a9cd7](https://github.com/somaz94/bash-pilot/commit/76a9cd76cef948342a9d067b7365fa66f0487574))
- add tests for config, report, and ping packages ([29edb16](https://github.com/somaz94/bash-pilot/commit/29edb168bcf3c7065ce55d4ef8c7c9d4b50fd168))

### Continuous Integration

- add GitHub Actions workflows and GoReleaser config ([232f242](https://github.com/somaz94/bash-pilot/commit/232f242cd85f8970884cc2b3017c5922481702b7))

### Contributors

- somaz

<br/>

