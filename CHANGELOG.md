# Changelog

All notable changes to this project are documented in this file.

The format follows Keep a Changelog and Semantic Versioning.

## [Unreleased]

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
