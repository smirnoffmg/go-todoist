package todoist

import (
	"context"
	"iter"
	"net/http"
	"testing"
)

// combinedBody is valid JSON for decoding into any resource type or into a
// Page[T]: json.Unmarshal ignores the fields a given target does not declare.
const combinedBody = `{"id":"1","results":[{"id":"1"}],"next_cursor":"","name":"x","content":"x","query":"q","item_id":"i"}`

func allMethodsClient(t *testing.T) *Client {
	t.Helper()
	return testClient(t, func(w http.ResponseWriter, r *http.Request) {
		// GetWorkspaces returns a plain array rather than a paginated object.
		if r.URL.Path == "/workspaces" && r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`[{"id":"1","name":"x"}]`))
			return
		}
		_, _ = w.Write([]byte(combinedBody))
	})
}

func drain[T any](t *testing.T, seq iter.Seq2[T, error]) {
	t.Helper()
	for _, err := range seq {
		if err != nil {
			t.Fatalf("iteration: %v", err)
		}
	}
}

func TestAllResourceMethods(t *testing.T) {
	c := allMethodsClient(t)
	ctx := context.Background()

	// Fully-populated list args exercise every query() branch; nil args (via the
	// iterators below) exercise the nil branch.
	calls := []func() error{
		// Tasks
		func() error {
			_, e := c.GetTasks(ctx, &GetTasksArgs{ProjectID: "p", SectionID: "s", ParentID: "pa", Label: "l", Cursor: "c", Limit: 1})
			return e
		},
		func() error { _, e := c.GetTask(ctx, "1"); return e },
		func() error { _, e := c.CreateTask(ctx, CreateTaskArgs{Content: "x"}); return e },
		func() error { _, e := c.UpdateTask(ctx, "1", UpdateTaskArgs{}); return e },
		func() error { return c.DeleteTask(ctx, "1") },
		func() error { return c.CloseTask(ctx, "1") },
		func() error { return c.ReopenTask(ctx, "1") },
		func() error { _, e := c.QuickAddTask(ctx, "buy milk tomorrow"); return e },
		// Projects
		func() error { _, e := c.GetProjects(ctx, &GetProjectsArgs{Cursor: "c", Limit: 1}); return e },
		func() error { _, e := c.GetProject(ctx, "1"); return e },
		func() error { _, e := c.CreateProject(ctx, CreateProjectArgs{Name: "x"}); return e },
		func() error { _, e := c.UpdateProject(ctx, "1", UpdateProjectArgs{}); return e },
		func() error { return c.DeleteProject(ctx, "1") },
		func() error { return c.ArchiveProject(ctx, "1") },
		func() error { return c.UnarchiveProject(ctx, "1") },
		func() error { _, e := c.GetCollaborators(ctx, "1", &GetProjectsArgs{Cursor: "c", Limit: 1}); return e },
		// Sections
		func() error {
			_, e := c.GetSections(ctx, &GetSectionsArgs{ProjectID: "p", Cursor: "c", Limit: 1})
			return e
		},
		func() error { _, e := c.GetSection(ctx, "1"); return e },
		func() error { _, e := c.CreateSection(ctx, CreateSectionArgs{Name: "x", ProjectID: "p"}); return e },
		func() error { _, e := c.UpdateSection(ctx, "1", UpdateSectionArgs{Name: "y"}); return e },
		func() error { return c.DeleteSection(ctx, "1") },
		// Labels
		func() error { _, e := c.GetLabels(ctx, &GetLabelsArgs{Cursor: "c", Limit: 1}); return e },
		func() error { _, e := c.GetLabel(ctx, "1"); return e },
		func() error { _, e := c.CreateLabel(ctx, CreateLabelArgs{Name: "x"}); return e },
		func() error { _, e := c.UpdateLabel(ctx, "1", UpdateLabelArgs{}); return e },
		func() error { return c.DeleteLabel(ctx, "1") },
		// Comments
		func() error {
			_, e := c.GetComments(ctx, &GetCommentsArgs{TaskID: "t", ProjectID: "p", Cursor: "c", Limit: 1})
			return e
		},
		func() error { _, e := c.GetComment(ctx, "1"); return e },
		func() error { _, e := c.CreateComment(ctx, CreateCommentArgs{Content: "x"}); return e },
		func() error { _, e := c.UpdateComment(ctx, "1", UpdateCommentArgs{Content: "y"}); return e },
		func() error { return c.DeleteComment(ctx, "1") },
		// Reminders
		func() error { _, e := c.GetReminders(ctx, &GetRemindersArgs{Cursor: "c", Limit: 1}); return e },
		func() error {
			_, e := c.CreateReminder(ctx, CreateReminderArgs{ItemID: "i", Type: "absolute"})
			return e
		},
		func() error { _, e := c.UpdateReminder(ctx, "1", UpdateReminderArgs{}); return e },
		func() error { return c.DeleteReminder(ctx, "1") },
		// Workspaces
		func() error { _, e := c.GetWorkspaces(ctx); return e },
		func() error { _, e := c.GetWorkspace(ctx, "1"); return e },
		func() error { _, e := c.CreateWorkspace(ctx, CreateWorkspaceArgs{Name: "x"}); return e },
		func() error { _, e := c.UpdateWorkspace(ctx, "1", UpdateWorkspaceArgs{}); return e },
		func() error {
			_, e := c.GetWorkspaceUsers(ctx, &GetWorkspaceUsersArgs{WorkspaceID: "1", Cursor: "c", Limit: 1})
			return e
		},
		// User
		func() error { _, e := c.GetUser(ctx); return e },
		func() error { _, e := c.GetProductivityStats(ctx); return e },
	}
	for i, fn := range calls {
		if err := fn(); err != nil {
			t.Errorf("call %d: %v", i, err)
		}
	}

	// Iterators with nil args cover both the iterator bodies and the query() nil
	// branch of each resource.
	drain(t, c.Tasks(ctx, nil))
	drain(t, c.Projects(ctx, nil))
	drain(t, c.Sections(ctx, nil))
	drain(t, c.Labels(ctx, nil))
	drain(t, c.Comments(ctx, nil))
	drain(t, c.Reminders(ctx, nil))

	// GetWorkspaceUsers is paginated but has a different args type; exercise its
	// nil-args query branch directly.
	if _, err := c.GetWorkspaceUsers(ctx, nil); err != nil {
		t.Errorf("GetWorkspaceUsers(nil): %v", err)
	}
}
