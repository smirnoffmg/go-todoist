package todoist

import (
	"context"
	"iter"
	"net/url"
	"strconv"
)

// Project is a Todoist project.
type Project struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Color          string `json:"color"`
	ParentID       string `json:"parent_id"`
	Order          int    `json:"child_order"`
	CommentCount   int    `json:"comment_count"`
	IsShared       bool   `json:"is_shared"`
	IsFavorite     bool   `json:"is_favorite"`
	IsInboxProject bool   `json:"is_inbox_project"`
	IsArchived     bool   `json:"is_archived"`
	ViewStyle      string `json:"view_style"`
	WorkspaceID    string `json:"workspace_id"`
	URL            string `json:"url"`
}

// Collaborator is a user who shares a project.
type Collaborator struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetProjectsArgs controls listing and pagination of projects.
type GetProjectsArgs struct {
	Cursor string
	Limit  int
}

func (a *GetProjectsArgs) query() url.Values {
	q := url.Values{}
	if a == nil {
		return q
	}
	if a.Cursor != "" {
		q.Set("cursor", a.Cursor)
	}
	if a.Limit > 0 {
		q.Set("limit", strconv.Itoa(a.Limit))
	}
	return q
}

// CreateProjectArgs are the parameters for creating a project. Name is required.
type CreateProjectArgs struct {
	Name       string  `json:"name"`
	ParentID   *string `json:"parent_id,omitempty"`
	Color      *string `json:"color,omitempty"`
	IsFavorite *bool   `json:"is_favorite,omitempty"`
	ViewStyle  *string `json:"view_style,omitempty"`
}

// UpdateProjectArgs are the parameters for updating a project. Only non-nil
// fields are sent.
type UpdateProjectArgs struct {
	Name       *string `json:"name,omitempty"`
	Color      *string `json:"color,omitempty"`
	IsFavorite *bool   `json:"is_favorite,omitempty"`
	ViewStyle  *string `json:"view_style,omitempty"`
}

// GetProjects returns a single page of projects.
func (c *Client) GetProjects(ctx context.Context, args *GetProjectsArgs) (Page[Project], error) {
	return doList[Project](ctx, c, "/projects", args.query())
}

// Projects returns an iterator over all projects, following pagination.
func (c *Client) Projects(ctx context.Context, args *GetProjectsArgs) iter.Seq2[Project, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[Project], error) {
		return doList[Project](ctx, c, "/projects", setCursor(base, cursor))
	})
}

// GetArchivedProjects returns a single page of archived projects.
func (c *Client) GetArchivedProjects(ctx context.Context, args *GetProjectsArgs) (Page[Project], error) {
	return doList[Project](ctx, c, "/projects/archived", args.query())
}

// ArchivedProjects returns an iterator over all archived projects, following
// pagination.
func (c *Client) ArchivedProjects(ctx context.Context, args *GetProjectsArgs) iter.Seq2[Project, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[Project], error) {
		return doList[Project](ctx, c, "/projects/archived", setCursor(base, cursor))
	})
}

// GetProject returns a single project by ID.
func (c *Client) GetProject(ctx context.Context, id string) (Project, error) {
	return doGet[Project](ctx, c, "/projects/"+id)
}

// JoinProject adds the current user to a shared project.
func (c *Client) JoinProject(ctx context.Context, id string) error {
	return doAction(ctx, c, "/projects/"+id+"/join")
}

// CreateProject creates a new project.
func (c *Client) CreateProject(ctx context.Context, args CreateProjectArgs) (Project, error) {
	return doPost[Project](ctx, c, "/projects", args)
}

// UpdateProject updates an existing project.
func (c *Client) UpdateProject(ctx context.Context, id string, args UpdateProjectArgs) (Project, error) {
	return doPost[Project](ctx, c, "/projects/"+id, args)
}

// DeleteProject deletes a project.
func (c *Client) DeleteProject(ctx context.Context, id string) error {
	return doDelete(ctx, c, "/projects/"+id)
}

// ArchiveProject archives a project.
func (c *Client) ArchiveProject(ctx context.Context, id string) error {
	return doAction(ctx, c, "/projects/"+id+"/archive")
}

// UnarchiveProject restores an archived project.
func (c *Client) UnarchiveProject(ctx context.Context, id string) error {
	return doAction(ctx, c, "/projects/"+id+"/unarchive")
}

// GetCollaborators returns a page of collaborators for a shared project.
func (c *Client) GetCollaborators(ctx context.Context, projectID string, args *GetProjectsArgs) (Page[Collaborator], error) {
	return doList[Collaborator](ctx, c, "/projects/"+projectID+"/collaborators", args.query())
}
