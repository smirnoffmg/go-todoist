package todoist

import (
	"context"
	"iter"
	"net/http"
	"net/url"
)

// WorkspaceInvitation is a pending invitation for a user to join a workspace.
type WorkspaceInvitation struct {
	ID             string `json:"id"`
	InviterID      string `json:"inviter_id"`
	UserEmail      string `json:"user_email"`
	WorkspaceID    string `json:"workspace_id"`
	Role           string `json:"role"`
	IsExistingUser bool   `json:"is_existing_user"`
}

// JoinWorkspaceArgs identifies the workspace to join, by invite code or by
// workspace ID (for open workspaces). At least one should be set.
type JoinWorkspaceArgs struct {
	InviteCode  *string `json:"invite_code,omitempty"`
	WorkspaceID *string `json:"workspace_id,omitempty"`
}

// GetWorkspaceInvitations returns the current user's pending invitations for a
// workspace.
func (c *Client) GetWorkspaceInvitations(ctx context.Context, workspaceID string) ([]WorkspaceInvitation, error) {
	q := url.Values{}
	q.Set("workspace_id", workspaceID)
	var out []WorkspaceInvitation
	err := c.do(ctx, http.MethodGet, "/workspaces/invitations", q, nil, &out)
	return out, err
}

// GetAllWorkspaceInvitations returns all invitations for a workspace (admin
// view).
func (c *Client) GetAllWorkspaceInvitations(ctx context.Context, workspaceID string) ([]WorkspaceInvitation, error) {
	q := url.Values{}
	q.Set("workspace_id", workspaceID)
	var out []WorkspaceInvitation
	err := c.do(ctx, http.MethodGet, "/workspaces/invitations/all", q, nil, &out)
	return out, err
}

// DeleteWorkspaceInvitation revokes a pending invitation.
func (c *Client) DeleteWorkspaceInvitation(ctx context.Context, workspaceID int64, userEmail string) (WorkspaceInvitation, error) {
	body := map[string]any{"workspace_id": workspaceID, "user_email": userEmail}
	return doPost[WorkspaceInvitation](ctx, c, "/workspaces/invitations/delete", body)
}

// AcceptWorkspaceInvitation accepts an invitation by its invite code.
func (c *Client) AcceptWorkspaceInvitation(ctx context.Context, inviteCode string) (WorkspaceInvitation, error) {
	return doPut[WorkspaceInvitation](ctx, c, "/workspaces/invitations/"+inviteCode+"/accept", nil)
}

// RejectWorkspaceInvitation rejects an invitation by its invite code.
func (c *Client) RejectWorkspaceInvitation(ctx context.Context, inviteCode string) (WorkspaceInvitation, error) {
	return doPut[WorkspaceInvitation](ctx, c, "/workspaces/invitations/"+inviteCode+"/reject", nil)
}

// JoinWorkspace adds the current user to a workspace and returns their
// membership.
func (c *Client) JoinWorkspace(ctx context.Context, args JoinWorkspaceArgs) (WorkspaceUser, error) {
	return doPost[WorkspaceUser](ctx, c, "/workspaces/join", args)
}

// GetWorkspaceActiveProjects returns a page of active projects in a workspace.
func (c *Client) GetWorkspaceActiveProjects(ctx context.Context, workspaceID string, args *GetProjectsArgs) (Page[Project], error) {
	return doWorkspaceProjectsList(ctx, c, "/workspaces/"+workspaceID+"/projects/active", args.query())
}

// WorkspaceActiveProjects iterates over all active projects in a workspace.
func (c *Client) WorkspaceActiveProjects(ctx context.Context, workspaceID string, args *GetProjectsArgs) iter.Seq2[Project, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[Project], error) {
		return doWorkspaceProjectsList(ctx, c, "/workspaces/"+workspaceID+"/projects/active", setCursor(base, cursor))
	})
}

// GetWorkspaceArchivedProjects returns a page of archived projects in a
// workspace.
func (c *Client) GetWorkspaceArchivedProjects(ctx context.Context, workspaceID string, args *GetProjectsArgs) (Page[Project], error) {
	return doWorkspaceProjectsList(ctx, c, "/workspaces/"+workspaceID+"/projects/archived", args.query())
}

// WorkspaceArchivedProjects iterates over all archived projects in a workspace.
func (c *Client) WorkspaceArchivedProjects(ctx context.Context, workspaceID string, args *GetProjectsArgs) iter.Seq2[Project, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[Project], error) {
		return doWorkspaceProjectsList(ctx, c, "/workspaces/"+workspaceID+"/projects/archived", setCursor(base, cursor))
	})
}
