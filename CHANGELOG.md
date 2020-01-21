# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.5.0]
- Added cassandra row_id with TIMEUUID for long-polling [#25](https://github.com/xmidt-org/codex-db/pull/25)

## [v0.4.0]
- Modified retry package to use backoff package for exponential backoffs on retries [#21](https://github.com/xmidt-org/codex-db/pull/21)
- Added automated releases using travis [#22](https://github.com/xmidt-org/codex-db/pull/22)

## [v0.3.3]
- fix read error causing corrupt data

## [v0.3.2]
- Updated batchInserter to have a configurable amount of batchers [#18](https://github.com/xmidt-org/codex-db/pull/18)

## [v0.3.1]
- Fixed typo in variable name [[#15](https://github.com/xmidt-org/codex-db/pull/15)]
- Fix metric cardinality [[#17](https://github.com/xmidt-org/codex-db/pull/17)]

## [v0.3.0]
- Bump webpa-common [[#12](https://github.com/xmidt-org/codex-db/pull/12)]
- Bump go-kit [[#12](https://github.com/xmidt-org/codex-db/pull/12)]
- Fix go-health version [[#12](https://github.com/xmidt-org/codex-db/pull/12)]
- Change cassandra deviceList getter functionality [[#12](https://github.com/xmidt-org/codex-db/pull/12)]

## [v0.2.0]
- adding ycql support

## [v0.1.2]
- switched webpa-common/wrp to wrp-go
- bumped webpa-common
- bumped capacityset

## [v0.1.1]
- Changed ugorji dependency

## [v0.1.0]
- Initial creation, moved from: https://github.com/xmidt-org/codex-deploy

[Unreleased]: https://github.com/xmidt-org/codex-db/compare/v0.5.0..HEAD
[v0.5.0]: https://github.com/xmidt-org/codex-db/compare/v0.4.0..v0.5.0
[v0.4.0]: https://github.com/xmidt-org/codex-db/compare/v0.3.3..v0.4.0
[v0.3.3]: https://github.com/xmidt-org/codex-db/compare/v0.3.2..v0.3.3
[v0.3.2]: https://github.com/xmidt-org/codex-db/compare/v0.3.1..v0.3.2
[v0.3.1]: https://github.com/xmidt-org/codex-db/compare/v0.3.0..v0.3.1
[v0.3.0]: https://github.com/xmidt-org/codex-db/compare/v0.2.0..v0.3.0
[v0.2.0]: https://github.com/xmidt-org/codex-db/compare/0.1.2...v0.2.0
[v0.1.2]: https://github.com/xmidt-org/codex-db/compare/0.1.1...v0.1.2
[v0.1.1]: https://github.com/xmidt-org/codex-db/compare/0.1.0...v0.1.1
[v0.1.0]: https://github.com/xmidt-org/codex-db/compare/0.0.0...v0.1.0
