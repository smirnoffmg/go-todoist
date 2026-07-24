package todoist

import (
	"context"
	"iter"
	"net/url"
	"strconv"
)

// Folder groups projects within a workspace.
type Folder struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	WorkspaceID  string `json:"workspace_id"`
	DefaultOrder int    `json:"default_order"`
	ChildOrder   int    `json:"child_order"`
	IsDeleted    bool   `json:"is_deleted"`
}

// GetFoldersArgs controls listing and pagination of folders.
type GetFoldersArgs struct {
	Cursor string
	Limit  int
}

func (a *GetFoldersArgs) query(workspaceID string) url.Values {
	q := url.Values{}
	q.Set("workspace_id", workspaceID)
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

// CreateFolderArgs are the parameters for creating a folder. Name and
// WorkspaceID are required.
type CreateFolderArgs struct {
	Name         string `json:"name"`
	WorkspaceID  int64  `json:"workspace_id"`
	DefaultOrder *int   `json:"default_order,omitempty"`
	ChildOrder   *int   `json:"child_order,omitempty"`
}

// UpdateFolderArgs are the parameters for updating a folder. Only non-nil fields
// are sent.
type UpdateFolderArgs struct {
	Name         *string `json:"name,omitempty"`
	DefaultOrder *int    `json:"default_order,omitempty"`
}

// GetFolders returns a single page of folders in a workspace.
func (c *Client) GetFolders(ctx context.Context, workspaceID string, args *GetFoldersArgs) (Page[Folder], error) {
	return doList[Folder](ctx, c, "/folders", args.query(workspaceID))
}

// Folders returns an iterator over all folders in a workspace, following
// pagination.
func (c *Client) Folders(ctx context.Context, workspaceID string, args *GetFoldersArgs) iter.Seq2[Folder, error] {
	base := args.query(workspaceID)
	return paginate(func(cursor string) (Page[Folder], error) {
		return doList[Folder](ctx, c, "/folders", setCursor(base, cursor))
	})
}

// GetFolder returns a single folder by ID.
func (c *Client) GetFolder(ctx context.Context, id string) (Folder, error) {
	return doGet[Folder](ctx, c, "/folders/"+id)
}

// CreateFolder creates a new folder.
func (c *Client) CreateFolder(ctx context.Context, args CreateFolderArgs) (Folder, error) {
	return doPost[Folder](ctx, c, "/folders", args)
}

// UpdateFolder updates an existing folder.
func (c *Client) UpdateFolder(ctx context.Context, id string, args UpdateFolderArgs) (Folder, error) {
	return doPost[Folder](ctx, c, "/folders/"+id, args)
}

// DeleteFolder deletes a folder.
func (c *Client) DeleteFolder(ctx context.Context, id string) error {
	return doDelete(ctx, c, "/folders/"+id)
}
