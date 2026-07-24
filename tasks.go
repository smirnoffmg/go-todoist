package todoist

import (
	"context"
	"iter"
	"net/url"
	"strconv"
)

// Task is a Todoist task (an "item" in Sync API terms).
type Task struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	SectionID    string    `json:"section_id"`
	ParentID     string    `json:"parent_id"`
	Content      string    `json:"content"`
	Description  string    `json:"description"`
	Priority     int       `json:"priority"`
	Labels       []string  `json:"labels"`
	Due          *Due      `json:"due"`
	Deadline     *Deadline `json:"deadline"`
	Duration     *Duration `json:"duration"`
	AssigneeID   string    `json:"assignee_id"`
	AssignerID   string    `json:"assigner_id"`
	Order        int       `json:"child_order"`
	CommentCount int       `json:"comment_count"`
	IsCompleted  bool      `json:"checked"`
	AddedAt      string    `json:"added_at"`
	CompletedAt  string    `json:"completed_at"`
	URL          string    `json:"url"`
}

// GetTasksArgs are the filters and pagination controls for listing tasks. All
// fields are optional.
type GetTasksArgs struct {
	ProjectID string
	SectionID string
	ParentID  string
	Label     string
	Cursor    string
	Limit     int
}

func (a *GetTasksArgs) query() url.Values {
	q := url.Values{}
	if a == nil {
		return q
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
	if a.Label != "" {
		q.Set("label", a.Label)
	}
	if a.Cursor != "" {
		q.Set("cursor", a.Cursor)
	}
	if a.Limit > 0 {
		q.Set("limit", strconv.Itoa(a.Limit))
	}
	return q
}

// CreateTaskArgs are the parameters for creating a task. Content is required;
// pointer fields are omitted from the request when nil.
type CreateTaskArgs struct {
	Content      string   `json:"content"`
	Description  *string  `json:"description,omitempty"`
	ProjectID    *string  `json:"project_id,omitempty"`
	SectionID    *string  `json:"section_id,omitempty"`
	ParentID     *string  `json:"parent_id,omitempty"`
	Order        *int     `json:"order,omitempty"`
	Labels       []string `json:"labels,omitempty"`
	Priority     *int     `json:"priority,omitempty"`
	AssigneeID   *string  `json:"assignee_id,omitempty"`
	DueString    *string  `json:"due_string,omitempty"`
	DueDate      *string  `json:"due_date,omitempty"`
	DueDatetime  *string  `json:"due_datetime,omitempty"`
	DueLang      *string  `json:"due_lang,omitempty"`
	Deadline     *string  `json:"deadline_date,omitempty"`
	Duration     *int     `json:"duration,omitempty"`
	DurationUnit *string  `json:"duration_unit,omitempty"`
}

// UpdateTaskArgs are the parameters for updating a task. Only non-nil fields are
// sent, so the zero value leaves the task unchanged.
type UpdateTaskArgs struct {
	Content      *string  `json:"content,omitempty"`
	Description  *string  `json:"description,omitempty"`
	Labels       []string `json:"labels,omitempty"`
	Priority     *int     `json:"priority,omitempty"`
	AssigneeID   *string  `json:"assignee_id,omitempty"`
	DueString    *string  `json:"due_string,omitempty"`
	DueDate      *string  `json:"due_date,omitempty"`
	DueDatetime  *string  `json:"due_datetime,omitempty"`
	DueLang      *string  `json:"due_lang,omitempty"`
	Deadline     *string  `json:"deadline_date,omitempty"`
	Duration     *int     `json:"duration,omitempty"`
	DurationUnit *string  `json:"duration_unit,omitempty"`
}

// GetTasks returns a single page of tasks matching args.
func (c *Client) GetTasks(ctx context.Context, args *GetTasksArgs) (Page[Task], error) {
	return doList[Task](ctx, c, "/tasks", args.query())
}

// Tasks returns an iterator over all tasks matching args, transparently
// following pagination cursors.
func (c *Client) Tasks(ctx context.Context, args *GetTasksArgs) iter.Seq2[Task, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[Task], error) {
		return doList[Task](ctx, c, "/tasks", setCursor(base, cursor))
	})
}

// GetTask returns a single task by ID.
func (c *Client) GetTask(ctx context.Context, id string) (Task, error) {
	return doGet[Task](ctx, c, "/tasks/"+id)
}

// CreateTask creates a new task.
func (c *Client) CreateTask(ctx context.Context, args CreateTaskArgs) (Task, error) {
	return doPost[Task](ctx, c, "/tasks", args)
}

// UpdateTask updates an existing task and returns the updated resource.
func (c *Client) UpdateTask(ctx context.Context, id string, args UpdateTaskArgs) (Task, error) {
	return doPost[Task](ctx, c, "/tasks/"+id, args)
}

// DeleteTask deletes a task.
func (c *Client) DeleteTask(ctx context.Context, id string) error {
	return doDelete(ctx, c, "/tasks/"+id)
}

// CloseTask marks a task as complete.
func (c *Client) CloseTask(ctx context.Context, id string) error {
	return doAction(ctx, c, "/tasks/"+id+"/close")
}

// ReopenTask reopens a previously completed task.
func (c *Client) ReopenTask(ctx context.Context, id string) error {
	return doAction(ctx, c, "/tasks/"+id+"/reopen")
}

// QuickAddTask creates a task using Todoist's natural-language quick-add syntax.
func (c *Client) QuickAddTask(ctx context.Context, text string) (Task, error) {
	return doPost[Task](ctx, c, "/tasks/quick", map[string]string{"text": text})
}
