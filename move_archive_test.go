package todoist

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestMoveTask(t *testing.T) {
	var gotPath string
	var body map[string]any
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = w.Write([]byte(`{"id":"1","section_id":"s2"}`))
	})

	_, err := c.MoveTask(context.Background(), "1", MoveTaskArgs{SectionID: Ptr("s2")})
	if err != nil {
		t.Fatalf("MoveTask: %v", err)
	}
	if gotPath != "/tasks/1/move" {
		t.Errorf("path = %q", gotPath)
	}
	if body["section_id"] != "s2" {
		t.Errorf("body = %+v", body)
	}
	if _, ok := body["project_id"]; ok {
		t.Error("project_id should be omitted when unset")
	}
}

func TestArchivedProjectsPathAndIterator(t *testing.T) {
	var gotPath string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"results":[{"id":"p1","is_archived":true}],"next_cursor":""}`))
	})

	page, err := c.GetArchivedProjects(context.Background(), &GetProjectsArgs{Limit: 10})
	if err != nil {
		t.Fatalf("GetArchivedProjects: %v", err)
	}
	if gotPath != "/projects/archived" {
		t.Errorf("path = %q", gotPath)
	}
	if len(page.Results) != 1 {
		t.Errorf("results = %+v", page.Results)
	}

	var count int
	for _, err := range c.ArchivedProjects(context.Background(), nil) {
		if err != nil {
			t.Fatalf("iteration: %v", err)
		}
		count++
	}
	if count != 1 {
		t.Errorf("count = %d", count)
	}
}

func TestSectionArchiveAndProjectJoinPaths(t *testing.T) {
	type call struct{ method, path string }
	var got call
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = call{r.Method, r.URL.Path}
	})
	ctx := context.Background()

	cases := []struct {
		name             string
		run              func() error
		method, wantPath string
	}{
		{"archive-section", func() error { return c.ArchiveSection(ctx, "s1") }, http.MethodPost, "/sections/s1/archive"},
		{"unarchive-section", func() error { return c.UnarchiveSection(ctx, "s1") }, http.MethodPost, "/sections/s1/unarchive"},
		{"join-project", func() error { return c.JoinProject(ctx, "p1") }, http.MethodPost, "/projects/p1/join"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if got.method != tc.method || got.path != tc.wantPath {
				t.Errorf("got %s %s, want %s %s", got.method, got.path, tc.method, tc.wantPath)
			}
		})
	}
}
