package todoist

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestWorkspaceInvitationsGetAndAll(t *testing.T) {
	var gotPath, gotQuery string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[{"id":"i1","user_email":"a@b.c","role":"MEMBER"}]`))
	})

	inv, err := c.GetWorkspaceInvitations(context.Background(), "ws1")
	if err != nil {
		t.Fatalf("GetWorkspaceInvitations: %v", err)
	}
	if gotPath != "/workspaces/invitations" || gotQuery != "workspace_id=ws1" {
		t.Errorf("path=%q query=%q", gotPath, gotQuery)
	}
	if len(inv) != 1 || inv[0].ID != "i1" {
		t.Errorf("inv = %+v", inv)
	}

	if _, err := c.GetAllWorkspaceInvitations(context.Background(), "ws1"); err != nil {
		t.Fatalf("GetAllWorkspaceInvitations: %v", err)
	}
	if gotPath != "/workspaces/invitations/all" {
		t.Errorf("path = %q", gotPath)
	}
}

func TestWorkspaceInvitationDelete(t *testing.T) {
	var gotPath string
	var body map[string]any
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = w.Write([]byte(`{"id":"i1","user_email":"a@b.c"}`))
	})

	if _, err := c.DeleteWorkspaceInvitation(context.Background(), 42, "a@b.c"); err != nil {
		t.Fatalf("DeleteWorkspaceInvitation: %v", err)
	}
	if gotPath != "/workspaces/invitations/delete" {
		t.Errorf("path = %q", gotPath)
	}
	// JSON numbers decode to float64.
	wid, _ := body["workspace_id"].(float64)
	if wid != 42 || body["user_email"] != "a@b.c" {
		t.Errorf("body = %+v", body)
	}
}

func TestWorkspaceInvitationAcceptRejectUsePUT(t *testing.T) {
	type call struct{ method, path string }
	var got call
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = call{r.Method, r.URL.Path}
		_, _ = w.Write([]byte(`{"id":"i1"}`))
	})
	ctx := context.Background()

	if _, err := c.AcceptWorkspaceInvitation(ctx, "code1"); err != nil {
		t.Fatalf("Accept: %v", err)
	}
	if got.method != http.MethodPut || got.path != "/workspaces/invitations/code1/accept" {
		t.Errorf("accept got %+v", got)
	}

	if _, err := c.RejectWorkspaceInvitation(ctx, "code1"); err != nil {
		t.Fatalf("Reject: %v", err)
	}
	if got.method != http.MethodPut || got.path != "/workspaces/invitations/code1/reject" {
		t.Errorf("reject got %+v", got)
	}
}

func TestJoinWorkspace(t *testing.T) {
	var gotPath string
	var body map[string]any
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = w.Write([]byte(`{"user_id":"u1","workspace_id":"ws1"}`))
	})

	wu, err := c.JoinWorkspace(context.Background(), JoinWorkspaceArgs{InviteCode: Ptr("code1")})
	if err != nil {
		t.Fatalf("JoinWorkspace: %v", err)
	}
	if gotPath != "/workspaces/join" {
		t.Errorf("path = %q", gotPath)
	}
	if body["invite_code"] != "code1" {
		t.Errorf("body = %+v", body)
	}
	if _, ok := body["workspace_id"]; ok {
		t.Error("workspace_id should be omitted when unset")
	}
	if wu.UserID != "u1" {
		t.Errorf("wu = %+v", wu)
	}
}

func TestWorkspaceProjectsListsAndIterators(t *testing.T) {
	var gotPath string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		// Response wraps items under "workspace_projects".
		_, _ = w.Write([]byte(`{"workspace_projects":[{"id":"p1"}],"next_cursor":"","has_more":false}`))
	})
	ctx := context.Background()

	page, err := c.GetWorkspaceActiveProjects(ctx, "ws1", &GetProjectsArgs{Limit: 10})
	if err != nil {
		t.Fatalf("GetWorkspaceActiveProjects: %v", err)
	}
	if gotPath != "/workspaces/ws1/projects/active" {
		t.Errorf("path = %q", gotPath)
	}
	if len(page.Results) != 1 || page.Results[0].ID != "p1" {
		t.Errorf("results = %+v", page.Results)
	}

	if _, err := c.GetWorkspaceArchivedProjects(ctx, "ws1", nil); err != nil {
		t.Fatalf("GetWorkspaceArchivedProjects: %v", err)
	}
	if gotPath != "/workspaces/ws1/projects/archived" {
		t.Errorf("path = %q", gotPath)
	}

	var active, archived int
	for _, err := range c.WorkspaceActiveProjects(ctx, "ws1", nil) {
		if err != nil {
			t.Fatalf("active iter: %v", err)
		}
		active++
	}
	for _, err := range c.WorkspaceArchivedProjects(ctx, "ws1", nil) {
		if err != nil {
			t.Fatalf("archived iter: %v", err)
		}
		archived++
	}
	if active != 1 || archived != 1 {
		t.Errorf("active=%d archived=%d", active, archived)
	}
}
