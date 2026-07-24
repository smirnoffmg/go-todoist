package todoist

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestFolderLifecycle(t *testing.T) {
	type call struct{ method, path, query string }
	var got call
	var body map[string]any
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = call{r.Method, r.URL.Path, r.URL.RawQuery}
		body = map[string]any{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = w.Write([]byte(`{"id":"f1","name":"Team","workspace_id":"ws1","results":[{"id":"f1"}],"next_cursor":""}`))
	})
	ctx := context.Background()

	t.Run("list", func(t *testing.T) {
		if _, err := c.GetFolders(ctx, "ws1", &GetFoldersArgs{Limit: 10}); err != nil {
			t.Fatal(err)
		}
		if got.method != http.MethodGet || got.path != "/folders" {
			t.Errorf("got %s %s", got.method, got.path)
		}
		if got.query != "limit=10&workspace_id=ws1" {
			t.Errorf("query = %q", got.query)
		}
	})

	t.Run("get", func(t *testing.T) {
		if _, err := c.GetFolder(ctx, "f1"); err != nil {
			t.Fatal(err)
		}
		if got.path != "/folders/f1" {
			t.Errorf("path = %q", got.path)
		}
	})

	t.Run("create", func(t *testing.T) {
		if _, err := c.CreateFolder(ctx, CreateFolderArgs{Name: "Team", WorkspaceID: 99}); err != nil {
			t.Fatal(err)
		}
		if got.method != http.MethodPost || got.path != "/folders" {
			t.Errorf("got %s %s", got.method, got.path)
		}
		// workspace_id must serialize as an integer.
		if wid, _ := body["workspace_id"].(float64); wid != 99 || body["name"] != "Team" {
			t.Errorf("body = %+v", body)
		}
	})

	t.Run("update", func(t *testing.T) {
		if _, err := c.UpdateFolder(ctx, "f1", UpdateFolderArgs{Name: Ptr("Renamed")}); err != nil {
			t.Fatal(err)
		}
		if got.path != "/folders/f1" {
			t.Errorf("path = %q", got.path)
		}
		if body["name"] != "Renamed" {
			t.Errorf("body = %+v", body)
		}
	})

	t.Run("delete", func(t *testing.T) {
		if err := c.DeleteFolder(ctx, "f1"); err != nil {
			t.Fatal(err)
		}
		if got.method != http.MethodDelete || got.path != "/folders/f1" {
			t.Errorf("got %s %s", got.method, got.path)
		}
	})
}

func TestFoldersIteratorAndNilArgs(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("workspace_id") != "ws1" {
			t.Errorf("missing workspace_id: %q", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"results":[{"id":"f1"}],"next_cursor":""}`))
	})

	var count int
	for _, err := range c.Folders(context.Background(), "ws1", nil) {
		if err != nil {
			t.Fatalf("iteration: %v", err)
		}
		count++
	}
	if count != 1 {
		t.Errorf("count = %d", count)
	}
}
