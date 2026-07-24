package todoist

import (
	"context"
	"iter"
	"net/url"
	"strconv"
	"time"
)

// GetCompletedTasksArgs are the parameters for the completed-task endpoints.
// Since and Until define the (inclusive, exclusive) time range and are required;
// the remaining fields are optional filters and pagination controls.
type GetCompletedTasksArgs struct {
	Since       time.Time
	Until       time.Time
	WorkspaceID string
	ProjectID   string
	SectionID   string
	ParentID    string
	FilterQuery string
	FilterLang  string
	Cursor      string
	Limit       int
}

func (a *GetCompletedTasksArgs) query() url.Values {
	q := url.Values{}
	if a == nil {
		return q
	}
	if !a.Since.IsZero() {
		q.Set("since", a.Since.UTC().Format(time.RFC3339))
	}
	if !a.Until.IsZero() {
		q.Set("until", a.Until.UTC().Format(time.RFC3339))
	}
	if a.WorkspaceID != "" {
		q.Set("workspace_id", a.WorkspaceID)
	}
	if a.ProjectID != "" {
		q.Set("project_id", a.ProjectID)
	}
	if a.SectionID != "" {
		q.Set("section_id", a.SectionID)
	}
	if a.ParentID != "" {
		q.Set("parent_id", a.ParentID)
	}
	if a.FilterQuery != "" {
		q.Set("filter_query", a.FilterQuery)
	}
	if a.FilterLang != "" {
		q.Set("filter_lang", a.FilterLang)
	}
	if a.Cursor != "" {
		q.Set("cursor", a.Cursor)
	}
	if a.Limit > 0 {
		q.Set("limit", strconv.Itoa(a.Limit))
	}
	return q
}

// GetCompletedByCompletionDate returns a page of tasks completed within the
// args' time range, keyed by completion date.
func (c *Client) GetCompletedByCompletionDate(ctx context.Context, args *GetCompletedTasksArgs) (Page[Task], error) {
	return doItemsList[Task](ctx, c, "/tasks/completed/by_completion_date", args.query())
}

// CompletedByCompletionDate iterates over all tasks completed within the time
// range, following pagination.
func (c *Client) CompletedByCompletionDate(ctx context.Context, args *GetCompletedTasksArgs) iter.Seq2[Task, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[Task], error) {
		return doItemsList[Task](ctx, c, "/tasks/completed/by_completion_date", setCursor(base, cursor))
	})
}

// GetCompletedByDueDate returns a page of tasks completed within the args' time
// range, keyed by due date.
func (c *Client) GetCompletedByDueDate(ctx context.Context, args *GetCompletedTasksArgs) (Page[Task], error) {
	return doItemsList[Task](ctx, c, "/tasks/completed/by_due_date", args.query())
}

// CompletedByDueDate iterates over all tasks completed within the time range,
// keyed by due date, following pagination.
func (c *Client) CompletedByDueDate(ctx context.Context, args *GetCompletedTasksArgs) iter.Seq2[Task, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[Task], error) {
		return doItemsList[Task](ctx, c, "/tasks/completed/by_due_date", setCursor(base, cursor))
	})
}

// GetTasksByFilterArgs are the parameters for querying tasks with a filter
// string. Query is required.
type GetTasksByFilterArgs struct {
	Query  string
	Lang   string
	Cursor string
	Limit  int
}

func (a *GetTasksByFilterArgs) query() url.Values {
	q := url.Values{}
	if a == nil {
		return q
	}
	if a.Query != "" {
		q.Set("query", a.Query)
	}
	if a.Lang != "" {
		q.Set("lang", a.Lang)
	}
	if a.Cursor != "" {
		q.Set("cursor", a.Cursor)
	}
	if a.Limit > 0 {
		q.Set("limit", strconv.Itoa(a.Limit))
	}
	return q
}

// GetTasksByFilter returns a page of tasks matching a Todoist filter query. This
// is the v1 replacement for the removed filters REST endpoints.
func (c *Client) GetTasksByFilter(ctx context.Context, args *GetTasksByFilterArgs) (Page[Task], error) {
	return doList[Task](ctx, c, "/tasks/filter", args.query())
}

// TasksByFilter iterates over all tasks matching a filter query, following
// pagination.
func (c *Client) TasksByFilter(ctx context.Context, args *GetTasksByFilterArgs) iter.Seq2[Task, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[Task], error) {
		return doList[Task](ctx, c, "/tasks/filter", setCursor(base, cursor))
	})
}
