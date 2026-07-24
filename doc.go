// Package todoist is a client for the unified Todoist API v1
// (https://developer.todoist.com/api/v1/).
//
// It has no external dependencies, takes a context.Context on every call, and
// exposes cursor pagination both as raw pages and as range-over-function
// iterators.
//
// # Authentication
//
// Construct a client with a personal or OAuth2 token:
//
//	api := todoist.New(os.Getenv("TODOIST_TOKEN"))
//
// # Pagination
//
// List endpoints have two forms: a low-level page fetch (for example GetTasks)
// returning a [Page], and an auto-paginating iterator (for example Tasks)
// returning an [iter.Seq2] that transparently follows the cursor:
//
//	for task, err := range api.Tasks(ctx, nil) {
//	    if err != nil {
//	        return err
//	    }
//	    fmt.Println(task.Content)
//	}
//
// # Errors
//
// Non-2xx responses are returned as [*Error], which carries the status, body,
// and (on HTTP 429) a RetryAfter duration.
//
// # Filters
//
// Todoist API v1 has no REST endpoints for filters; they are managed through the
// Sync API. See [Client.Sync] and [NewCommand].
package todoist
