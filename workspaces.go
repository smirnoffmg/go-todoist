package todoist

import (
	"context"
	"net/url"
	"strconv"
)

// Workspace is a Todoist workspace (team space).
type Workspace struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Role        string `json:"role"`
	Plan        string `json:"current_active_plan"`
	IsGuest     bool   `json:"is_guest_allowed"`
	MemberCount int    `json:"member_count"`
	LogoBig     string `json:"logo_big"`
}

// WorkspaceUser is a member of a workspace.
type WorkspaceUser struct {
	UserID      string `json:"user_id"`
	WorkspaceID string `json:"workspace_id"`
	UserEmail   string `json:"user_email"`
	FullName    string `json:"full_name"`
	Role        string `json:"role"`
	ImageID     string `json:"image_id"`
}

// GetWorkspaceUsersArgs controls listing and pagination of workspace members.
type GetWorkspaceUsersArgs struct {
	WorkspaceID string
	Cursor      string
	Limit       int
}

func (a *GetWorkspaceUsersArgs) query() url.Values {
	q := url.Values{}
	if a == nil {
		return q
	}
	if a.WorkspaceID != "" {
		q.Set("workspace_id", a.WorkspaceID)
	}
	if a.Cursor != "" {
		q.Set("cursor", a.Cursor)
	}
	if a.Limit > 0 {
		q.Set("limit", strconv.Itoa(a.Limit))
	}
	return q
}

// CreateWorkspaceArgs are the parameters for creating a workspace. Name is
// required.
type CreateWorkspaceArgs struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// UpdateWorkspaceArgs are the parameters for updating a workspace.
type UpdateWorkspaceArgs struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// GetWorkspaces returns the workspaces the user belongs to. Unlike the other
// list endpoints this one is not paginated and returns a plain array.
func (c *Client) GetWorkspaces(ctx context.Context) ([]Workspace, error) {
	return doGet[[]Workspace](ctx, c, "/workspaces")
}

// GetWorkspace returns a single workspace by ID.
func (c *Client) GetWorkspace(ctx context.Context, id string) (Workspace, error) {
	return doGet[Workspace](ctx, c, "/workspaces/"+id)
}

// CreateWorkspace creates a new workspace.
func (c *Client) CreateWorkspace(ctx context.Context, args CreateWorkspaceArgs) (Workspace, error) {
	return doPost[Workspace](ctx, c, "/workspaces", args)
}

// UpdateWorkspace updates an existing workspace.
func (c *Client) UpdateWorkspace(ctx context.Context, id string, args UpdateWorkspaceArgs) (Workspace, error) {
	return doPost[Workspace](ctx, c, "/workspaces/"+id, args)
}

// GetWorkspaceUsers returns a page of members of a workspace.
func (c *Client) GetWorkspaceUsers(ctx context.Context, args *GetWorkspaceUsersArgs) (Page[WorkspaceUser], error) {
	return doList[WorkspaceUser](ctx, c, "/workspaces/users", args.query())
}
