# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-07-24

### Added
- Completed-task endpoints: `GetCompletedByCompletionDate` /
  `GetCompletedByDueDate` (+ iterators) and `GetTasksByFilter` /`TasksByFilter`
  (the practical replacement for the absent filters REST).
- `MoveTask`, `GetArchivedProjects` (+ iterator), `ArchiveSection` /
  `UnarchiveSection`, and `JoinProject`.
- Shared labels: `GetSharedLabels` (+ iterator), `RenameSharedLabel`,
  `RemoveSharedLabel`.
- `location_reminders` resource (list/get/create/update/delete + iterator).
- Opt-in retry/backoff on HTTP 429 and 5xx via `WithRetry`, honoring
  `Retry-After` and respecting context cancellation.
- OAuth helpers: `RevokeToken` and `MigratePersonalToken`.
- Templates: `CreateProjectFromFile`, `ImportIntoProjectFromFile`,
  `ImportIntoProjectFromTemplateID`, `GetTemplateFile`, `GetTemplateURL`.
- Workspace invitations (`GetWorkspaceInvitations`, `GetAllWorkspaceInvitations`,
  `DeleteWorkspaceInvitation`, `AcceptWorkspaceInvitation`,
  `RejectWorkspaceInvitation`), `JoinWorkspace`, and workspace project listings
  (`GetWorkspaceActiveProjects` / `GetWorkspaceArchivedProjects` + iterators).

### Changed
- Lowered the minimum Go version in `go.mod` to 1.23 (was 1.26).

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

[Unreleased]: https://github.com/smirnoffmg/go-todoist/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/smirnoffmg/go-todoist/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/smirnoffmg/go-todoist/releases/tag/v0.1.0
