package todoist

import (
	"context"
	"net/http"
	"testing"
)

// These tests pin bugs found by exercising the real Todoist v1 API, where the
// documented paths/shapes differed from the live behaviour.

func TestUserDecodesTZInfoObject(t *testing.T) {
	// tz_info is a nested object, not a string.
	body := `{"id":"42","email":"a@b.c","tz_info":{"timezone":"Europe/Moscow","gmt_string":"+03:00","hours":3,"minutes":0,"is_dst":0}}`
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	})

	u, err := c.GetUser(context.Background())
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if u.TZInfo.Timezone != "Europe/Moscow" || u.TZInfo.Hours != 3 {
		t.Errorf("TZInfo = %+v", u.TZInfo)
	}
}

func TestQuickAddTaskPath(t *testing.T) {
	var gotPath string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"id":"1","content":"x"}`))
	})
	if _, err := c.QuickAddTask(context.Background(), "buy milk"); err != nil {
		t.Fatalf("QuickAddTask: %v", err)
	}
	if gotPath != "/tasks/quick" {
		t.Errorf("path = %q, want /tasks/quick", gotPath)
	}
}

func TestProductivityStatsPath(t *testing.T) {
	var gotPath string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"completed_count":204,"karma":7982.0}`))
	})
	stats, err := c.GetProductivityStats(context.Background())
	if err != nil {
		t.Fatalf("GetProductivityStats: %v", err)
	}
	if gotPath != "/tasks/completed/stats" {
		t.Errorf("path = %q, want /tasks/completed/stats", gotPath)
	}
	if stats.CompletedCount != 204 {
		t.Errorf("CompletedCount = %d", stats.CompletedCount)
	}
}

func TestGetWorkspacesDecodesBareArray(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"1","name":"Team"},{"id":"2","name":"Personal"}]`))
	})
	ws, err := c.GetWorkspaces(context.Background())
	if err != nil {
		t.Fatalf("GetWorkspaces: %v", err)
	}
	if len(ws) != 2 || ws[0].Name != "Team" {
		t.Errorf("workspaces = %+v", ws)
	}
}

func TestGetWorkspaceUsersPathAndQuery(t *testing.T) {
	var gotPath, gotQuery string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"results":[],"next_cursor":null}`))
	})
	_, err := c.GetWorkspaceUsers(context.Background(), &GetWorkspaceUsersArgs{WorkspaceID: "ws1"})
	if err != nil {
		t.Fatalf("GetWorkspaceUsers: %v", err)
	}
	if gotPath != "/workspaces/users" {
		t.Errorf("path = %q, want /workspaces/users", gotPath)
	}
	if gotQuery != "workspace_id=ws1" {
		t.Errorf("query = %q, want workspace_id=ws1", gotQuery)
	}
}

func TestNullNextCursorStopsIteration(t *testing.T) {
	// The API returns next_cursor: null on the last page.
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"1"}],"next_cursor":null}`))
	})
	var count int
	for _, err := range c.Tasks(context.Background(), nil) {
		if err != nil {
			t.Fatalf("iteration: %v", err)
		}
		count++
	}
	if count != 1 {
		t.Errorf("count = %d, want 1 (null cursor must stop iteration)", count)
	}
}
