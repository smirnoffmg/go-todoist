# go-todoist

A clean, idiomatic Go client for the [Todoist API v1](https://developer.todoist.com/api/v1/) —
the unified API that merges the former REST and Sync APIs.

- Zero external dependencies (standard library only).
- `context.Context` on every call.
- Cursor pagination exposed both as raw pages and as `iter.Seq2` iterators.
- REST coverage for tasks, projects, sections, labels, comments, reminders,
  workspaces, and user, plus the batched Sync API. (Filters have no REST
  endpoint in v1 and are managed via Sync.)

Requires Go 1.23+ (uses range-over-function iterators).

## Install

```sh
go get github.com/smirnoffmg/go-todoist
```

## Authentication

Create a client with a personal API token (Todoist → Settings → Integrations →
Developer) or an OAuth2 access token:

```go
api := todoist.New(os.Getenv("TODOIST_TOKEN"))
```

Options: `WithHTTPClient`, `WithBaseURL`, `WithUserAgent`.

## Tasks and pagination

Each list resource has two forms: a low-level page fetch (`GetTasks`) and an
auto-paginating iterator (`Tasks`) that transparently follows the cursor.

```go
ctx := context.Background()
api := todoist.New(os.Getenv("TODOIST_TOKEN"))

// Iterate over every task, following pagination automatically.
for task, err := range api.Tasks(ctx, &todoist.GetTasksArgs{ProjectID: "220474322"}) {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(task.ID, task.Content)
}

// Or fetch a single page and manage the cursor yourself.
page, err := api.GetTasks(ctx, &todoist.GetTasksArgs{Limit: 50})
if err != nil {
    log.Fatal(err)
}
fmt.Println(len(page.Results), page.NextCursor)
```

Create, update, and complete a task:

```go
task, err := api.CreateTask(ctx, todoist.CreateTaskArgs{
    Content:   "Buy milk",
    DueString: todoist.Ptr("tomorrow at 9am"),
    Priority:  todoist.Ptr(todoist.PriorityHigh),
})
if err != nil {
    log.Fatal(err)
}

_, _ = api.UpdateTask(ctx, task.ID, todoist.UpdateTaskArgs{Content: todoist.Ptr("Buy oat milk")})
_ = api.CloseTask(ctx, task.ID)
```

Optional request fields are pointers so an unset field is omitted rather than
sent as a zero value. Use the `Ptr` helper to set them inline.

## Error handling

Non-2xx responses return an `*todoist.Error` carrying the status and body. On
rate limiting (429) it also exposes `RetryAfter`.

```go
_, err := api.GetTask(ctx, "bad-id")
var apiErr *todoist.Error
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.RetryAfter)
    if apiErr.Temporary() {
        // safe to retry after apiErr.RetryAfter
    }
}
```

## Sync API

The Sync endpoint batches multiple write commands in one round trip and supports
incremental sync via a persisted token. Commands can reference each other's
`temp_id`.

```go
addProject := todoist.NewCommand("project_add", "proj-tmp", map[string]any{"name": "Launch"})
addTask := todoist.NewCommand("item_add", "", map[string]any{
    "content":    "Draft announcement",
    "project_id": "proj-tmp", // resolved via temp_id mapping
})

resp, err := api.Sync(ctx, todoist.SyncRequest{
    ResourceTypes: []string{"projects", "items"},
    Commands:      []todoist.Command{addProject, addTask},
})
if err != nil {
    log.Fatal(err)
}
if err := resp.Err(); err != nil {
    log.Fatal(err) // aggregates per-command failures
}

realProjectID := resp.TempIDMapping["proj-tmp"]
fmt.Println("created project", realProjectID)

// Persist resp.SyncToken and pass it back next time for an incremental sync.
```

## Development

```sh
make lint   # golangci-lint run
make test   # go test -race -cover ./... (offline; 100% coverage)
```

### Testing against real Todoist

Integration tests hit the live API. They are gated behind the `integration`
build tag so they never run in the normal offline suite or CI, and they skip
automatically unless `TODOIST_TOKEN` is set. Each test creates its own scratch
project/label and deletes it on cleanup, so your existing data is left untouched.

Get a personal token from Todoist → Settings → Integrations → Developer, then:

```sh
export TODOIST_TOKEN=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
make test-integration
# or:
go test -tags=integration -run Integration -v ./...
```

These tests cover the full round trip: authenticating (`GetUser`), the task
lifecycle (create → get → update → close → reopen → delete), sections and
comments, labels, project pagination via the iterator, and a read-only `Sync`.

## License

See [LICENSE](LICENSE).
