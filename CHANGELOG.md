# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-07-24

Initial release.

### Added

- Client for the unified Todoist API v1 with functional options
  (`WithHTTPClient`, `WithBaseURL`, `WithUserAgent`) and `context.Context` on
  every call.
- REST coverage: tasks (incl. close/reopen/quick-add), projects (incl.
  archive/collaborators), sections, labels, comments, reminders, workspaces, and
  user (incl. completed/productivity stats).
- Cursor pagination exposed both as raw `Page[T]` fetches and as
  auto-paginating `iter.Seq2` iterators.
- Batched Sync API (`Sync`, `Command`, `NewCommand`, `NewUUID`) with `temp_id`
  mapping and per-command error aggregation via `SyncResponse.Err`.
- Typed `*Error` with HTTP status, body, and `Retry-After` handling.
- Zero external runtime dependencies (standard library only).

### Notes

- Filters have no REST endpoint in Todoist API v1 and are managed through the
  Sync API.

[Unreleased]: https://github.com/smirnoffmg/go-todoist/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/smirnoffmg/go-todoist/releases/tag/v0.1.0
