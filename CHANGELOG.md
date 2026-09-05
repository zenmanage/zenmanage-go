# Changelog

All notable changes to this project are documented in this file.

The format follows Keep a Changelog and Semantic Versioning.

## [Unreleased]

### Added

- `zenmanage.Error` interface, implemented by every error type this SDK returns (`ConfigurationError`, `EvaluationError`, `FetchRulesError`, `InvalidRulesError`), matching the shared error base class in the JavaScript, PHP, and Python SDKs. Use `errors.As(err, &zmErr)` to distinguish SDK errors from other errors.
- `FlagManager.ReportUsage()` — an explicit, manually-callable usage reporting method, matching the JavaScript/PHP/Python SDKs. `Single()` continues to report usage automatically; this is for reporting usage outside of a `Single`/`All` evaluation path.
- Cross-SDK CRC32B/rollout-bucketing test vectors (`rollout_test.go`), matching the fixed spec vectors used in `zenmanage-php`'s `RolloutBucketTest.php` and `zenmanage-javascript`'s `rollout.test.ts`, to lock in deterministic rollout bucketing across languages explicitly.
- Test suites for the `middleware/gin` and `middleware/echo` submodules (previously untested).
- CI now builds and tests `middleware/gin` and `middleware/echo` as part of the standard test matrix, and enforces a per-module code coverage threshold.
- README: "Key Compatibility", "Error Handling", "Fetch All Flags", and "Manual Usage Reporting" sections, and `examples/README.md` now lists `middleware.go`.

### Fixed

- `middleware/gin` and `middleware/echo` referenced `zenmanage-go v0.1.0` from the Go module proxy, but that version was never tagged, so both submodules failed to resolve their dependency and could not build or be tested at all. Each submodule's `go.mod` now carries a `replace` directive pointing at the sibling module source, which only takes effect when building that submodule directly (ignored by downstream consumers) — this unblocks local development and CI until a real tagged release exists.
- `FlagManager.All()` no longer reports per-flag usage as it evaluates the collection. Usage reporting stays on the `Single()` path, matching every other SDK — `All()` is a retrieval, not a usage signal.
- `FlagManager.Single()` now reports the effective default value (inline parameter, falling back to a `DefaultsCollection` entry) on every usage report, including when the flag is found and evaluated normally. Previously the default was only sent on the fallback paths.

### Changed

- Renamed request headers to use a consistent `X-ZEN-` prefix, matching the server-side convention and the JavaScript/PHP SDKs: `X-API-Key` → `X-ZEN-API-KEY`, `X-ZENMANAGE-CONTEXT` → `X-ZEN-CONTEXT`, and `X-Default-Value` → `X-ZEN-DEFAULT-VALUE`. The API still accepts the old header names as legacy aliases, so this is non-breaking.

## [0.1.0] - 2026-05-07

### Added

- Initial Zenmanage Go SDK implementation
- Config builder with environment loading support
- API client with retry logic and usage reporting
- Flag manager with cache-backed rule loading
- Rule engine with parity operators and context targeting
- Percentage rollout bucketing via CRC32B
- In-memory, filesystem, and null cache backends
- Default value collection and fallback precedence
- Full parity examples aligned with other SDKs
- High-coverage unit test suite
- Publishing guide for public distribution
