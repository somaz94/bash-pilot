# Changelog

All notable changes to this project will be documented in this file.

## [v0.9.1](https://github.com/somaz94/bash-pilot/compare/v0.9.0...v0.9.1) (2026-09-28)

### Bug Fixes

- **ssh:** read every ForwardAgent form and carry agent sockets through migrate ([563d838](https://github.com/somaz94/bash-pilot/commit/563d83895ad969f51c20bfbd944ec8fb56c53692))

### Documentation

- describe ForwardAgent socket export in USAGE ([394009d](https://github.com/somaz94/bash-pilot/commit/394009d23d1b6581d368279ecc79e5016c54e784))

### Contributors

- somaz

<br/>

## [v0.9.0](https://github.com/somaz94/bash-pilot/compare/v0.8.0...v0.9.0) (2026-09-28)

### Bug Fixes

- **ssh:** read ssh_config values the way ssh does and re-quote them on migrate import ([63d7a78](https://github.com/somaz94/bash-pilot/commit/63d7a78fa4df39588bf45862ea692ad439fc3a3f))
- **ssh:** use the first value of a repeated ssh_config parameter ([3ca19e8](https://github.com/somaz94/bash-pilot/commit/3ca19e8b218467b206ec010c4d8e86d8773914fb))

### Documentation

- document the migrate export format version ([3a149e5](https://github.com/somaz94/bash-pilot/commit/3a149e5e92f76f94a9904dfb2fe982e9d90188c6))

### Contributors

- somaz

<br/>

## [v0.8.0](https://github.com/somaz94/bash-pilot/compare/v0.7.1...v0.8.0) (2026-09-28)

### Bug Fixes

- **cli:** reject stray arguments and suggest the intended subcommand on a typo ([2330e90](https://github.com/somaz94/bash-pilot/commit/2330e90bc2b8c11354a8f530aaa3420c5a28c9ca))
- **ssh:** ignore trailing comments and Match blocks when parsing ssh_config ([135eab1](https://github.com/somaz94/bash-pilot/commit/135eab16d975c8983dde13738b28085f8db6f446))

### Documentation

- note Include, Match and trailing comment handling in SETUP troubleshooting ([193809d](https://github.com/somaz94/bash-pilot/commit/193809d392238be1c84621a6c20992127cb9be41))

### Contributors

- somaz

<br/>

## [v0.7.1](https://github.com/somaz94/bash-pilot/compare/v0.7.0...v0.7.1) (2026-09-28)

### Bug Fixes

- **cli:** reject more than one pattern in ssh ping ([2f801bb](https://github.com/somaz94/bash-pilot/commit/2f801bb35ffa6379a5d01dec3a206f73d528d6c7))
- **migrate:** detect existing SSH hosts with the ssh_config parser and match each pattern ([bb80b28](https://github.com/somaz94/bash-pilot/commit/bb80b28ab07f79c25a37c4411ecf14ba26c1bd83))
- **git:** keep the final newline when clean drops a trailing empty [safe] section ([c4d9cf2](https://github.com/somaz94/bash-pilot/commit/c4d9cf29917ba39745c7b613edbe1e88344b8249))

### Documentation

- replace the hand-written coverage table and document ping and import rules ([b6ea6ad](https://github.com/somaz94/bash-pilot/commit/b6ea6ad479a7484a52b33a052cbe27a27cfa8fe6))
- sync CLAUDE.md and docs with the current CLI commands, output format and JSON fields ([6527869](https://github.com/somaz94/bash-pilot/commit/65278694cabfb05b29f2a1877cdc55fddb7c79fa))

### Continuous Integration

- run golangci-lint pinned to v2.13 in the unit-tests job ([e8be78f](https://github.com/somaz94/bash-pilot/commit/e8be78f63f318b0f81190be94a02b001fef803cc))

### Chores

- add make lint target for golangci-lint ([1281717](https://github.com/somaz94/bash-pilot/commit/1281717f7d647819d498fa6a8be63bbe89a955b2))

### Contributors

- somaz

<br/>

## [v0.7.0](https://github.com/somaz94/bash-pilot/compare/v0.6.3...v0.7.0) (2026-09-28)

### Bug Fixes

- **cli:** fail on an unreadable --config, report an unreadable SSH config in doctor, and print errors once ([8534972](https://github.com/somaz94/bash-pilot/commit/8534972b1db95213f99aad686e29bddb4326d0da))
- **snapshot:** scope includeIf email lookup to its own section and drop removed kubectl --short ([9481688](https://github.com/somaz94/bash-pilot/commit/94816886f751acdad7758072cd487878621623a7))
- **ssh:** follow ssh_config(5) syntax, split ProxyCommand from ProxyJump, and stabilize audit output ([87d14c6](https://github.com/somaz94/bash-pilot/commit/87d14c6959507a376e5629087271797671ac17e0))

### Documentation

- document config fallback, ProxyCommand export, and prefixed audit messages ([eb2ff84](https://github.com/somaz94/bash-pilot/commit/eb2ff84a31b5bb9d5c186efcec04fbe0ef43d149))

### Tests

- make MakeDir umask-proof and assert exact gitconfig clean output ([63d394a](https://github.com/somaz94/bash-pilot/commit/63d394ab51adac3fd319ae215bd3bd6718701c0a))

### Continuous Integration

- retry mirror pushes on transient remote failures ([bfc0c0a](https://github.com/somaz94/bash-pilot/commit/bfc0c0a9da76b697f6ef43340b82bb409204e813))
- drop the dead issue-close trigger from changelog generation ([fe726e7](https://github.com/somaz94/bash-pilot/commit/fe726e711828bef4f5b23eb3b67648c696d1dc66))

### Chores

- replace routable IPs, LAN-shaped octets and personal paths in examples and fixtures ([1e96fdd](https://github.com/somaz94/bash-pilot/commit/1e96fdd97d01f0bb5dd23ea75aa5f09337c38c78))
- tighten remaining comments to their non-obvious why ([5bc33f4](https://github.com/somaz94/bash-pilot/commit/5bc33f496b0c31a99fd097070b41178e6d3fbcef))
- trim stale and redundant comments in internal git, migrate and snapshot ([b2f4276](https://github.com/somaz94/bash-pilot/commit/b2f42764f43aea6770b4d05bf869886f6e9eca3c))
- trim stale and redundant comments in internal ssh, env, prompt, report and config ([519a0e9](https://github.com/somaz94/bash-pilot/commit/519a0e97a2445c6ce273d7b768373064d5250a04))
- trim stale and redundant comments in cmd/cli ([1302da2](https://github.com/somaz94/bash-pilot/commit/1302da23d3ca92edb7e6c199acfb6c9bdbce7df5))
- trim stale and redundant comments in scripts ([f73b9c8](https://github.com/somaz94/bash-pilot/commit/f73b9c8939f43f778da1d89cafd2ec548f252806))

### Contributors

- somaz

<br/>

## [v0.6.3](https://github.com/somaz94/bash-pilot/compare/v0.6.2...v0.6.3) (2026-08-14)

### Bug Fixes

- report a failed SSH or gitconfig write instead of claiming success ([a12bda6](https://github.com/somaz94/bash-pilot/commit/a12bda6d128e7a9ac9bd321dc9cdc03ea91016e0))

### Tests

- replace real cloud host IPs in fixtures with example values ([19e1020](https://github.com/somaz94/bash-pilot/commit/19e10206553dd5e522a79a826c599a0e8be1a23c))
- replace real host inventory in fixtures with example values ([3e14db3](https://github.com/somaz94/bash-pilot/commit/3e14db39df5e7bab58e24ac2438d31da92d3aa68))

### Continuous Integration

- add a golangci-lint config scoped to defect-finding linters ([c92cd9c](https://github.com/somaz94/bash-pilot/commit/c92cd9c6f9731b023f460e7428e37091993b7603))
- remove DCO workflow ([68062c6](https://github.com/somaz94/bash-pilot/commit/68062c6253273877c31df03a9586e65831d4cfb6))
- adopt semantic-pr, labels, lock-threads, PR size, and auto-assign reusables ([53318b7](https://github.com/somaz94/bash-pilot/commit/53318b705b2884f1bc0209667fe984e13de4d481))
- use reusable stale-issues workflow ([cec2232](https://github.com/somaz94/bash-pilot/commit/cec2232c9f86fb780fa52b3cca7eba194ca68341))
- use reusable issue-greeting workflow ([f65886f](https://github.com/somaz94/bash-pilot/commit/f65886f8726e191a9078d6f1aab725499bae9640))
- use reusable dependabot-auto-merge workflow ([608dc40](https://github.com/somaz94/bash-pilot/commit/608dc404fda739056a55ec04246971d3ca80f37e))
- use reusable contributors workflow ([133eebf](https://github.com/somaz94/bash-pilot/commit/133eebf34535f652921aa803edd5508cdfa01665))
- add ok-to-test workflow stub ([35d8261](https://github.com/somaz94/bash-pilot/commit/35d82613e31e3fb1d70993e4986dd3ed30a23da7))
- add PR welcome workflow stub ([c9cad12](https://github.com/somaz94/bash-pilot/commit/c9cad1285cf5b10893f0b33ca56b6b757d487265))
- add DCO check via shared reusable workflow ([512df00](https://github.com/somaz94/bash-pilot/commit/512df00d1e81bb9a6965f75bb34e1340438e7db9))

### Chores

- publish the Homebrew package as a cask instead of a deprecated formula ([159c072](https://github.com/somaz94/bash-pilot/commit/159c07276476164898ea44aea9795dee996e9f48))
- **deps:** bump actions/setup-go from 6 to 7 (#14) ([#14](https://github.com/somaz94/bash-pilot/pull/14)) ([5636449](https://github.com/somaz94/bash-pilot/commit/5636449b1c93ea6ceb6f9805ef9dd0698eff1010))
- **deps:** bump actions/checkout from 6 to 7 (#13) ([#13](https://github.com/somaz94/bash-pilot/pull/13)) ([e49177c](https://github.com/somaz94/bash-pilot/commit/e49177cd2cb052d0869a6b8cff4e004c6544bca3))

### Contributors

- somaz

<br/>

## [v0.6.2](https://github.com/somaz94/bash-pilot/compare/v0.6.1...v0.6.2) (2026-05-28)

### Code Refactoring

- use internal/testutil in snapshot tests ([62e22eb](https://github.com/somaz94/bash-pilot/commit/62e22eb3fd28496af9c3fef249f3b3835ed971ee))
- use internal/testutil in env tests ([9f1ff93](https://github.com/somaz94/bash-pilot/commit/9f1ff933062b58e304741067277a5fa7f8cad15b))
- introduce internal/testutil and use it in git tests ([b24de53](https://github.com/somaz94/bash-pilot/commit/b24de539126c6a2dba01830d03241052fa2398a5))
- extract file permission magic numbers to config.Perm* ([bc0bdeb](https://github.com/somaz94/bash-pilot/commit/bc0bdeb00085252589a4d216e031509b7936d6fa))
- split git.Clean into backup + rewrite helpers ([b25b0af](https://github.com/somaz94/bash-pilot/commit/b25b0af4dc659b377ff75adda6cb32a4b738eed5))
- split snapshot.captureGit, extract parseIncludeIfProfiles ([4020971](https://github.com/somaz94/bash-pilot/commit/40209712f683af0046654647d7562ce312956b51))
- unify severity rendering in report.RenderSeverity ([79b39f5](https://github.com/somaz94/bash-pilot/commit/79b39f525cf114ba7605497418edabebcbb320bd))
- reuse snapshot.ParseOnly in migrate import ([736a42d](https://github.com/somaz94/bash-pilot/commit/736a42d63a08a35092045aea20c9667638b586c9))
- make shell scripts bash+zsh portable and safer ([b692262](https://github.com/somaz94/bash-pilot/commit/b692262150d1d55069624ed3686f1baa934cfd5c))

### Continuous Integration

- add concurrency guards to recurring workflows ([e9c158d](https://github.com/somaz94/bash-pilot/commit/e9c158d7ee43496cf425c4a39886506c84dc0525))

### Contributors

- somaz

<br/>

## [v0.6.1](https://github.com/somaz94/bash-pilot/compare/v0.6.0...v0.6.1) (2026-03-31)

### Bug Fixes

- correct ping error message, git active profile detection, doctor gitconfig path ([d90bbc8](https://github.com/somaz94/bash-pilot/commit/d90bbc8200a42cf90412895974aa18d0041fe04f))

### Code Refactoring

- replace eval with direct exec in demo.sh, cross-platform cover-html, add stars badge ([b524a8d](https://github.com/somaz94/bash-pilot/commit/b524a8d09ad0fb3b7c796c51a14a59b1825f3470))

### Documentation

- remove duplicate rules covered by global CLAUDE.md ([4529d78](https://github.com/somaz94/bash-pilot/commit/4529d786f662520c3a0b5bbf23ec43596107281f))

### Tests

- add connection refused and git active profile false positive tests ([cbfa2fa](https://github.com/somaz94/bash-pilot/commit/cbfa2fa6321cfbd61e7842f77a417f45bed9cd7d))

### Continuous Integration

- add changelog category groups in goreleaser config ([0161bb2](https://github.com/somaz94/bash-pilot/commit/0161bb2b729f45bdf773e045fd69311252cdcce7))

### Chores

- remove duplicate rules from CLAUDE.md (moved to global) ([11aed46](https://github.com/somaz94/bash-pilot/commit/11aed4639e07e64fa93910494178b61c4e9fa427))
- add git config protection to CLAUDE.md ([218c329](https://github.com/somaz94/bash-pilot/commit/218c329bdfd9b8df85aa638fbb7a188a05a15b53))

### Contributors

- somaz

<br/>

## [v0.6.0](https://github.com/somaz94/bash-pilot/compare/v0.5.0...v0.6.0) (2026-03-20)

### Features

- add auto-generated PR body script for make pr ([570d4f8](https://github.com/somaz94/bash-pilot/commit/570d4f88650974b73954fc40743d593f4553f51a))
- wire --only flag to diff, setup, migrate import CLI commands ([032b705](https://github.com/somaz94/bash-pilot/commit/032b70579bdcdbce374f004d896e545437503575))
- add --only flag for section filtering in diff, setup, migrate import ([153be71](https://github.com/somaz94/bash-pilot/commit/153be7186dca6f4ce430168c34a90d20cdb4e450))
- add migrate command for cross-machine config migration ([8ad451c](https://github.com/somaz94/bash-pilot/commit/8ad451c4f8bc7dcf4850eab5e1accca128741d6d))
- add setup command to install missing tools from snapshot ([78f99f2](https://github.com/somaz94/bash-pilot/commit/78f99f2fecebe695d856a0299409649e43e74886))
- add snapshot and diff commands for environment comparison ([644cace](https://github.com/somaz94/bash-pilot/commit/644cace92e47c999e2eff6dc399dcffa4b70a81b))

### Bug Fixes

- add user.name support in migrate git profile export/import ([0d98376](https://github.com/somaz94/bash-pilot/commit/0d983765b92cfa16d3eba05e4ac74c48b82d4910))
- copy SSH config to ~/.ssh/config in demo for migrate export ([63e681c](https://github.com/somaz94/bash-pilot/commit/63e681c62943d4d99a06b9b0be89d2881b5628e1))
- prevent Header panic when title exceeds 50 characters ([fcf6d2e](https://github.com/somaz94/bash-pilot/commit/fcf6d2e17699f6fa616ee376495f7693acc0037c))

### Documentation

- add --only flag usage and update goreleaser caveats ([a5cebd1](https://github.com/somaz94/bash-pilot/commit/a5cebd18396e824bcfdd7e15f4de020d4116b206))
- add snapshot/diff/setup use cases ([4c607a6](https://github.com/somaz94/bash-pilot/commit/4c607a696ecf85dc7f5ad6bba9698d16c4d9eb0b))

### Tests

- add tests for --only flag in diff and migrate import ([bea038a](https://github.com/somaz94/bash-pilot/commit/bea038a15c1038ae323232da5b8eb0cc9627f8cf))

### Continuous Integration

- limit push trigger to main branch only ([06597cb](https://github.com/somaz94/bash-pilot/commit/06597cbe9001532beb22359b1191a0d5e61acc75))

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

