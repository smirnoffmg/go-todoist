package todoist

import (
	"context"
	"iter"
	"net/url"
	"strconv"
)

// Comment is a comment on a task or project.
type Comment struct {
	ID         string      `json:"id"`
	TaskID     string      `json:"task_id"`
	ProjectID  string      `json:"project_id"`
	Content    string      `json:"content"`
	PostedAt   string      `json:"posted_at"`
	PostedUID  string      `json:"posted_uid"`
	Attachment *Attachment `json:"attachment"`
}

// GetCommentsArgs controls listing of comments. Exactly one of TaskID or
// ProjectID must be set.
type GetCommentsArgs struct {
	TaskID    string
	ProjectID string
	Cursor    string
	Limit     int
}

func (a *GetCommentsArgs) query() url.Values {
	q := url.Values{}
	if a == nil {
		return q
	}
	if a.TaskID != "" {
		q.Set("task_id", a.TaskID)
	}
	if a.ProjectID != "" {
		q.Set("project_id", a.ProjectID)
	}
	if a.Cursor != "" {
		q.Set("cursor", a.Cursor)
	}
	if a.Limit > 0 {
		q.Set("limit", strconv.Itoa(a.Limit))
	}
	return q
}

// CreateCommentArgs are the parameters for creating a comment. Content and
// exactly one of TaskID or ProjectID are required.
type CreateCommentArgs struct {
	Content    string      `json:"content"`
	TaskID     *string     `json:"task_id,omitempty"`
	ProjectID  *string     `json:"project_id,omitempty"`
	Attachment *Attachment `json:"attachment,omitempty"`
}

// UpdateCommentArgs are the parameters for updating a comment.
type UpdateCommentArgs struct {
	Content string `json:"content"`
}

// GetComments returns a single page of comments.
func (c *Client) GetComments(ctx context.Context, args *GetCommentsArgs) (Page[Comment], error) {
	return doList[Comment](ctx, c, "/comments", args.query())
}

// Comments returns an iterator over all comments matching args, following
// pagination.
func (c *Client) Comments(ctx context.Context, args *GetCommentsArgs) iter.Seq2[Comment, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[Comment], error) {
		return doList[Comment](ctx, c, "/comments", setCursor(base, cursor))
	})
}

// GetComment returns a single comment by ID.
func (c *Client) GetComment(ctx context.Context, id string) (Comment, error) {
	return doGet[Comment](ctx, c, "/comments/"+id)
}

// CreateComment creates a new comment.
func (c *Client) CreateComment(ctx context.Context, args CreateCommentArgs) (Comment, error) {
	return doPost[Comment](ctx, c, "/comments", args)
}

// UpdateComment updates an existing comment.
func (c *Client) UpdateComment(ctx context.Context, id string, args UpdateCommentArgs) (Comment, error) {
	return doPost[Comment](ctx, c, "/comments/"+id, args)
}

// DeleteComment deletes a comment.
func (c *Client) DeleteComment(ctx context.Context, id string) error {
	return doDelete(ctx, c, "/comments/"+id)
}
