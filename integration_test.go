//go:build integration

// Package-level integration tests that exercise the client against the real
// Todoist API. They are excluded from normal builds by the "integration" build
// tag and only run when TODOIST_TOKEN is set. They create their own scratch
// project and delete it on cleanup, so they do not touch your existing data.
//
// Run with:
//
//	TODOIST_TOKEN=xxxxx go test -tags=integration -run Integration -v ./...
package todoist

import (
	"context"
	"os"
	"testing"
	"time"
)

func integrationClient(t *testing.T) *Client {
	t.Helper()
	token := os.Getenv("TODOIST_TOKEN")
	if token == "" {
		t.Skip("TODOIST_TOKEN not set; skipping integration tests")
	}
	return New(token, WithUserAgent("go-todoist-integration-test"))
}

// scratchProject creates a uniquely named project and registers its deletion.
func scratchProject(ctx context.Context, t *testing.T, c *Client) Project {
	t.Helper()
	name := "go-todoist integration " + time.Now().UTC().Format("20060102T150405.000")
	p, err := c.CreateProject(ctx, CreateProjectArgs{Name: name})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() {
		if err := c.DeleteProject(context.Background(), p.ID); err != nil {
			t.Logf("cleanup: DeleteProject(%s): %v", p.ID, err)
		}
	})
	return p
}

func TestIntegrationUser(t *testing.T) {
	c := integrationClient(t)
	u, err := c.GetUser(context.Background())
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if u.ID == "" {
		t.Error("expected a non-empty user ID")
	}
	t.Logf("authenticated as %s (%s)", u.FullName, u.Email)
}

func TestIntegrationTaskLifecycle(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()
	project := scratchProject(ctx, t, c)

	created, err := c.CreateTask(ctx, CreateTaskArgs{
		Content:   "integration task",
		ProjectID: Ptr(project.ID),
		Priority:  Ptr(PriorityHigh),
		DueString: Ptr("tomorrow"),
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if created.Content != "integration task" {
		t.Errorf("content = %q", created.Content)
	}

	got, err := c.GetTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("GetTask id = %q, want %q", got.ID, created.ID)
	}

	updated, err := c.UpdateTask(ctx, created.ID, UpdateTaskArgs{Content: Ptr("integration task (edited)")})
	if err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if updated.Content != "integration task (edited)" {
		t.Errorf("updated content = %q", updated.Content)
	}

	if err := c.CloseTask(ctx, created.ID); err != nil {
		t.Fatalf("CloseTask: %v", err)
	}
	if err := c.ReopenTask(ctx, created.ID); err != nil {
		t.Fatalf("ReopenTask: %v", err)
	}
	if err := c.DeleteTask(ctx, created.ID); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
}

func TestIntegrationSectionsAndComments(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()
	project := scratchProject(ctx, t, c)

	section, err := c.CreateSection(ctx, CreateSectionArgs{Name: "Phase 1", ProjectID: project.ID})
	if err != nil {
		t.Fatalf("CreateSection: %v", err)
	}

	task, err := c.CreateTask(ctx, CreateTaskArgs{
		Content:   "task with a comment",
		ProjectID: Ptr(project.ID),
		SectionID: Ptr(section.ID),
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	comment, err := c.CreateComment(ctx, CreateCommentArgs{Content: "hello from Go", TaskID: Ptr(task.ID)})
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	var found bool
	for cm, err := range c.Comments(ctx, &GetCommentsArgs{TaskID: task.ID}) {
		if err != nil {
			t.Fatalf("Comments iterator: %v", err)
		}
		if cm.ID == comment.ID {
			found = true
		}
	}
	if !found {
		t.Error("created comment not returned by Comments iterator")
	}
}

func TestIntegrationLabelLifecycle(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	label, err := c.CreateLabel(ctx, CreateLabelArgs{Name: "go-todoist-itest-" + time.Now().UTC().Format("150405.000")})
	if err != nil {
		t.Fatalf("CreateLabel: %v", err)
	}
	t.Cleanup(func() {
		if err := c.DeleteLabel(context.Background(), label.ID); err != nil {
			t.Logf("cleanup: DeleteLabel(%s): %v", label.ID, err)
		}
	})

	if _, err := c.GetLabel(ctx, label.ID); err != nil {
		t.Fatalf("GetLabel: %v", err)
	}
}

func TestIntegrationProjectsPagination(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()
	scratchProject(ctx, t, c) // ensure at least one project exists

	var count int
	for p, err := range c.Projects(ctx, &GetProjectsArgs{Limit: 50}) {
		if err != nil {
			t.Fatalf("Projects iterator: %v", err)
		}
		if p.ID == "" {
			t.Error("project with empty ID")
		}
		count++
	}
	if count == 0 {
		t.Error("expected at least one project")
	}
}

func TestIntegrationProductivityStats(t *testing.T) {
	c := integrationClient(t)
	stats, err := c.GetProductivityStats(context.Background())
	if err != nil {
		t.Fatalf("GetProductivityStats: %v", err)
	}
	t.Logf("completed_count=%d karma=%.0f days=%d weeks=%d",
		stats.CompletedCount, stats.Karma, len(stats.DaysItems), len(stats.WeekItems))
}

func TestIntegrationWorkspaces(t *testing.T) {
	c := integrationClient(t)
	ws, err := c.GetWorkspaces(context.Background())
	if err != nil {
		t.Fatalf("GetWorkspaces: %v", err)
	}
	t.Logf("workspaces: %d", len(ws))
}

func TestIntegrationQuickAdd(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	task, err := c.QuickAddTask(ctx, "go-todoist quick-add smoke test tomorrow")
	if err != nil {
		t.Fatalf("QuickAddTask: %v", err)
	}
	t.Cleanup(func() {
		if err := c.DeleteTask(context.Background(), task.ID); err != nil {
			t.Logf("cleanup: DeleteTask(%s): %v", task.ID, err)
		}
	})
	if task.ID == "" {
		t.Error("quick-added task has no ID")
	}
}

func TestIntegrationTasksByFilter(t *testing.T) {
	c := integrationClient(t)
	var count int
	for task, err := range c.TasksByFilter(context.Background(), &GetTasksByFilterArgs{Query: "today | overdue", Limit: 50}) {
		if err != nil {
			t.Fatalf("TasksByFilter: %v", err)
		}
		_ = task
		count++
	}
	t.Logf("filter matched %d tasks", count)
}

func TestIntegrationCompletedByCompletionDate(t *testing.T) {
	c := integrationClient(t)
	now := time.Now().UTC()
	page, err := c.GetCompletedByCompletionDate(context.Background(), &GetCompletedTasksArgs{
		Since: now.AddDate(0, 0, -7),
		Until: now,
		Limit: 50,
	})
	if err != nil {
		t.Fatalf("GetCompletedByCompletionDate: %v", err)
	}
	t.Logf("completed in last 7 days: %d (next_cursor=%q)", len(page.Results), page.NextCursor)
}

func TestIntegrationArchivedProjectsAndSharedLabels(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()

	ap, err := c.GetArchivedProjects(ctx, &GetProjectsArgs{Limit: 50})
	if err != nil {
		t.Fatalf("GetArchivedProjects: %v", err)
	}
	sl, err := c.GetSharedLabels(ctx, &GetSharedLabelsArgs{Limit: 50})
	if err != nil {
		t.Fatalf("GetSharedLabels: %v", err)
	}
	t.Logf("archived projects=%d shared labels=%d", len(ap.Results), len(sl.Results))
}

func TestIntegrationTemplateURL(t *testing.T) {
	c := integrationClient(t)
	ctx := context.Background()
	project := scratchProject(ctx, t, c)

	res, err := c.GetTemplateURL(ctx, project.ID, false)
	if err != nil {
		t.Fatalf("GetTemplateURL: %v", err)
	}
	if res.FileURL == "" {
		t.Error("expected a non-empty template file URL")
	}
	t.Logf("template url: %s", res.FileURL)
}

func TestIntegrationWorkspaceInvitations(t *testing.T) {
	c := integrationClient(t)
	ws, err := c.GetWorkspaces(context.Background())
	if err != nil {
		t.Fatalf("GetWorkspaces: %v", err)
	}
	if len(ws) == 0 {
		t.Skip("no workspaces on this account")
	}
	inv, err := c.GetWorkspaceInvitations(context.Background(), ws[0].ID)
	if err != nil {
		t.Fatalf("GetWorkspaceInvitations: %v", err)
	}
	t.Logf("workspace %s has %d pending invitations", ws[0].ID, len(inv))
}

func TestIntegrationSyncReadOnly(t *testing.T) {
	c := integrationClient(t)
	resp, err := c.Sync(context.Background(), SyncRequest{
		SyncToken:     "*",
		ResourceTypes: []string{"projects"},
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if resp.SyncToken == "" {
		t.Error("expected a sync token in the response")
	}
	t.Logf("full sync returned %d projects", len(resp.Projects))
}
